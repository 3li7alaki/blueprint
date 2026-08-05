package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"blueprint/internal/assets"
	"blueprint/internal/atomicfile"
	"blueprint/internal/gates"
	"blueprint/internal/model"
	"blueprint/internal/parser"
	"blueprint/internal/questions"
	"blueprint/internal/repo"
	"blueprint/internal/scan"
)

type App struct {
	In      io.Reader
	Out     io.Writer
	Err     io.Writer
	Now     func() time.Time
	Version string
}

type options struct {
	root string
	json bool
}

var slugRE = regexp.MustCompile(`^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`)

const usage = `blueprint: the spec layer above the code.

Specs are source of truth, code is derivative. Read a requirement, never the whole spec file.

read
  spec list                          every feature, with uncovered counts
  spec show <feature> [--section intent|surfaces|requirements|edges|scope|deps]
  req list [--feature f] [--uncovered] [--derived] [--blocked]
  req show <feature>/<slug>          confidence, EARS line, fit line
  req next                           the next startable requirement
  open list [--blocking <f|f/slug>] [--status OPEN|DEFERRED]
  trace <feature>/<slug>             files carrying the tag, code and test
  map [--path p]                     coverage per directory
  inventory <path>                   files, kind, tags, and line counts
  ask <pass> [--depth quick|standard|paranoid] [--batch n] [--for value] [--all]
  ask --confirm [--path p] [--batch n]
  look show [--section register|platform|users|positioning|principles|...]
  check [--gate coverage|orphan|budget|blocked|derived|shape|dash|unmapped|bug|drift|look|tokens]
  mint <feature>/<slug>              prints the mint unit command, unexecuted

write
  init [--hooks]                     install blueprint into this repo
  look new                           PRODUCT.md from the template, for the look pass to fill
  spec new <feature>
  req add <feature>/<slug> --ears <s> --fit <s> --confidence stated|derived [--evidence <s>]
  req confirm <feature>/<slug>       derived becomes stated. The only way.
  req correct <feature>/<slug> --ears <s> --reason <s>
  req bug <feature>/<slug> --reason <s>
  req fixed <feature>/<slug>
  ask done <question-slug>
  harvest scope <glob>
  hook session|pre-write|done
  amend <feature>/<slug> --ears <s> --reason <s>    the only legal edit
  open add <slug> --question <s> --cost <s> --blocks <globs> [--status] [--pass] [--owner]
  open resolve <slug>
  decide <slug> --context <s> --decision <s> --because <s>
  supersede <old-slug> <new-slug>

global
  --json                             machine output; human output is not the protocol
  --root <path>                      override the git root
  --version, --help

A derived requirement is unimplementable until a human confirms it. An unanswered question
belongs in OPEN.md with its cost, never inferred into a spec.
`

func (a App) Run(args []string) (int, error) {
	opts, args, err := common(args)
	if err != nil {
		return 1, err
	}
	if len(args) == 0 {
		fmt.Fprint(a.Out, usage)
		return 1, nil
	}
	if args[0] == "version" || args[0] == "--version" {
		return 0, a.output(opts.json, map[string]string{"version": a.Version}, a.Version)
	}
	if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		fmt.Fprint(a.Out, usage)
		return 0, nil
	}
	root, err := repo.Root(opts.root)
	if err != nil {
		return 1, err
	}
	switch args[0] {
	case "init":
		err = a.init(root, opts.json, args[1:])
	case "spec":
		err = a.spec(root, opts.json, args[1:])
	case "look":
		err = a.look(root, opts.json, args[1:])
	case "req":
		err = a.req(root, opts.json, args[1:])
	case "open":
		err = a.open(root, opts.json, args[1:])
	case "trace":
		err = a.trace(root, opts.json, args[1:])
	case "map":
		err = a.mapCoverage(root, opts.json, args[1:])
	case "inventory":
		err = a.inventory(root, opts.json, args[1:])
	case "ask":
		err = a.ask(root, opts.json, args[1:])
	case "harvest":
		err = a.harvest(root, opts.json, args[1:])
	case "hook":
		return a.hook(root, opts.json, args[1:])
	case "check":
		return a.check(root, opts.json, args[1:])
	case "mint":
		err = a.mint(root, opts.json, args[1:])
	case "decide":
		err = a.decide(root, opts.json, args[1:])
	case "supersede":
		err = a.supersede(root, opts.json, args[1:])
	case "amend":
		err = a.amend(root, opts.json, args[1:])
	default:
		err = fmt.Errorf("unknown command: %s", args[0])
	}
	if err != nil {
		return 1, err
	}
	return 0, nil
}

func common(args []string) (options, []string, error) {
	var o options
	var rest []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--json":
			o.json = true
		case "--root":
			if i+1 >= len(args) {
				return o, nil, fmt.Errorf("--root requires a value")
			}
			i++
			o.root = args[i]
		default:
			rest = append(rest, args[i])
		}
	}
	return o, rest, nil
}

func (a App) output(asJSON bool, value any, human string) error {
	if asJSON {
		enc := json.NewEncoder(a.Out)
		enc.SetEscapeHTML(false)
		return enc.Encode(value)
	}
	if human != "" {
		_, err := fmt.Fprintln(a.Out, human)
		return err
	}
	return nil
}

func (a App) spec(root string, asJSON bool, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("spec requires list, show, or new")
	}
	switch args[0] {
	case "list":
		if len(args) != 1 {
			return fmt.Errorf("spec list takes no arguments")
		}
		specs, errs := parser.Specs(root)
		if len(errs) > 0 {
			return errs[0]
		}
		tags, _ := scan.Tags(root)
		type item struct {
			Slug         string `json:"slug"`
			Status       string `json:"status"`
			Depth        string `json:"depth"`
			Requirements int    `json:"requirements"`
			Uncovered    int    `json:"uncovered"`
		}
		out := []item{}
		var lines []string
		for _, spec := range specs {
			n := 0
			for _, req := range spec.Requirements {
				if !covered(tags, req.Qualified()) {
					n++
				}
			}
			out = append(out, item{spec.Feature, spec.Status, spec.Depth, len(spec.Requirements), n})
			lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%d\t%d", spec.Feature, spec.Status, spec.Depth, len(spec.Requirements), n))
		}
		return a.output(asJSON, out, strings.Join(lines, "\n"))
	case "show":
		pos, flags, err := parseFlags(args[1:], map[string]bool{"--section": true})
		if err != nil || len(pos) != 1 {
			if err != nil {
				return err
			}
			return fmt.Errorf("spec show requires one feature")
		}
		spec, err := parser.Spec(filepath.Join(root, "blueprint", "spec", pos[0]+".md"))
		if err != nil {
			return err
		}
		section := flags["--section"]
		if section == "" {
			return fmt.Errorf("--section is required")
		}
		key := section
		if section == "scope" {
			key = "out of scope"
		} else if section == "deps" {
			key = "depends on"
		}
		if !oneOf(key, "intent", "surfaces", "requirements", "edges", "out of scope", "depends on") {
			return fmt.Errorf("invalid section: %s", section)
		}
		text := strings.Join(spec.Sections[key], "\n")
		return a.output(asJSON, map[string]any{"feature": spec.Feature, "section": section, "lines": spec.Sections[key]}, text)
	case "new":
		if len(args) != 2 || !validSlug(args[1]) {
			return fmt.Errorf("spec new requires a valid feature slug")
		}
		path := filepath.Join(root, "blueprint", "spec", args[1]+".md")
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%s:1: spec already exists", path)
		}
		template, err := assets.FS.ReadFile("templates/spec.md")
		if err != nil {
			return err
		}
		data := emptySpec(args[1], string(template))
		if err := atomicfile.Write(path, []byte(data)); err != nil {
			return err
		}
		return a.output(asJSON, map[string]string{"path": rel(root, path)}, rel(root, path))
	default:
		return fmt.Errorf("unknown spec command: %s", args[0])
	}
}

