// Package lint checks requirement text before writers save it.
package lint

import (
	"regexp"
	"strings"

	"blueprint/internal/assets"
)

type Problem struct{ Rule, Detail string }

func (p Problem) Error() string {
	if p.Detail == "" {
		return p.Rule
	}
	return p.Rule + ": " + p.Detail
}

var (
	shallRE      = regexp.MustCompile(`\bSHALL\b`)
	observableRE = regexp.MustCompile("\\p{Nd}|(?i:\\b(zero|one|two|three|four|five|six|seven|eight|nine|ten|eleven|twelve|first|second|third|fourth|fifth|no|none|exactly|every|each)\\b)|`[^`]+`|\"[^\"]+\"|/")
	vagueTerms   = loadVagueTerms()
)

type vagueTerm struct {
	text    string
	pattern *regexp.Regexp
}

func loadVagueTerms() []vagueTerm {
	data, err := assets.FS.ReadFile("vague.txt")
	if err != nil {
		panic(err)
	}
	var terms []vagueTerm
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Hyphens and slashes belong to words, so fast-paced does not match fast.
		pattern := `(?i)(^|[^\pL\pN/-])` + regexp.QuoteMeta(line) + `($|[^\pL\pN/-])`
		terms = append(terms, vagueTerm{line, regexp.MustCompile(pattern)})
	}
	return terms
}

// @spec lint/ears-has-shall-and-full-stop
// @spec lint/one-shall-per-requirement
// @spec lint/no-vague-terms
func EARS(line string) []Problem {
	var problems []Problem
	if !strings.Contains(line, " SHALL ") || !strings.HasSuffix(line, ".") {
		problems = append(problems, Problem{Rule: "ears-shape"})
	}
	if len(shallRE.FindAllStringIndex(line, -1)) > 1 {
		problems = append(problems, Problem{Rule: "one-shall"})
	}
	return append(problems, vague(line)...)
}

// @spec lint/no-vague-terms
// @spec lint/fit-is-observable
func Fit(line string) []Problem {
	problems := vague(line)
	if !observableRE.MatchString(line) {
		problems = append(problems, Problem{Rule: "fit-not-observable"})
	}
	return problems
}

func vague(line string) []Problem {
	var problems []Problem
	for _, term := range vagueTerms {
		if term.pattern.MatchString(line) {
			problems = append(problems, Problem{Rule: "vague-term", Detail: term.text})
		}
	}
	return problems
}
