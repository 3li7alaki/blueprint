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
| `blueprint check [--gate <name>]` | gate results, exit 1 on any failure |
| `blueprint mint <feature>/<slug>` | the `mint spec new` command line for that requirement, on stdout, unexecuted |

`req next` is the command an idle agent loop calls. It is the whole scheduler.

## Write

Writers are the only legal mutation path for `blueprint/`. They preserve section order, never
reformat untouched lines, and refuse to produce an em dash or en dash.

| command | does |
|---|---|
| `blueprint spec new <feature>` | writes `blueprint/spec/<feature>.md` from the template |
| `blueprint req add <feature>/<slug> --ears <s> --fit <s> --confidence stated\|derived` | appends a requirement |
| `blueprint req confirm <feature>/<slug>` | flips `derived` to `stated`. The only way. |
| `blueprint open add <slug> --question <s> --cost <s> --blocks <globs> [--status OPEN\|DEFERRED] [--pass <p>] [--owner <s>]` | appends an unknown |
| `blueprint open resolve <slug>` | removes the entry. Refuses unless the answer already exists in a spec or decision file. |
| `blueprint decide <slug> --context <s> --decision <s> --because <s>` | writes an immutable decision record |
| `blueprint supersede <old-slug> <new-slug>` | marks a decision or requirement superseded |
| `blueprint amend <feature>/<slug> --ears <s> --reason <s>` | the only legal edit to an existing requirement. Records the old text and the reason. |

`blueprint init` writes `AGENTS.md` (appending the block if the file exists), `CLAUDE.md`,
`blueprint/` from templates, and the hook entries. It never overwrites an existing spec.

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
| `drift` | a spec file is newer than the newest file carrying one of its `@spec` tags |

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

## Edges
| edge | answer |

## Out of scope
## Depends on
```

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