func (a App) req(root string, asJSON bool, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("req requires list, show, next, add, or confirm")
	}
	specs, errs := parser.Specs(root)
	if len(errs) > 0 && args[0] != "add" {
		return errs[0]
	}
	reqs := allRequirements(specs)
	switch args[0] {
	case "list":
		pos, flags, err := parseFlags(args[1:], map[string]bool{"--feature": true, "--uncovered": false, "--derived": false, "--blocked": false})
		if err != nil || len(pos) > 0 {
			if err != nil {
				return err
			}
			return fmt.Errorf("req list accepts flags only")
		}
		tags, _ := scan.Tags(root)
		opens, err := parser.Opens(filepath.Join(root, "blueprint", "OPEN.md"))
		if err != nil {
			return err
		}
		out := []model.Requirement{}
		var lines []string
		for _, req := range reqs {
			if flags["--feature"] != "" && req.Feature != flags["--feature"] {
				continue
			}
			if flags["--uncovered"] == "true" && covered(tags, req.Qualified()) {
				continue
			}
			if flags["--derived"] == "true" && req.Confidence != "derived" {
				continue
			}
			if flags["--blocked"] == "true" && !gates.IsBlocked(opens, req.Qualified()) {
				continue
			}
			out = append(out, req)
			lines = append(lines, req.Qualified())
		}
		return a.output(asJSON, out, strings.Join(lines, "\n"))
	case "show":
		if len(args) != 2 {
			return fmt.Errorf("req show requires feature/slug")
		}
		req, ok := reqs[args[1]]
		if !ok {
			return fmt.Errorf("blueprint/spec:1: requirement not found: %s", args[1])
		}
		lines := []string{req.Confidence, req.EARS, "fit: " + req.Fit}
		if req.Evidence != "" {
			lines = append(lines, "evidence: "+req.Evidence)
		}
		if req.Bug != "" {
			lines = append(lines, "bug: "+req.Bug)
		}
		human := strings.Join(lines, "\n")
		return a.output(asJSON, req, human)
	case "next":
		if len(args) != 1 {
			return fmt.Errorf("req next takes no arguments")
		}
		tags, _ := scan.Tags(root)
		opens, err := parser.Opens(filepath.Join(root, "blueprint", "OPEN.md"))
		if err != nil {
			return err
		}
		for _, spec := range specs {
			if spec.Status == "shipped" || !depsMet(root, spec, specs) {
				continue
			}
			for _, req := range spec.Requirements {
				if req.Confidence == "stated" && req.SupersededBy == "" && !covered(tags, req.Qualified()) && !gates.IsBlocked(opens, req.Qualified()) {
					return a.output(asJSON, req, req.Qualified())
				}
			}
		}
		return a.output(asJSON, nil, "")
	case "add":
		pos, flags, err := parseFlags(args[1:], map[string]bool{"--ears": true, "--fit": true, "--confidence": true, "--evidence": true})
		if err != nil || len(pos) != 1 {
			if err != nil {
				return err
			}
			return fmt.Errorf("req add requires feature/slug")
		}
		feature, slug, err := qualified(pos[0])
		if err != nil {
			return err
		}
		if !oneOf(flags["--confidence"], "stated", "derived") || flags["--ears"] == "" || flags["--fit"] == "" {
			return fmt.Errorf("--ears, --fit, and valid --confidence are required")
		}
		if !strings.HasSuffix(flags["--ears"], ".") {
			return fmt.Errorf("--ears must end in a full stop")
		}
		if err := noDash(flags); err != nil {
			return err
		}
		path := filepath.Join(root, "blueprint", "spec", feature+".md")
		spec, err := parser.Spec(path)
		if err != nil {
			return err
		}
		for _, req := range spec.Requirements {
			if req.Slug == slug {
				return fmt.Errorf("%s:%d: requirement already exists", path, req.Line)
			}
		}
		block := fmt.Sprintf("\n### %s\n`%s`\n%s\nfit: %s\n", slug, flags["--confidence"], flags["--ears"], flags["--fit"])
		if flags["--evidence"] != "" {
			block = strings.TrimSuffix(block, "\n") + "\nevidence: " + flags["--evidence"] + "\n"
		}
		if err := insertBefore(path, "## Edges", block); err != nil {
			return err
		}
		return a.output(asJSON, map[string]string{"requirement": pos[0]}, pos[0])
	case "confirm":
		if len(args) != 2 {
			return fmt.Errorf("req confirm requires feature/slug")
		}
		req, ok := reqs[args[1]]
		if !ok {
			return fmt.Errorf("blueprint/spec:1: requirement not found: %s", args[1])
		}
		if req.Confidence != "derived" {
			return fmt.Errorf("%s:%d: requirement is not derived", filepath.Join(root, "blueprint", "spec", req.Feature+".md"), req.Line)
		}
		path := filepath.Join(root, "blueprint", "spec", req.Feature+".md")
		if err := replaceOnce(path, "`derived`", "`stated`", req.Line); err != nil {
			return err
		}
		return a.output(asJSON, map[string]string{"requirement": args[1], "confidence": "stated"}, args[1])
	case "correct":
		pos, flags, err := parseFlags(args[1:], map[string]bool{"--ears": true, "--reason": true})
		if err != nil || len(pos) != 1 {
			if err != nil {
				return err
			}
			return fmt.Errorf("req correct requires feature/slug")
		}
		if flags["--ears"] == "" || flags["--reason"] == "" || !strings.HasSuffix(flags["--ears"], ".") {
			return fmt.Errorf("--ears ending in a full stop and --reason are required")
		}
		if err := noDash(flags); err != nil {
			return err
		}
		req, ok := reqs[pos[0]]
		if !ok {
			return fmt.Errorf("blueprint/spec:1: requirement not found: %s", pos[0])
		}
		path := filepath.Join(root, "blueprint", "spec", req.Feature+".md")
		if err := replaceOnce(path, req.EARS, flags["--ears"], req.Line); err != nil {
			return err
		}
		if req.Confidence == "derived" {
			if err := replaceOnce(path, "`derived`", "`stated`", req.Line); err != nil {
				return err
			}
		}
		return a.output(asJSON, map[string]string{"requirement": pos[0], "reason": flags["--reason"]}, pos[0])
	case "bug":
		pos, flags, err := parseFlags(args[1:], map[string]bool{"--reason": true})
		if err != nil || len(pos) != 1 || flags["--reason"] == "" {
			if err != nil {
				return err
			}
			return fmt.Errorf("req bug requires feature/slug and --reason")
		}
		if err := noDash(flags); err != nil {
			return err
		}
		req, ok := reqs[pos[0]]
		if !ok {
			return fmt.Errorf("blueprint/spec:1: requirement not found: %s", pos[0])
		}
		path := filepath.Join(root, "blueprint", "spec", req.Feature+".md")
		if req.Bug != "" {
			return fmt.Errorf("%s:%d: requirement already has a bug marker", path, req.Line)
		}
		anchor := "fit: " + req.Fit
		if req.Evidence != "" {
			anchor = "evidence: " + req.Evidence
		}
		if err := replaceOnce(path, anchor, anchor+"\nbug: "+flags["--reason"], req.Line); err != nil {
			return err
		}
		if req.Confidence == "derived" {
			if err := replaceOnce(path, "`derived`", "`stated`", req.Line); err != nil {
				return err
			}
		}
		return a.output(asJSON, map[string]string{"requirement": pos[0], "bug": flags["--reason"]}, pos[0])
	case "fixed":
		if len(args) != 2 {
			return fmt.Errorf("req fixed requires feature/slug")
		}
		req, ok := reqs[args[1]]
		if !ok {
			return fmt.Errorf("blueprint/spec:1: requirement not found: %s", args[1])
		}
		if req.Bug == "" {
			return fmt.Errorf("%s:%d: requirement has no bug marker", filepath.Join(root, "blueprint", "spec", req.Feature+".md"), req.Line)
		}
		path := filepath.Join(root, "blueprint", "spec", req.Feature+".md")
		if err := replaceOnce(path, "bug: "+req.Bug, "", req.Line); err != nil {
			return err
		}
		return a.output(asJSON, map[string]string{"requirement": args[1], "fixed": "true"}, args[1])
	default:
		return fmt.Errorf("unknown req command: %s", args[0])
	}
}

