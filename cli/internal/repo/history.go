package repo

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Revision is one committed version of a file, plus its commit time in unix seconds.
type Revision struct {
	Time    int64
	Content string
}

// History returns every version of rel, oldest first, ending with the working tree when it holds
// uncommitted changes.
//
// @spec traceability/drift-follows-the-amendment
//
// A caller asking when one requirement last changed cannot get that from a file timestamp: adding
// a second requirement to a spec would date every requirement in it. Comparing consecutive
// revisions is the only answer that stays true per requirement.
//
// An empty result means git could not answer, and a caller must treat that as no information
// rather than as no change.
func History(root, rel string) []Revision {
	cmd := exec.Command("git", "log", "--reverse", "--format=%ct %H", "--", rel)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	var revs []Revision
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		secs, sha, ok := strings.Cut(strings.TrimSpace(line), " ")
		if !ok {
			continue
		}
		when, err := strconv.ParseInt(secs, 10, 64)
		if err != nil {
			continue
		}
		show := exec.Command("git", "show", sha+":"+rel)
		show.Dir = root
		content, err := show.Output()
		if err != nil {
			continue
		}
		revs = append(revs, Revision{Time: when, Content: string(content)})
	}
	if len(revs) == 0 {
		return nil
	}
	if dirtySet(root)[filepath.ToSlash(rel)] {
		if content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel))); err == nil {
			revs = append(revs, Revision{Time: modTime(root, rel), Content: string(content)})
		}
	}
	return revs
}
