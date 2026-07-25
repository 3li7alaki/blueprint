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

var Names = []string{"coverage", "orphan", "budget", "blocked", "derived", "shape", "dash", "drift"}

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
		case "drift":
			for _, spec := range specs {
				info, err := os.Stat(spec.Path)
				if err != nil {
					continue
				}
				for _, req := range spec.Requirements {
					var newest int64
					for _, tag := range tags {
						if tag.Qualified == req.Qualified() && tag.ModTime > newest {
							newest = tag.ModTime
						}
					}
					if newest > 0 && info.ModTime().UnixNano() > newest {
						offenders = append(offenders, req.Qualified())
					}
				}
			}
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
