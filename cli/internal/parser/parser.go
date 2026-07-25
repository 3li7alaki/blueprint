package parser

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"blueprint/internal/model"
)

var specSections = []string{"Intent", "Surfaces", "Requirements", "Edges", "Out of scope", "Depends on"}

func Specs(root string) ([]model.Spec, []error) {
	paths, _ := filepath.Glob(filepath.Join(root, "blueprint", "spec", "*.md"))
	var specs []model.Spec
	var errs []error
	for _, path := range paths {
		s, err := Spec(path)
		if err != nil {
			errs = append(errs, err)
		} else {
			specs = append(specs, s)
		}
	}
	return specs, errs
}

func Spec(path string) (model.Spec, error) {
	lines, err := readLines(path)
	if err != nil {
		return model.Spec{}, err
	}
	bad := func(line int, msg string) (model.Spec, error) {
		return model.Spec{}, model.ParseError{Path: path, Line: line, Msg: msg}
	}
	if len(lines) < 4 || !strings.HasPrefix(lines[0], "# ") {
		return bad(1, "expected feature heading")
	}
	feature := strings.TrimSpace(strings.TrimPrefix(lines[0], "# "))
	if feature != strings.TrimSuffix(filepath.Base(path), ".md") {
		return bad(1, "feature slug does not match filename")
	}
	statusLine := nextContent(lines, 1)
	if statusLine < 0 || !strings.HasPrefix(lines[statusLine], "status: ") {
		return bad(statusLine+1, "expected status")
	}
	status := trimTicks(strings.TrimSpace(strings.TrimPrefix(lines[statusLine], "status: ")))
	if !oneOf(status, "drafting", "ready", "building", "shipped") {
		return bad(statusLine+1, "invalid status")
	}
	depthLine := nextContent(lines, statusLine+1)
	if depthLine < 0 || !strings.HasPrefix(lines[depthLine], "depth: ") {
		return bad(depthLine+1, "expected depth")
	}
	depth := trimTicks(strings.TrimSpace(strings.TrimPrefix(lines[depthLine], "depth: ")))
	if !oneOf(depth, "quick", "standard", "paranoid") {
		return bad(depthLine+1, "invalid depth")
	}
	s := model.Spec{Path: path, Feature: feature, Status: status, Depth: depth, Sections: map[string][]string{}}
	current := ""
	wantSection := 0
	for i := depthLine + 1; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "## ") {
			name := strings.TrimSpace(strings.TrimPrefix(line, "## "))
			if wantSection >= len(specSections) || name != specSections[wantSection] {
				return bad(i+1, "unexpected section "+name)
			}
			current = name
			wantSection++
			continue
		}
		if current == "" {
			if ignorable(line) {
				continue
			}
			return bad(i+1, "content before first section")
		}
		s.Sections[strings.ToLower(current)] = append(s.Sections[strings.ToLower(current)], line)
	}
	if wantSection != len(specSections) {
		return bad(len(lines), "missing required section "+specSections[wantSection])
	}
	reqLines := s.Sections["requirements"]
	for i := 0; i < len(reqLines); i++ {
		if !strings.HasPrefix(reqLines[i], "### ") {
			if ignorable(reqLines[i]) {
				continue
			}
			return bad(sectionLine(lines, "Requirements")+i+1, "expected requirement heading")
		}
		slug := trimTicks(strings.TrimSpace(strings.TrimPrefix(reqLines[i], "### ")))
		start := i
		conf := nextContent(reqLines, i+1)
		ears := nextContent(reqLines, conf+1)
		fit := nextContent(reqLines, ears+1)
		if conf < 0 || !oneOf(trimTicks(strings.TrimSpace(reqLines[conf])), "stated", "derived") {
			return bad(sectionLine(lines, "Requirements")+start+2, "missing requirement confidence")
		}
		if ears < 0 || strings.HasPrefix(reqLines[ears], "### ") || !strings.HasSuffix(reqLines[ears], ".") || !strings.Contains(reqLines[ears], " SHALL ") || (!strings.Contains(reqLines[ears], "WHEN ") && !strings.Contains(reqLines[ears], "WHILE ")) {
			return bad(sectionLine(lines, "Requirements")+start+3, "missing EARS line ending in a full stop")
		}
		if fit < 0 || !strings.HasPrefix(reqLines[fit], "fit: ") {
			return bad(sectionLine(lines, "Requirements")+start+4, "missing fit line")
		}
		r := model.Requirement{Feature: feature, Slug: slug, Confidence: trimTicks(strings.TrimSpace(reqLines[conf])), EARS: reqLines[ears], Fit: strings.TrimPrefix(reqLines[fit], "fit: "), Line: sectionLine(lines, "Requirements") + start + 1}
		i = fit
		next := nextContent(reqLines, i+1)
		if next >= 0 && strings.HasPrefix(reqLines[next], "superseded-by: ") {
			r.SupersededBy = strings.TrimSpace(strings.TrimPrefix(reqLines[next], "superseded-by: "))
			i = next
		}
		s.Requirements = append(s.Requirements, r)
	}
	surfaceLines := s.Sections["surfaces"]
	for i := 0; i < len(surfaceLines); i++ {
		if ignorable(surfaceLines[i]) {
			continue
		}
		if !strings.HasPrefix(surfaceLines[i], "### ") {
			return bad(sectionLine(lines, "Surfaces")+i+1, "expected surface heading")
		}
		keys := []string{"- who:", "- data:", "- empty:", "- loading:", "- error:", "- denied:"}
		at := i
		for k, key := range keys {
			at = nextContent(surfaceLines, at+1)
			if at < 0 || !strings.HasPrefix(surfaceLines[at], key) || strings.TrimSpace(strings.TrimPrefix(surfaceLines[at], key)) == "" {
				return bad(sectionLine(lines, "Surfaces")+i+k+2, "expected surface "+strings.Trim(key, "-: "))
			}
		}
		i = at
	}
	return s, nil
}

