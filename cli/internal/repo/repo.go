package repo

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func Root(explicit string) (string, error) {
	if explicit != "" {
		return filepath.Abs(explicit)
	}
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(out)), nil
	}
	return os.Getwd()
}

// Files lists every file a gate should see.
//
// --others is not optional. Plain `git ls-files` shows only tracked files, so brand new work
// is invisible and every scan-based gate passes on an empty list. A gate that cannot see the
// code it guards is worse than no gate: it reports green. --exclude-standard keeps ignored
// build output and vendored trees out.
func Files(root string) ([]string, error) {
	cmd := exec.Command("git", "ls-files", "--cached", "--others", "--exclude-standard")
	cmd.Dir = root
	if out, err := cmd.Output(); err == nil {
		seen := map[string]bool{}
		var files []string
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if line != "" && !seen[line] {
				seen[line] = true
				files = append(files, filepath.ToSlash(line))
			}
		}
		sort.Strings(files)
		return files, nil
	}
	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		if d.Type().IsRegular() {
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			files = append(files, filepath.ToSlash(rel))
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}
