package questions

import (
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"

	"blueprint/internal/assets"
)

type Bank struct {
	Pass      string     `json:"pass"`
	Produces  string     `json:"produces"`
	Note      string     `json:"note"`
	Questions []Question `json:"questions"`
}

type Question struct {
	Slug   string   `json:"slug"`
	Depth  string   `json:"depth"`
	Ask    string   `json:"ask"`
	Exit   string   `json:"exit"`
	Cost   string   `json:"cost"`
	Blocks []string `json:"blocks"`
}

func Load(pass string) (Bank, error) {
	names, err := fs.Glob(assets.FS, "questions/*.toml")
	if err != nil {
		return Bank{}, err
	}
	sort.Strings(names)
	for _, name := range names {
		data, err := assets.FS.ReadFile(name)
		if err != nil {
			return Bank{}, err
		}
		bank, err := parse(name, string(data))
		if err != nil {
			return Bank{}, err
		}
		if bank.Pass == pass {
			return bank, nil
		}
	}
	return Bank{}, fmt.Errorf("questions/%s:1: unknown pass", pass)
}

func parse(path, data string) (Bank, error) {
	var bank Bank
	var current *Question
	for i, raw := range strings.Split(data, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if line == "[[q]]" {
			bank.Questions = append(bank.Questions, Question{})
			current = &bank.Questions[len(bank.Questions)-1]
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return Bank{}, fmt.Errorf("%s:%d: expected key = value", path, i+1)
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if current == nil {
			s, err := quoted(value)
			if err != nil {
				return Bank{}, fmt.Errorf("%s:%d: %v", path, i+1, err)
			}
			switch key {
			case "pass":
				bank.Pass = s
			case "produces":
				bank.Produces = s
			case "note":
				bank.Note = s
			default:
				return Bank{}, fmt.Errorf("%s:%d: unknown key %s", path, i+1, key)
			}
			continue
		}
		if key == "blocks" {
			a, err := quotedArray(value)
			if err != nil {
				return Bank{}, fmt.Errorf("%s:%d: %v", path, i+1, err)
			}
			current.Blocks = a
			continue
		}
		s, err := quoted(value)
		if err != nil {
			return Bank{}, fmt.Errorf("%s:%d: %v", path, i+1, err)
		}
		switch key {
		case "slug":
			current.Slug = s
		case "depth":
			current.Depth = s
		case "ask":
			current.Ask = s
		case "exit":
			current.Exit = s
		case "cost":
			current.Cost = s
		default:
			return Bank{}, fmt.Errorf("%s:%d: unknown key %s", path, i+1, key)
		}
	}
	return bank, nil
}

func quoted(value string) (string, error) {
	s, err := strconv.Unquote(value)
	if err != nil {
		return "", fmt.Errorf("expected quoted string")
	}
	return s, nil
}

func quotedArray(value string) ([]string, error) {
	if !strings.HasPrefix(value, "[") || !strings.HasSuffix(value, "]") {
		return nil, fmt.Errorf("expected array of quoted strings")
	}
	body := strings.TrimSpace(value[1 : len(value)-1])
	if body == "" {
		return []string{}, nil
	}
	parts := strings.Split(body, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		s, err := quoted(strings.TrimSpace(part))
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}
