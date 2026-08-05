package gates

import (
	"strings"
	"testing"
)

// @spec look/surfaces-need-a-product
func TestLookGate(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		product string
		want    string
	}{
		{name: "no surfaces, no product needed", spec: validSpec(), want: "pass"},
		{name: "surface without a product", spec: withSurface(), want: "fail"},
		{name: "surface with an unanswered section", spec: withSurface(), product: strings.Replace(fullProduct(), "web\n", "\n", 1), want: "fail"},
		{name: "surface with a product", spec: withSurface(), product: fullProduct(), want: "pass"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			write(t, root, "blueprint/spec/feature.md", tt.spec)
			if tt.product != "" {
				write(t, root, "PRODUCT.md", tt.product)
			}
			results := Run(root, "look")
			if len(results) != 1 || results[0].Status != tt.want {
				t.Fatalf("results = %#v", results)
			}
		})
	}
}

// @spec look/tokens-stay-home
func TestTokensGate(t *testing.T) {
	conventions := "# Conventions\n\n## Tokens\nhome: src/tokens.css\ncomponents: src/**/*.tsx\n"
	tests := []struct {
		name        string
		conventions string
		source      string
		want        string
	}{
		{name: "unscoped repo is not checked", source: "const c = \"#ff0000\"\n", want: "pass"},
		{name: "raw colour in a component", conventions: conventions, source: "const c = \"#ff0000\"\n", want: "fail"},
		{name: "raw font stack in a component", conventions: conventions, source: "const s = {fontFamily: 1}\nconst t = \"font-family: Inter\"\n", want: "fail"},
		{name: "an exception is allowed to say so", conventions: conventions, source: "const c = \"#ff0000\" // blueprint:allow-raw\n", want: "pass"},
		{name: "tokens live in the home file", conventions: conventions, source: "const c = \"var(--brand)\"\n", want: "pass"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			write(t, root, "blueprint/spec/feature.md", validSpec())
			write(t, root, "src/tokens.css", ":root { --brand: #ff0000 }\n")
			write(t, root, "src/button.tsx", tt.source)
			if tt.conventions != "" {
				write(t, root, "blueprint/CONVENTIONS.md", tt.conventions)
			}
			results := Run(root, "tokens")
			if len(results) != 1 || results[0].Status != tt.want {
				t.Fatalf("results = %#v", results)
			}
		})
	}
}

func withSurface() string {
	return strings.Replace(validSpec(), "## Surfaces\n", "## Surfaces\n### dashboard\n- who: operator\n- data: reads orders\n- empty: no orders yet\n- loading: skeleton rows\n- error: retry\n- denied: 404\n", 1)
}

func fullProduct() string {
	return "# Product\n\n## Register\n\nproduct\n\n## Platform\n\nweb\n\n## Users\n\noperators at a desk\n\n## Positioning\n\nthe only one that shows the queue\n\n## Accessibility & Inclusion\n\nWCAG 2.2 AA\n"
}
