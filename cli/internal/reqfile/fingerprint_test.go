package reqfile

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestFingerprintVectors(t *testing.T) {
	type input struct{ ID, Source string }
	var vectors []struct {
		Name, Target, Canonical, Fingerprint string
		Term                                 *input
		Terms, Requirements                  []input
		SameAs                               string `json:"same_as"`
		DifferentFrom                        string `json:"different_from"`
	}
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "formats", "vectors", "fingerprint.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &vectors); err != nil {
		t.Fatal(err)
	}
	if len(vectors) < 6 {
		t.Fatalf("vector count = %d, want at least 6", len(vectors))
	}
	results := map[string]string{}
	for _, v := range vectors {
		t.Run(v.Name, func(t *testing.T) {
			var canonical, got string
			if v.Term != nil {
				term, problems := ParseTerm(v.Term.ID, []byte(v.Term.Source))
				if len(problems) != 0 {
					t.Fatal(problems)
				}
				canonical, got = termCanonical(term), TermFingerprint(term)
			} else {
				terms := map[string]Term{}
				termHashes := map[string]string{}
				for _, in := range v.Terms {
					term, problems := ParseTerm(in.ID, []byte(in.Source))
					if len(problems) != 0 {
						t.Fatal(problems)
					}
					terms[in.ID], termHashes[in.ID] = term, TermFingerprint(term)
				}
				var reqs []Requirement
				var target Requirement
				for _, in := range v.Requirements {
					r, problems := Parse(in.ID, []byte(in.Source))
					if len(problems) != 0 {
						t.Fatal(problems)
					}
					reqs = append(reqs, r)
					if r.ID == v.Target {
						target = r
					}
				}
				hashes, problems := Fingerprints(reqs, terms)
				if len(problems) != 0 || len(hashes) != len(reqs) || target.ID == "" {
					t.Fatalf("hashes = %v, problems = %v, target = %q", hashes, problems, target.ID)
				}
				canonical, got = requirementCanonical(target, termHashes, hashes), hashes[v.Target]
			}
			if canonical != v.Canonical {
				t.Fatalf("canonical = %q, want %q", canonical, v.Canonical)
			}
			if got != v.Fingerprint {
				t.Fatalf("fingerprint = %q, want %q", got, v.Fingerprint)
			}
			results[v.Name] = got
		})
	}
	for _, v := range vectors {
		if v.SameAs != "" && results[v.Name] != results[v.SameAs] {
			t.Errorf("%s differs from %s", v.Name, v.SameAs)
		}
		if v.DifferentFrom != "" && results[v.Name] == results[v.DifferentFrom] {
			t.Errorf("%s equals %s", v.Name, v.DifferentFrom)
		}
	}
}

func TestTermFingerprintMeaning(t *testing.T) {
	base := Term{ID: "money:amount", Definition: "A café  amount.", Status: "planned", States: []string{"draft", "active", "draft"}}
	for _, tt := range []struct {
		name   string
		change func(*Term)
		same   bool
	}{
		{"id", func(v *Term) { v.ID = "other:term" }, true},
		{"status", func(v *Term) { v.Status = "existing" }, true},
		{"body", func(v *Term) { v.Body = "Different rationale.\n" }, true},
		{"double spaces", func(v *Term) { v.Definition = "A café amount." }, false},
		{"unicode normalization", func(v *Term) { v.Definition = "A cafe\u0301  amount." }, false},
		{"state order", func(v *Term) { v.States = []string{"draft", "draft", "active"} }, false},
		{"duplicate states", func(v *Term) { v.States = []string{"draft", "active"} }, false},
		{"definition trailing space", func(v *Term) { v.Definition += " " }, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			changed := base
			tt.change(&changed)
			if same := TermFingerprint(changed) == TermFingerprint(base); same != tt.same {
				t.Fatalf("fingerprint equality = %t, want %t", same, tt.same)
			}
		})
	}
}

