package reqfile

import (
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

type Term struct {
	ID, Definition, Status, Body string
	States                       []string
}

type Open struct {
	ID, Question, Cost, Status, Asked, Owner, Resolution, Body string
	Blocks                                                     []string
}

var (
	slugPattern     = regexp.MustCompile(`^` + slug + `$`)
	termIDPattern   = regexp.MustCompile(`^` + slug + `:` + slug + `$`)
	termPathPattern = regexp.MustCompile(`^` + slug + `/` + slug + `$`)
	blockPattern    = regexp.MustCompile(`^` + slug + `(/` + slug + `)*/\*$`)
)

// TermIDFromPath checks lexical containment without resolving symbolic links.
func TermIDFromPath(termsRoot, path string) (string, error) {
	id, err := objectIDFromPath(termsRoot, path, termPathPattern)
	return strings.ReplaceAll(id, "/", ":"), err
}

// OpenIDFromPath checks lexical containment without resolving symbolic links.
func OpenIDFromPath(openRoot, path string) (string, error) {
	return objectIDFromPath(openRoot, path, slugPattern)
}

func objectIDFromPath(root, path string, pattern *regexp.Regexp) (string, error) {
	rel, err := filepath.Rel(root, path)
	id := strings.TrimSuffix(filepath.ToSlash(rel), ".md")
	if err != nil || !strings.HasSuffix(rel, ".md") || !pattern.MatchString(id) {
		return "", Problem{Rule: "bad-id", Msg: "expected an object file under its root with the required slug depth"}
	}
	return id, nil
}

func ParseTerm(id string, data []byte) (Term, []Problem) {
	r := Term{ID: id}
	var problems []Problem
	if !termIDPattern.MatchString(id) {
		problems = append(problems, Problem{Rule: "bad-id", Msg: "expected namespace:term with slug parts"})
	}
	var states string
	fields := map[string]*string{"definition": &r.Definition, "status": &r.Status, "states": &states}
	lines := parseHeader(data, fields, &r.Body, &problems)
	if lines == nil {
		return r, problems
	}
	for _, key := range []string{"definition", "status"} {
		if lines[key] == 0 || (key == "definition" && strings.TrimSpace(r.Definition) == "") {
			problems = append(problems, Problem{"missing-" + key, lines[key], "required nonempty key " + key + " is missing"})
		}
	}
	if lines["status"] != 0 && !slices.Contains([]string{"existing", "planned"}, r.Status) {
		problems = append(problems, Problem{"bad-term-status", lines["status"], "expected existing or planned"})
	}
	if lines["states"] != 0 {
		var valid bool
		r.States, valid = parseList(states, slugPattern.MatchString)
		if !valid {
			problems = append(problems, Problem{"bad-states", lines["states"], "expected comma-separated state slugs"})
		}
	}
	return r, problems
}

func ParseOpen(id string, data []byte) (Open, []Problem) {
	r := Open{ID: id}
	var problems []Problem
	if !slugPattern.MatchString(id) {
		problems = append(problems, Problem{Rule: "bad-id", Msg: "expected one slug"})
	}
	var blocks string
	fields := map[string]*string{
		"question": &r.Question, "cost": &r.Cost, "blocks": &blocks, "status": &r.Status,
		"asked": &r.Asked, "owner": &r.Owner, "resolution": &r.Resolution,
	}
	lines := parseHeader(data, fields, &r.Body, &problems)
	if lines == nil {
		return r, problems
	}
	for _, key := range []string{"question", "cost", "blocks", "status", "asked"} {
		if lines[key] == 0 || ((key == "question" || key == "cost") && strings.TrimSpace(*fields[key]) == "") {
			problems = append(problems, Problem{"missing-" + key, lines[key], "required nonempty key " + key + " is missing"})
		}
	}
	check := func(key, rule string, ok bool) {
		if lines[key] != 0 && !ok {
			problems = append(problems, Problem{rule, lines[key], "invalid " + key + " value"})
		}
	}
	check("question", "bad-question", strings.TrimSpace(r.Question) == "" || strings.HasSuffix(r.Question, "?"))
	if lines["blocks"] != 0 {
		var valid bool
		r.Blocks, valid = parseList(blocks, func(s string) bool { return idPattern.MatchString(s) || blockPattern.MatchString(s) })
		check("blocks", "bad-blocks", valid)
	}
	check("status", "bad-open-status", slices.Contains([]string{"open", "deferred", "resolved"}, r.Status))
	_, err := time.Parse("2006-01-02", r.Asked)
	check("asked", "bad-asked", err == nil)
	check("owner", "empty-owner", strings.TrimSpace(r.Owner) != "")
	if r.Status == "resolved" {
		if strings.TrimSpace(r.Resolution) == "" {
			problems = append(problems, Problem{"resolved-without-resolution", lines["status"], "resolved status requires a nonempty resolution"})
		}
	} else {
		check("resolution", "resolution-without-resolved", false)
	}
	return r, problems
}

func parseList(value string, valid func(string) bool) ([]string, bool) {
	items := strings.Split(value, ",")
	ok := true
	for i, item := range items {
		items[i] = strings.TrimSpace(item)
		ok = valid(items[i]) && ok
	}
	return items, ok
}
