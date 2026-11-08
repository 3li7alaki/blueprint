// Package reqfile parses and validates version 0 requirement, term and open question files.
package reqfile

import (
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

type Problem struct {
	Rule string
	Line int
	Msg  string
}

func (p Problem) Error() string { return fmt.Sprintf("%s:%d: %s", p.Rule, p.Line, p.Msg) }

type Requirement struct {
	ID, EARS, FitKind, Fit, ReferencePath, ReferenceDigest string
	Confidence, Evidence, Tier, SupersededBy, Body         string
	Scope, Needs, After, Terms                             []string
}

const slug = `[a-z0-9]+(-[a-z0-9]+)*`

var (
	idPattern        = regexp.MustCompile(`^` + slug + `(/` + slug + `)+$`)
	keyPattern       = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	triggerPattern   = regexp.MustCompile(`(WHEN|WHILE|IF|WHERE) `)
	referencePattern = regexp.MustCompile(`^(.+) sha256:([0-9a-f]{64})$`)
	termPattern      = regexp.MustCompile(`^\[\[(` + slug + `:` + slug + `)\]\]`)
)

// IDFromPath checks lexical containment; it does not access the filesystem.
func IDFromPath(specRoot, path string) (string, error) {
	rel, err := filepath.Rel(specRoot, path)
	id := strings.TrimSuffix(filepath.ToSlash(rel), ".md")
	if err != nil || !strings.HasSuffix(rel, ".md") || !idPattern.MatchString(id) {
		return "", Problem{Rule: "bad-id", Msg: "expected a requirement file under the spec root with at least two slug segments"}
	}
	return id, nil
}

func Parse(id string, data []byte) (Requirement, []Problem) {
	r := Requirement{ID: id}
	var problems []Problem
	if !idPattern.MatchString(id) {
		problems = append(problems, Problem{Rule: "bad-id", Msg: "expected at least two slug segments"})
	}
	var fit, reference, scope, needs, after string
	fields := map[string]*string{
		"ears": &r.EARS, "fit": &fit, "reference": &reference,
		"confidence": &r.Confidence, "evidence": &r.Evidence, "scope": &scope,
		"needs": &needs, "after": &after, "tier": &r.Tier, "superseded-by": &r.SupersededBy,
	}
	lines := parseHeader(data, fields, &r.Body, &problems)
	if lines == nil {
		return r, problems
	}
	bad := func(rule, key, msg string) {
		problems = append(problems, Problem{Rule: rule, Line: lines[key], Msg: msg})
	}
	for _, key := range []string{"ears", "fit", "confidence", "scope"} {
		if lines[key] == 0 {
			bad("missing-"+key, key, "required key "+key+" is missing")
		}
	}
	check := func(key, rule string, ok bool) {
		if lines[key] != 0 && !ok {
			bad(rule, key, "invalid "+key+" value")
		}
	}
	check("ears", "bad-ears-shape", strings.HasSuffix(r.EARS, ".") && strings.Contains(r.EARS, " SHALL ") && triggerPattern.MatchString(r.EARS))
	r.FitKind, r.Fit, _ = strings.Cut(fit, ": ")
	check("fit", "bad-fit-kind", strings.Contains(fit, ": ") && slices.Contains([]string{"test", "property", "judged"}, r.FitKind))
	if strings.Contains(fit, ": ") {
		check("fit", "empty-fit-criterion", strings.TrimSpace(r.Fit) != "")
	}
	if r.FitKind == "judged" && lines["reference"] == 0 {
		bad("judged-without-reference", "fit", "judged fit requires reference")
	}
	check("reference", "reference-without-judged", r.FitKind == "judged")
	if match := referencePattern.FindStringSubmatch(reference); match != nil && strings.TrimSpace(match[1]) != "" {
		r.ReferencePath, r.ReferenceDigest = match[1], match[2]
	} else {
		check("reference", "bad-reference", false)
	}
	check("confidence", "bad-confidence", slices.Contains([]string{"stated", "derived"}, r.Confidence))
	check("tier", "bad-tier", slices.Contains([]string{"chore", "normal", "consequential"}, r.Tier))
	for _, list := range []struct {
		key, value, rule string
		target           *[]string
	}{{"scope", scope, "bad-scope", &r.Scope}, {"needs", needs, "bad-dependency-id", &r.Needs}, {"after", after, "bad-dependency-id", &r.After}} {
		if lines[list.key] == 0 {
			continue
		}
		valid := true
		for _, item := range strings.Split(list.value, ",") {
			item = strings.TrimSpace(item)
			*list.target = append(*list.target, item)
			valid = valid && item != "" && (list.key == "scope" || idPattern.MatchString(item))
		}
		check(list.key, list.rule, valid)
	}
	check("superseded-by", "bad-superseded-by", idPattern.MatchString(r.SupersededBy))
	for _, field := range []struct{ key, value string }{{"ears", r.EARS}, {"fit", r.Fit}} {
		for rest := field.value; ; {
			at := strings.Index(rest, "[[")
			if at < 0 {
				break
			}
			rest = rest[at:]
			match := termPattern.FindStringSubmatch(rest)
			if match == nil {
				bad("bad-term-ref", field.key, "expected [[namespace:term]] with slug parts")
				break
			}
			r.Terms = append(r.Terms, match[1])
			rest = rest[len(match[0]):]
		}
	}
	slices.Sort(r.Terms)
	r.Terms = slices.Compact(r.Terms)
	return r, problems
}

func CheckGraph(reqs []Requirement) []Problem {
	edges := map[string][]string{}
	for _, r := range reqs {
		edges[r.ID] = append(edges[r.ID], append(slices.Clone(r.Needs), r.After...)...)
	}
	ids := make([]string, 0, len(edges))
	for id := range edges {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	var problems []Problem
	for _, id := range ids {
		slices.Sort(edges[id])
		valid := []string{}
		for _, dep := range slices.Compact(edges[id]) {
			rule := ""
			if _, exists := edges[dep]; !exists {
				rule = "unknown-dependency"
			} else if dep == id {
				rule = "self-dependency"
			}
			if rule != "" {
				problems = append(problems, Problem{Rule: rule, Msg: id + " -> " + dep})
			} else {
				valid = append(valid, dep)
			}
		}
		edges[id] = valid
	}
	for _, start := range ids {
		visiting := map[string]bool{}
		var walk func(string, []string)
		walk = func(id string, path []string) {
			visiting[id] = true
			path = append(path, id)
			for _, dep := range edges[id] {
				if dep == start {
					problems = append(problems, Problem{Rule: "dependency-cycle", Msg: strings.Join(append(path, start), " -> ")})
				} else if dep > start && !visiting[dep] {
					// Starting at the smallest id avoids reporting cycle rotations.
					walk(dep, path)
				}
			}
			delete(visiting, id)
		}
		walk(start, nil)
	}
	return problems
}
