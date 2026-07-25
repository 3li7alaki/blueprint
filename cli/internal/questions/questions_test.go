package questions

import "testing"

func TestLoad(t *testing.T) {
	tests := []struct {
		pass string
		want int
	}{
		{pass: "frame", want: 10},
		{pass: "boundaries", want: 8},
		{pass: "nouns", want: 10},
		{pass: "surfaces", want: 10},
		{pass: "rules", want: 10},
		{pass: "edges", want: 10},
		{pass: "gates", want: 8},
	}
	for _, tt := range tests {
		t.Run(tt.pass, func(t *testing.T) {
			bank, err := Load(tt.pass)
			if err != nil {
				t.Fatal(err)
			}
			if len(bank.Questions) != tt.want {
				t.Fatalf("questions = %d, want %d", len(bank.Questions), tt.want)
			}
		})
	}
}

func TestParseRejectsUnsupportedTOML(t *testing.T) {
	tests := []string{"value = 4", "[q]", "unknown = \"x\""}
	for _, input := range tests {
		if _, err := parse("bank.toml", input); err == nil {
			t.Fatalf("parse(%q) succeeded", input)
		}
	}
}
