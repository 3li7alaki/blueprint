# blueprint CLI contract

An agent should never read a 150-line spec to implement one requirement. It should read four
lines. That is the whole reason this binary exists.

`blueprint` is a parser and a writer for the artifacts described in `AGENTS.md`. It holds no
state of its own: the repo files are the state. It launches nothing, calls no model, and has no
network access.

Every command supports `--json`. Human output is never the integration protocol.

## Read

| command | returns |
|---|---|
| `blueprint spec list` | every feature: slug, status, depth, requirement count, uncovered count |
| `blueprint spec show <feature> [--section intent\|surfaces\|requirements\|edges\|scope\|deps]` | one section, not the file |
| `blueprint req list [--feature <f>] [--uncovered] [--derived] [--blocked]` | qualified slugs, one per line |
| `blueprint req show <feature>/<slug>` | confidence, EARS line, fit line, nothing else |
| `blueprint req next` | the next implementable requirement: not covered, not blocked, deps met |
| `blueprint open list [--blocking <feature\|feature/slug>] [--status OPEN\|DEFERRED]` | matching entries |
| `blueprint trace <feature>/<slug>` | files carrying the tag, split into code and test |
| `blueprint ask <pass> [--depth quick\|standard\|paranoid] [--batch 5]` | next unanswered questions from the bank |
| `blueprint look show [--section <name>]` | one section of `PRODUCT.md`, or the section names |
| `blueprint check [--gate <name>]` | gate results, exit 1 on any failure |
| `blueprint mint <feature>/<slug>` | the `mint spec new` command line for that requirement, on stdout, unexecuted |
| `blueprint map [--path <p>]` | coverage per directory: files, mapped, unmapped, derived, open |
| `blueprint inventory <path>` | every file under path: test or code, tagged or not, line count |

`req next` is the command an idle agent loop calls. It is the whole scheduler.

`ask` substitutes `{braces}` in a question with `--for <value>`. The bank asks about `{entity}`
once per entity, and the caller owns that loop. Teaching the binary to enumerate entities would
mean parsing prose, so it does not try.

`ask` skips questions already recorded in `blueprint/.grill`, so a grill resumes across sessions
instead of restarting at question one. `--all` ignores the ledger.

## Brownfield

An existing codebase is not a blank page, and pretending otherwise is why spec tools get
abandoned in week two. Code is evidence of what a system does. It is never evidence of what it
should do, so everything harvested from it is `derived` and stays unimplementable until a human
confirms it.

The binary does not read code semantically. It ships the ledger; an agent running the `harvest`
skill reads the source and calls the writers. Teaching a Go binary to parse every framework is a
bottomless pit and the model is better at it anyway.

| command | does |
|---|---|
| `blueprint harvest scope <glob>` | adds a glob to the harvested set in `PROJECT.md`. The progress ledger. |
| `blueprint req correct <feature>/<slug> --ears <s> --reason <s>` | the human says the code is wrong: rewrite to intended behaviour, mark `stated` |
| `blueprint req bug <feature>/<slug> --reason <s>` | the human says the current behaviour is a defect: mark `stated`, record `bug:` |
| `blueprint req fixed <feature>/<slug>` | clears the `bug:` marker once the code matches |
| `blueprint ask --confirm [--path <p>] [--batch 5]` | derived requirements needing a human, ordered by blast radius, each with its evidence |

`ask --confirm` orders by consequence, never by file order: money, then auth and permissions,
then data loss, then everything else. People skim when question twelve feels as trivial as
question eleven. The classifier is a keyword match over the slug, EARS text and evidence paths,
which is crude on purpose; it decides ordering only, never whether something is asked.

Three answers, not two. `confirm` when the code is right, `correct` when it should behave
differently, `bug` when nobody meant this. The last two are why harvesting an existing codebase
pays for itself on day one: each produces a `stated` requirement the current code fails, which
is a red gate and a slug to fix against, rather than a vague sense that something is off.