func Decisions(root string) []error {
	paths, _ := filepath.Glob(filepath.Join(root, "blueprint", "decisions", "*.md"))
	var errs []error
	for _, path := range paths {
		lines, err := readLines(path)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		fail := func(line int, msg string) {
			errs = append(errs, model.ParseError{Path: path, Line: line, Msg: msg})
		}
		if len(lines) < 4 || lines[0] != "# "+strings.TrimSuffix(filepath.Base(path), ".md") {
			fail(1, "decision slug does not match filename")
			continue
		}
		expected := []string{"status: ", "superseded-by:", "date: "}
		at := 1
		valid := true
		for _, prefix := range expected {
			at = nextContent(lines, at)
			if at < 0 || !strings.HasPrefix(lines[at], prefix) {
				fail(at+1, "invalid decision header")
				valid = false
				break
			}
			at++
		}
		if !valid {
			continue
		}
		status := strings.TrimSpace(strings.TrimPrefix(lines[nextContent(lines, 1)], "status: "))
		if !oneOf(status, "accepted", "superseded") {
			fail(2, "invalid decision status")
			continue
		}
		sections := []string{"## Context", "## Decision", "## Because", "## Consequences"}
		cursor := at
		for _, section := range sections {
			found := false
			for cursor < len(lines) {
				if lines[cursor] == section {
					found = true
					cursor++
					break
				}
				cursor++
			}
			if !found {
				fail(len(lines), "missing decision section "+strings.TrimPrefix(section, "## "))
				break
			}
		}
	}
	return errs
}

func Opens(path string) ([]model.Open, error) {
	lines, err := readLines(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []model.Open
	for i := 0; i < len(lines); {
		if !strings.HasPrefix(lines[i], "## ") || strings.HasPrefix(lines[i], "### ") {
			i++
			continue
		}
		o := model.Open{Slug: strings.TrimSpace(strings.TrimPrefix(lines[i], "## ")), Line: i + 1}
		keys := []*string{&o.Status, &o.Pass, &o.Asked, &o.Owner, &o.Question, &o.Cost, &o.Blocks}
		names := []string{"status", "pass", "asked", "owner", "question", "cost", "blocks"}
		for k := range names {
			i++
			if i >= len(lines) || !strings.HasPrefix(lines[i], names[k]+":") {
				return nil, model.ParseError{Path: path, Line: i + 1, Msg: "expected " + names[k] + " key"}
			}
			*keys[k] = strings.TrimSpace(strings.TrimPrefix(lines[i], names[k]+":"))
		}
		if !oneOf(o.Status, "OPEN", "DEFERRED") {
			return nil, model.ParseError{Path: path, Line: o.Line + 1, Msg: "invalid open status"}
		}
		out = append(out, o)
		i++
	}
	return out, nil
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	inComment := false
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		for {
			if inComment {
				end := strings.Index(line, "-->")
				if end < 0 {
					line = ""
					break
				}
				line = line[end+3:]
				inComment = false
				continue
			}
			start := strings.Index(line, "<!--")
			if start < 0 {
				break
			}
			end := strings.Index(line[start+4:], "-->")
			if end < 0 {
				line = line[:start]
				inComment = true
				break
			}
			line = line[:start] + line[start+4+end+3:]
		}
		lines = append(lines, line)
	}
	return lines, s.Err()
}

func nextContent(lines []string, start int) int {
	if start < 0 {
		return -1
	}
	for i := start; i < len(lines); i++ {
		if !ignorable(lines[i]) {
			return i
		}
	}
	return -1
}

func ignorable(s string) bool {
	t := strings.TrimSpace(s)
	return t == "" || strings.HasPrefix(t, "<!--") || strings.HasPrefix(t, "-->") || strings.HasPrefix(t, "Written by ") || strings.HasPrefix(t, "One key ") || strings.HasPrefix(t, "Every unresolved ") || strings.HasPrefix(t, "An entry ") || strings.HasPrefix(t, "Blocked work ") || strings.HasPrefix(t, "Resolving an ") || strings.HasPrefix(t, "then run ")
}
func trimTicks(s string) string { return strings.Trim(s, "`") }
func oneOf(s string, values ...string) bool {
	for _, v := range values {
		if s == v {
			return true
		}
	}
	return false
}
func sectionLine(lines []string, name string) int {
	for i, line := range lines {
		if line == "## "+name {
			return i + 1
		}
	}
	return 0
}
