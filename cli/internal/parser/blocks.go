package parser

import "strings"

// RequirementBlocks maps each requirement slug in a spec to the text of its entry.
//
// @spec traceability/drift-follows-the-amendment
//
// Deliberately not the strict parser. This reads historical revisions of a file, and a revision
// written before a grammar rule existed must still yield its requirements rather than one parse
// error that hides the whole history. Comments and blank lines are dropped so a reflow is not an
// amendment.
func RequirementBlocks(content string) map[string]string {
	blocks := map[string]string{}
	slug := ""
	var body []string
	flush := func() {
		if slug != "" {
			blocks[slug] = strings.Join(body, "\n")
		}
		slug, body = "", nil
	}
	in := false
	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimRight(raw, " \t")
		switch {
		case strings.HasPrefix(line, "## "):
			flush()
			in = strings.TrimSpace(line) == "## Requirements"
		case in && strings.HasPrefix(line, "### "):
			flush()
			slug = trimTicks(strings.TrimSpace(strings.TrimPrefix(line, "### ")))
		case in && slug != "" && !ignorable(line):
			body = append(body, strings.TrimSpace(line))
		}
	}
	flush()
	return blocks
}