`harvest scope` is required before the `unmapped` gate reports anything. Scope one area at a
time, the one you are about to touch. A repo-wide harvest produces hundreds of unconfirmed
requirements nobody reads, which is worse than no spec because it looks like one.

## Adopting an existing spec

Some repos already have specs, and better ones than a first grill would produce. Harvesting
requirements from code in that situation is the wrong move twice over: it re-derives what a
human already stated, and it produces a second spec that looks as authoritative as the real one
while being strictly weaker.

Adoption is a one-time conversion, not a sync. Two spec systems always drift, and the drift is
invisible precisely because both look official. So the existing document is converted and then
stops being normative. It may be deleted, or left as narrative that no gate reads, but it is
never a second place requirements live.

What already has a home, so needs no new grammar:

| in an existing spec | becomes |
|---|---|
| acceptance criteria | requirements, `stated`, because a human wrote them |
| a blocked or unanswered item | an `OPEN.md` entry, priced |
| rationale, research notes, rejected options | a decision record |
| an unmitigated risk | an `OPEN.md` entry, since that is a question with a cost |
| a mitigated risk | the `because` line of the decision that mitigated it |
| in scope, out of scope | the spec's own `Out of scope` section |

Everything adopted carries `evidence:` pointing at the source document and line, so a reader can
check the conversion rather than trust it.

Confidence is `stated`, not `derived`. This is the one place `stated` is correct without a live
human in the loop: the requirement was written by a person, and adoption is a translation of
their words, not an inference from behaviour. Adopting as `derived` would force a human to
re-confirm decisions they already made, which is how a migration stalls.

Existing citations in code are rewritten in the same pass, not kept alive by a mapping. A repo
that cites `AC-5` in comments gets those comments replaced with the `@spec` tag for the
requirement that criterion became. A compatibility layer would be a second citation scheme that
still counts, which is the same redundancy as a second spec, one layer down. The edit is small:
citations cluster in a handful of files, and anything the conversion misses shows up as an
uncovered requirement rather than a silent wrong link.

The reading is done by an agent under the `adopt` skill, for the same reason harvest is: the
binary must not learn to parse thirteen house styles of markdown. Nothing in adoption needs a
new command; it is `req add`, `open add` and `decide` driven from what the document already
says.

## Write

Writers are the only legal mutation path for `blueprint/`. They preserve section order, never
reformat untouched lines, and refuse to produce an em dash or en dash.

| command | does |
|---|---|
| `blueprint look new` | writes root `PRODUCT.md` from the template. Refuses to replace one |
| `blueprint spec new <feature>` | writes `blueprint/spec/<feature>.md` from the template |
| `blueprint req add <feature>/<slug> --ears <s> --fit <s> --confidence stated\|derived [--evidence <s>]` | appends a requirement |
| `blueprint req confirm <feature>/<slug>` | flips `derived` to `stated`. The only way. |
| `blueprint open add <slug> --question <s> --cost <s> --blocks <globs> [--status OPEN\|DEFERRED] [--pass <p>] [--owner <s>]` | appends an unknown |
| `blueprint open resolve <slug>` | removes the entry. Refuses unless the answer already exists in a spec or decision file. |
| `blueprint decide <slug> --context <s> --decision <s> --because <s>` | writes an immutable decision record |
| `blueprint supersede <old-slug> <new-slug>` | marks a decision or requirement superseded |
| `blueprint amend <feature>/<slug> --ears <s> --reason <s>` | the only legal edit to an existing requirement. Records the old text and the reason. |

`blueprint init` writes `AGENTS.md` (appending the block if the file exists), `CLAUDE.md`, and
`blueprint/` from templates. It never overwrites an existing spec.

`blueprint init --hooks` additionally merges three entries into `.claude/settings.json`, creating
it if absent and leaving unrelated keys untouched:

