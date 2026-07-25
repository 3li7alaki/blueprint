package repo

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// A gate that cannot see new files reports green on code it never read. Plain `git ls-files`
// lists only tracked files, so every scan-based gate passed on an empty list until Files
// started asking for untracked ones too.
func TestFilesIncludesUntracked(t *testing.T) {
	root := t.TempDir()
	if out, err := exec.Command("git", "-C", root, "init", "-q").CombinedOutput(); err != nil {
		t.Skipf("git unavailable: %v: %s", err, out)
	}

	tracked := filepath.Join(root, "tracked.js")
	if err := os.WriteFile(tracked, []byte("// @spec a/b\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "-C", root, "add", "tracked.js").CombinedOutput(); err != nil {
		t.Fatalf("git add: %v: %s", err, out)
	}
	if err := os.WriteFile(filepath.Join(root, "untracked.js"), []byte("// @spec a/c\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("ignored.js\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "ignored.js"), []byte("// @spec a/d\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	files, err := Files(root)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, f := range files {
		got[f] = true
	}
	for _, want := range []string{"tracked.js", "untracked.js"} {
		if !got[want] {
			t.Errorf("Files() missing %q, got %v", want, files)
		}
	}
	if got["ignored.js"] {
		t.Errorf("Files() returned an ignored file, got %v", files)
	}
}
