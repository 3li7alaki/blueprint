package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWriteAndReadCommands(t *testing.T) {
	root := t.TempDir()
	fixed := time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC)
	run := func(args ...string) (int, string, error) {
		t.Helper()
		var out bytes.Buffer
		instance := App{Out: &out, Err: &bytes.Buffer{}, Now: func() time.Time { return fixed }, Version: "test"}
		args = append([]string{"--root", root}, args...)
		code, err := instance.Run(args)
		return code, out.String(), err
	}
	commands := [][]string{
		{"init"},
		{"spec", "new", "checkout"},
		{"req", "add", "checkout/accepts-card", "--ears", "WHEN a valid card is submitted, THE system SHALL accept it.", "--fit", "returns success", "--confidence", "stated"},
		{"open", "add", "currency", "--question", "Which currency?", "--cost", "payment rewrite", "--blocks", "checkout/accepts-card"},
		{"decide", "currency-choice", "--context", "A currency is needed.", "--decision", "Use USD.", "--because", "The founder selected it."},
	}
	for _, command := range commands {
		if code, _, err := run(command...); code != 0 || err != nil {
			t.Fatalf("%v: code=%d err=%v", command, code, err)
		}
	}
	tests := []struct {
		args []string
		want string
	}{
		{args: []string{"spec", "list"}, want: "checkout"},
		{args: []string{"req", "show", "checkout/accepts-card"}, want: "returns success"},
		{args: []string{"req", "list", "--blocked"}, want: "checkout/accepts-card"},
		{args: []string{"open", "list", "--status", "OPEN"}, want: "currency"},
		{args: []string{"ask", "frame", "--depth", "quick", "--batch", "1"}, want: "who-hurts"},
		{args: []string{"mint", "checkout/accepts-card"}, want: "mint spec new"},
		{args: []string{"trace", "checkout/accepts-card"}, want: "code:"},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			code, out, err := run(tt.args...)
			if code != 0 || err != nil || !strings.Contains(out, tt.want) {
				t.Fatalf("code=%d out=%q err=%v, want %q", code, out, err, tt.want)
			}
		})
	}
}

func TestEveryCommandAcceptsJSON(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer
	instance := App{Out: &out, Err: &bytes.Buffer{}, Now: time.Now, Version: "test"}
	if code, err := instance.Run([]string{"--root", root, "--json", "init"}); code != 0 || err != nil {
		t.Fatal(code, err)
	}
	out.Reset()
	if code, err := instance.Run([]string{"--root", root, "--json", "spec", "list"}); code != 0 || err != nil {
		t.Fatal(code, err)
	}
	if !strings.HasPrefix(out.String(), "[") {
		t.Fatalf("output is not JSON: %q", out.String())
	}
}

func TestWritersPreserveUntouchedLines(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "blueprint", "spec", "feature.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	before := "# feature\nstatus: ready\ndepth: quick\n\n## Intent\ncustom spacing  \n\n## Surfaces\n\n## Requirements\n\n## Edges\n| edge | answer |\n\n## Out of scope\n\n## Depends on\n"
	if err := os.WriteFile(path, []byte(before), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	instance := App{Out: &out, Err: &bytes.Buffer{}, Now: time.Now}
	code, err := instance.Run([]string{"--root", root, "req", "add", "feature/works", "--ears", "WHEN called, THE system SHALL work.", "--fit", "returns success", "--confidence", "stated"})
	if code != 0 || err != nil {
		t.Fatal(code, err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(after), "custom spacing  \n") {
		t.Fatalf("untouched line changed:\n%s", after)
	}
}