| event | command | effect |
|---|---|---|
| `SessionStart` | `blueprint hook session` | injects open blockers, failing gates, next requirement |
| `PreToolUse` on `Edit\|Write\|MultiEdit` | `blueprint hook pre-write` | refuses hand edits to `blueprint/spec/`, `blueprint/decisions/`, and `PRODUCT.md` once it exists |
| `Stop` | `blueprint hook done` | refuses to end a turn while a gate is red |

The hooks are subcommands, not scripts written into the repo. A copied script goes stale the
day the binary changes, and every repo would carry its own drifting copy. Each reads the hook
payload on stdin and exits 0 to allow or 2 to block, with the reason on stderr.

`blueprint hook pre-write` allows the edit when `BLUEPRINT_AMEND=1`, which the writers set for
their own atomic replacements.

## Gates

`blueprint check` runs all of these. `--gate <name>` runs one.

| gate | fails when |
|---|---|
| `coverage` | a requirement slug appears in no file under a test path |
| `orphan` | a `@spec` tag names a requirement that does not exist |
| `budget` | a doc exceeds its line budget |
| `blocked` | a `@spec` tag matches a `blocks:` glob of a live `OPEN.md` entry |
| `derived` | a `derived` requirement has implementing code |
| `shape` | a requirement is missing its confidence, EARS or fit line |
| `dash` | any tracked file contains an em dash or en dash |
| `unmapped` | a non-test file inside a harvested scope carries no `@spec` tag |
| `bug` | a requirement carries a `bug:` marker, meaning the code is known to violate it |
| `drift` | a requirement was reworded after the last change to the code carrying its tag |
| `look` | a spec declares a surface and `PRODUCT.md` is missing or a required section is empty |
| `tokens` | a component file carries a raw colour or font stack instead of using the token home |

`drift` compares requirement text across commits, never file timestamps. Three consequences, and
each one is a reason a timestamp cannot do this job:

- Adding a requirement dates the spec file but not the requirements already in it, so a new
  sibling never drags its neighbours red.
- A requirement written for code that already satisfies it is an introduction, not an amendment,
  so harvesting a brownfield repo does not light the gate up on day one.
- A clone, a checkout, and a task worktree rewrite every mtime in whatever order they walk the
  tree. Commit times survive all three, so the gate answers the same in every copy.

Uncommitted work is dated by mtime, having no commit yet. Where git cannot answer at all the gate
reports nothing: no history means no information, which is not the same as no change.

`look` requires `register`, `platform`, `users`, `positioning` and `accessibility & inclusion`.
The rest of `PRODUCT.md` is worth having and none of it is worth a red gate; a check that fires on
a missing adjective teaches people to ignore checks. It stays silent until a spec declares a
surface, so a CLI or a worker never sees it.

`tokens` is opt-in through a machine-read section of `blueprint/CONVENTIONS.md`, the same shape as
`## Harvested` in `PROJECT.md`:

```md
## Tokens
home: src/styles/tokens.css
components: src/**/*.tsx, src/**/*.css
```

It matches raw hex colours and font stacks only. Spacing is deliberately excluded: a 1px border is
legitimate in any component, so a px rule would cry wolf until somebody switched the gate off. A
line carrying `blueprint:allow-raw` is skipped, because an SVG fill or a third party embed is a
real exception, and a gate with no exit is a gate people delete.

Exit code is 0 or 1. `--json` emits `{gate, status, offenders[]}` per gate.

## Test path detection

A file is a test file when its path matches any of:

```
(^|/)tests?/         (^|/)__tests__/       (^|/)spec/
\.test\.[a-z]+$      \.spec\.[a-z]+$       _test\.[a-z]+$      (^|/)test_[^/]+\.py$
```

`blueprint/spec/` is excluded. It is specs, not tests.

Coverage means the literal string `@spec <feature>/<slug>` appears in at least one test file.
No name normalisation, no heuristics. One rule, greppable by hand when the binary is absent.

## File grammar

The parser is strict. A file that does not match is a `shape` failure, never a silent skip.

### `blueprint/spec/<feature>.md`