func (a App) open(root string, asJSON bool, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("open requires list, add, or resolve")
	}
	openPath := filepath.Join(root, "blueprint", "OPEN.md")
	entries, err := parser.Opens(openPath)
	if err != nil {
		return err
	}
	switch args[0] {
	case "list":
		pos, flags, err := parseFlags(args[1:], map[string]bool{"--blocking": true, "--status": true})
		if err != nil || len(pos) != 0 {
			if err != nil {
				return err
			}
			return fmt.Errorf("open list accepts flags only")
		}
		if flags["--status"] != "" && !oneOf(flags["--status"], "OPEN", "DEFERRED") {
			return fmt.Errorf("invalid open status")
		}
		out := []model.Open{}
		var lines []string
		for _, entry := range entries {
			if flags["--status"] != "" && entry.Status != flags["--status"] {
				continue
			}
			if flags["--blocking"] != "" && !openMatches(entry.Blocks, flags["--blocking"]) {
				continue
			}
			out = append(out, entry)
			lines = append(lines, entry.Slug)
		}
		return a.output(asJSON, out, strings.Join(lines, "\n"))
	case "add":
		pos, flags, err := parseFlags(args[1:], map[string]bool{"--question": true, "--cost": true, "--blocks": true, "--status": true, "--pass": true, "--owner": true})
		if err != nil || len(pos) != 1 || !validSlug(first(pos)) {
			if err != nil {
				return err
			}
			return fmt.Errorf("open add requires a valid slug")
		}
		if flags["--question"] == "" || flags["--cost"] == "" {
			return fmt.Errorf("--question and --cost are required")
		}
		if flags["--status"] == "" {
			flags["--status"] = "OPEN"
		}
		if !oneOf(flags["--status"], "OPEN", "DEFERRED") {
			return fmt.Errorf("invalid open status")
		}
		if flags["--pass"] == "" {
			flags["--pass"] = "unknown"
		}
		if flags["--owner"] == "" {
			flags["--owner"] = "unassigned"
		}
		if err := noDash(flags); err != nil {
			return err
		}
		for _, entry := range entries {
			if entry.Slug == pos[0] {
				return fmt.Errorf("%s:%d: open entry already exists", openPath, entry.Line)
			}
		}
		block := fmt.Sprintf("\n## %s\nstatus: %s\npass: %s\nasked: %s\nowner: %s\nquestion: %s\ncost: %s\nblocks: %s\n",
			pos[0], flags["--status"], flags["--pass"], a.Now().Format("2006-01-02"), flags["--owner"], flags["--question"], flags["--cost"], flags["--blocks"])
		if err := appendFile(openPath, block); err != nil {
			return err
		}
		return a.output(asJSON, map[string]string{"slug": pos[0]}, pos[0])
	case "resolve":
		if len(args) != 2 {
			return fmt.Errorf("open resolve requires a slug")
		}
		var target *model.Open
		for i := range entries {
			if entries[i].Slug == args[1] {
				target = &entries[i]
			}
		}
		if target == nil {
			return fmt.Errorf("%s:1: open entry not found: %s", openPath, args[1])
		}
		if !answerExists(root, args[1]) {
			return fmt.Errorf("%s:%d: answer must exist in a spec or decision file", openPath, target.Line)
		}
		if err := removeOpen(openPath, target.Line); err != nil {
			return err
		}
		return a.output(asJSON, map[string]string{"resolved": args[1]}, args[1])
	default:
		return fmt.Errorf("unknown open command: %s", args[0])
	}
}

