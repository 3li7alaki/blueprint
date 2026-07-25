package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBrownfieldWriters(t *testing.T) {
	root := t.TempDir()
	run := testRunner(t, root)
	for _, command := range [][]string{
		{"init"},
		{"spec", "new", "pay"},
		{"req", "add", "pay/charges-once", "--ears", "WHEN charged, THE system SHALL charge once.", "--fit", "one charge exists", "--confidence", "derived", "--evidence", "src/pay.go:12"},
		{"req", "correct", "pay/charges-once", "--ears", "WHEN charged, THE system SHALL reject duplicates.", "--reason", "The owner clarified the rule."},
		{"req", "bug", "pay/charges-once", "--reason", "Duplicates are currently accepted."},
	} {
		if code, _, err := run(command...); code != 0 || err != nil {
			t.Fatalf("%v: code=%d err=%v", command, code, err)
		}
	}
	_, shown, err := run("req", "show", "pay/charges-once")
	if err != nil || !strings.Contains(shown, "stated\nWHEN charged, THE system SHALL reject duplicates.") || !strings.Contains(shown, "bug: Duplicates are currently accepted.") {
		t.Fatalf("show = %q, err=%v", shown, err)
	}
	if code, _, err := run("req", "fixed", "pay/charges-once"); code != 0 || err != nil {
		t.Fatal(code, err)
	}
	_, shown, err = run("req", "show", "pay/charges-once")
	if err != nil || strings.Contains(shown, "bug:") {
		t.Fatalf("show after fixed = %q, err=%v", shown, err)
	}
}

func TestHarvestScopeIsIdempotent(t *testing.T) {
	root := t.TempDir()
	run := testRunner(t, root)
	if code, _, err := run("init"); code != 0 || err != nil {
		t.Fatal(code, err)
	}
	for i := 0; i < 2; i++ {
		if code, _, err := run("harvest", "scope", "src/auth/**"); code != 0 || err != nil {
			t.Fatal(code, err)
		}
	}
	data, err := os.ReadFile(filepath.Join(root, "blueprint", "PROJECT.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(data), "- src/auth/**") != 1 {
		t.Fatalf("PROJECT.md = %s", data)
	}
}

func TestAskLedgerAndBraceSubstitution(t *testing.T) {
	root := t.TempDir()
	run := testRunner(t, root)
	if code, _, err := run("init"); code != 0 || err != nil {
		t.Fatal(code, err)
	}
	if code, _, err := run("ask", "done", "who-hurts"); code != 0 || err != nil {
		t.Fatal(code, err)
	}
	_, out, err := run("ask", "frame", "--depth", "quick", "--batch", "1")
	if err != nil || strings.Contains(out, "who-hurts") {
		t.Fatalf("ledger was ignored: %q, %v", out, err)
	}
	_, out, err = run("ask", "nouns", "--depth", "paranoid", "--batch", "20", "--for", "invoice", "--all")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "{entity}") || !strings.Contains(out, "invoice") {
		t.Fatalf("brace substitution failed: %q", out)
	}
}

func TestInitHooksPreservesExistingHook(t *testing.T) {
	root := t.TempDir()
	settings := filepath.Join(root, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settings), 0o755); err != nil {
		t.Fatal(err)
	}
	before := `{"theme":"dark","hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"echo keep"}]}]}}`
	if err := os.WriteFile(settings, []byte(before), 0o644); err != nil {
		t.Fatal(err)
	}
	run := testRunner(t, root)
	if code, _, err := run("init", "--hooks"); code != 0 || err != nil {
		t.Fatal(code, err)
	}
	data, err := os.ReadFile(settings)
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, command := range []string{"echo keep", "blueprint hook session", "blueprint hook pre-write", "blueprint hook done"} {
		if strings.Count(text, command) != 1 {
			t.Fatalf("%q count in settings = %d\n%s", command, strings.Count(text, command), text)
		}
	}
	if value["theme"] != "dark" {
		t.Fatalf("unrelated key changed: %#v", value)
	}
}

func testRunner(t *testing.T, root string) func(...string) (int, string, error) {
	t.Helper()
	return func(args ...string) (int, string, error) {
		t.Helper()
		var out bytes.Buffer
		instance := App{In: strings.NewReader("{}"), Out: &out, Err: &bytes.Buffer{}, Now: func() time.Time {
			return time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC)
		}, Version: "test"}
		code, err := instance.Run(append([]string{"--root", root}, args...))
		return code, out.String(), err
	}
}
