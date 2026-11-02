package reqfile

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const minimalTerm = "definition: An amount is a quantity of money.\nstatus: existing\n"
const minimalOpen = "question: Which rounding rule applies?\ncost: Incorrect totals require ledger repairs.\nblocks: billing/rounds-half-even\nstatus: open\nasked: 2026-11-02\n"

func TestTermOpenFixtures(t *testing.T) {
	for _, kind := range []struct {
		name           string
		valid, invalid int
	}{{"term", 2, 8}, {"open", 3, 16}} {
		for _, category := range []string{"valid", "invalid"} {
			paths, err := filepath.Glob(filepath.Join("..", "..", "..", "formats", "fixtures", category, kind.name+"--*.md"))
			if err != nil {
				t.Fatal(err)
			}
			if (category == "valid" && len(paths) < kind.valid) || (category == "invalid" && len(paths) != kind.invalid) {
				t.Fatalf("%s %s fixture count = %d", kind.name, category, len(paths))
			}
			for _, path := range paths {
				name := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(path), kind.name+"--"), ".md")
				t.Run(kind.name+"/"+category+"/"+name, func(t *testing.T) {
					data, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					var problems []Problem
					if kind.name == "term" {
						_, problems = ParseTerm("fixture:"+name, data)
					} else {
						_, problems = ParseOpen(name, data)
					}
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
}

func TestTermOpenParsedFields(t *testing.T) {
	body := "  café rationale  \n\n[[invalid\nstatus: unvalidated\n"
	for _, newline := range []string{"\n", "\r\n"} {
		termData := minimalTerm + "states: draft , active,closed\n\n" + body
		term, problems := ParseTerm("money:amount", []byte(strings.ReplaceAll(termData, "\n", newline)))
		wantTerm := Term{ID: "money:amount", Definition: "An amount is a quantity of money.", Status: "existing", States: []string{"draft", "active", "closed"}, Body: body}
		if len(problems) != 0 || !reflect.DeepEqual(term, wantTerm) {
			t.Fatalf("newline %q: term = %#v, %v; want %#v", newline, term, problems, wantTerm)
		}
		openData := strings.Replace(minimalOpen, "status: open", "status: resolved", 1) + "owner: billing-team\nresolution: decisions/rounding-half-even\n\n" + body
		open, problems := ParseOpen("rounding-rule", []byte(strings.ReplaceAll(openData, "\n", newline)))
		wantOpen := Open{ID: "rounding-rule", Question: "Which rounding rule applies?", Cost: "Incorrect totals require ledger repairs.", Status: "resolved", Asked: "2026-11-02", Owner: "billing-team", Resolution: "decisions/rounding-half-even", Blocks: []string{"billing/rounds-half-even"}, Body: body}
		if len(problems) != 0 || !reflect.DeepEqual(open, wantOpen) {
			t.Fatalf("newline %q: open = %#v, %v; want %#v", newline, open, problems, wantOpen)
		}
	}
	open, problems := ParseOpen("prefixes", []byte(strings.Replace(minimalOpen, "billing/rounds-half-even", "billing/* , billing/checkout/*, review/loads-proposal", 1)))
	if len(problems) != 0 || !reflect.DeepEqual(open.Blocks, []string{"billing/*", "billing/checkout/*", "review/loads-proposal"}) {
		t.Fatalf("blocks = %v, %v", open.Blocks, problems)
	}
}

func TestObjectHeaderBoundaries(t *testing.T) {
	for _, parser := range []struct {
		name, minimal string
		parse         func([]byte) (string, []Problem)
	}{
		{"requirement", minimal, func(data []byte) (string, []Problem) { r, p := Parse("fixture/header", data); return r.Body, p }},
		{"term", minimalTerm, func(data []byte) (string, []Problem) { r, p := ParseTerm("fixture:header", data); return r.Body, p }},
		{"open", minimalOpen, func(data []byte) (string, []Problem) { r, p := ParseOpen("header", data); return r.Body, p }},
	} {
		t.Run(parser.name, func(t *testing.T) {
			for _, body := range []string{"", "\n", "  rationale  \n\n\ufeffunknown: body\n"} {
				got, problems := parser.parse([]byte(parser.minimal + "\n" + body))
				if len(problems) != 0 || got != body {
					t.Fatalf("body = %q, %v; want %q", got, problems, body)
				}
			}
			for _, data := range []string{parser.minimal, strings.TrimSuffix(parser.minimal, "\n")} {
				if body, problems := parser.parse([]byte(data)); len(problems) != 0 || body != "" {
					t.Fatalf("no body = %q, %v", body, problems)
				}
			}
			for _, encoding := range []struct{ data, rule string }{
				{"\xef\xbb\xbf" + parser.minimal, "byte-order-mark"},
				{"\xef\xbb\xbf", "byte-order-mark"},
				{parser.minimal + "\xff", "malformed-line"},
			} {
				_, problems := parser.parse([]byte(encoding.data))
				if len(problems) != 1 || problems[0].Rule != encoding.rule || problems[0].Line != 0 {
					t.Fatalf("encoding problems = %v, want %s at zero", problems, encoding.rule)
				}
			}
		})
	}
	term, problems := ParseTerm("fixture:duplicate", []byte(minimalTerm+"status: planned\n"))
	if len(problems) != 1 || problems[0].Line != 3 || term.Status != "existing" {
		t.Fatalf("duplicate term = %#v, %v", term, problems)
	}
	open, problems := ParseOpen("duplicate", []byte(minimalOpen+"status: deferred\n"))
	if len(problems) != 1 || problems[0].Line != 6 || open.Status != "open" {
		t.Fatalf("duplicate open = %#v, %v", open, problems)
	}
}

func TestTermOpenValidation(t *testing.T) {
	for _, tt := range []struct {
		name, data, rule string
		line             int
	}{
		{"definition empty", "definition: \nstatus: existing\n", "missing-definition", 1},
		{"empty status", "definition: An amount.\nstatus: \n", "bad-term-status", 2},
		{"states empty", minimalTerm + "states: \n", "bad-states", 3},
		{"states trailing comma", minimalTerm + "states: active,\n", "bad-states", 3},
	} {
		t.Run("term/"+tt.name, func(t *testing.T) {
			_, problems := ParseTerm("fixture:validation", []byte(tt.data))
			if len(problems) != 1 || problems[0].Rule != tt.rule || problems[0].Line != tt.line {
				t.Fatalf("problems = %v, want %s at %d", problems, tt.rule, tt.line)
			}
		})
	}
	for _, tt := range []struct {
		name, old, value, extra, rule string
		line                          int
	}{
		{name: "empty question", old: "Which rounding rule applies?", rule: "missing-question", line: 1},
		{name: "empty cost", old: "Incorrect totals require ledger repairs.", rule: "missing-cost", line: 2},
		{name: "trailing question space", old: "applies?", value: "applies? ", rule: "bad-question", line: 1},
		{name: "empty blocks", old: "billing/rounds-half-even", rule: "bad-blocks", line: 3},
		{name: "empty status", old: "status: open", value: "status: ", rule: "bad-open-status", line: 4},
		{name: "empty date", old: "2026-11-02", rule: "bad-asked", line: 5},
		{name: "nonleap date", old: "2026-11-02", value: "2025-02-29", rule: "bad-asked", line: 5},
		{name: "unpadded date", old: "2026-11-02", value: "2026-1-02", rule: "bad-asked", line: 5},
		{name: "empty resolution", old: "status: open", value: "status: resolved", extra: "resolution: \n", rule: "resolved-without-resolution", line: 4},
		{name: "deferred resolution", old: "status: open", value: "status: deferred", extra: "resolution: decisions/rounding\n", rule: "resolution-without-resolved", line: 6},
		{name: "blank forbidden resolution", extra: "resolution: \n", rule: "resolution-without-resolved", line: 6},
	} {
		t.Run("open/"+tt.name, func(t *testing.T) {
			data := minimalOpen
			if tt.old != "" {
				data = strings.Replace(data, tt.old, tt.value, 1)
			}
			_, problems := ParseOpen("validation", []byte(data+tt.extra))
			if len(problems) != 1 || problems[0].Rule != tt.rule || problems[0].Line != tt.line {
				t.Fatalf("problems = %v, want %s at %d", problems, tt.rule, tt.line)
			}
		})
	}
	for _, value := range []string{"billing", "billing/**", "*", "/*", "billing/*/extra", "billing/Bad/*", "billing/valid,", "../billing/*"} {
		_, problems := ParseOpen("blocks", []byte(strings.Replace(minimalOpen, "billing/rounds-half-even", value, 1)))
		if len(problems) != 1 || problems[0].Rule != "bad-blocks" {
			t.Fatalf("blocks %q: %v", value, problems)
		}
	}
	_, termProblems := ParseTerm("bad/term", []byte(minimalTerm))
	_, openProblems := ParseOpen("bad/open", []byte(minimalOpen))
	for _, problems := range [][]Problem{termProblems, openProblems} {
		if len(problems) != 1 || problems[0].Rule != "bad-id" || problems[0].Line != 0 {
			t.Fatalf("bad id problems = %v", problems)
		}
	}
}

func TestResolutionWithoutStatus(t *testing.T) {
	for _, value := range []string{"decisions/rounding", ""} {
		data := strings.Replace(minimalOpen, "status: open\n", "", 1) + "resolution: " + value + "\n"
		_, problems := ParseOpen("missing-status", []byte(data))
		if len(problems) != 2 || problems[0].Rule != "missing-status" || problems[0].Line != 0 || problems[1].Rule != "resolution-without-resolved" || problems[1].Line != 5 {
			t.Fatalf("resolution %q: problems = %v, want missing-status at zero and resolution-without-resolved at 5", value, problems)
		}
	}
}

func TestTermOpenIDFromPath(t *testing.T) {
	for _, kind := range []struct {
		name, good, want string
		idFromPath       func(string, string) (string, error)
	}{
		{"term", "money/amount.md", "money:amount", TermIDFromPath},
		{"open", "rounding-rule.md", "rounding-rule", OpenIDFromPath},
	} {
		t.Run(kind.name, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "blueprint", kind.name)
			for _, tt := range []struct{ name, path, want string }{
				{"good", filepath.Join(root, kind.good), kind.want},
				{"bad slug", filepath.Join(root, strings.ToUpper(kind.good)), ""},
				{"double hyphen", filepath.Join(root, strings.Replace(kind.good, ".md", "--bad.md", 1)), ""},
				{"too deep", filepath.Join(root, "nested", kind.good), ""},
				{"too shallow", root + ".md", ""},
				{"term without namespace", filepath.Join(root, "amount.md"), ""},
				{"traversal", root + "/../other/" + kind.good, ""},
				{"not under root", filepath.Join(filepath.Dir(root), "other", kind.good), ""},
				{"sibling prefix", filepath.Join(root+"-other", kind.good), ""},
				{"wrong extension", filepath.Join(root, strings.TrimSuffix(kind.good, ".md")+".txt"), ""},
				{"mixed absolute relative", kind.good, ""},
			} {
				if tt.name == "term without namespace" && kind.name != "term" {
					continue
				}
				t.Run(tt.name, func(t *testing.T) {
					got, err := kind.idFromPath(root, tt.path)
					if tt.want != "" {
						if err != nil || got != tt.want {
							t.Fatalf("id = %q, error = %v, want %q", got, err, tt.want)
						}
					} else if problem, ok := err.(Problem); !ok || problem.Rule != "bad-id" || problem.Line != 0 || got != "" {
						t.Fatalf("id = %q, error = %v, want bad-id", got, err)
					}
				})
			}
			id, err := kind.idFromPath("blueprint/"+kind.name, "blueprint/"+kind.name+"/"+kind.good)
			if err != nil || id != kind.want {
				t.Fatalf("relative id = %q, error = %v", id, err)
			}
		})
	}
}
