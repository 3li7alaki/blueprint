package gates

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// @spec traceability/coverage-is-literal
// @spec confidence/derived-is-unimplementable
func TestGateFailures(t *testing.T) {
	tests := []struct {
		name string
		file string
		data string
		gate string
	}{
		{name: "coverage", file: "blueprint/spec/feature.md", data: validSpec(), gate: "coverage"},
		{name: "orphan", file: "app.go", data: "// @spec missing/thing\n", gate: "orphan"},
		{name: "derived", file: "app.go", data: "// @spec feature/derived-rule\n", gate: "derived"},
		{name: "dash", file: "notes.txt", data: "bad\u2014dash\n", gate: "dash"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			write(t, root, "blueprint/spec/feature.md", validSpec())
			write(t, root, tt.file, tt.data)
			results := Run(root, tt.gate)
			if len(results) != 1 || results[0].Status != "fail" {
				t.Fatalf("results = %#v", results)
			}
		})
	}
}

// @spec unknowns/open-entries-block
func TestUnmappedGate(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{name: "passes mapped", source: "// @spec feature/uncovered\n", want: "pass"},
		{name: "fails unmapped", source: "package app\n", want: "fail"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			write(t, root, "blueprint/spec/feature.md", validSpec())
			write(t, root, "blueprint/PROJECT.md", "# Project\n\n## Harvested\n- src/**\n")
			write(t, root, "src/app.go", tt.source)
			results := Run(root, "unmapped")
			if len(results) != 1 || results[0].Status != tt.want {
				t.Fatalf("results = %#v", results)
			}
		})
	}
}

// Both directions are asserted with the mtimes set to the opposite of the commit order, because a
// test that only proves "no drift" would also pass on a gate that silently reported nothing.
//
// @spec traceability/drift-survives-checkout
func TestDriftUsesCommitTimeNotModTime(t *testing.T) {
	root := t.TempDir()
	git(t, root, "init")
	git(t, root, "config", "user.email", "test@example.com")
	git(t, root, "config", "user.name", "test")

	write(t, root, "blueprint/spec/feature.md", validSpec())
	commit(t, root, "2026-01-01T00:00:00Z", "spec")
	write(t, root, "src/app.go", "// @spec feature/uncovered\npackage app\n")
	commit(t, root, "2026-01-02T00:00:00Z", "code")

	// A checkout rewrites every file at the moment it runs, in its own order. Reproduce the worst
	// case: the spec landing on disk after the code it was committed before.
	touch(t, root, "src/app.go", 1000)
	touch(t, root, "blueprint/spec/feature.md", 2000)
	if results := Run(root, "drift"); results[0].Status != "pass" {
		t.Fatalf("spec committed first must not be drift: %#v", results)
	}

	write(t, root, "blueprint/spec/feature.md", reworded())
	commit(t, root, "2026-01-03T00:00:00Z", "reword the rule")
	touch(t, root, "blueprint/spec/feature.md", 1000)
	touch(t, root, "src/app.go", 2000)
	if results := Run(root, "drift"); results[0].Status != "fail" {
		t.Fatalf("a rule reworded after its code is drift whatever the mtimes say: %#v", results)
	}
}

// @spec traceability/drift-follows-the-amendment
func TestDriftFollowsTheAmendment(t *testing.T) {
	root := t.TempDir()
	git(t, root, "init")
	git(t, root, "config", "user.email", "test@example.com")
	git(t, root, "config", "user.name", "test")

	// The spec lands after the code it describes, which is what writing a rule for existing
	// behaviour looks like in history. It is not drift.
	write(t, root, "src/app.go", "// @spec feature/uncovered\npackage app\n")
	commit(t, root, "2026-01-01T00:00:00Z", "code")
	write(t, root, "blueprint/spec/feature.md", validSpec())
	commit(t, root, "2026-01-02T00:00:00Z", "spec the existing behaviour")
	if results := Run(root, "drift"); results[0].Status != "pass" {
		t.Fatalf("a rule written for code that already satisfies it is not drift: %#v", results)
	}

	// A sibling arriving later dates the file, never the rules already in it.
	write(t, root, "blueprint/spec/feature.md", strings.Replace(validSpec(), "\n## Edges", "\n### later-arrival\n`stated`\nWHEN asked again, THE system SHALL answer.\nfit: passes\n\n## Edges", 1))
	commit(t, root, "2026-01-03T00:00:00Z", "add a sibling")
	if results := Run(root, "drift"); results[0].Status != "pass" {
		t.Fatalf("adding a sibling requirement is not drift: %#v", results)
	}

	write(t, root, "blueprint/spec/feature.md", reworded())
	commit(t, root, "2026-01-04T00:00:00Z", "reword the rule")
	results := Run(root, "drift")
	if results[0].Status != "fail" || len(results[0].Offenders) != 1 || results[0].Offenders[0] != "feature/uncovered" {
		t.Fatalf("only the reworded rule drifts: %#v", results)
	}
}

func git(t *testing.T, root string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func commit(t *testing.T, root, date, message string) {
	t.Helper()
	git(t, root, "add", "-A")
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE="+date, "GIT_COMMITTER_DATE="+date)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}
}

func touch(t *testing.T, root, rel string, unix int64) {
	t.Helper()
	when := time.Unix(unix, 0)
	if err := os.Chtimes(filepath.Join(root, filepath.FromSlash(rel)), when, when); err != nil {
		t.Fatal(err)
	}
}

func validSpec() string {
	return "# feature\nstatus: ready\ndepth: quick\n\n## Intent\n\n## Surfaces\n\n## Requirements\n### uncovered\n`stated`\nWHEN called, THE system SHALL answer.\nfit: passes\n\n### derived-rule\n`derived`\nWHEN called, THE system SHALL derive.\nfit: passes\n\n## Edges\n| edge | answer |\n\n## Out of scope\n\n## Depends on\n"
}

func reworded() string {
	return strings.Replace(validSpec(), "WHEN called, THE system SHALL answer.", "WHEN called, THE system SHALL answer twice.", 1)
}

func write(t *testing.T, root, rel, data string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}