func (a App) trace(root string, asJSON bool, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("trace requires feature/slug")
	}
	tags, err := scan.Tags(root)
	if err != nil {
		return err
	}
	result := struct {
		Requirement string   `json:"requirement"`
		Code        []string `json:"code"`
		Test        []string `json:"test"`
	}{Requirement: args[0]}
	for _, tag := range tags {
		if tag.Qualified != args[0] {
			continue
		}
		if tag.Test {
			result.Test = appendUnique(result.Test, tag.Path)
		} else {
			result.Code = appendUnique(result.Code, tag.Path)
		}
	}
	human := "code:\n" + strings.Join(result.Code, "\n") + "\ntest:\n" + strings.Join(result.Test, "\n")
	return a.output(asJSON, result, human)
}

type inventoryItem struct {
	Path   string `json:"path"`
	Kind   string `json:"kind"`
	Tagged bool   `json:"tagged"`
	Lines  int    `json:"lines"`
}

func (a App) inventory(root string, asJSON bool, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("inventory requires a path")
	}
	prefix := strings.Trim(strings.TrimPrefix(filepath.ToSlash(args[0]), "./"), "/")
	files, err := repo.Files(root)
	if err != nil {
		return err
	}
	tags, _ := scan.Tags(root)
	tagged := map[string]bool{}
	for _, tag := range tags {
		tagged[tag.Path] = true
	}
	var out []inventoryItem
	var lines []string
	for _, file := range files {
		if file != prefix && !strings.HasPrefix(file, prefix+"/") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file)))
		if err != nil {
			continue
		}
		count := 0
		if len(data) > 0 {
			count = len(strings.Split(strings.TrimSuffix(string(data), "\n"), "\n"))
		}
		kind := "code"
		if scan.IsTest(file) {
			kind = "test"
		}
		item := inventoryItem{Path: file, Kind: kind, Tagged: tagged[file], Lines: count}
		out = append(out, item)
		lines = append(lines, fmt.Sprintf("%s\t%s\t%t\t%d", file, kind, item.Tagged, count))
	}
	return a.output(asJSON, out, strings.Join(lines, "\n"))
}

type mapItem struct {
	Path     string `json:"path"`
	Files    int    `json:"files"`
	Mapped   int    `json:"mapped"`
	Unmapped int    `json:"unmapped"`
	Derived  int    `json:"derived"`
	Open     int    `json:"open"`
}

func (a App) mapCoverage(root string, asJSON bool, args []string) error {
	pos, flags, err := parseFlags(args, map[string]bool{"--path": true})
	if err != nil || len(pos) != 0 {
		if err != nil {
			return err
		}
		return fmt.Errorf("map accepts flags only")
	}
	prefix := strings.Trim(strings.TrimPrefix(filepath.ToSlash(flags["--path"]), "./"), "/")
	files, err := repo.Files(root)
	if err != nil {
		return err
	}
	tags, _ := scan.Tags(root)
	specs, errs := parser.Specs(root)
	if len(errs) > 0 {
		return errs[0]
	}
	reqs := allRequirements(specs)
	opens, err := parser.Opens(filepath.Join(root, "blueprint", "OPEN.md"))
	if err != nil {
		return err
	}
	byFile := map[string][]scan.Tag{}
	for _, tag := range tags {
		byFile[tag.Path] = append(byFile[tag.Path], tag)
	}
	grouped := map[string]*mapItem{}
	for _, file := range files {
		if strings.HasPrefix(file, "blueprint/") || scan.IsTest(file) || (prefix != "" && file != prefix && !strings.HasPrefix(file, prefix+"/")) {
			continue
		}
		dir := filepath.ToSlash(filepath.Dir(file))
		if dir == "." {
			dir = "."
		}
		item := grouped[dir]
		if item == nil {
			item = &mapItem{Path: dir}
			grouped[dir] = item
		}
		item.Files++
		if len(byFile[file]) == 0 {
			item.Unmapped++
			continue
		}
		item.Mapped++
		for _, tag := range byFile[file] {
			if reqs[tag.Qualified].Confidence == "derived" {
				item.Derived++
			}
			if gates.IsBlocked(opens, tag.Qualified) {
				item.Open++
			}
		}
	}
	var keys []string
	for key := range grouped {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var out []mapItem
	var lines []string
	for _, key := range keys {
		item := *grouped[key]
		out = append(out, item)
		lines = append(lines, fmt.Sprintf("%s\t%d\t%d\t%d\t%d\t%d", item.Path, item.Files, item.Mapped, item.Unmapped, item.Derived, item.Open))
	}
	return a.output(asJSON, out, strings.Join(lines, "\n"))
}

func (a App) harvest(root string, asJSON bool, args []string) error {
	if len(args) != 2 || args[0] != "scope" || strings.TrimSpace(args[1]) == "" {
		return fmt.Errorf("harvest scope requires a glob")
	}
	if strings.ContainsAny(args[1], "\u2013\u2014") {
		return fmt.Errorf("glob contains a forbidden dash character")
	}
	path := filepath.Join(root, "blueprint", "PROJECT.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	text := string(data)
	for _, existing := range harvestedScopes(text) {
		if existing == args[1] {
			return a.output(asJSON, map[string]string{"scope": args[1]}, args[1])
		}
	}
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	insert := len(lines)
	found := false
	for i, line := range lines {
		if line == "## Harvested" {
			found = true
			insert = i + 1
			for insert < len(lines) && !strings.HasPrefix(lines[insert], "## ") {
				insert++
			}
			break
		}
	}
	add := []string{"- " + args[1]}
	if !found {
		add = []string{"", "## Harvested", "- " + args[1]}
	}
	lines = append(lines[:insert], append(add, lines[insert:]...)...)
	if err := atomicfile.Write(path, []byte(strings.Join(lines, "\n")+"\n")); err != nil {
		return err
	}
	return a.output(asJSON, map[string]string{"scope": args[1]}, args[1])
}

func harvestedScopes(text string) []string {
	var out []string
	in := false
	for _, line := range strings.Split(text, "\n") {
		if line == "## Harvested" {
			in = true
			continue
		}
		if in && strings.HasPrefix(line, "## ") {
			break
		}
		if in && strings.HasPrefix(strings.TrimSpace(line), "- ") {
			out = append(out, strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "- ")))
		}
	}
	return out
}

