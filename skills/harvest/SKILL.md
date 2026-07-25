---
name: harvest
description: Extract a spec from a codebase that already exists. Use when blueprint is being introduced to a brownfield repo, when blueprint/spec is empty but the code is not, or when asked to map, document, reverse engineer, or catch loose ends in existing code. Produces derived requirements with evidence plus OPEN entries, never stated requirements.
---

# harvest

Code is evidence of what a system does. It is never evidence of what it should do.

So everything you produce here is `derived`, carries `evidence`, and is unimplementable until a
human confirms it. You are not writing a spec. You are building a finite, ordered list of things
to ask about, out of a codebase nobody can hold in their head.

Read the `blueprint` skill first. Every rule there still applies.

## Scope, before anything

```sh
blueprint harvest scope 'src/checkout/**'
blueprint inventory src/checkout
blueprint map
```

Harvest one area, the one about to be touched. Never the whole repo. A repo-wide pass produces
hundreds of unconfirmed requirements nobody reads, which is worse than no spec because it looks
like one. `map` gives you the percentage that makes the work finite.

## What to read, in this order

1. **Tests.** The strongest signal in any repo. A test is someone in the past caring enough to
   encode a fit criterion. Its name is close to a requirement and its assertion is close to a
   `fit:` line. Harvest these first and lean on them hardest.
2. **Schema, migrations, models.** Entities and their lifecycle. An enum column is a state
   machine somebody already wrote down.
3. **Routes, endpoints, screens.** Surfaces. Record the five states only where the code shows
   them; a missing empty state is a question, not an omission to invent.
4. **Guards and middleware.** The permission matrix, as actually enforced rather than as
   intended.
5. **Config and env.** Third parties, limits, and the flags that change behaviour at runtime.

Git history is for archaeology: it tells you when and sometimes who, rarely why. Treat what it
suggests as a question, never as an answer.

## Writing what you find

```sh
blueprint req add checkout/rejects-expired-card \
  --ears "WHEN a card's expiry is in the past, THE system SHALL refuse the charge." \
  --fit "a charge with a past expiry returns 402 and creates no order" \
  --confidence derived \
  --evidence "src/checkout/charge.ts:88, tests/charge.test.ts:41"
```

Evidence is not decoration. A human confirming a claim without a pointer is voting, not
reviewing. Cite `file:line` for every derived requirement.

Tag the code you harvested with `@spec <feature>/<slug>` as you go. That is what moves `map`
off zero and what the `unmapped` gate measures.

## The loose ends are the product

Documenting what already exists is low value; anyone can read the code. The value is what the
reading exposes:

- a state in an enum that nothing transitions into
- error handling on four of five sibling routes
- two code paths answering the same question differently
- a permission enforced in the interface and nowhere on the server
- a test asserting behaviour the code no longer has
- a symbol nothing calls

Each of these is a question from `questions/8-brownfield.toml`, and each one you surface is
worth more than ten requirements describing code that was already fine. Ask them with the
evidence attached: "your code appears to say X, confirm or correct" is far easier to answer than
a blank page.

## What code can never tell you

Do not infer these. Ever. They are `blueprint open add` entries:

- whether the current behaviour is correct, or a bug nobody noticed
- what was deliberately excluded, since non-goals leave no trace in source
- why a decision was made

## Then hand it back to a human

```sh
blueprint ask --confirm --path src/checkout --batch 5
```

Ordered by blast radius, not by file. For each one the human answers exactly one of:

| answer | command | means |
|---|---|---|
| confirm | `blueprint req confirm <f>/<slug>` | the code is right, this is now stated |
| correct | `blueprint req correct <f>/<slug> --ears "..." --reason "..."` | it should behave differently, and now the code fails a stated requirement |
| bug | `blueprint req bug <f>/<slug> --reason "..."` | nobody meant this; the gate goes red until it is fixed |

`correct` and `bug` are the point. They turn "something feels off in checkout" into a tracked
slug with a failing gate and a place to start.

Never answer on their behalf. A derived requirement you confirm yourself is an invented
requirement with extra steps.
