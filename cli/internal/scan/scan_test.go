package scan

import (
	"os"
	"path/filepath"
	"testing"
)

// @spec traceability/coverage-is-literal
func TestOnlyCommentedTagsCount(t *testing.T) {
	tests := []struct {
		name string
		line string
		want bool
	}{
		{name: "slash comment", line: "// @spec auth/login", want: true},
		{name: "indented", line: "    // @spec auth/login", want: true},
		{name: "hash", line: "# @spec auth/login", want: true},
		{name: "block continuation", line: " * @spec auth/login", want: true},
		{name: "html", line: "<!-- @spec auth/login -->", want: true},
		{name: "sql", line: "-- @spec auth/login", want: true},

		// The case this rule exists for: a fixture string that builds a test repo is data,
		// not a claim that the requirement exists.
		{name: "go string literal", line: `data: "// @spec auth/login\n",`, want: false},
		{name: "assignment", line: `const tag = "@spec auth/login"`, want: false},
		{name: "trailing after code", line: "call() // @spec auth/login", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "source.go")
			if err := os.WriteFile(path, []byte(tt.line+"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			tags, err := Tags(root)
			if err != nil {
				t.Fatal(err)
			}
			if got := len(tags) == 1; got != tt.want {
				t.Errorf("line %q: counted=%v want=%v (tags=%v)", tt.line, got, tt.want, tags)
			}
		})
	}
}