func (a App) ask(root string, asJSON bool, args []string) error {
	if len(args) > 0 && args[0] == "done" {
		if len(args) != 2 || !validSlug(args[1]) {
			return fmt.Errorf("ask done requires a question slug")
		}
		path := filepath.Join(root, "blueprint", ".grill")
		existing, err := os.ReadFile(path)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		for _, line := range strings.Split(string(existing), "\n") {
			if line == args[1] {
				return a.output(asJSON, map[string]string{"question": args[1]}, args[1])
			}
		}
		text := strings.TrimRight(string(existing), "\n")
		if text != "" {
			text += "\n"
		}
		text += args[1] + "\n"
		if err := atomicfile.Write(path, []byte(text)); err != nil {
			return err
		}
		return a.output(asJSON, map[string]string{"question": args[1]}, args[1])
	}
	pos, flags, err := parseFlags(args, map[string]bool{"--depth": true, "--batch": true, "--for": true, "--all": false, "--confirm": false, "--path": true})
	if err != nil {
		return err
	}
	if flags["--confirm"] == "true" {
		if len(pos) != 0 {
			return fmt.Errorf("ask --confirm accepts flags only")
		}
		return a.askConfirm(root, asJSON, flags)
	}
	if len(pos) != 1 {
		if err != nil {
			return err
		}
		return fmt.Errorf("ask requires one pass")
	}
	depth := flags["--depth"]
	if depth == "" {
		depth = "standard"
	}
	if !oneOf(depth, "quick", "standard", "paranoid") {
		return fmt.Errorf("invalid depth")
	}
	batch := 5
	if flags["--batch"] != "" {
		batch, err = strconv.Atoi(flags["--batch"])
		if err != nil || batch < 1 {
			return fmt.Errorf("--batch requires a positive integer")
		}
	}
	bank, err := questions.Load(pos[0])
	if err != nil {
		return err
	}
	opens, err := parser.Opens(filepath.Join(root, "blueprint", "OPEN.md"))
	if err != nil {
		return err
	}
	asked := map[string]bool{}
	for _, entry := range opens {
		asked[entry.Slug] = true
	}
	if flags["--all"] != "true" {
		if data, err := os.ReadFile(filepath.Join(root, "blueprint", ".grill")); err == nil {
			for _, slug := range strings.Split(string(data), "\n") {
				asked[strings.TrimSpace(slug)] = true
			}
		}
	}
	levels := map[string]int{"quick": 0, "standard": 1, "paranoid": 2}
	out := []questions.Question{}
	var lines []string
	for _, q := range bank.Questions {
		if !asked[q.Slug] && levels[q.Depth] <= levels[depth] && len(out) < batch {
			q.Ask = replaceBraces(q.Ask, flags["--for"])
			out = append(out, q)
			lines = append(lines, q.Slug+": "+q.Ask)
		}
	}
	return a.output(asJSON, out, strings.Join(lines, "\n"))
}

func (a App) askConfirm(root string, asJSON bool, flags map[string]string) error {
	batch := 5
	var err error
	if flags["--batch"] != "" {
		batch, err = strconv.Atoi(flags["--batch"])
		if err != nil || batch < 1 {
			return fmt.Errorf("--batch requires a positive integer")
		}
	}
	specs, errs := parser.Specs(root)
	if len(errs) > 0 {
		return errs[0]
	}
	var out []model.Requirement
	for _, spec := range specs {
		for _, req := range spec.Requirements {
			if req.Confidence == "derived" && (flags["--path"] == "" || strings.HasPrefix(req.Evidence, flags["--path"])) {
				out = append(out, req)
			}
		}
	}
	// This heuristic decides ordering only and never whether a question is asked.
	sort.SliceStable(out, func(i, j int) bool { return blastRadius(out[i]) < blastRadius(out[j]) })
	if len(out) > batch {
		out = out[:batch]
	}
	var lines []string
	for _, req := range out {
		lines = append(lines, req.Qualified()+": "+req.Evidence)
	}
	return a.output(asJSON, out, strings.Join(lines, "\n"))
}

func blastRadius(req model.Requirement) int {
	text := strings.ToLower(req.Slug + " " + req.EARS + " " + req.Evidence)
	groups := [][]string{
		{"money", "payment", "price", "billing", "charge", "refund", "currency", "invoice"},
		{"auth", "login", "permission", "access", "role", "session", "token"},
		{"delete", "deletion", "data loss", "erase", "purge", "destroy"},
	}
	for rank, words := range groups {
		for _, word := range words {
			if strings.Contains(text, word) {
				return rank
			}
		}
	}
	return len(groups)
}

var bracesRE = regexp.MustCompile(`\{[^{}]+\}`)

func replaceBraces(text, value string) string {
	if value == "" {
		return text
	}
	return bracesRE.ReplaceAllString(text, value)
}

func (a App) check(root string, asJSON bool, args []string) (int, error) {
	pos, flags, err := parseFlags(args, map[string]bool{"--gate": true})
	if err != nil || len(pos) != 0 {
		if err != nil {
			return 1, err
		}
		return 1, fmt.Errorf("check accepts flags only")
	}
	if flags["--gate"] != "" && !oneOf(flags["--gate"], gates.Names...) {
		return 1, fmt.Errorf("unknown gate: %s", flags["--gate"])
	}
	results := gates.Run(root, flags["--gate"])
	failed := false
	var lines []string
	for _, result := range results {
		failed = failed || result.Status == "fail"
		lines = append(lines, result.Gate+": "+result.Status)
		for _, offender := range result.Offenders {
			lines = append(lines, "  "+offender)
		}
	}
	if asJSON {
		enc := json.NewEncoder(a.Out)
		enc.SetEscapeHTML(false)
		for _, result := range results {
			if err := enc.Encode(result); err != nil {
				return 1, err
			}
		}
	} else if err := a.output(false, results, strings.Join(lines, "\n")); err != nil {
		return 1, err
	}
	if failed {
		return 1, nil
	}
	return 0, nil
}

func (a App) mint(root string, asJSON bool, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("mint requires feature/slug")
	}
	specs, errs := parser.Specs(root)
	if len(errs) > 0 {
		return errs[0]
	}
	req, ok := allRequirements(specs)[args[0]]
	if !ok {
		return fmt.Errorf("blueprint/spec:1: requirement not found: %s", args[0])
	}
	spec := findSpec(specs, req.Feature)
	scope := sectionValues(spec.Sections["out of scope"])
	scope = append(scope, "blueprint/spec/"+req.Feature+".md")
	cmd := "mint spec new " + shellQuote(req.EARS) +
		" --slug " + shellQuote(req.Feature+"--"+req.Slug) +
		" --scope " + shellQuote(strings.Join(scope, ",")) +
		" --acceptance " + shellQuote(req.EARS) +
		" --gate " + shellQuote("blueprint check") +
		" --reviews " + shellQuote("blueprint/REVIEW.md")
	return a.output(asJSON, map[string]string{"command": cmd}, cmd)
}

