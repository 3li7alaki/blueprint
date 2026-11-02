package reqfile

import (
	"strings"
	"unicode/utf8"
)

// A nil line map stops value validation after a file-wide encoding error.
func parseHeader(data []byte, fields map[string]*string, body *string, problems *[]Problem) map[string]int {
	if strings.HasPrefix(string(data), "\xef\xbb\xbf") {
		*problems = append(*problems, Problem{Rule: "byte-order-mark", Msg: "leading UTF-8 byte order mark is forbidden"})
		return nil
	}
	if !utf8.Valid(data) {
		*problems = append(*problems, Problem{Rule: "malformed-line", Msg: "expected UTF-8"})
		return nil
	}
	lines := map[string]int{}
	text := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	for i, line := range text {
		if line == "" {
			*body = strings.Join(text[i+1:], "\n")
			break
		}
		key, value, ok := strings.Cut(line, ": ")
		if !ok || !keyPattern.MatchString(key) || strings.HasPrefix(value, " ") || strings.HasPrefix(value, "\t") || strings.ContainsRune(line, '\r') {
			*problems = append(*problems, Problem{"malformed-line", i + 1, "expected key: value with one separator space"})
			continue
		}
		field, known := fields[key]
		if !known {
			*problems = append(*problems, Problem{"unknown-key", i + 1, "unknown key " + key})
			continue
		}
		if lines[key] != 0 {
			*problems = append(*problems, Problem{"duplicate-key", i + 1, "duplicate key " + key})
			continue
		}
		*field, lines[key] = value, i+1
	}
	return lines
}
