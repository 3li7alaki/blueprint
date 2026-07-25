---
name: adopt
description: Convert an existing spec, PRD, requirements doc or acceptance-criteria list into blueprint's format. Use when a repo already has written specs and blueprint is arriving, when blueprint/spec is empty but docs/ holds real requirements, or when asked to migrate, adopt, import or convert existing specs. Produces stated requirements with evidence, not derived ones.
---

# adopt

Some repos already have specs, and better ones than a first grill would produce. When that is
true, harvesting requirements out of the code is wrong twice: it re-derives what a human already
wrote down, and it leaves a second spec that looks as authoritative as the real one while being
strictly weaker.

Adoption is a translation. A person wrote these requirements, so they are `stated`. You are
changing their form, not their status.

Read the `blueprint` skill first. Every rule there still applies.

## Before touching anything

Read the whole source document. All of it, not the acceptance criteria section. The rationale,
the rejected options and the risks are where the reasons live, and reasons are the part that
gets lost in a migration and never comes back.

Then read the code that implements it, enough to know which criteria are real and which were
aspirational. A criterion nobody built is still a requirement; it is just uncovered.

## The mapping

| in the source document | becomes | why |
|---|---|---|
| an acceptance criterion | `req add --confidence stated` | a human wrote it |
| a blocked or unanswered item | `open add`, priced | it is still unanswered |
| rationale, research, a rejected option | `decide` | this is why, and why survives badly |
| an unmitigated risk | `open add` | a risk is a question with a cost |
| a mitigated risk | the `--because` of the decision that mitigated it | it explains that decision |
| in scope, out of scope | the spec's `Out of scope` section | already has a home |
| a numbered id such as AC-4 | a word slug, cited in `evidence:` | numbers go stale, words do not |

Every adopted requirement carries `--evidence "<source file>:<line> (<original id>)"`. A reader
must be able to check your conversion against the original rather than trust it.

## Converting a criterion

Source criteria are usually prose and rarely EARS. Rewrite the grammar, never the meaning.

```
AC-2. Products are vendor-owned. A product created by/for a merchant is linked to exactly
      one vendor. Querying products as merchant A never returns merchant B's products.
      (Tested with two seeded merchants.)
```

becomes

```sh
blueprint req add multitenancy/products-are-vendor-owned \
  --ears "WHEN merchant A queries products, THE system SHALL never return merchant B's products." \
  --fit "two seeded merchants, a cross-read returns empty" \
  --confidence stated \
  --evidence "docs/specs/01-multitenancy.md:29 (AC-2)"
```

Notice what happened to the parenthetical. "Tested with two seeded merchants" was the fit
criterion all along, sitting in prose. Most good specs already contain their fit criteria; your
job is to find them, not invent them.

One criterion often holds two requirements. "A product is linked to exactly one vendor" and
"merchant A never sees merchant B's products" are different claims that fail differently, so
they get different slugs. Split freely. Never merge: a merged requirement cannot be traced to
one test.

When a criterion has no checkable outcome and you cannot find one in the surrounding prose, do
not invent it. That is an `open add`, and the cost line is that it can never be verified.

## Rewriting citations in code

Code that cites the old ids gets rewritten in the same pass. A comment reading
`// The three-actor wall (E2, AC-1 / AC-8)` becomes two `@spec` tags naming the requirements
those criteria became.

Do not build a compatibility mapping that keeps old citations counting. That is a second
citation scheme that still works, which is the same redundancy as a second spec, one layer
down. Anything the rewrite misses shows up as an uncovered requirement, which is a visible
gap rather than a silent wrong link.

## Then the source document stops being normative

This is the part that decides whether adoption worked.

Delete it, or reduce it to a pointer at `blueprint/spec/<feature>.md`. If narrative is worth
keeping, keep it as narrative that states plainly it is not the spec.

What you must not do is leave two documents that both look official. They will drift, and the
drift is invisible precisely because each one looks like the source of truth. If the team is
not ready to delete the original, that hesitation is itself an open question worth recording,
because a half-finished migration is worse than either end state.

## Report

Say plainly: criteria adopted, criteria that became more than one requirement, criteria you
could not give a fit line, open questions recorded, and citations rewritten. Then name what you
deleted or what still needs deleting.

Never soften a criterion to make it fit EARS. If the original is vague, the requirement is
vague, and that vagueness is a finding worth raising rather than smoothing over.