func (a App) decide(root string, asJSON bool, args []string) error {
	pos, flags, err := parseFlags(args, map[string]bool{"--context": true, "--decision": true, "--because": true})
	if err != nil || len(pos) != 1 || !validSlug(first(pos)) {
		if err != nil {
			return err
		}
		return fmt.Errorf("decide requires a valid slug")
	}
	if flags["--context"] == "" || flags["--decision"] == "" || flags["--because"] == "" {
		return fmt.Errorf("--context, --decision, and --because are required")
	}
	if err := noDash(flags); err != nil {
		return err
	}
	path := filepath.Join(root, "blueprint", "decisions", pos[0]+".md")
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s:1: decision already exists and is immutable", path)
	}
	data := fmt.Sprintf("# %s\nstatus: accepted\nsuperseded-by: \ndate: %s\n\n## Context\n%s\n\n## Decision\n%s\n\n## Because\n%s\n\n## Consequences\n\n",
		pos[0], a.Now().Format("2006-01-02"), flags["--context"], flags["--decision"], flags["--because"])
	if err := atomicfile.Write(path, []byte(data)); err != nil {
		return err
	}
	return a.output(asJSON, map[string]string{"path": rel(root, path)}, rel(root, path))
}

func (a App) supersede(root string, asJSON bool, args []string) error {
	if len(args) != 2 || !validSlug(args[0]) || !validSlug(args[1]) {
		return fmt.Errorf("supersede requires two valid slugs")
	}
	decision := filepath.Join(root, "blueprint", "decisions", args[0]+".md")
	if data, err := os.ReadFile(decision); err == nil {
		text := string(data)
		if !strings.Contains(text, "status: accepted") {
			return fmt.Errorf("%s:2: decision is not accepted", decision)
		}
		text = strings.Replace(text, "status: accepted", "status: superseded", 1)
		text = strings.Replace(text, "superseded-by: ", "superseded-by: "+args[1], 1)
		if err := atomicfile.Write(decision, []byte(text)); err != nil {
			return err
		}
		return a.output(asJSON, map[string]string{"superseded": args[0], "by": args[1]}, args[0]+" -> "+args[1])
	}
	specs, errs := parser.Specs(root)
	if len(errs) > 0 {
		return errs[0]
	}
	for _, spec := range specs {
		for _, req := range spec.Requirements {
			if req.Slug != args[0] {
				continue
			}
			path := filepath.Join(root, "blueprint", "spec", spec.Feature+".md")
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			needle := "fit: " + req.Fit
			replacement := needle + "\nsuperseded-by: " + args[1]
			if strings.Contains(string(data), replacement) {
				return fmt.Errorf("%s:%d: requirement is already superseded", path, req.Line)
			}
			if err := atomicfile.Write(path, []byte(strings.Replace(string(data), needle, replacement, 1))); err != nil {
				return err
			}
			return a.output(asJSON, map[string]string{"superseded": args[0], "by": args[1]}, args[0]+" -> "+args[1])
		}
	}
	return fmt.Errorf("blueprint:1: decision or requirement not found: %s", args[0])
}

func (a App) amend(root string, asJSON bool, args []string) error {
	pos, flags, err := parseFlags(args, map[string]bool{"--ears": true, "--reason": true})
	if err != nil || len(pos) != 1 {
		if err != nil {
			return err
		}
		return fmt.Errorf("amend requires feature/slug")
	}
	if flags["--ears"] == "" || flags["--reason"] == "" || !strings.HasSuffix(flags["--ears"], ".") {
		return fmt.Errorf("--ears ending in a full stop and --reason are required")
	}
	if err := noDash(flags); err != nil {
		return err
	}
	specs, errs := parser.Specs(root)
	if len(errs) > 0 {
		return errs[0]
	}
	req, ok := allRequirements(specs)[pos[0]]
	if !ok {
		return fmt.Errorf("blueprint/spec:1: requirement not found: %s", pos[0])
	}
	recordSlug := "amend-" + req.Feature + "-" + req.Slug
	record := filepath.Join(root, "blueprint", "decisions", recordSlug+".md")
	for {
		if _, err := os.Stat(record); os.IsNotExist(err) {
			break
		}
		recordSlug += "-again"
		record = filepath.Join(root, "blueprint", "decisions", recordSlug+".md")
	}
	data := fmt.Sprintf("# %s\nstatus: accepted\nsuperseded-by: \ndate: %s\n\n## Context\nOld EARS: %s\n\n## Decision\nNew EARS: %s\n\n## Because\n%s\n\n## Consequences\n\n",
		recordSlug, a.Now().Format("2006-01-02"), req.EARS, flags["--ears"], flags["--reason"])
	if err := atomicfile.Write(record, []byte(data)); err != nil {
		return err
	}
	specPath := filepath.Join(root, "blueprint", "spec", req.Feature+".md")
	if err := replaceOnce(specPath, req.EARS, flags["--ears"], req.Line); err != nil {
		_ = os.Remove(record)
		return err
	}
	return a.output(asJSON, map[string]string{"requirement": pos[0], "record": rel(root, record)}, pos[0])
}

func (a App) init(root string, asJSON bool, args []string) error {
	pos, flags, err := parseFlags(args, map[string]bool{"--hooks": false})
	if err != nil || len(pos) != 0 {
		if err != nil {
			return err
		}
		return fmt.Errorf("init accepts --hooks only")
	}
	var written []string
	agentBlock, err := assets.FS.ReadFile("templates/AGENTS.blueprint.md")
	if err != nil {
		return err
	}
	agentsPath := filepath.Join(root, "AGENTS.md")
	if existing, err := os.ReadFile(agentsPath); err == nil {
		if !strings.Contains(string(existing), "## Blueprint") {
			data := append(existing, '\n')
			data = append(data, agentBlock...)
			if err := atomicfile.Write(agentsPath, data); err != nil {
				return err
			}
			written = append(written, "AGENTS.md")
		}
	} else if os.IsNotExist(err) {
		if err := atomicfile.Write(agentsPath, agentBlock); err != nil {
			return err
		}
		written = append(written, "AGENTS.md")
	} else {
		return err
	}
	targets := map[string]string{
		"templates/CLAUDE.md":      "CLAUDE.md",
		"templates/PROJECT.md":     "blueprint/PROJECT.md",
		"templates/OPEN.md":        "blueprint/OPEN.md",
		"templates/CONVENTIONS.md": "blueprint/CONVENTIONS.md",
		"templates/REVIEW.md":      "blueprint/REVIEW.md",
	}
	keys := make([]string, 0, len(targets))
	for key := range targets {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, source := range keys {
		target := targets[source]
		path := filepath.Join(root, filepath.FromSlash(target))
		if _, err := os.Stat(path); err == nil {
			continue
		}
		data, err := assets.FS.ReadFile(source)
		if err != nil {
			return err
		}
		if err := atomicfile.Write(path, data); err != nil {
			return err
		}
		written = append(written, target)
	}
	if err := os.MkdirAll(filepath.Join(root, "blueprint", "spec"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(root, "blueprint", "decisions"), 0o755); err != nil {
		return err
	}
	if flags["--hooks"] == "true" {
		changed, err := mergeHookSettings(root)
		if err != nil {
			return err
		}
		if changed {
			written = append(written, ".claude/settings.json")
		}
	}
	sort.Strings(written)
	return a.output(asJSON, map[string]any{"written": written}, strings.Join(written, "\n"))
}

func mergeHookSettings(root string) (bool, error) {
	path := filepath.Join(root, ".claude", "settings.json")
	settings := map[string]any{}
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &settings); err != nil {
			return false, fmt.Errorf("%s:1: invalid JSON: %v", path, err)
		}
	} else if !os.IsNotExist(err) {
		return false, err
	}
	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		hooks = map[string]any{}
		settings["hooks"] = hooks
	}
	entries := []struct {
		event   string
		matcher string
		command string
	}{
		{"SessionStart", "", "blueprint hook session"},
		{"PreToolUse", "Edit|Write|MultiEdit", "blueprint hook pre-write"},
		{"Stop", "", "blueprint hook done"},
	}
	changed := false
	for _, entry := range entries {
		list, _ := hooks[entry.event].([]any)
		if containsCommand(list, entry.command) {
			continue
		}
		item := map[string]any{"hooks": []any{map[string]any{"type": "command", "command": entry.command}}}
		if entry.matcher != "" {
			item["matcher"] = entry.matcher
		}
		hooks[entry.event] = append(list, item)
		changed = true
	}
	if !changed {
		return false, nil
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return false, err
	}
	data = append(data, '\n')
	return true, atomicfile.Write(path, data)
}

