package reqfile

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

const minimal = "ears: WHEN called, THE system SHALL answer.\nfit: test: The response is successful.\nconfidence: stated\nscope: cli/internal/example/**\n"

func fixture(t *testing.T, category, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "formats", "fixtures", category, "requirement--"+name+".md"))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestFixtures(t *testing.T) {
	for _, category := range []string{"valid", "invalid"} {
		paths, err := filepath.Glob(filepath.Join("..", "..", "..", "formats", "fixtures", category, "requirement--*.md"))
		if err != nil {
			t.Fatal(err)
		}
		if (category == "valid" && len(paths) < 4) || (category == "invalid" && len(paths) != 19) {
			t.Fatalf("%s fixture count = %d", category, len(paths))
		}
		for _, path := range paths {
			name := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(path), "requirement--"), ".md")
			t.Run(category+"/"+name, func(t *testing.T) {
				_, problems := Parse("fixture/"+name, fixture(t, category, name))
				if category == "valid" && len(problems) != 0 {
					t.Fatalf("problems = %v", problems)
				}
				if category == "invalid" && (len(problems) != 1 || problems[0].Rule != name) {
					t.Fatalf("problems = %v, want exactly %q", problems, name)
				}
			})
		}
	}
}

func TestParsedFields(t *testing.T) {
	tests := []struct {
		name string
		want Requirement
	}{
		{name: "minimal", want: Requirement{
			ID: "fixture/minimal", EARS: "WHEN called, THE system SHALL answer.",
			FitKind: "test", Fit: "The response is successful.", Confidence: "stated",
			Scope: []string{"cli/internal/example/**"},
		}},
		{name: "linked", want: Requirement{
			ID: "fixture/linked", EARS: "WHILE [[money:amount]] is pending, THE system SHALL retain [[billing:invoice]].",
			FitKind: "property", Fit: "Every [[billing:invoice]] preserves [[money:amount]].",
			Confidence: "derived", Evidence: "cli/internal/billing/amount.go:12", Tier: "consequential",
			SupersededBy: "billing/retains-complete-invoice",
			Scope:        []string{"cli/internal/billing/**", "cli/internal/checkout/**"},
			Needs:        []string{"billing/defines-amount", "billing/loads-invoice"},
			After:        []string{"billing/opens-checkout"}, Terms: []string{"billing:invoice", "money:amount"},
			Body: "Rationale keeps its spacing.  \n\n[[body text is not validated\nunknown: body content\n",
		}},
		{name: "judged", want: Requirement{
			ID: "fixture/judged", EARS: "IF called, THE system SHALL answer.",
			FitKind: "judged", Fit: "The response matches the reference.", Confidence: "stated",
			ReferencePath: "reference/answer.txt", ReferenceDigest: strings.Repeat("0123456789abcdef", 4),
			Scope: []string{"cli/internal/example/**"}, Tier: "normal",
			Body: "A reviewer compares the explanation.\n",
		}},
		{name: "where", want: Requirement{
			ID: "fixture/where", EARS: "WHERE responses are enabled, THE system SHALL answer.",
			FitKind: "test", Fit: "The response is successful.", Confidence: "stated",
			Scope: []string{"cli/internal/example/**"}, Tier: "chore", Evidence: "A stated rule with café text.",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := fixture(t, "valid", tt.name)
			for _, newline := range []string{"\n", "\r\n"} {
				got, problems := Parse(tt.want.ID, []byte(strings.ReplaceAll(string(data), "\n", newline)))
				if len(problems) != 0 || !reflect.DeepEqual(got, tt.want) {
					t.Fatalf("newline %q: got %#v, %v; want %#v", newline, got, problems, tt.want)
				}
			}
		})
	}
}

