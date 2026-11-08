package lint

import (
	"reflect"
	"testing"
)

// @spec lint/ears-has-shall-and-full-stop
func TestShape(t *testing.T) {
	for _, tt := range []struct {
		name string
		line string
		want bool
	}{
		{name: "IF THEN", line: "IF a card is declined, THEN THE system SHALL show the decline code.", want: true},
		{name: "ubiquitous", line: "THE system SHALL log every refund.", want: true},
		{name: "WHERE", line: "WHERE refunds are enabled, THE system SHALL log every refund.", want: true},
		{name: "missing SHALL", line: "THE system logs every refund."},
		{name: "missing full stop", line: "THE system SHALL log every refund"},
		{name: "missing spaced SHALL", line: "THE system SHALL\tlog every refund."},
		{name: "trailing space", line: "THE system SHALL log every refund. "},
		{name: "other lint violations", line: "THE system SHALL respond fast and SHALL log refunds.", want: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := Shape(tt.line); got != tt.want {
				t.Fatalf("Shape(%q) = %v, want %v", tt.line, got, tt.want)
			}
			shapeProblem := false
			for _, problem := range EARS(tt.line) {
				if problem.Rule == "ears-shape" {
					shapeProblem = true
				}
			}
			if shapeProblem != !tt.want {
				t.Fatalf("EARS(%q) reports ears-shape = %v, want %v", tt.line, shapeProblem, !tt.want)
			}
		})
	}
}

func TestLint(t *testing.T) {
	for _, tt := range []struct {
		name  string
		check func(string) []Problem
		line  string
		want  []Problem
	}{
		{name: "ears missing SHALL", check: EARS, line: "THE system accepts cards.", want: []Problem{{Rule: "ears-shape"}}},
		{name: "ears missing full stop", check: EARS, line: "THE system SHALL accept cards", want: []Problem{{Rule: "ears-shape"}}},
		{name: "ears missing spaced SHALL", check: EARS, line: "THE system SHALL\taccept cards.", want: []Problem{{Rule: "ears-shape"}}},
		{name: "ubiquitous ears", check: EARS, line: "THE system SHALL accept cards."},
		{name: "triggered ears", check: EARS, line: "WHEN called, THE system SHALL return 201."},
		{name: "two SHALL clauses", check: EARS, line: "THE system SHALL accept cards and SHALL record payment.", want: []Problem{{Rule: "one-shall"}}},
		{name: "SHALL word boundary", check: EARS, line: "THE system SHALL accept SHALLOW cards."},
		{name: "vague ears", check: EARS, line: "THE system SHALL respond fast.", want: []Problem{{Rule: "vague-term", Detail: "fast"}}},
		{name: "all ears problems in order", check: EARS, line: "THE system SHALL be robust and SHALL respond fast", want: []Problem{{Rule: "ears-shape"}, {Rule: "one-shall"}, {Rule: "vague-term", Detail: "fast"}, {Rule: "vague-term", Detail: "robust"}}},
		{name: "vague fit", check: Fit, line: "responds fast", want: []Problem{{Rule: "vague-term", Detail: "fast"}, {Rule: "fit-not-observable"}}},
		{name: "breakfast", check: Fit, line: "breakfast is served at 8"},
		{name: "hyphenated boundary", check: Fit, line: "1 fast-paced response"},
		{name: "slash boundary", check: Fit, line: "fast/slow"},
		{name: "letter and digit boundaries", check: Fit, line: "fast2 2fast fasté éfast"},
		{name: "user-friendly unit", check: Fit, line: "1 user-friendly response", want: []Problem{{Rule: "vague-term", Detail: "user-friendly"}}},
		{name: "user-friendly boundary", check: Fit, line: "1 user-friendly-style response"},
		{name: "and/or unit", check: Fit, line: "accepts cards and/or cash", want: []Problem{{Rule: "vague-term", Detail: "and/or"}}},
		{name: "and/or boundary", check: Fit, line: "accepts cards/and/or/cash"},
		{name: "phrase", check: Fit, line: "1 response as far as possible", want: []Problem{{Rule: "vague-term", Detail: "as far as possible"}}},
		{name: "phrase boundary", check: Fit, line: "1 response as far as possible-now"},
		{name: "overlapping phrases", check: Fit, line: "1 response as appropriate", want: []Problem{{Rule: "vague-term", Detail: "appropriate"}, {Rule: "vague-term", Detail: "as appropriate"}}},
		{name: "case folding and distinct terms in list order", check: Fit, line: "1 ROBUST FAST fast", want: []Problem{{Rule: "vague-term", Detail: "fast"}, {Rule: "vague-term", Detail: "robust"}}},
		{name: "unobservable fit", check: Fit, line: "returns success", want: []Problem{{Rule: "fit-not-observable"}}},
		{name: "numeric fit", check: Fit, line: "returns 201"},
		{name: "unicode digit", check: Fit, line: "returns ١"},
		{name: "backticked fit", check: Fit, line: "returns `ok`"},
		{name: "quoted fit", check: Fit, line: `returns "ok"`},
		{name: "path fit", check: Fit, line: "writes out/report.json"},
		{name: "quantity fit", check: Fit, line: "exactly one row"},
		{name: "empty backticks", check: Fit, line: "returns ``", want: []Problem{{Rule: "fit-not-observable"}}},
		{name: "empty quotes", check: Fit, line: `returns ""`, want: []Problem{{Rule: "fit-not-observable"}}},
		{name: "unclosed backtick", check: Fit, line: "returns `ok", want: []Problem{{Rule: "fit-not-observable"}}},
		{name: "unclosed quote", check: Fit, line: `returns "ok`, want: []Problem{{Rule: "fit-not-observable"}}},
		{name: "quantity word boundary", check: Fit, line: "returns someone", want: []Problem{{Rule: "fit-not-observable"}}},
		{name: "quantity outside fixed set", check: Fit, line: "returns thirteen rows", want: []Problem{{Rule: "fit-not-observable"}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.check(tt.line); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("lint(%q) = %#v, want %#v", tt.line, got, tt.want)
			}
		})
	}
}

func TestFitQuantityWords(t *testing.T) {
	for _, word := range []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten", "eleven", "twelve", "first", "second", "third", "fourth", "fifth", "no", "none", "exactly", "every", "each", "EACH"} {
		t.Run(word, func(t *testing.T) {
			if got := Fit(word + " row"); len(got) != 0 {
				t.Fatalf("quantity word rejected: %v", got)
			}
		})
	}
}

func TestProblemError(t *testing.T) {
	for _, tt := range []struct {
		problem Problem
		want    string
	}{
		{Problem{Rule: "ears-shape"}, "ears-shape"},
		{Problem{Rule: "vague-term", Detail: "fast"}, "vague-term: fast"},
	} {
		if got := tt.problem.Error(); got != tt.want {
			t.Fatalf("Error() = %q, want %q", got, tt.want)
		}
	}
}