func containsCommand(value any, command string) bool {
	switch value := value.(type) {
	case map[string]any:
		if value["command"] == command {
			return true
		}
		for _, child := range value {
			if containsCommand(child, command) {
				return true
			}
		}
	case []any:
		for _, child := range value {
			if containsCommand(child, command) {
				return true
			}
		}
	}
	return false
}

func (a App) hook(root string, asJSON bool, args []string) (int, error) {
	if len(args) != 1 || !oneOf(args[0], "session", "pre-write", "done") {
		return 1, fmt.Errorf("hook requires session, pre-write, or done")
	}
	in := a.In
	if in == nil {
		in = os.Stdin
	}
	payload, err := io.ReadAll(io.LimitReader(in, 1<<20))
	if err != nil {
		return 1, err
	}
	switch args[0] {
	case "pre-write":
		if os.Getenv("BLUEPRINT_AMEND") == "1" {
			return 0, nil
		}
		var value any
		if len(strings.TrimSpace(string(payload))) > 0 {
			if err := json.Unmarshal(payload, &value); err != nil {
				return 1, fmt.Errorf("stdin:1: invalid hook JSON: %v", err)
			}
		}
		file := findJSONText(value, "file_path")
		if strings.Contains(filepath.ToSlash(file), "/blueprint/spec/") || strings.HasPrefix(filepath.ToSlash(file), "blueprint/spec/") {
			return hookBlock(a.Err, "BLOCKED: blueprint/spec/ is not directly editable.")
		}
		if strings.Contains(filepath.ToSlash(file), "/blueprint/decisions/") || strings.HasPrefix(filepath.ToSlash(file), "blueprint/decisions/") {
			return hookBlock(a.Err, "BLOCKED: decision records are immutable.")
		}
		if file != "" && productGuarded(root, file) {
			return hookBlock(a.Err, "BLOCKED: PRODUCT.md holds what a human said. Run the look pass or amend it, never write it by hand.")
		}
	case "session":
		var out strings.Builder
		out.WriteString("BLUEPRINT ACTIVE. Specs in blueprint/ are source of truth; code is derivative.\n")
		opens, _ := parser.Opens(filepath.Join(root, "blueprint", "OPEN.md"))
		n := 0
		for _, entry := range opens {
			if entry.Status == "OPEN" && n < 5 {
				if n == 0 {
					out.WriteString("\nOpen questions blocking work:\n")
				}
				out.WriteString(entry.Slug + "\n")
				n++
			}
		}
		results := gates.Run(root, "")
		n = 0
		for _, result := range results {
			if result.Status == "fail" && n < 5 {
				if n == 0 {
					out.WriteString("\nFailing gates:\n")
				}
				out.WriteString(result.Gate + ": fail\n")
				n++
			}
		}
		specs, errs := parser.Specs(root)
		if len(errs) == 0 {
			tags, _ := scan.Tags(root)
			for _, spec := range specs {
				if spec.Status == "shipped" || !depsMet(root, spec, specs) {
					continue
				}
				found := false
				for _, req := range spec.Requirements {
					if req.Confidence == "stated" && req.SupersededBy == "" && !covered(tags, req.Qualified()) && !gates.IsBlocked(opens, req.Qualified()) {
						out.WriteString("\nNext implementable requirement: " + req.Qualified() + "\n")
						found = true
						break
					}
				}
				if found {
					break
				}
			}
		}
		fmt.Fprint(a.Out, out.String())
	case "done":
		var value any
		_ = json.Unmarshal(payload, &value)
		session := findJSONText(value, "session_id")
		if session == "" {
			session = os.Getenv("CLAUDE_CODE_SESSION_ID")
		}
		if session == "" {
			session = "default"
		}
		sentinel := filepath.Join(os.TempDir(), "blueprint-done-gate-"+safeName(session))
		if info, err := os.Stat(sentinel); err == nil && time.Since(info.ModTime()) < time.Minute {
			return 0, nil
		}
		results := gates.Run(root, "")
		var failures []string
		for _, result := range results {
			if result.Status == "fail" {
				failures = append(failures, result.Gate+": fail")
			}
		}
		if len(failures) > 0 {
			if err := atomicfile.Write(sentinel, []byte("blocked\n")); err != nil {
				return 1, err
			}
			return hookBlock(a.Err, "BLOCKED: blueprint gates are failing, so this is not done.\n\n"+strings.Join(failures, "\n"))
		}
	}
	if asJSON {
		return 0, a.output(true, map[string]bool{"allow": true}, "")
	}
	return 0, nil
}

func findJSONText(value any, key string) string {
	switch value := value.(type) {
	case map[string]any:
		if text, ok := value[key].(string); ok {
			return text
		}
		for _, child := range value {
			if text := findJSONText(child, key); text != "" {
				return text
			}
		}
	case []any:
		for _, child := range value {
			if text := findJSONText(child, key); text != "" {
				return text
			}
		}
	}
	return ""
}

