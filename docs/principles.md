# blueprint principles

## One purpose

blueprint removes ambiguity before it becomes code. Every surviving feature must either extract
an answer from a human, record that no answer exists and price it, or make an existing answer
mechanically checkable.

If a feature does not do one of those three, it does not belong here.

## Ownership

Drivers own workflow: projects, tickets, terminals, agent selection, worktrees, branches,
commits, PRs, retries, deployment.

mint owns completion: evidence, provenance, independence, the floor, snapshots, receipts,
freshness.

blueprint owns definition: the question bank, specs, requirement slugs, the unknowns register,
conventions, review lenses, repo-side traceability gates, and the stated half of the design
contract in `PRODUCT.md`.

Design tools own execution: the screens, and the `DESIGN.md` they derive from the code they wrote.
blueprint records the choice and its reason. The tool records the result. One writer per file.

Three layers, no overlap. blueprint never issues a receipt. mint never asks a founder a question.

## Invariants

1. A requirement is a slug, an EARS statement, and a fit criterion. Missing any one, it is not
   a requirement yet.
2. Slugs are words. Numbers carry no meaning and go stale on insert.
3. A slug is immutable once implemented. Renaming is superseding.
4. Confidence is explicit. `derived` requirements are not implementable until confirmed by a
   human. The model may propose; it may not decide and then forget it decided.
5. Unknowns are recorded, priced, and blocking. Never resolved by inference.
6. Code follows spec. A spec is never edited to match code outside an explicit amend.
7. Every doc has one reader-moment and one line budget. A fact lives in exactly one file; other
   files reference it by slug and never restate it.
8. Enforcement is mechanical. Prose is the weakest layer and is budgeted accordingly.
9. The gates are grep and set math. If a check needs a database, it is the wrong check.
10. A spec at 80% with the rest priced in `OPEN.md` beats a spec at 100% that never shipped.

## Why the budgets exist

Frontier models follow roughly 150 to 200 instructions with consistency. A coding harness spends
about 50 of those before your first line. Config files get selectively ignored past ~80 lines.
Every line of guidance competes with every other line.

So budgets are not tidiness. A doc over budget is not verbose. It is inert, and worse, it makes
its neighbours inert too. Split it or delete it.

## The failure this exists to prevent

A founder holds a detailed product in their head, describes 30% of it, and gets a confident
implementation of a plausible-but-different product. Nobody notices until it is expensive.

The model is not at fault for that and no amount of model capability fixes it. The missing 70%
was never said out loud. blueprint's job is to make someone say it, on the record, before the
first commit, or to write down, with a price attached, that they refused to.
