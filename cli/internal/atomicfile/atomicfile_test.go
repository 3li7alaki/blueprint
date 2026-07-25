package atomicfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWrite(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{name: "plain", content: "plain\n"},
		{name: "em dash", content: "bad\u2014value", wantErr: true},
		{name: "en dash", content: "bad\u2013value", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "file")
			err := Write(path, []byte(tt.content))
			if (err != nil) != tt.wantErr {
				t.Fatalf("Write() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !strings.Contains(err.Error(), path+":1:") {
				t.Fatalf("error %q does not name file and line", err)
			}
			if !tt.wantErr {
				got, err := os.ReadFile(path)
				if err != nil || string(got) != tt.content {
					t.Fatalf("file = %q, %v", got, err)
				}
			}
		})
	}
}
