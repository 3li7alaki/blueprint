package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSpecShapes(t *testing.T) {
	valid := "# feature\nstatus: ready\ndepth: quick\n\n## Intent\ntext\n\n## Surfaces\n### api\n- who: member\n- data: rows\n- empty: none\n- loading: wait\n- error: retry\n- denied: refuse\n\n## Requirements\n### works\n`stated`\nWHEN called, THE system SHALL answer.\nfit: returns success\n\n## Edges\n| edge | answer |\n\n## Out of scope\n\n## Depends on\n"
	tests := []struct {
		name    string
		edit    func(string) string
		wantErr string
	}{
		{name: "valid", edit: func(s string) string { return s }},
		{name: "filename", edit: func(s string) string { return strings.Replace(s, "# feature", "# other", 1) }, wantErr: ":1:"},
		{name: "confidence", edit: func(s string) string { return strings.Replace(s, "`stated`", "", 1) }, wantErr: "confidence"},
		{name: "surface state", edit: func(s string) string { return strings.Replace(s, "- denied: refuse", "", 1) }, wantErr: "surface denied"},
		{name: "ears", edit: func(s string) string { return strings.Replace(s, "WHEN called, THE system SHALL answer.", "answer", 1) }, wantErr: "EARS"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "feature.md")
			if err := os.WriteFile(path, []byte(tt.edit(valid)), 0o644); err != nil {
				t.Fatal(err)
			}
			_, err := Spec(path)
			if tt.wantErr == "" && err != nil {
				t.Fatal(err)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Fatalf("error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestOpenIgnoresTemplateCommentAndParsesEntries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "OPEN.md")
	data := "# Open\n<!--\n## sample\nstatus: OPEN\n-->\n\n## real\nstatus: OPEN\npass: frame\nasked: 2026-01-01\nowner: founder\nquestion: Why?\ncost: rewrite\nblocks: feature/*\n"
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	entries, err := Opens(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Slug != "real" {
		t.Fatalf("entries = %#v", entries)
	}
}
