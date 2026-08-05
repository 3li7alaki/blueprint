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
PRODUCT.md                   # register, platform, users, personality, principles  (<= 60 lines)
DESIGN.md                    # what the interface is today. Not ours: see Look below
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

## The passes

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
| look | PRODUCT.md, plus a decision per visual choice | the look gate passes |

Depth dial `quick | standard | paranoid` filters the bank. It never skips a pass. `look` is the
one pass that is skipped whole, and only where no spec declares a surface: a CLI has no register.

## Look

The seven settle what the product does. `look` settles what it is like to meet.

It does not own the interview. Where a design skill is installed, that skill asks, because its
questions are better than a bank maintained beside a spec parser. What `look` owns is the half no
interview has: a shrug becomes an `OPEN.md` entry with a price and a `blocks` glob, and every
visual choice becomes a decision record carrying its reason. An interview ends when the talking
stops. Only a price makes silence expensive.

Two files, split exactly as `stated` and `derived` are:

- `PRODUCT.md` is what a human said. Register, platform, users, positioning, personality,
  anti-references, principles, accessibility. Root, because every design tool reads it. Whoever
  runs the interview may create it; once it exists `hook pre-write` guards it, so it changes
  through an amend and not through a helpful edit.
- `DESIGN.md` is what the interface turned out to be, generated from the code by whatever builds
  the screens. blueprint never writes it and no gate reads it. Checking code against a document
  derived from that same code is a circle.

Visual choices are decision records, never prose in a spec. A colour strategy is a choice with a
reason, and the reason is the only thing that stops the next agent silently re-deciding it. What
is mechanically checkable is the token home: `blueprint/CONVENTIONS.md` names it, and the `tokens`
gate then keeps colour and type out of components.

## Brownfield

When the code exists first, the frame and nouns passes are replaced by a harvest. The rule that
makes it safe is already in the confidence model: code is evidence of what a system does and
never of what it should do, so everything harvested is `derived`, carries `evidence: file:line`,
and stays unimplementable until a human confirms it.

Confirmation has three outcomes, not two. `req confirm` when the code is right, `req correct`
when it should behave differently, `req bug` when nobody meant this. The last two produce a
`stated` requirement the code fails, which is the point: a red gate and a slug beat a suspicion.

Harvest is scoped with `harvest scope`, one area at a time, and the `unmapped` gate only looks
inside those globs. A repo-wide harvest yields hundreds of unconfirmed requirements nobody
reads, which is worse than no spec because it looks like one.

The binary never reads code semantically. It ships `inventory` and `map`; the `harvest` skill
does the reading. Teaching a Go binary to parse every framework is a bottomless pit.

## Adopting an existing spec

A repo that already has written acceptance criteria does not get harvested. Harvesting there
re-derives what a human already stated and leaves a second, weaker spec that looks equally
official. Adoption converts the criteria instead, as `stated`, because a person wrote them, each
carrying `evidence:` back to the source line and its original id.

Adoption is a one-time conversion, never a sync. The source document stops being normative:
deleted, or reduced to narrative that says plainly it is not the spec. Two live specs drift, and
the drift is invisible because both look authoritative. Old citations in code are rewritten in
the same pass rather than kept alive by a mapping, since a second citation scheme that still
counts is the same redundancy one layer down.

Nothing in adoption needs a new command. It is `req add`, `open add` and `decide`, driven by the
`adopt` skill reading what the document already says.

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
