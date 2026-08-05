package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// @spec look/product-is-stated
func TestLookCommands(t *testing.T) {
	root := t.TempDir()
	run := func(stdin string, args ...string) (int, string, string, error) {
		t.Helper()
		var out, errs bytes.Buffer
		instance := App{Out: &out, Err: &errs, In: strings.NewReader(stdin), Now: time.Now, Version: "test"}
		code, err := instance.Run(append([]string{"--root", root}, args...))
		return code, out.String(), errs.String(), err
	}

	// Whoever runs the interview may create the file, including a design skill with better
	// questions than a bank. Only what comes after is guarded.
	payload := `{"tool_input":{"file_path":"` + filepath.ToSlash(filepath.Join(root, "PRODUCT.md")) + `"}}`
	if code, _, _, err := run(payload, "hook", "pre-write"); code != 0 || err != nil {
		t.Fatalf("creating PRODUCT.md must be allowed: code=%d err=%v", code, err)
	}

	if code, out, _, err := run("", "look", "new"); code != 0 || err != nil || !strings.Contains(out, "PRODUCT.md") {
		t.Fatalf("look new: code=%d out=%q err=%v", code, out, err)
	}
	if _, _, _, err := run("", "look", "new"); err == nil {
		t.Fatal("look new must refuse to replace an existing PRODUCT.md")
	}

	// The template ships every section unanswered, which is what the look gate reports on. A
	// reader asking for one section must get that section and nothing else.
	path := filepath.Join(root, "PRODUCT.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(strings.Replace(string(data), "## Register\n", "## Register\n\nproduct\n", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	code, out, _, err := run("", "look", "show", "--section", "register")
	if code != 0 || err != nil || strings.TrimSpace(out) != "product" {
		t.Fatalf("look show: code=%d out=%q err=%v", code, out, err)
	}

	code, _, errs, err := run(payload, "hook", "pre-write")
	if code != 2 || err != nil || !strings.Contains(errs, "BLOCKED") {
		t.Fatalf("pre-write must block a hand edit to PRODUCT.md: code=%d err=%q %v", code, errs, err)
	}
}