func TestRequirementFingerprintMeaning(t *testing.T) {
	r, problems := Parse("feature/alpha", []byte(minimal))
	if len(problems) != 0 {
		t.Fatal(problems)
	}
	base, _ := Fingerprints([]Requirement{r}, nil)
	for _, tt := range []struct {
		name   string
		change func(*Requirement)
		same   bool
	}{
		{"id", func(v *Requirement) { v.ID = "feature/other" }, true},
		{"ears", func(v *Requirement) { v.EARS += " " }, false},
		{"fit kind", func(v *Requirement) { v.FitKind = "property" }, false},
		{"fit", func(v *Requirement) { v.Fit += " " }, false},
		{"confidence", func(v *Requirement) { v.Confidence = "derived" }, true},
		{"evidence", func(v *Requirement) { v.Evidence = "Different evidence." }, true},
		{"tier", func(v *Requirement) { v.Tier = "consequential" }, true},
		{"scope", func(v *Requirement) { v.Scope = []string{"other/**"} }, true},
		{"after", func(v *Requirement) { v.After = []string{v.ID, "unknown/target"} }, true},
		{"superseded by", func(v *Requirement) { v.SupersededBy = "unknown/target" }, true},
		{"body", func(v *Requirement) { v.Body = "Different rationale.\n" }, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			changed := r
			tt.change(&changed)
			got, problems := Fingerprints([]Requirement{changed}, nil)
			if len(problems) != 0 || (got[changed.ID] == base[r.ID]) != tt.same {
				t.Fatalf("hashes = %v, problems = %v, want equality %t", got, problems, tt.same)
			}
		})
	}
	r.FitKind, r.ReferenceDigest, r.ReferencePath = "judged", strings.Repeat("a", 64), "reference/first"
	base, _ = Fingerprints([]Requirement{r}, nil)
	r.ReferencePath = "reference/second"
	got, _ := Fingerprints([]Requirement{r}, nil)
	if got[r.ID] != base[r.ID] {
		t.Fatal("reference path changed fingerprint")
	}
	r.ReferenceDigest = strings.Repeat("b", 64)
	got, _ = Fingerprints([]Requirement{r}, nil)
	if got[r.ID] == base[r.ID] {
		t.Fatal("reference digest did not change fingerprint")
	}
}

func TestFingerprintPropagation(t *testing.T) {
	const a, b, c = "feature/alpha", "feature/beta", "feature/gamma"
	reqs := []Requirement{
		{ID: a, EARS: "WHEN called, THE system SHALL answer.", FitKind: "test", Fit: "It answers.", Needs: []string{b}},
		{ID: b, EARS: "WHEN called, THE system SHALL use [[money:amount]].", FitKind: "test", Fit: "It uses an amount.", Needs: []string{c}, Terms: []string{"money:amount"}},
		{ID: c, EARS: "WHEN called, THE system SHALL load.", FitKind: "test", Fit: "It loads."},
	}
	terms := map[string]Term{"money:amount": {Definition: "An amount.", Status: "planned"}}
	base, problems := Fingerprints(reqs, terms)
	if len(problems) != 0 {
		t.Fatal(problems)
	}
	terms["money:amount"] = Term{Definition: "A different amount.", Status: "planned"}
	got, problems := Fingerprints(reqs, terms)
	if len(problems) != 0 || got[a] == base[a] || got[b] == base[b] || got[c] != base[c] {
		t.Fatalf("term propagation: got %v, base %v, problems %v", got, base, problems)
	}
	base = got
	reqs[2].Fit = "It loads a different value."
	got, problems = Fingerprints(reqs, terms)
	if len(problems) != 0 || got[a] == base[a] || got[b] == base[b] || got[c] == base[c] {
		t.Fatalf("needs propagation: got %v, base %v, problems %v", got, base, problems)
	}
	base = got
	reqs[0].Needs = []string{c}
	got, problems = Fingerprints(reqs, terms)
	if len(problems) != 0 || got[a] == base[a] || got[b] != base[b] || got[c] != base[c] {
		t.Fatalf("needs target change: got %v, base %v, problems %v", got, base, problems)
	}
}