func TestParseValidation(t *testing.T) {
	tests := []struct {
		name, data, rule string
		line             int
	}{
		{name: "missing period", data: strings.Replace(minimal, "answer.", "answer", 1), rule: "bad-ears-shape", line: 1},
		{name: "missing SHALL", data: strings.Replace(minimal, " SHALL ", " will ", 1), rule: "bad-ears-shape", line: 1},
		{name: "empty ears", data: strings.Replace(minimal, "WHEN called, THE system SHALL answer.", "", 1), rule: "bad-ears-shape", line: 1},
		{name: "empty confidence", data: strings.Replace(minimal, "stated", "", 1), rule: "bad-confidence", line: 3},
		{name: "empty scope", data: strings.Replace(minimal, "cli/internal/example/**", "", 1), rule: "bad-scope", line: 4},
		{name: "fit separator", data: strings.Replace(minimal, "test: The response is successful.", "test", 1), rule: "bad-fit-kind", line: 2},
		{name: "blank criterion", data: strings.Replace(minimal, "The response is successful.", " \t", 1), rule: "empty-fit-criterion", line: 2},
		{name: "after id", data: minimal + "after: billing/valid, invalid\n", rule: "bad-dependency-id", line: 5},
		{name: "empty dependency", data: minimal + "needs: \n", rule: "bad-dependency-id", line: 5},
		{name: "empty supersession", data: minimal + "superseded-by: \n", rule: "bad-superseded-by", line: 5},
		{name: "fit term", data: strings.Replace(minimal, "The response is successful.", "[[billing:]] is valid.", 1), rule: "bad-term-ref", line: 2},
		{name: "nested term", data: strings.Replace(minimal, "The response is successful.", "[[[[billing:amount]] is valid.", 1), rule: "bad-term-ref", line: 2},
		{name: "uppercase key", data: minimal + "Evidence: source\n", rule: "malformed-line", line: 5},
		{name: "two spaces", data: minimal + "evidence:  source\n", rule: "malformed-line", line: 5},
		{name: "tab separator", data: minimal + "evidence:\tsource\n", rule: "malformed-line", line: 5},
		{name: "missing separator", data: minimal + "evidence:source\n", rule: "malformed-line", line: 5},
		{name: "bare carriage return", data: minimal + "evidence: source\r\nmore\rtext\n", rule: "malformed-line", line: 6},
		{name: "duplicate", data: minimal + "confidence: derived\n", rule: "duplicate-key", line: 5},
		{name: "unknown", data: minimal + "other-key: value\n", rule: "unknown-key", line: 5},
		{name: "invalid UTF-8", data: minimal + "\xff", rule: "malformed-line", line: 0},
		{name: "uppercase digest", data: strings.Replace(minimal, "fit: test:", "fit: judged:", 1) + "reference: reference/answer.txt sha256:" + strings.Repeat("A", 64), rule: "bad-reference", line: 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, problems := Parse("fixture/validation", []byte(tt.data))
			if len(problems) != 1 || problems[0].Rule != tt.rule || problems[0].Line != tt.line {
				t.Fatalf("problems = %v, want %s at %d", problems, tt.rule, tt.line)
			}
		})
	}
	_, problems := Parse("invalid", []byte(minimal))
	if len(problems) != 1 || problems[0].Rule != "bad-id" || problems[0].Line != 0 {
		t.Fatalf("bad id problems = %v", problems)
	}
}

func TestBodyAndHeaderBoundaries(t *testing.T) {
	for _, body := range []string{"", "\n", "  rationale  \n\n[[invalid\nfit: unvalidated\n"} {
		got, problems := Parse("fixture/body", []byte(minimal+"\n"+body))
		if len(problems) != 0 || got.Body != body {
			t.Fatalf("body = %q, %v; want %q", got.Body, problems, body)
		}
	}
	_, problems := Parse("fixture/empty", nil)
	want := []Problem{
		{Rule: "missing-ears", Msg: "required key ears is missing"},
		{Rule: "missing-fit", Msg: "required key fit is missing"},
		{Rule: "missing-confidence", Msg: "required key confidence is missing"},
		{Rule: "missing-scope", Msg: "required key scope is missing"},
	}
	if !reflect.DeepEqual(problems, want) {
		t.Fatalf("problems = %v, want %v", problems, want)
	}
	got, problems := Parse("fixture/no-newline", []byte(strings.TrimSuffix(minimal, "\n")))
	if len(problems) != 0 || got.Body != "" {
		t.Fatalf("no final newline: %#v, %v", got, problems)
	}
}