func safeName(value string) string {
	return regexp.MustCompile(`[^a-zA-Z0-9._-]+`).ReplaceAllString(value, "_")
}

func hookBlock(errOut io.Writer, reason string) (int, error) {
	if errOut == nil {
		errOut = os.Stderr
	}
	_, err := fmt.Fprintln(errOut, reason)
	return 2, err
}

func parseFlags(args []string, allowed map[string]bool) ([]string, map[string]string, error) {
	flags := map[string]string{}
	var pos []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		takesValue, ok := allowed[arg]
		if !ok {
			if strings.HasPrefix(arg, "--") {
				return nil, flags, fmt.Errorf("unknown flag: %s", arg)
			}
			pos = append(pos, arg)
			continue
		}
		if takesValue {
			if i+1 >= len(args) {
				return nil, flags, fmt.Errorf("%s requires a value", arg)
			}
			i++
			flags[arg] = args[i]
		} else {
			flags[arg] = "true"
		}
	}
	return pos, flags, nil
}

func allRequirements(specs []model.Spec) map[string]model.Requirement {
	out := map[string]model.Requirement{}
	for _, spec := range specs {
		for _, req := range spec.Requirements {
			out[req.Qualified()] = req
		}
	}
	return out
}

func covered(tags []scan.Tag, qualified string) bool {
	for _, tag := range tags {
		if tag.Qualified == qualified && tag.Test {
			return true
		}
	}
	return false
}

func depsMet(root string, spec model.Spec, specs []model.Spec) bool {
	for _, dep := range sectionValues(spec.Sections["depends on"]) {
		for _, candidate := range specs {
			if candidate.Feature == dep && candidate.Status == "shipped" {
				dep = ""
				break
			}
		}
		if dep == "" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, "blueprint", "decisions", dep+".md"))
		if err != nil || !strings.Contains(string(data), "status: accepted") {
			return false
		}
	}
	return true
}

func sectionValues(lines []string) []string {
	var out []string
	for _, line := range lines {
		line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "-"))
		line = strings.Trim(line, "`")
		if line != "" && !strings.HasPrefix(line, "<!--") && !strings.HasPrefix(line, "-->") {
			out = append(out, line)
		}
	}
	return out
}

func findSpec(specs []model.Spec, feature string) model.Spec {
	for _, spec := range specs {
		if spec.Feature == feature {
			return spec
		}
	}
	return model.Spec{}
}

func qualified(value string) (string, string, error) {
	feature, slug, ok := strings.Cut(value, "/")
	if !ok || !validSlug(feature) || !validSlug(slug) {
		return "", "", fmt.Errorf("expected valid feature/slug")
	}
	return feature, slug, nil
}

func validSlug(value string) bool { return slugRE.MatchString(value) }

func noDash(values map[string]string) error {
	for flag, value := range values {
		if strings.ContainsAny(value, "\u2013\u2014") {
			return fmt.Errorf("%s: value contains a forbidden dash character", flag)
		}
	}
	return nil
}

func insertBefore(path, marker, block string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	index := strings.Index(string(data), marker)
	if index < 0 {
		return fmt.Errorf("%s:1: missing section %s", path, marker)
	}
	text := string(data[:index]) + strings.TrimSuffix(block, "\n") + "\n\n" + string(data[index:])
	return atomicfile.Write(path, []byte(text))
}

func appendFile(path, block string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	text := strings.TrimRight(string(data), "\n") + "\n" + block
	return atomicfile.Write(path, []byte(text))
}

// replaceOnce rewrites one line of a requirement block, starting the search at the
// requirement's heading line.
//
// Searching the whole file instead would be wrong: a spec template explains the format in
// prose, so the literal tokens it writes about (`derived`, an example EARS sentence) appear
// in guidance comments above the real requirements. A whole-file replace edits the comment,
// reports success, and leaves the requirement untouched.
func replaceOnce(path, old, new string, line int) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	if line < 1 || line > len(lines) {
		return fmt.Errorf("%s:%d: line out of range", path, line)
	}
	for i := line - 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == strings.TrimSpace(old) {
			lines[i] = strings.Replace(lines[i], strings.TrimSpace(old), strings.TrimSpace(new), 1)
			return atomicfile.Write(path, []byte(strings.Join(lines, "\n")))
		}
		// Stop at the next requirement so an edit can never land in a neighbour's block.
		if i > line-1 && strings.HasPrefix(lines[i], "### ") {
			break
		}
	}
	return fmt.Errorf("%s:%d: expected text not found in this requirement", path, line)
}

func emptySpec(feature, template string) string {
	text := strings.Replace(template, "# <feature-slug>", "# "+feature, 1)
	text = strings.Replace(text, "status: `drafting` | `ready` | `building` | `shipped`", "status: drafting", 1)
	text = strings.Replace(text, "depth: `quick` | `standard` | `paranoid`", "depth: standard", 1)
	if start := strings.Index(text, "### `<requirement-slug>`"); start >= 0 {
		if end := strings.Index(text[start:], "## Edges"); end >= 0 {
			text = text[:start] + text[start+end:]
		}
	}
	return text
}

func rel(root, path string) string {
	value, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(value)
}

func first(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func oneOf(value string, choices ...string) bool {
	for _, choice := range choices {
		if value == choice {
			return true
		}
	}
	return false
}

func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func openMatches(patterns, target string) bool {
	if !strings.Contains(target, "/") {
		target += "/*"
	}
	for _, pattern := range strings.Split(patterns, ",") {
		pattern = strings.TrimSpace(pattern)
		if pattern == "*" || pattern == target {
			return true
		}
		if strings.HasSuffix(pattern, "/*") && strings.HasPrefix(target, strings.TrimSuffix(pattern, "*")) {
			return true
		}
		if strings.HasSuffix(target, "/*") && strings.HasPrefix(pattern, strings.TrimSuffix(target, "*")) {
			return true
		}
	}
	return false
}

func answerExists(root, slug string) bool {
	for _, pattern := range []string{filepath.Join(root, "blueprint", "spec", "*.md"), filepath.Join(root, "blueprint", "decisions", "*.md")} {
		paths, _ := filepath.Glob(pattern)
		for _, path := range paths {
			if data, err := os.ReadFile(path); err == nil && strings.Contains(string(data), slug) {
				return true
			}
		}
	}
	return false
}

func removeOpen(path string, line int) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	start := line - 1
	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			end = i
			break
		}
	}
	for start > 0 && strings.TrimSpace(lines[start-1]) == "" {
		start--
	}
	return atomicfile.Write(path, []byte(strings.Join(append(lines[:start], lines[end:]...), "\n")))
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}
