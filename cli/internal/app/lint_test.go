package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func lintFixture(t *testing.T) (string, func(...string) (int, string, error)) {
	t.Helper()
	root := t.TempDir()
	run := testRunner(t, root)
	if code, out, err := run("spec", "new", "checkout"); code != 0 || err != nil {
		t.Fatalf("spec new: code=%d out=%q err=%v", code, out, err)
	}
	return root, run
}

func lintAddArgs(ears, fit string) []string {
	return []string{"req", "add", "checkout/accepts-card", "--ears", ears, "--fit", fit, "--confidence", "stated"}
}

// @spec lint/ears-has-shall-and-full-stop
func TestReqAddRefusesEARSShapeWithoutWriting(t *testing.T) {
	root, run := lintFixture(t)
	path := filepath.Join(root, "blueprint", "spec", "checkout.md")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, ears := range []string{"THE system accepts cards.", "THE system SHALL accept cards"} {
		code, out, err := run(lintAddArgs(ears, "returns 201")...)
		if code != 1 || err == nil || !strings.Contains(err.Error(), "ears-shape") || out != "" {
			t.Fatalf("code=%d out=%q err=%v, want ears-shape refusal", code, out, err)
		}
		after, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(before, after) {
			t.Fatal("refused req add changed spec bytes")
		}
	}
}

// @spec lint/one-shall-per-requirement
func TestReqAddRefusesTwoSHALLClauses(t *testing.T) {
	_, run := lintFixture(t)
	code, out, err := run(lintAddArgs("WHEN a card is submitted, THE system SHALL accept it and SHALL record payment.", "returns 201")...)
	if code != 1 || err == nil || !strings.Contains(err.Error(), "one-shall") || out != "" {
		t.Fatalf("code=%d out=%q err=%v, want one-shall refusal", code, out, err)
	}
}

// @spec lint/no-vague-terms
func TestReqAddRefusesVagueFitAndAcceptsBreakfast(t *testing.T) {
	_, run := lintFixture(t)
	ears := "WHEN a card is submitted, THE system SHALL accept it."
	code, out, err := run(lintAddArgs(ears, "responds fast")...)
	if code != 1 || err == nil || !strings.Contains(err.Error(), "vague-term: fast") || out != "" {
		t.Fatalf("code=%d out=%q err=%v, want vague-term: fast refusal", code, out, err)
	}
	code, out, err = run(lintAddArgs(ears, "breakfast is served at 8")...)
	if code != 0 || err != nil || !strings.Contains(out, "checkout/accepts-card") {
		t.Fatalf("breakfast fit: code=%d out=%q err=%v", code, out, err)
	}
}

// @spec lint/fit-is-observable
func TestReqAddRefusesUnobservableFitAndAcceptsStatus(t *testing.T) {
	_, run := lintFixture(t)
	ears := "WHEN a card is submitted, THE system SHALL accept it."
	code, out, err := run(lintAddArgs(ears, "returns success")...)
	if code != 1 || err == nil || !strings.Contains(err.Error(), "fit-not-observable") || out != "" {
		t.Fatalf("code=%d out=%q err=%v, want fit-not-observable refusal", code, out, err)
	}
	code, out, err = run(lintAddArgs(ears, "returns 201")...)
	if code != 0 || err != nil || !strings.Contains(out, "checkout/accepts-card") {
		t.Fatalf("status fit: code=%d out=%q err=%v", code, out, err)
	}
}

func TestReqAddListsEveryLintProblemBeforeReadingSpec(t *testing.T) {
	// A missing spec must not hide the input diagnostics.
	run := testRunner(t, t.TempDir())
	code, out, err := run(lintAddArgs("THE system SHALL respond fast and SHALL record payment", "returns success")...)
	want := "ears-shape; one-shall; vague-term: fast; fit-not-observable"
	if code != 1 || err == nil || err.Error() != want || out != "" {
		t.Fatalf("code=%d out=%q err=%v, want %q", code, out, err, want)
	}
}

func TestEARSEditWritersRefuseLintBeforeWriting(t *testing.T) {
	for _, command := range [][]string{{"req", "correct"}, {"amend"}} {
		t.Run(strings.Join(command, " "), func(t *testing.T) {
			root, run := lintFixture(t)
			if code, out, err := run(lintAddArgs("WHEN called, THE system SHALL accept cards.", "returns 201")...); code != 0 || err != nil {
				t.Fatalf("req add: code=%d out=%q err=%v", code, out, err)
			}
			path := filepath.Join(root, "blueprint", "spec", "checkout.md")
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			for _, tt := range []struct{ ears, rule string }{
				{"THE system accepts cards.", "ears-shape"},
				{"THE system SHALL accept cards", "ears-shape"},
				{"THE system SHALL accept cards and SHALL record payment.", "one-shall"},
				{"THE system SHALL respond fast.", "vague-term: fast"},
			} {
				args := append(append([]string{}, command...), "checkout/accepts-card", "--ears", tt.ears, "--reason", "Owner clarified behavior.")
				code, out, err := run(args...)
				if code != 1 || err == nil || !strings.Contains(err.Error(), tt.rule) || out != "" {
					t.Fatalf("%q: code=%d out=%q err=%v, want %s", tt.ears, code, out, err, tt.rule)
				}
				after, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(before, after) {
					t.Fatal("refused edit changed spec bytes")
				}
				if _, err := os.Stat(filepath.Join(root, "blueprint", "decisions")); !os.IsNotExist(err) {
					t.Fatalf("refused edit created decisions directory: %v", err)
				}
			}
			code, out, err := run(append(append([]string{}, command...), "checkout/accepts-card", "--ears", "WHEN called, THE system SHALL return 201.", "--reason", "Owner clarified behavior.")...)
			if code != 0 || err != nil {
				t.Fatalf("valid edit: code=%d out=%q err=%v", code, out, err)
			}
		})
	}
}