func TestIDFromPath(t *testing.T) {
	root := filepath.Join(t.TempDir(), "blueprint", "spec")
	tests := []struct {
		name, path, want string
	}{
		{name: "good", path: filepath.Join(root, "billing", "checkout", "rounds-half-even.md"), want: "billing/checkout/rounds-half-even"},
		{name: "digits", path: filepath.Join(root, "v2", "rounds-2-times.md"), want: "v2/rounds-2-times"},
		{name: "one segment", path: filepath.Join(root, "billing.md")},
		{name: "bad segment", path: filepath.Join(root, "billing", "Bad.md")},
		{name: "double hyphen", path: filepath.Join(root, "billing", "bad--slug.md")},
		{name: "outside root", path: filepath.Join(filepath.Dir(root), "other", "billing", "rounds.md")},
		{name: "sibling prefix", path: filepath.Join(root+"-other", "billing", "rounds.md")},
		{name: "escape", path: root + "/../other/billing/rounds.md"},
		{name: "wrong extension", path: filepath.Join(root, "billing", "rounds.txt")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IDFromPath(root, tt.path)
			if tt.want != "" {
				if err != nil || got != tt.want {
					t.Fatalf("id = %q, error = %v, want %q", got, err, tt.want)
				}
			} else if problem, ok := err.(Problem); !ok || problem.Rule != "bad-id" || problem.Line != 0 || !strings.Contains(err.Error(), "bad-id") {
				t.Fatalf("error = %v, want bad-id", err)
			}
		})
	}
	id, err := IDFromPath("blueprint/spec", "blueprint/spec/billing/rounds.md")
	if err != nil || id != "billing/rounds" {
		t.Fatalf("relative id = %q, error = %v", id, err)
	}
}

func TestCheckGraph(t *testing.T) {
	const a, b, c = "feature/alpha", "feature/beta", "feature/gamma"
	tests := []struct {
		name string
		reqs []Requirement
		want []Problem
	}{
		{name: "empty"},
		{name: "no cycle", reqs: []Requirement{{ID: a, Needs: []string{b}}, {ID: b, After: []string{c}}, {ID: c}}},
		{name: "two node cycle", reqs: []Requirement{{ID: b, Needs: []string{a}}, {ID: a, Needs: []string{b}}}, want: []Problem{{Rule: "dependency-cycle", Msg: a + " -> " + b + " -> " + a}}},
		{name: "mixed three node cycle", reqs: []Requirement{{ID: c, After: []string{a}}, {ID: a, Needs: []string{b}}, {ID: b, After: []string{c}}}, want: []Problem{{Rule: "dependency-cycle", Msg: a + " -> " + b + " -> " + c + " -> " + a}}},
		{name: "unknown", reqs: []Requirement{{ID: a, After: []string{b}}}, want: []Problem{{Rule: "unknown-dependency", Msg: a + " -> " + b}}},
		{name: "self", reqs: []Requirement{{ID: a, Needs: []string{a}, After: []string{a}}}, want: []Problem{{Rule: "self-dependency", Msg: a + " -> " + a}}},
		{name: "duplicate edges", reqs: []Requirement{{ID: a, Needs: []string{b, b}, After: []string{b}}, {ID: b, After: []string{a}}}, want: []Problem{{Rule: "dependency-cycle", Msg: a + " -> " + b + " -> " + a}}},
		{name: "unknown and self order", reqs: []Requirement{{ID: b, Needs: []string{c, b, c}}, {ID: a, Needs: []string{c}}}, want: []Problem{
			{Rule: "unknown-dependency", Msg: a + " -> " + c},
			{Rule: "self-dependency", Msg: b + " -> " + b},
			{Rule: "unknown-dependency", Msg: b + " -> " + c},
		}},
		{name: "overlapping cycles", reqs: []Requirement{{ID: c, Needs: []string{b, a}}, {ID: b, Needs: []string{c, a}}, {ID: a, Needs: []string{c, b}}}, want: []Problem{
			{Rule: "dependency-cycle", Msg: a + " -> " + b + " -> " + a},
			{Rule: "dependency-cycle", Msg: a + " -> " + b + " -> " + c + " -> " + a},
			{Rule: "dependency-cycle", Msg: a + " -> " + c + " -> " + a},
			{Rule: "dependency-cycle", Msg: a + " -> " + c + " -> " + b + " -> " + a},
			{Rule: "dependency-cycle", Msg: b + " -> " + c + " -> " + b},
		}},
		{name: "supersession is not an edge", reqs: []Requirement{{ID: a, SupersededBy: b}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, reverse := range []bool{false, true} {
				reqs := slices.Clone(tt.reqs)
				if reverse {
					slices.Reverse(reqs)
				}
				for i := range reqs {
					reqs[i].Needs = slices.Clone(reqs[i].Needs)
					reqs[i].After = slices.Clone(reqs[i].After)
					if reverse {
						slices.Reverse(reqs[i].Needs)
						slices.Reverse(reqs[i].After)
					}
				}
				got := CheckGraph(reqs)
				if !reflect.DeepEqual(got, tt.want) {
					t.Fatalf("reversed %t: problems = %v, want %v", reverse, got, tt.want)
				}
				if again := CheckGraph(reqs); !reflect.DeepEqual(again, got) {
					t.Fatalf("repeated check changed: %v, then %v", got, again)
				}
			}
		})
	}
}
