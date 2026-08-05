package parser

import (
	"os"
	"path/filepath"
	"strings"
)

// ProductNames and productDirs mirror where every design tool already looks for this file. A
// second convention would mean a repo with two of them and no way to tell which one is read.
var productNames = []string{"PRODUCT.md", "Product.md", "product.md"}
var productDirs = []string{".", ".agents/context", "docs"}

// Product is the strategic half of the design contract: what a human said about who this is for
// and how it should come across. Sections are lower-cased headings mapped to their content lines,
// comments and blanks already stripped, so an unanswered section is an empty slice.
//
// @spec look/product-is-stated
type Product struct {
	Path     string              `json:"path"`
	Sections map[string][]string `json:"sections"`
}

// ProductPath returns where PRODUCT.md is, or where it should be written, and whether it exists.
func ProductPath(root string) (string, bool) {
	for _, dir := range productDirs {
		for _, name := range productNames {
			path := filepath.Join(root, filepath.FromSlash(dir), name)
			if _, err := os.Stat(path); err == nil {
				return path, true
			}
		}
	}
	return filepath.Join(root, "PRODUCT.md"), false
}

func ReadProduct(root string) (Product, bool) {
	path, ok := ProductPath(root)
	if !ok {
		return Product{Path: path}, false
	}
	lines, err := readLines(path)
	if err != nil {
		return Product{Path: path}, false
	}
	product := Product{Path: path, Sections: map[string][]string{}}
	section := ""
	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			section = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(line, "## ")))
			if _, seen := product.Sections[section]; !seen {
				product.Sections[section] = nil
			}
			continue
		}
		if section == "" || ignorable(line) {
			continue
		}
		product.Sections[section] = append(product.Sections[section], strings.TrimSpace(line))
	}
	return product, true
}

// Answered reports whether a section exists and holds content. An empty section is the shape a
// half-run grill leaves behind, and it has to read as unanswered rather than as answered blank.
func (p Product) Answered(section string) bool {
	return len(p.Sections[strings.ToLower(section)]) > 0
}
