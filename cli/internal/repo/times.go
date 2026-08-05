package repo

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Times returns an effective change time, in unix seconds, for each repo-relative path.
//
// @spec traceability/drift-survives-checkout
//
// mtime is the wrong clock for asking which of two files changed last. A clone, a checkout, or a
// fresh worktree writes every file at the moment it runs, in whatever order it walks the tree, so
// a spec can look newer than the code implementing it on a tree nobody has edited. The commit that
// last touched a file survives all three. A file with uncommitted changes has no such commit and
// is genuinely newer than anything committed, so it keeps its mtime.
//
// One `git log` per path. The path count is the specs plus the files carrying their tags, which is
// tens in a real repo. If that ever measures slow, the replacement is a single `git log
// --name-only` walk taking the first occurrence of each path, at the cost of reading whole history.
func Times(root string, paths []string) map[string]int64 {
	times := make(map[string]int64, len(paths))
	dirty := dirtySet(root)
	for _, p := range paths {
		rel := filepath.ToSlash(p)
		if _, seen := times[rel]; seen {
			continue
		}
		if !dirty[rel] {
			if secs, ok := lastCommit(root, rel); ok {
				times[rel] = secs
				continue
			}
		}
		times[rel] = modTime(root, rel)
	}
	return times
}

// dirtySet lists every path git reports as changed, staged, or untracked. A nil map is the honest
// answer when git is unavailable: every path then falls through to lastCommit, which fails for the
// same reason, and the whole gate degrades to mtime rather than reporting a wrong answer loudly.
func dirtySet(root string) map[string]bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	dirty := map[string]bool{}
	for _, line := range strings.Split(string(out), "\n") {
		if len(line) < 4 {
			continue
		}
		path := line[3:]
		// A rename reads `R  old -> new`. The new name is the one on disk.
		if i := strings.LastIndex(path, " -> "); i >= 0 {
			path = path[i+4:]
		}
		dirty[strings.Trim(path, `"`)] = true
	}
	return dirty
}

func lastCommit(root, rel string) (int64, bool) {
	cmd := exec.Command("git", "log", "-1", "--format=%ct", "--", rel)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return 0, false
	}
	secs, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return 0, false
	}
	return secs, true
}

func modTime(root, rel string) int64 {
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return 0
	}
	return info.ModTime().Unix()
}
