package app

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"blueprint/internal/assets"
	"blueprint/internal/atomicfile"
	"blueprint/internal/parser"
)

// look reads and creates PRODUCT.md, the strategic half of the design contract.
//
// @spec look/product-is-stated
//
// There is no writer for the individual sections, and that is deliberate. The answers arrive as
// prose from an interview, so a flag per field would be a worse editor than a text file. What the
// binary owns is the shape: creating the file from the template, reading one section back without
// a caller opening the whole document, and failing the gate when a surface exists and the file
// does not.
func (a App) look(root string, asJSON bool, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("look requires show or new")
	}
	switch args[0] {
	case "show":
		pos, flags, err := parseFlags(args[1:], map[string]bool{"--section": true})
		if err != nil || len(pos) != 0 {
			if err != nil {
				return err
			}
			return fmt.Errorf("look show accepts --section only")
		}
		product, ok := parser.ReadProduct(root)
		if !ok {
			return fmt.Errorf("%s:1: no PRODUCT.md, run the look pass", rel(root, product.Path))
		}
		section := strings.ToLower(flags["--section"])
		if section == "" {
			names := make([]string, 0, len(product.Sections))
			for name := range product.Sections {
				names = append(names, name)
			}
			sort.Strings(names)
			return a.output(asJSON, product, strings.Join(names, "\n"))
		}
		if section == "principles" {
			section = "design principles"
		}
		lines, seen := product.Sections[section]
		if !seen {
			return fmt.Errorf("%s:1: no section named %s", rel(root, product.Path), section)
		}
		return a.output(asJSON, map[string]any{"section": section, "lines": lines}, strings.Join(lines, "\n"))
	case "new":
		path, exists := parser.ProductPath(root)
		if exists {
			return fmt.Errorf("%s:1: PRODUCT.md already exists, amend it instead of replacing it", rel(root, path))
		}
		template, err := assets.FS.ReadFile("templates/PRODUCT.md")
		if err != nil {
			return err
		}
		if err := atomicfile.Write(path, template); err != nil {
			return err
		}
		return a.output(asJSON, map[string]string{"path": rel(root, path)}, rel(root, path))
	default:
		return fmt.Errorf("unknown look command: %s", args[0])
	}
}

// productGuarded reports whether a write has to be refused because PRODUCT.md already exists.
//
// @spec look/product-is-stated
//
// Creating the file is open to whoever runs the interview, including a design skill that does it
// better than a question bank would. What is guarded is everything after: a silent edit to what a
// human said is indistinguishable from an answer they never gave. Once one exists, a second copy
// under another accepted path is refused too, because two of these is the failure the single file
// was chosen to avoid.
func productGuarded(root, path string) bool {
	target, exists := parser.ProductPath(root)
	if !exists {
		return false
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	if same, err := filepath.Abs(target); err == nil && abs == same {
		return true
	}
	return strings.EqualFold(filepath.Base(abs), "product.md") && insideRoot(root, abs)
}

func insideRoot(root, path string) bool {
	value, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(value, "..") && !filepath.IsAbs(value)
}
