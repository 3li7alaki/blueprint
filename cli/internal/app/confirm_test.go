package app

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// The spec template explains the format in prose, so the literal token `derived` appears in a
// guidance comment above the requirements. A whole-file replace edits that comment, reports
// success, and leaves the requirement derived: the anti-hallucination clause silently off.
func TestConfirmEditsTheRequirementNotTheGuidance(t *testing.T) {
	root := t.TempDir()
	fixed := time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC)
	run := func(args ...string) (int, string, error) {
		t.Helper()
		var out bytes.Buffer
		instance := App{Out: &out, Err: &bytes.Buffer{}, Now: func() time.Time { return fixed }, Version: "test"}
		return func() (int, string, error) {
			code, err := instance.Run(append([]string{"--root", root}, args...))
			return code, out.String(), err
		}()
	}

	for _, command := range [][]string{
		{"init"},
		{"spec", "new", "pay"},
		{"req", "add", "pay/refund-is-idempotent",
			"--ears", "WHEN a refund is requested twice, THE system SHALL apply it once.",
			"--fit", "the second call returns the first refund id",
			"--confidence", "derived"},
		{"req", "add", "pay/logs-every-attempt",
			"--ears", "WHEN a refund is attempted, THE system SHALL record it.",
			"--fit", "an audit row exists",
			"--confidence", "derived"},
	} {
		if code, _, err := run(command...); code != 0 || err != nil {
			t.Fatalf("%v: code=%d err=%v", command, code, err)
		}
	}

	if code, _, err := run("req", "confirm", "pay/refund-is-idempotent"); code != 0 || err != nil {
		t.Fatalf("confirm: code=%d err=%v", code, err)
	}

	_, out, err := run("req", "show", "pay/refund-is-idempotent")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, "stated") {
		t.Errorf("confirm reported success but confidence is still %q", strings.SplitN(out, "\n", 2)[0])
	}

	// The neighbouring requirement must be untouched: an edit that walks into the next block
	// confirms something no human confirmed.
	_, out, err = run("req", "show", "pay/logs-every-attempt")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, "derived") {
		t.Errorf("confirm leaked into the next requirement, which now reads %q", strings.SplitN(out, "\n", 2)[0])
	}

	// Assert on parsed state, not on the raw text: the template's guidance comment contains
	// example tokens, so a textual count would measure the documentation.
	_, out, err = run("req", "list", "--derived")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "pay/logs-every-attempt" {
		t.Errorf("expected exactly one requirement left derived, got %q", strings.TrimSpace(out))
	}
}
