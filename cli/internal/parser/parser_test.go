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

// @spec lint/read-accepts-what-write-accepts
func TestSpecEARSShapeAndReadLeniency(t *testing.T) {
	for _, tt := range []struct {
		name    string
		ears    string
		fit     string
		wantErr bool
	}{
		{name: "IF THEN", ears: "IF a card is declined, THEN THE system SHALL show the decline code.", fit: "returns 201"},
		{name: "ubiquitous", ears: "THE system SHALL log every refund.", fit: "returns 201"},
		{name: "WHERE", ears: "WHERE refunds are enabled, THE system SHALL log every refund.", fit: "returns 201"},
		{name: "WHILE", ears: "WHILE running, THE system SHALL log every refund.", fit: "returns 201"},
		{name: "missing SHALL", ears: "THE system logs every refund.", fit: "returns 201", wantErr: true},
		{name: "missing full stop", ears: "THE system SHALL log every refund", fit: "returns 201", wantErr: true},
		{name: "heading instead of EARS", ears: "### THE system SHALL log every refund.", fit: "returns 201", wantErr: true},
		{name: "old repeated SHALL", ears: "WHEN called, THE system SHALL accept cards and SHALL log payment.", fit: "returns 201"},
		{name: "old vague EARS", ears: "WHEN called, THE system SHALL respond fast.", fit: "returns 201"},
		{name: "old vague fit", ears: "WHEN called, THE system SHALL answer.", fit: "1 fast response"},
		{name: "old unobservable fit", ears: "WHEN called, THE system SHALL answer.", fit: "returns success"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "feature.md")
			data := "# feature\nstatus: ready\ndepth: quick\n\n## Intent\n\n## Surfaces\n\n## Requirements\n### works\n`stated`\n" + tt.ears + "\nfit: " + tt.fit + "\n\n## Edges\n\n## Out of scope\n\n## Depends on\n"
			if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
				t.Fatal(err)
			}
			spec, err := Spec(path)
			if tt.wantErr {
				want := path + ":12: missing EARS line ending in a full stop"
				if err == nil || err.Error() != want {
					t.Fatalf("error = %v, want %q", err, want)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(spec.Requirements) != 1 {
				t.Fatalf("requirements = %#v, want one requirement", spec.Requirements)
			}
			got := spec.Requirements[0]
			if got.EARS != tt.ears || got.Fit != tt.fit {
				t.Fatalf("requirement = %#v, want EARS %q and fit %q", got, tt.ears, tt.fit)
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

func TestRequirementOptionalLinesRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		optional string
		evidence string
		bug      string
	}{
		{name: "absent"},
		{name: "evidence", optional: "evidence: src/pay.go:12\n", evidence: "src/pay.go:12"},
		{name: "evidence and bug", optional: "evidence: src/pay.go:12\nbug: charges twice\n", evidence: "src/pay.go:12", bug: "charges twice"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "feature.md")
			data := "# feature\nstatus: ready\ndepth: quick\n\n## Intent\n\n## Surfaces\n\n## Requirements\n### works\n`derived`\nWHEN called, THE system SHALL answer.\nfit: returns success\n" + tt.optional + "\n## Edges\n| edge | answer |\n\n## Out of scope\n\n## Depends on\n"
			if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
				t.Fatal(err)
			}
			spec, err := Spec(path)
			if err != nil {
				t.Fatal(err)
			}
			got := spec.Requirements[0]
			if got.Evidence != tt.evidence || got.Bug != tt.bug {
				t.Fatalf("requirement = %#v", got)
			}
		})
	}
}