func TestFingerprintProblems(t *testing.T) {
	const a, b, c, d = "feature/alpha", "feature/beta", "feature/gamma", "feature/delta"
	for _, tt := range []struct {
		name string
		reqs []Requirement
		want []Problem
		ids  []string
	}{
		{name: "empty"},
		{name: "unknown term with shared invalid dependency", reqs: []Requirement{{ID: a, Needs: []string{b, c}}, {ID: b, Needs: []string{c}}, {ID: c, Terms: []string{"money:missing", "money:missing"}}, {ID: d}}, want: []Problem{{Rule: "unknown-term", Msg: c + " -> money:missing"}}, ids: []string{d}},
		{name: "unknown needs", reqs: []Requirement{{ID: a, Needs: []string{b}}, {ID: b, Needs: []string{c, c}}, {ID: d}}, want: []Problem{{Rule: "unknown-dependency", Msg: b + " -> " + c}}, ids: []string{d}},
		{name: "independent problems", reqs: []Requirement{{ID: a, Needs: []string{c, b}, Terms: []string{"z:last", "a:first", "a:first"}}}, want: []Problem{
			{Rule: "unknown-term", Msg: a + " -> a:first"}, {Rule: "unknown-term", Msg: a + " -> z:last"},
			{Rule: "unknown-dependency", Msg: a + " -> " + b}, {Rule: "unknown-dependency", Msg: a + " -> " + c},
		}},
		{name: "self needs", reqs: []Requirement{{ID: a, Needs: []string{a, a}}}, want: []Problem{{Rule: "dependency-cycle", Msg: a + " -> " + a}}},
		{name: "needs cycle", reqs: []Requirement{{ID: b, Needs: []string{a}}, {ID: a, Needs: []string{b}}, {ID: c}}, want: []Problem{{Rule: "dependency-cycle", Msg: a + " -> " + b + " -> " + a}}, ids: []string{c}},
		{name: "dependent of cycle", reqs: []Requirement{{ID: a, Needs: []string{b}}, {ID: b, Needs: []string{c}}, {ID: c, Needs: []string{b}}, {ID: d}}, want: []Problem{{Rule: "dependency-cycle", Msg: b + " -> " + c + " -> " + b}}, ids: []string{d}},
		{name: "scheduling cycle", reqs: []Requirement{{ID: a, Needs: []string{b}}, {ID: b, After: []string{a, c}}}, ids: []string{a, b}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			for _, reverse := range []bool{false, true} {
				reqs := slices.Clone(tt.reqs)
				if reverse {
					slices.Reverse(reqs)
				}
				for i := range reqs {
					reqs[i].Needs, reqs[i].Terms = slices.Clone(reqs[i].Needs), slices.Clone(reqs[i].Terms)
					if reverse {
						slices.Reverse(reqs[i].Needs)
						slices.Reverse(reqs[i].Terms)
					}
				}
				before, _ := json.Marshal(reqs)
				got, problems := Fingerprints(reqs, nil)
				if !reflect.DeepEqual(problems, tt.want) || len(got) != len(tt.ids) {
					t.Fatalf("reverse %t: hashes %v, problems %v; want ids %v, problems %v", reverse, got, problems, tt.ids, tt.want)
				}
				for _, id := range tt.ids {
					if got[id] == "" {
						t.Errorf("missing fingerprint for %s", id)
					}
				}
				after, _ := json.Marshal(reqs)
				if string(before) != string(after) {
					t.Fatal("Fingerprints mutated input")
				}
			}
		})
	}
}

