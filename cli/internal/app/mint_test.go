package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const mintReview = `# Review

1. **Before.** Ignore items outside Order.

## Order
1. **Trace.** Check requirement tags.
4. **Scope creep.** Check scope.
7. **Spec integrity.** Check protected files.

## Notes
1. **After.** Ignore items after Order.
`

func newMintFixture(t *testing.T, confidence string) string {
	t.Helper()
	root := t.TempDir()
	path := filepath.Join(root, "blueprint", "spec", "checkout.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	spec := fmt.Sprintf("# checkout\nstatus: ready\ndepth: quick\n\n## Intent\nAccept cards.\n\n## Surfaces\n\n## Requirements\n### accepts-card\n`%s`\nWHEN a customer's valid card is submitted, THE system SHALL accept it.\nfit: returns success for the customer's card\n\n## Edges\n\n## Out of scope\n- forbidden/**\n\n## Depends on\n", confidence)
	if err := os.WriteFile(path, []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func runMint(root string, args ...string) (int, string, error) {
	var out bytes.Buffer
	instance := App{Out: &out, Err: &bytes.Buffer{}, Now: time.Now, Version: "test"}
	code, err := instance.Run(append([]string{"--root", root, "mint"}, args...))
	return code, out.String(), err
}

func TestMintCommand(t *testing.T) {
	const want = `mint spec new 'checkout/accepts-card' --slug 'checkout--accepts-card' --goal 'Implement checkout/accepts-card so that its fit holds: returns success for the customer'"'"'s card' --scope 'cli/**,docs/**,blueprint/spec/checkout.md' --protected 'blueprint/spec/checkout.md' --acceptance 'WHEN a customer'"'"'s valid card is submitted, THE system SHALL accept it.' --gate 'blueprint: blueprint check'`
	for _, tt := range []struct {
		name   string
		review string
		json   bool
		lenses string
	}{
		{name: "review lenses", review: mintReview, lenses: "trace,scope-creep,spec-integrity"},
		{name: "missing review"},
		{name: "no order items", review: "# Review\n\n## Order\nNo lenses yet.\n\n## Notes\n1. **Trace.** Outside Order.\n"},
		{name: "json", review: mintReview, json: true, lenses: "trace,scope-creep,spec-integrity"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root := newMintFixture(t, "stated")
			if tt.review != "" {
				if err := os.WriteFile(filepath.Join(root, "blueprint", "REVIEW.md"), []byte(tt.review), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			args := []string{"checkout/accepts-card", "--scope", " cli/** , docs/** "}
			if tt.json {
				args = append(args, "--json")
			}
			code, out, err := runMint(root, args...)
			if code != 0 || err != nil {
				t.Fatalf("code=%d out=%q err=%v", code, out, err)
			}
			command := strings.TrimSuffix(out, "\n")
			if tt.json {
				var result map[string]string
				if err := json.Unmarshal([]byte(out), &result); err != nil {
					t.Fatal(err)
				}
				if len(result) != 1 {
					t.Fatalf("unexpected JSON shape: %s", out)
				}
				command = result["command"]
			}
			expected := want
			if tt.lenses != "" {
				expected += " --reviews '" + tt.lenses + "'"
			}
			if command != expected {
				t.Fatalf("command:\n%s\nwant:\n%s", command, expected)
			}
		})
	}
}

func TestMintRefusals(t *testing.T) {
	for _, tt := range []struct {
		name       string
		args       []string
		confidence string
		blocks     string
		superseded string
		want       string
	}{
		{name: "missing scope", want: "--scope"},
		{name: "scope without value", args: []string{"--scope"}, want: "--scope"},
		{name: "empty scope", args: []string{"--scope", ""}, want: "--scope"},
		{name: "whitespace scope", args: []string{"--scope", " \t "}, want: "--scope"},
		{name: "empty scope entry", args: []string{"--scope", "a,,b"}, want: "--scope"},
		{name: "whitespace scope entry", args: []string{"--scope", "a, \t ,b"}, want: "--scope"},
		{name: "leading empty entry", args: []string{"--scope", ",a"}, want: "--scope"},
		{name: "trailing empty entry", args: []string{"--scope", "a,"}, want: "--scope"},
		{name: "derived", args: []string{"--scope", "cli/**"}, confidence: "derived", want: "blueprint req confirm checkout/accepts-card"},
		{name: "superseded", args: []string{"--scope", "cli/**"}, superseded: "accepts-bank-transfer", want: "superseded by accepts-bank-transfer"},
		{name: "blocked requirement", args: []string{"--scope", "cli/**"}, blocks: "checkout/accepts-card", want: "blocked"},
		{name: "blocked feature", args: []string{"--scope", "cli/**"}, blocks: "checkout/*", want: "blocked"},
		{name: "blocked project", args: []string{"--scope", "cli/**"}, blocks: "*", want: "blocked"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			confidence := tt.confidence
			if confidence == "" {
				confidence = "stated"
			}
			root := newMintFixture(t, confidence)
			if tt.superseded != "" {
				path := filepath.Join(root, "blueprint", "spec", "checkout.md")
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				data = bytes.Replace(data, []byte("\n\n## Edges"), []byte("\nsuperseded-by: "+tt.superseded+"\n\n## Edges"), 1)
				if err := os.WriteFile(path, data, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if tt.blocks != "" {
				open := "# Open\n\n## currency\nstatus: OPEN\npass: rules\nasked: 2026-11-08\nowner: founder\nquestion: Which currency?\ncost: payment rewrite\nblocks: " + tt.blocks + "\n"
				if err := os.WriteFile(filepath.Join(root, "blueprint", "OPEN.md"), []byte(open), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			args := append([]string{"checkout/accepts-card"}, tt.args...)
			code, out, err := runMint(root, args...)
			if code != 1 || err == nil || !strings.Contains(err.Error(), tt.want) || out != "" {
				t.Fatalf("code=%d out=%q err=%v, want error containing %q and no output", code, out, err, tt.want)
			}
		})
	}
}

func TestMintCommandAcceptedByMint(t *testing.T) {
	if _, err := exec.LookPath("mint"); err != nil {
		t.Skip("mint is not on PATH; cannot check blueprint/mint compatibility")
	}
	root := newMintFixture(t, "stated")
	if err := os.WriteFile(filepath.Join(root, "blueprint", "REVIEW.md"), []byte(mintReview), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MINT_STATE_HOME", t.TempDir())
	run := func(name string, args ...string) []byte {
		t.Helper()
		cmd := exec.Command(name, args...)
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%s %v: %v\n%s", name, args, err, out)
		}
		return out
	}
	run("git", "-c", "init.defaultBranch=main", "init")
	run("git", "add", "blueprint")
	run("git", "-c", "user.name=Blueprint test", "-c", "user.email=blueprint@example.invalid", "-c", "commit.gpgsign=false", "-c", "core.hooksPath=/dev/null", "commit", "-m", "fixture")
	code, command, err := runMint(root, "checkout/accepts-card", "--scope", "cli/**")
	if code != 0 || err != nil {
		t.Fatalf("code=%d out=%q err=%v", code, command, err)
	}
	out := run("sh", "-c", command)
	var created struct {
		SpecPath string `json:"specPath"`
	}
	if err := json.Unmarshal(out, &created); err != nil {
		t.Fatalf("mint spec new output is not JSON: %v\n%s", err, out)
	}
	if created.SpecPath == "" {
		t.Fatalf("mint spec new reported no specPath: %s", out)
	}
	validated := run("mint", "spec", "validate", created.SpecPath)
	t.Logf("mint accepted generated command and validated reported specPath: %s", validated)
}
