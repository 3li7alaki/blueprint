package gates

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"blueprint/internal/model"
	"blueprint/internal/parser"
	"blueprint/internal/repo"
	"blueprint/internal/scan"
)

var Names = []string{"coverage", "orphan", "budget", "blocked", "derived", "shape", "dash", "unmapped", "bug", "drift", "look", "tokens"}

type Result struct {
	Gate      string   `json:"gate"`
	Status    string   `json:"status"`
	Offenders []string `json:"offenders"`
}

func Run(root, only string) []Result {
	specs, shapeErrs := parser.Specs(root)
	opens, openErr := parser.Opens(filepath.Join(root, "blueprint", "OPEN.md"))
	if openErr != nil {
		shapeErrs = append(shapeErrs, openErr)
	}
	shapeErrs = append(shapeErrs, parser.Decisions(root)...)
	tags, _ := scan.Tags(root)
	reqs := map[string]model.Requirement{}
	specByFeature := map[string]model.Spec{}
	for _, spec := range specs {
		specByFeature[spec.Feature] = spec
		for _, req := range spec.Requirements {
			reqs[req.Qualified()] = req
		}
	}
	results := make([]Result, 0, len(Names))
	for _, name := range Names {
		if only != "" && name != only {
			continue
		}
		var offenders []string
		switch name {
		case "coverage":
			for q := range reqs {
				found := false
				for _, tag := range tags {
					found = found || tag.Qualified == q && tag.Test
				}
				if !found {
					offenders = append(offenders, q)
				}
			}
		case "orphan":
			for _, tag := range tags {
				if _, ok := reqs[tag.Qualified]; !ok {
					offenders = append(offenders, tag.Path+": "+tag.Qualified)
				}
			}
		case "budget":
			budgets := map[string]int{"blueprint/PROJECT.md": 60, "blueprint/CONVENTIONS.md": 80, "blueprint/REVIEW.md": 60}
			if path, ok := parser.ProductPath(root); ok {
				budgets[rel(root, path)] = 60
			}
			files, _ := filepath.Glob(filepath.Join(root, "blueprint", "spec", "*.md"))
			for _, file := range files {
				rel, _ := filepath.Rel(root, file)
				budgets[filepath.ToSlash(rel)] = 150
			}
			for rel, limit := range budgets {
				if data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel))); err == nil {
					count := len(strings.Split(strings.TrimSuffix(string(data), "\n"), "\n"))
					if count > limit {
						offenders = append(offenders, fmt.Sprintf("%s: %d lines, limit %d", rel, count, limit))
					}
				}
			}
		case "blocked":
			for _, tag := range tags {
				for _, o := range opens {
					if blocked(o.Blocks, tag.Qualified) {
						offenders = append(offenders, tag.Path+": "+tag.Qualified+" blocked by "+o.Slug)
					}
				}
			}
		case "derived":
			for _, tag := range tags {
				if req, ok := reqs[tag.Qualified]; ok && req.Confidence == "derived" && !tag.Test {
					offenders = append(offenders, tag.Path+": "+tag.Qualified)
				}
			}
		case "shape":
			for _, err := range shapeErrs {
				offenders = append(offenders, err.Error())
			}
		case "dash":
			offenders = dash(root)
		case "unmapped":
			offenders = unmapped(root, tags)
		case "bug":
			for _, req := range reqs {
				if req.Bug != "" {
					offenders = append(offenders, req.Qualified())
				}
			}
		case "drift":
			offenders = drift(root, specs, tags)
		case "look":
			offenders = look(root, specs)
		case "tokens":
			offenders = tokens(root)
		}
		sort.Strings(offenders)
		status := "pass"
		if len(offenders) > 0 {
			status = "fail"
		}
		results = append(results, Result{Gate: name, Status: status, Offenders: offenders})
	}
	return results
}