func TestVerifyReference(t *testing.T) {
	parent := t.TempDir()
	root, outside := filepath.Join(parent, "repo"), filepath.Join(parent, "repo-other")
	for _, dir := range []string{filepath.Join(root, "refs"), outside} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	data := []byte("The reference answer.\r\n\x00\xff")
	digest := fmt.Sprintf("%x", sha256.Sum256(data))
	for _, path := range []string{filepath.Join(root, "refs", "answer.txt"), filepath.Join(outside, "answer.txt")} {
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "empty.txt"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(parent, "alias")
	for _, link := range []struct{ name, target string }{
		{alias, root},
		{filepath.Join(root, "inside"), "refs/answer.txt"},
		{filepath.Join(root, "chain"), "inside"},
		{filepath.Join(root, "escape"), filepath.Join(outside, "answer.txt")},
		{filepath.Join(root, "escape-dir"), outside},
		{filepath.Join(root, "dangling"), "not-found"},
		{filepath.Join(root, "loop"), "loop"},
	} {
		if err := os.Symlink(link.target, link.name); err != nil {
			t.Fatal(err)
		}
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	relative, err := filepath.Rel(cwd, root)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name, root, path, digest, rule string
	}{
		{name: "matching", path: "refs/answer.txt"},
		{name: "cleaned", path: "refs/../refs/answer.txt"},
		{name: "relative root", root: relative, path: "refs/answer.txt"},
		{name: "symlink root", root: alias, path: "refs/answer.txt"},
		{name: "inside symlink", path: "inside"},
		{name: "symlink chain", path: "chain"},
		{name: "symlink root and file", root: alias, path: "chain"},
		{name: "empty file", path: "empty.txt", digest: fmt.Sprintf("%x", sha256.Sum256(nil))},
		{name: "mismatch", path: "refs/answer.txt", digest: strings.Repeat("0", 64), rule: "reference-digest-mismatch"},
		{name: "raw bytes", path: "refs/answer.txt", digest: fmt.Sprintf("%x", sha256.Sum256([]byte(strings.ReplaceAll(string(data), "\r\n", "\n")))), rule: "reference-digest-mismatch"},
		{name: "missing", path: "refs/absent.txt", rule: "reference-missing"},
		{name: "empty path", rule: "reference-missing"},
		{name: "directory", path: "refs", rule: "reference-missing"},
		{name: "missing root", root: filepath.Join(parent, "absent"), path: "answer.txt", rule: "reference-missing"},
		{name: "root is file", root: filepath.Join(root, "refs", "answer.txt"), path: "answer.txt", rule: "reference-missing"},
		{name: "absolute inside", path: filepath.Join(root, "refs", "answer.txt"), rule: "reference-outside-root"},
		{name: "absolute outside", path: filepath.Join(outside, "answer.txt"), rule: "reference-outside-root"},
		{name: "lexical escape", path: "../repo-other/answer.txt", rule: "reference-outside-root"},
		{name: "missing lexical escape", path: "../absent/answer.txt", rule: "reference-outside-root"},
		{name: "symlink escape", path: "escape", rule: "reference-outside-root"},
		{name: "directory symlink escape", path: "escape-dir/answer.txt", rule: "reference-outside-root"},
		{name: "symlink root escape", root: alias, path: "escape", rule: "reference-outside-root"},
		{name: "dangling symlink", path: "dangling", rule: "reference-missing"},
		{name: "symlink loop", path: "loop", rule: "reference-missing"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if tt.root == "" {
				tt.root = root
			}
			if tt.digest == "" {
				tt.digest = digest
			}
			r := Requirement{ID: "feature/judged", FitKind: "judged", ReferencePath: tt.path, ReferenceDigest: tt.digest}
			problems := VerifyReference(tt.root, r)
			if tt.rule == "" {
				if len(problems) != 0 {
					t.Fatal(problems)
				}
			} else if len(problems) != 1 || problems[0].Rule != tt.rule || problems[0].Line != 0 || !strings.Contains(problems[0].Msg, r.ID) {
				t.Fatalf("problems = %v, want %s at zero naming %s", problems, tt.rule, r.ID)
			}
		})
	}
	for _, kind := range []string{"test", "property", ""} {
		if problems := VerifyReference(filepath.Join(parent, "missing"), Requirement{FitKind: kind, ReferencePath: "../escape"}); len(problems) != 0 {
			t.Fatalf("non-judged kind %q: %v", kind, problems)
		}
	}
}
