package app

import (
	"testing"

	"blueprint/internal/parser"
)

func TestSupersedeRequirement(t *testing.T) {
	tests := []struct {
		name         string
		evidence     string
		bug          string
		duplicateFit bool
	}{
		{name: "no optional lines"},
		{name: "evidence", evidence: "src/checkout.go:12"},
		{name: "bug", bug: "Cards are currently rejected."},
		{name: "evidence and bug", evidence: "src/checkout.go:12", bug: "Cards are currently rejected."},
		{name: "identical fit lines", duplicateFit: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			run := testRunner(t, root)
			commands := [][]string{{"spec", "new", "checkout"}}
			if tt.duplicateFit {
				commands = append(commands, []string{"req", "add", "checkout/keeps-card",
					"--ears", "WHEN a card is stored, THE system SHALL keep it.",
					"--fit", "returns status 201", "--confidence", "stated"})
			}
			add := []string{"req", "add", "checkout/accepts-card",
				"--ears", "WHEN a card is submitted, THE system SHALL accept it.",
				"--fit", "returns status 201", "--confidence", "stated"}
			if tt.evidence != "" {
				add = append(add, "--evidence", tt.evidence)
			}
			commands = append(commands, add)
			if tt.bug != "" {
				commands = append(commands, []string{"req", "bug", "checkout/accepts-card", "--reason", tt.bug})
			}
			for _, command := range commands {
				if code, _, err := run(command...); code != 0 || err != nil {
					t.Fatalf("%v: code=%d err=%v", command, code, err)
				}
			}
			before, errs := parser.Specs(root)
			if len(errs) != 0 || len(before) != 1 {
				t.Fatalf("before supersede: specs=%v errors=%v", before, errs)
			}

			if code, _, err := run("supersede", "accepts-card", "accepts-token"); code != 0 || err != nil {
				t.Fatalf("supersede: code=%d err=%v", code, err)
			}

			after, errs := parser.Specs(root)
			if len(errs) != 0 || len(after) != 1 {
				t.Fatalf("after supersede: specs=%v errors=%v", after, errs)
			}
			wantCount := 1
			if tt.duplicateFit {
				wantCount++
			}
			if len(after[0].Requirements) != wantCount || len(before[0].Requirements) != wantCount {
				t.Fatalf("requirement counts: before=%d after=%d, want %d", len(before[0].Requirements), len(after[0].Requirements), wantCount)
			}
			for i, got := range after[0].Requirements {
				want := before[0].Requirements[i]
				if want.Slug == "accepts-card" {
					want.SupersededBy = "accepts-token"
				}
				if got != want {
					t.Errorf("requirement changed: got %+v, want %+v", got, want)
				}
			}
		})
	}
}
