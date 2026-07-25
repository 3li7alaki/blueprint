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

type Tag struct {
	Qualified string
	Path      string
	Test      bool
	ModTime   int64
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
		info, _ := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
		for _, match := range tagRE.FindAllStringSubmatch(string(data), -1) {
			tags = append(tags, Tag{Qualified: match[1], Path: rel, Test: IsTest(rel), ModTime: info.ModTime().UnixNano()})
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