// drift names every requirement whose wording was amended after the last change to the code
// carrying its tag. Three things it deliberately stays quiet about:
//
// @spec traceability/drift-follows-the-amendment
//
// A requirement that was only ever introduced is not drift. Writing a rule for code that already
// satisfies it is the whole brownfield flow, and a gate that reddened on it would make harvest
// unusable. A sibling requirement added to the same file is not drift either, which is why this
// compares requirement text and never file timestamps. A requirement with no tag at all is
// uncovered, and the coverage gate owns that.
//
// @spec traceability/drift-survives-checkout
//
// Times come from commits, so a clone, a checkout and a task worktree all answer the same.
func drift(root string, specs []model.Spec, tags []scan.Tag) []string {
	var paths []string
	for _, tag := range tags {
		paths = append(paths, tag.Path)
	}
	times := repo.Times(root, paths)
	var offenders []string
	for _, spec := range specs {
		rel, err := filepath.Rel(root, spec.Path)
		if err != nil {
			continue
		}
		amended := amendments(root, filepath.ToSlash(rel))
		for _, req := range spec.Requirements {
			at := amended[req.Slug]
			if at == 0 {
				continue
			}
			var newest int64
			for _, tag := range tags {
				if tag.Qualified == req.Qualified() && times[tag.Path] > newest {
					newest = times[tag.Path]
				}
			}
			if newest > 0 && at > newest {
				offenders = append(offenders, req.Qualified())
			}
		}
	}
	return offenders
}

// amendments dates the last rewording of each requirement in a spec. A requirement that has never
// changed since the revision that introduced it maps to zero, meaning no amendment, not no answer.
func amendments(root, rel string) map[string]int64 {
	amended := map[string]int64{}
	previous := map[string]string{}
	for i, rev := range repo.History(root, rel) {
		blocks := parser.RequirementBlocks(rev.Content)
		for slug, text := range blocks {
			if was, existed := previous[slug]; i > 0 && existed && was != text {
				amended[slug] = rev.Time
			} else if _, seen := amended[slug]; !seen {
				amended[slug] = 0
			}
		}
		previous = blocks
	}
	return amended
}

func unmapped(root string, tags []scan.Tag) []string {
	data, err := os.ReadFile(filepath.Join(root, "blueprint", "PROJECT.md"))
	if err != nil {
		return nil
	}
	scopes := harvested(string(data))
	if len(scopes) == 0 {
		return nil
	}
	tagged := map[string]bool{}
	for _, tag := range tags {
		if !tag.Test {
			tagged[tag.Path] = true
		}
	}
	files, _ := repo.Files(root)
	var offenders []string
	for _, file := range files {
		if scan.IsTest(file) || strings.HasPrefix(file, "blueprint/") {
			continue
		}
		for _, scope := range scopes {
			if scan.MatchGlob(scope, file) && !tagged[file] {
				offenders = append(offenders, file)
				break
			}
		}
	}
	return offenders
}

func harvested(data string) []string {
	var out []string
	in := false
	for _, line := range strings.Split(data, "\n") {
		if line == "## Harvested" {
			in = true
			continue
		}
		if in && strings.HasPrefix(line, "## ") {
			break
		}
		if in && strings.HasPrefix(strings.TrimSpace(line), "- ") {
			out = append(out, strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "- ")))
		}
	}
	return out
}

func IsBlocked(opens []model.Open, qualified string) bool {
	for _, o := range opens {
		if blocked(o.Blocks, qualified) {
			return true
		}
	}
	return false
}

func blocked(patterns, qualified string) bool {
	for _, pattern := range strings.Split(patterns, ",") {
		pattern = strings.TrimSpace(pattern)
		if pattern == "*" {
			return true
		}
		if ok, _ := path.Match(pattern, qualified); ok {
			return true
		}
	}
	return false
}

func dash(root string) []string {
	files, _ := repo.Files(root)
	var offenders []string
	for _, rel := range files {
		if strings.HasPrefix(rel, "cli/") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			continue
		}
		for i, line := range strings.Split(string(data), "\n") {
			if strings.ContainsAny(line, "\u2013\u2014") {
				offenders = append(offenders, fmt.Sprintf("%s:%d", rel, i+1))
			}
		}
	}
	return offenders
}