```md
# <feature-slug>
status: drafting | ready | building | shipped
depth: quick | standard | paranoid

## Intent
<free text>

## Surfaces
### <surface-slug>
- who: <role-slug>, <role-slug>
- data: <text>
- empty: <text>
- loading: <text>
- error: <text>
- denied: <text>

## Requirements
### <requirement-slug>
`stated` | `derived`
<EARS line, ending in a full stop>
fit: <one line>
evidence: <optional, one line>
bug: <optional, one line>

## Edges
| edge | answer |

## Out of scope
## Depends on
```

`evidence:` is optional and free text. `harvest` writes `file:line` references so a human can
check a derived claim instead of taking it on faith; a grill may instead cite the question that
produced it. Confirmation without a pointer is a vote, not a review.

`bug:` is optional and written only by `req bug`. Its presence means the requirement states the
intended behaviour and the code is known to violate it, so the `bug` gate stays red until
`req fixed` clears it. Both lines, when present, follow `fit:` in that order.

`###` headings mean a surface under `## Surfaces` and a requirement under `## Requirements`.
Backticks around a heading slug are stripped. Feature slug must equal the filename stem.

### `blueprint/OPEN.md`

One entry per `##` heading, one key per line, no inline separators:

```md
## <question-slug>
status: OPEN
pass: frame
asked: 2026-07-25
owner: founder
question: <one line, verbatim as asked>
cost: <what has to be rebuilt if this is guessed wrong>
blocks: checkout-split-payment/*, ledger/*
```

`blocks: *` blocks the whole project. An empty `blocks:` blocks nothing and is still recorded.

### `blueprint/decisions/<slug>.md`

```md
# <slug>
status: accepted | superseded
superseded-by: <slug>
date: 2026-07-25

## Context
## Decision
## Because
## Consequences
```

Immutable once written. A change is a new file plus `blueprint supersede`.

### `PRODUCT.md`

Root, or `.agents/context/` or `docs/`, which is where every design tool already looks. One
`## ` heading per section, free text under each, and an empty section reads as unanswered rather
than as answered blank:

```md
## Register
product
## Platform
web
## Users
## Product Purpose
## Positioning
## Brand Personality
## Anti-references
## Design Principles
## Accessibility & Inclusion
## Conversion & proof
```

`Register` is `brand` or `product`, `Platform` is `web`, `ios`, `android` or `adaptive`, both as
bare words. `Conversion & proof` belongs to the brand register and is deleted outright for a
product.

Created by `look new` and filled in by the look pass, which keeps priced `OPEN.md` entries for
what nobody decided and a decision record per visual choice. Once the file exists it is guarded,
so later changes go through an amend rather than a helpful edit.

`DESIGN.md` beside it is not blueprint's. It records what the interface actually is, generated
from the code by whatever builds the screens, and no gate reads it.

### `blueprint/PROJECT.md`

Free text except one machine-read section, absent in a greenfield project:

```md
## Harvested
- src/auth/**
- src/checkout/**
```

Written by `blueprint harvest scope`. The `unmapped` gate only looks inside these globs, so an
un-harvested legacy tree never fails a gate and the list itself is the progress ledger.

### `blueprint/.grill`

An append-only ledger, one question slug per line, written when a bank question has been asked
and dealt with. Not a document, never read by a human, and the reason a grill resumes where it
stopped instead of restarting at question one.


## Implementation notes

- Go, one module under `cli/`, standard library only. Zero third-party dependencies, ever.
  Single static binary, `CGO_ENABLED=0`. The question bank uses one flat TOML subset
  (`key = "string"`, `key = ["a", "b"]`, `[[q]]` tables) so a small internal reader covers it.
- Distribution matches mint: every push to `main` rebuilds four platform binaries and
  republishes the rolling `latest` release. No semver, no tags. `install.sh` fetches `latest`.
- No network at runtime, no telemetry, no config file, no daemon.
- Resolve the repo root from git, or `--root`.
- Every writer is atomic: temp file plus rename, same directory.
- Errors name the file and line. "parse error" alone is a defect.
