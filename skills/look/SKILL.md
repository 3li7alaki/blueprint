---
name: look
description: Settle how a product looks and sounds before anyone builds a screen, and make the unanswered parts cost something. Use when a repo has surfaces and no PRODUCT.md, when the look gate fails, when asked to define brand, visual identity, design principles, palette, typography or motion, or when starting an interface from scratch. Produces PRODUCT.md, decision records and priced OPEN entries.
---

# look

The seven passes settle what the product does. This one settles what it is like to meet.

It runs only where there are surfaces. A CLI, a worker and a library have no register and no
palette, and grilling them for one is how a tool earns a reputation for ceremony.

Read the `blueprint` skill first. Every rule there still applies, especially the one that matters
most here: never answer your own question. Taste is the easiest thing in the world to supply on
someone else's behalf, and a palette nobody chose is a palette nobody will defend in review.

## Use the best interview available

**If the `impeccable` skill is installed, it runs the questions.** Invoke `/impeccable init` and
follow it. Its interview is more developed than anything this bank holds: it knows the register
split, the platform rulebooks, the saturated defaults to refuse, and how to push past an adjective
that would fit any competitor. Reimplementing that here would be a worse copy that also has to be
maintained.

Without impeccable, ask from the bank instead:

```sh
blueprint ask look --depth standard --batch 4
```

Either way, `register` and `platform` are settled first, because every later question is graded
against them, and the brand only questions are dropped once the answer is `product`.

## Then do the part an interview cannot

An interview ends when the talking stops. What it cannot do is make silence expensive. That is
this pass's whole contribution, and it is not optional:

**Every question that got a shrug becomes a priced entry.** "Whatever you think" and "we'll decide
later" are the same answer, and neither belongs in a document that reads as decided.

```sh
blueprint open add motion-budget \
  --question "Does anything animate on the dashboard, or is stillness the point?" \
  --cost "motion added per component at the end, uniformly, which is the reflex everyone recognises" \
  --blocks "dashboard/*"
```

Take the cost line from the matching question in `questions/9-look.toml`. That file is the price
list even when impeccable asked the questions: the wording may be its, the consequence of guessing
is ours, and `blocks` is what stops work starting on a surface nobody has decided.

**Every visual choice becomes a decision record, with its reason.**

```sh
blueprint decide colour-committed-oxblood \
  --context "Operators triage the queue in a bright room, and the brand is not the SaaS cream default." \
  --decision "Committed strategy: oxblood carries roughly 40 percent of the surface." \
  --because "The scene rules out a dark theme, and one saturated colour is what keeps it off the near-white default."
```

Theme, colour strategy, type direction, motion energy and the token home each get one. A choice
without a written reason gets silently re-decided by the next agent, which is exactly how a design
system dissolves into per-component taste.

## Where things live

| answer | home | written by |
|---|---|---|
| register, platform, users, positioning, personality, anti-references, principles, accessibility | `PRODUCT.md` at the root | the interview, whichever ran it |
| theme, colour, type, motion, token home | `blueprint/decisions/<slug>.md` | `blueprint decide`, here |
| anything refused or unknown | `blueprint/OPEN.md`, priced | `blueprint open add`, here |
| what the interface actually is today | `DESIGN.md` | impeccable, from the code, later |

`blueprint look new` writes `PRODUCT.md` from the template if nothing else has. Once it exists it
is guarded: edits go through `blueprint amend`, so an answer nobody gave cannot appear in it.

`PRODUCT.md` and `DESIGN.md` are the same split as `stated` and `derived`. One is what somebody
said. The other is what the code turned out to be. Never write the second from the first.

## Then hand it to the builders

Once `blueprint check --gate look` passes, interface work has a contract:

- `blueprint look show --section principles` gives a builder four lines, not a document.
- The token home named in `blueprint/CONVENTIONS.md` switches on the `tokens` gate, so colour and
  type stay in one file instead of spreading through components.
- impeccable builds the screens and writes `DESIGN.md` for what it built.

Report at the end: what was captured, what is open with its cost, and which features are blocked
by an open entry.
