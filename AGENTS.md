# blueprint: the spec contract

> blueprint decides what is worth building and forbids ambiguity. mint decides whether a built
> thing is allowed to be called done.

This file is blueprint's own contract, for agents working *on* blueprint. The block installed
into a consuming repo lives in `templates/AGENTS.blueprint.md` and is deliberately shorter.

## Boundaries

blueprint owns: the question bank, specs, requirement slugs, the unknowns register, conventions,
review lenses, and the repo-side gates that keep code traceable to a requirement.

blueprint does not own: tickets, worktrees, branches, agent lifecycle, retries, deployment
(drivers), or completion evidence, provenance, receipts and freshness (mint). Never re-implement
either.

## Artifacts in a consuming repo

```text
AGENTS.md                    # 15-line blueprint block, pointers only
CLAUDE.md                    # @AGENTS.md, nothing else
blueprint/
  PROJECT.md                 # problem, users, non-goals, success metric   (<= 60 lines)
  OPEN.md                    # unresolved questions, priced, blocking      (grows)
  CONVENTIONS.md             # one rule per recurring thing + exemplar path (<= 80 lines)
  REVIEW.md                  # adversarial lenses only                     (<= 60 lines)
  spec/<feature>.md          # EARS requirements, surfaces, states, edges  (<= 150 lines)
  decisions/<slug>.md        # immutable; superseded by slug, never edited
```

## Slugs

Words, never numbers. A number carries no meaning and renumbers on insert.

- feature slug = spec filename stem: `auth-magic-link`
- requirement slug = `### ` heading inside that file: `link-expires-in-15-min`
- qualified: `auth-magic-link/link-expires-in-15-min`
- decision: `blueprint/decisions/postgres-over-dynamo.md`

A slug is immutable once implemented. Renaming is superseding: new slug, old one marked
`superseded-by: <slug>`. A rename is a semantic change and must break traces loudly.

## Requirement shape

Every requirement is one `### <slug>` heading followed by exactly three things:

```md
### link-expires-in-15-min
`stated`
WHILE a magic link is unused, WHEN it is older than 15 minutes, THE system SHALL reject it.
fit: request with a 16-minute-old link returns 410 and no session is created.
```

- **EARS** grammar. No `shall` without a trigger or a precondition.
- **fit criterion** (Volere): a mechanical pass/fail. A requirement without one is a mood.
- **confidence**: `stated` (the human said it) or `derived` (inferred). `derived` is not
  implementable until a human confirms it. This is the anti-hallucination clause.

## Operating rules

1. No spec, no code. Missing spec means run the grill, not guess.
2. Never answer your own question. "You decide" produces a decision record, visible, with
   rationale, not a silent choice buried in an implementation.
3. Anything unresolved goes to `OPEN.md` verbatim with its cost line. Never into a spec.
4. `OPEN.md` blocking a feature means that feature is not startable. Report; do not route around.
5. Tag implementing symbols `@spec <feature>/<slug>`. Tag the covering test the same.
6. Build only what a requirement asks. Unrequested work fails review even if it works.
7. Code follows spec. Never edit a spec to match code. That is `/blueprint amend`, human-gated.
8. `blueprint check` passes before any claim of done. mint issues the receipt, not you.
9. Never write an em dash or an en dash, in this repo or in anything blueprint generates.
   Use a comma, a colon, parentheses, or a full stop. `blueprint check` fails on one.

## The seven passes

Ordered; each feeds the next. Question bank lives in `questions/`, one file per pass, loaded
on demand.

| pass | produces | stop condition |
|---|---|---|
| frame | problem, who hurts, cost of nothing | a weak answer kills the project here |
| boundaries | non-goals, v1 cut line | non-goals are explicit and written |
| nouns | entities, lifecycle states, definitions | every state has in/out transitions or is terminal |
| surfaces | pages, screens, endpoints | each has empty/loading/error/denied states |
| rules | EARS requirements + slugs | every verb has a trigger and a response |
| edges | adversarial sweep per rule | concurrent, empty, huge, hostile, deleted-parent covered |
| gates | fit criteria | every requirement is mechanically checkable |

Depth dial `quick | standard | paranoid` filters the bank. It never skips a pass.

## Handoff to mint

One requirement, one mint unit:

```bash
mint spec new "<goal>" --slug <feature>--<requirement> \
  --scope "<paths from the spec>,blueprint/spec/<feature>.md" \
  --acceptance "<the EARS line verbatim>" \
  --gate "<from AGENTS.md>" --reviews "<from REVIEW.md>"
```

The spec file is inside `--scope` on purpose: editing the spec then invalidates the receipt,
so spec drift is caught by mint's existing freshness rule rather than new code.
