package reqfile

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// TermFingerprint hashes the parsed definition and ordered states, excluding status.
func TermFingerprint(t Term) string { return fingerprint(termCanonical(t)) }

func termCanonical(t Term) string {
	return "blueprint-term-fingerprint-v0\ndefinition\n" + t.Definition + "\nstates\n" + strings.Join(t.States, ",") + "\n"
}

func fingerprint(canonical string) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(canonical)))
}

func sortedDistinct(ids []string) []string {
	ids = slices.Clone(ids)
	slices.Sort(ids)
	return slices.Compact(ids)
}

func requirementCanonical(r Requirement, terms, needs map[string]string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "blueprint-requirement-fingerprint-v0\ncontract\n0\nears\n%s\nfit\n%s: %s\nreference\n%s\n", r.EARS, r.FitKind, r.Fit, r.ReferenceDigest)
	for _, id := range sortedDistinct(r.Terms) {
		fmt.Fprintf(&b, "term %s %s\n", id, terms[id])
	}
	for _, id := range sortedDistinct(r.Needs) {
		fmt.Fprintf(&b, "needs %s %s\n", id, needs[id])
	}
	return b.String()
}

// Fingerprints hashes parsed requirements with unique IDs, memoizing needs targets.
// Unknown terms, unknown needs targets and needs cycles prevent fingerprints,
// including for every requirement that transitively needs an invalid requirement.
func Fingerprints(reqs []Requirement, terms map[string]Term) (map[string]string, []Problem) {
	byID := map[string]Requirement{}
	ids := make([]string, 0, len(reqs))
	for _, r := range reqs {
		byID[r.ID] = r
		ids = append(ids, r.ID)
	}
	slices.Sort(ids)
	termHashes := map[string]string{}
	for id, term := range terms {
		termHashes[id] = TermFingerprint(term)
	}
	hashes := map[string]string{}
	state := map[string]int{}
	var problems []Problem
	var path []string
	var visit func(string) string
	visit = func(id string) string {
		if state[id] == 2 {
			return hashes[id]
		}
		if state[id] == 1 {
			cycle := append(slices.Clone(path[slices.Index(path, id):]), id)
			problems = append(problems, Problem{Rule: "dependency-cycle", Msg: strings.Join(cycle, " -> ")})
			return ""
		}
		state[id] = 1
		path = append(path, id)
		r, valid := byID[id], true
		for _, term := range sortedDistinct(r.Terms) {
			if _, ok := terms[term]; !ok {
				problems = append(problems, Problem{Rule: "unknown-term", Msg: id + " -> " + term})
				valid = false
			}
		}
		for _, dep := range sortedDistinct(r.Needs) {
			if _, ok := byID[dep]; !ok {
				problems = append(problems, Problem{Rule: "unknown-dependency", Msg: id + " -> " + dep})
				valid = false
			} else if visit(dep) == "" {
				valid = false
			}
		}
		path = path[:len(path)-1]
		state[id] = 2
		if valid {
			hashes[id] = fingerprint(requirementCanonical(r, termHashes, hashes))
		}
		return hashes[id]
	}
	for _, id := range ids {
		visit(id)
	}
	return hashes, problems
}

// VerifyReference checks a judged reference's raw file bytes after lexical and
// resolved containment checks. Other fit kinds require no filesystem access.
func VerifyReference(repoRoot string, r Requirement) []Problem {
	if r.FitKind != "judged" {
		return nil
	}
	bad := func(rule, msg string) []Problem {
		return []Problem{{Rule: rule, Msg: r.ID + ": " + msg}}
	}
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		return bad("reference-missing", err.Error())
	}
	path := filepath.Join(root, r.ReferencePath)
	if filepath.IsAbs(r.ReferencePath) || !insideRoot(root, path) {
		return bad("reference-outside-root", r.ReferencePath)
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return bad("reference-missing", err.Error())
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return bad("reference-missing", err.Error())
	}
	if !insideRoot(root, path) {
		return bad("reference-outside-root", r.ReferencePath)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return bad("reference-missing", err.Error())
	}
	if fingerprint(string(data)) != "sha256:"+r.ReferenceDigest {
		return bad("reference-digest-mismatch", r.ReferencePath)
	}
	return nil
}

func insideRoot(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
