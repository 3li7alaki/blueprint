package scan

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"blueprint/internal/repo"
)

var tagRE = regexp.MustCompile(`@spec ([a-z0-9][a-z0-9-]*/[a-z0-9][a-z0-9-]*)`)
var testRE = regexp.MustCompile(`(^|/)(tests?|__tests__|spec)/|(\.test\.[a-z]+|\.spec\.[a-z]+|_test\.[a-z]+)$|(^|/)test_[^/]+\.py$`)

// A tag counts only on a line that begins, after whitespace, with a comment marker. See
// blueprint/decisions/tags-live-in-comments.md.
//
// Without this rule a tag inside a string literal reads as a claim about a requirement, so any
// test that builds a fixture repo trips the orphan gate. The obvious alternative, exempting
// test paths, is worse: tests are exactly where real tags must live, because coverage is
// measured there, and any path a gate cannot see is a place untraceable code can hide.
var commentMarkers = []string{"//", "#", "--", "/*", "*", "<!--", ";", "%", "\"\"\""}

func inComment(line string) bool {
	trimmed := strings.TrimLeft(line, " \t")
	for _, marker := range commentMarkers {
		if strings.HasPrefix(trimmed, marker) {
			return true
		}
	}
	return false
}

type Tag struct {
	Qualified string
	Path      string
	Test      bool
}

func Tags(root string) ([]Tag, error) {
	files, err := repo.Files(root)
	if err != nil {
		return nil, err
	}
	var tags []Tag
	for _, rel := range files {
		if strings.HasPrefix(rel, "blueprint/spec/") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			if !inComment(line) {
				continue
			}
			for _, match := range tagRE.FindAllStringSubmatch(line, -1) {
				tags = append(tags, Tag{Qualified: match[1], Path: rel, Test: IsTest(rel)})
			}
		}
	}
	sort.Slice(tags, func(i, j int) bool {
		if tags[i].Path == tags[j].Path {
			return tags[i].Qualified < tags[j].Qualified
		}
		return tags[i].Path < tags[j].Path
	})
	return tags, nil
}

func IsTest(path string) bool {
	path = filepath.ToSlash(path)
	return !strings.HasPrefix(path, "blueprint/spec/") && testRE.MatchString(path)
}

func MatchGlob(pattern, name string) bool {
	var b strings.Builder
	b.WriteByte('^')
	for i := 0; i < len(pattern); i++ {
		switch pattern[i] {
		case '*':
			if i+1 < len(pattern) && pattern[i+1] == '*' {
				// `src/**/*.tsx` has to match `src/button.tsx`. Everyone who writes that pattern
				// means "at any depth, including none", and a literal `.*` silently demands a
				// directory in between, so the glob quietly covers less than its author asked for.
				if i+2 < len(pattern) && pattern[i+2] == '/' {
					b.WriteString("(?:.*/)?")
					i += 2
					continue
				}
				b.WriteString(".*")
				i++
			} else {
				b.WriteString("[^/]*")
			}
		case '?':
			b.WriteString("[^/]")
		default:
			b.WriteString(regexp.QuoteMeta(string(pattern[i])))
		}
	}
	b.WriteByte('$')
	return regexp.MustCompile(b.String()).MatchString(filepath.ToSlash(name))
}
