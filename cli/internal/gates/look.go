package gates

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"blueprint/internal/model"
	"blueprint/internal/parser"
	"blueprint/internal/repo"
	"blueprint/internal/scan"
)

// required names the sections nobody can build a surface without. The rest of PRODUCT.md is worth
// having and none of it is worth a red gate, because a gate that fires on a missing adjective
// teaches people to ignore gates.
var required = []string{"register", "platform", "users", "positioning", "accessibility & inclusion"}

// look fires only once a spec declares a surface. A CLI, a library and a worker have no register
// and no palette, and asking them for one would make the gate a tax rather than a check.
//
// @spec look/surfaces-need-a-product
func look(root string, specs []model.Spec) []string {
	surfaces := false
	for _, spec := range specs {
		for _, line := range spec.Sections["surfaces"] {
			surfaces = surfaces || strings.HasPrefix(line, "### ")
		}
	}
	if !surfaces {
		return nil
	}
	product, ok := parser.ReadProduct(root)
	if !ok {
		return []string{"a spec declares a surface and there is no PRODUCT.md: run the look pass"}
	}
	var offenders []string
	for _, section := range required {
		if !product.Answered(section) {
			offenders = append(offenders, rel(root, product.Path)+": "+section+" is unanswered")
		}
	}
	return offenders
}

var hexRE = regexp.MustCompile(`#[0-9a-fA-F]{3}([0-9a-fA-F]{3}([0-9a-fA-F]{2})?)?\b`)
var fontRE = regexp.MustCompile(`(?i)font-family\s*[:=]`)

// tokens keeps colour and type in one file instead of scattered through components, which is what
// makes a rebrand an edit rather than an excavation. It reads its scope from CONVENTIONS.md:
//
//	## Tokens
//	home: src/styles/tokens.css
//	components: src/**/*.tsx, src/**/*.css
//
// @spec look/tokens-stay-home
//
// No section, no gate. Opt-in matches `unmapped`, and for the same reason: a check nobody scoped
// is a check that fires on a tree nobody was working in.
//
// Only raw hex and font stacks are matched. Spacing is deliberately excluded: a 1px border and a
// 2px offset are legitimate everywhere, so a px rule would cry wolf until it was switched off. A
// line carrying `blueprint:allow-raw` is skipped, because an SVG fill or a third party embed is a
// real exception and a gate with no exit is a gate people delete.
func tokens(root string) []string {
	home, globs := tokenScope(root)
	if len(globs) == 0 {
		return nil
	}
	files, _ := repo.Files(root)
	var offenders []string
	for _, file := range files {
		if file == home || scan.IsTest(file) || strings.HasPrefix(file, "blueprint/") {
			continue
		}
		matched := false
		for _, glob := range globs {
			matched = matched || scan.MatchGlob(glob, file)
		}
		if !matched {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file)))
		if err != nil {
			continue
		}
		for i, line := range strings.Split(string(data), "\n") {
			if strings.Contains(line, "blueprint:allow-raw") {
				continue
			}
			switch {
			case hexRE.MatchString(line):
				offenders = append(offenders, positioned(file, i, "raw colour, belongs in "+home))
			case fontRE.MatchString(line):
				offenders = append(offenders, positioned(file, i, "raw font stack, belongs in "+home))
			}
		}
	}
	return offenders
}

func rel(root, path string) string {
	value, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(value)
}

func positioned(file string, index int, message string) string {
	return file + ":" + strconv.Itoa(index+1) + ": " + message
}

func tokenScope(root string) (string, []string) {
	data, err := os.ReadFile(filepath.Join(root, "blueprint", "CONVENTIONS.md"))
	if err != nil {
		return "", nil
	}
	home := ""
	var globs []string
	in := false
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "## ") {
			in = strings.TrimSpace(line) == "## Tokens"
			continue
		}
		if !in {
			continue
		}
		key, value, ok := strings.Cut(strings.TrimSpace(line), ":")
		if !ok {
			continue
		}
		switch strings.TrimSpace(key) {
		case "home":
			home = strings.TrimSpace(value)
		case "components":
			for _, glob := range strings.Split(value, ",") {
				if glob = strings.TrimSpace(glob); glob != "" {
					globs = append(globs, glob)
				}
			}
		}
	}
	if home == "" {
		return "", nil
	}
	return home, globs
}
