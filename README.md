# blueprint

> Grill the idea until it has no holes, then hand your agents a spec they can't misread.

AI doesn't fail your project at the coding step. It fails at the step where nobody said what
"done" meant, so the model guessed: plausibly, confidently, and differently each session.

blueprint is the layer above the code. It interrogates the idea until every feature, every
screen, every rule and every edge case is written down with a slug on it. Then it wires that
spec into the repo so agents can't quietly drift from it.

```
   idea  ──grill──▶  spec  ──cook──▶  code  ──check──▶  receipt
    ▲                  │                │                  │
    └──── OPEN.md ◀────┘         @spec tags ───────────────┘
         unknowns, priced         traceability
```

```bash
curl -fsSL https://raw.githubusercontent.com/3li7alaki/blueprint/main/install.sh | sh
```

One static binary, no runtime dependencies. [Plugin and first steps below.](#install)

## The problem, precisely

Three failures, all the same root:

- **Hallucinated requirements.** Nobody specified it, so the model invented it. It looked fine.
- **Context loss.** Session two doesn't know what session one decided. Neither do you.
- **Founder ignorance, deferred.** The gap between what was in their head and what they said
  surfaces after it's built, priced as a rewrite.

blueprint attacks all three at the source: the ambiguity that made guessing necessary.

## How it works

**1. Grill.** `/blueprint grill <feature>` runs a question bank, not a freestyle chat. Seven
passes: frame, boundaries, nouns, surfaces, rules, edges, gates. Each question has an exit
condition. It doesn't move on until the answer clears it.

Unanswered questions don't vanish and they don't get invented. They land in `OPEN.md` with a
price tag:

```md
## checkout-currency-rounding
status: DEFERRED BY FOUNDER · asked 2026-07-25
question: half-up or banker's rounding on split payments?
if wrong, cost: payments module + ledger rewrite, all money tests
blocks: checkout-split-payment/*
```

`blocks` is enforced. Blocked work can't be started. Ignorance now has a price and a signature.

**2. Spec.** Requirements land in EARS syntax, *while `<precondition>`, when `<trigger>`, the
system shall `<response>`*, each with a slug and a fit criterion. Slugs are words, never
numbers. `auth-magic-link/link-expires-in-15-min` survives inserts, deletes and reorders;
`REQ-047` does not.

**3. Cook.** `/blueprint cook <slug>` is TDD with the argument removed. The requirement slug
*is* the test. Red test = a named unmet requirement. Every implementing symbol carries
`@spec <feature>/<slug>`.

**4. Check.** `blueprint check` is grep and set math, not a service:

| gate | fails when |
|---|---|
| coverage | a requirement has no test |
| orphan | a `@spec` tag points at a requirement that doesn't exist |
| budget | a doc grew past its line limit |
| blocked | work started on something `OPEN.md` blocks |
| drift | code moved but the spec didn't |

Your traceability matrix is `rg '@spec' | sort`. No database, no tool, no ceremony.

**5. Ship.** Each requirement becomes a [mint](https://github.com/3li7alaki/mint) unit. Slug,
EARS acceptance, scope and gates map one-to-one. blueprint says what's worth building. mint
says whether it's allowed to be called done.

## Your codebase already exists

Most projects are not blank pages, and a spec tool that assumes otherwise gets abandoned in week
two. So blueprint harvests.

Code is evidence of what a system does. It is never evidence of what it should do. Everything
harvested is therefore `derived`: written into the spec, carrying `file:line` evidence, and
unimplementable until a human confirms it. The spec never lies about what it knows.

```sh
blueprint harvest scope 'src/checkout/**'    # one area, never the whole repo
blueprint map                                 # 11% mapped, 9 open
/blueprint-harvest src/checkout               # the agent reads, you answer
```

Then the confirm pass, ordered by blast radius rather than by file, because people skim when
question twelve feels as trivial as question eleven. Three answers, not two:

| answer | means |
|---|---|
| **confirm** | the code is right, this becomes `stated` |
| **correct** | it should behave differently, and the code now fails a stated requirement |
| **bug** | nobody meant this; the gate goes red until it is fixed |

The last two are why this pays for itself on day one. The moment someone says "no, sessions
should expire in 7 days", you have a tracked slug, a red gate and a place to start, instead of a
vague sense that something is off in auth.

What harvesting actually finds is rarely the requirements. It is the loose ends: a state nothing
transitions into, error handling on four of five sibling routes, a permission enforced in the UI
and nowhere on the server, a test asserting behaviour the code no longer has. Those are worth
more than a hundred requirements describing code that was already fine.

You never spec the whole codebase. You harvest the area you are about to touch, and `map` gives
you a number that goes up.

### Unless you already have specs

Then harvesting is the wrong move. It re-derives what someone already wrote and leaves a second
spec that looks as official as the real one while being weaker.

```sh
/blueprint-adopt docs/specs/01-multitenancy.md
```

Acceptance criteria become `stated` requirements, because a person wrote them. Blocked items and
unmitigated risks become priced `OPEN.md` entries. Rationale and rejected options become decision
records. Everything carries `evidence:` back to the source line and its original id, so the
conversion can be checked rather than trusted.

Then the original stops being normative: deleted, or reduced to narrative that says so plainly.
Old citations in the code get rewritten in the same pass, not kept alive by a mapping. Two live
specs drift, and the drift is invisible precisely because both look authoritative.

## Why agents actually obey it

Because it isn't written as a request. Models follow roughly 150 to 200 instructions before
quality degrades, and config files get selectively ignored past ~80 lines. A polite
`CLAUDE.md` is a wish list.

blueprint spends 15 lines of prose and puts the rest in hooks:

| layer | mechanism |
|---|---|
| prose | 15 lines in `AGENTS.md`, the loop and not a rulebook |
| session | injects live status: open blockers, uncovered requirements |
| pre-write | specs are immutable outside `/blueprint amend`; code follows spec, never the reverse |
| stop | `blueprint check` fails → "done" is refused |
| review | adversarial pass from a different vendor, feeding mint's independence clause |

The prose is for the compliant case. The hooks are for every other case.

## Install

The CLI, a single static binary with no runtime dependencies:

```bash
curl -fsSL https://raw.githubusercontent.com/3li7alaki/blueprint/main/install.sh | sh
```

The Claude Code plugin, for the slash commands and skills:

```
/plugin marketplace add 3li7alaki/blueprint
/plugin install blueprint@blueprint
```

Then, in any repo:

```bash
blueprint init --hooks   # AGENTS.md, CLAUDE.md, blueprint/, and the enforcement hooks
/blueprint-grill         # greenfield: the interrogation
/blueprint-harvest src/  # brownfield: read what exists
blueprint check          # the gates
```

## Depth

`quick` · `standard` · `paranoid`. Set per project, override per feature.

Ship the spec at 80% and park the rest in `OPEN.md`. A spec that exists beats a spec that's
perfect. The 20% is still written down, still priced, still blocking. That's the whole trick.

## What blueprint is not

Not a task tracker. Not a build tool. Not a completion authority; that's mint's job, and
blueprint never re-implements it. It owns exactly one thing: making sure the thing you're
about to build has been fully described first.

See [AGENTS.md](AGENTS.md) for the operating contract and
[docs/principles.md](docs/principles.md) for the boundaries.

Prior art worth reading: [EARS](https://alistairmavin.com/ears/) ·
[Volere](https://www.volere.org/templates/volere-requirements-specification-template/) ·
[Spec Kit](https://github.com/github/spec-kit) · [AGENTS.md](https://agents.md)
