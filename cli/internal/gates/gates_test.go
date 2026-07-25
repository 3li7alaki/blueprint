package gates

import (
	"os"
	"path/filepath"
	"testing"
)

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
		{name: "dash", file: "notes.txt", data: "bad—dash\n", gate: "dash"},
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

func validSpec() string {
	return "# feature\nstatus: ready\ndepth: quick\n\n## Intent\n\n## Surfaces\n\n## Requirements\n### uncovered\n`stated`\nWHEN called, THE system SHALL answer.\nfit: passes\n\n### derived-rule\n`derived`\nWHEN called, THE system SHALL derive.\nfit: passes\n\n## Edges\n| edge | answer |\n\n## Out of scope\n\n## Depends on\n"
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
