---
description: Settle how the product looks and sounds, and price whatever nobody decided. Writes PRODUCT.md, decision records and OPEN entries.
argument-hint: "[--depth quick|standard|paranoid]"
---

Run the look pass. Load the `look` skill first and follow it exactly.

$ARGUMENTS

This pass runs only in a repo with surfaces. If no spec declares one, say so and stop; a CLI has
no register.

Two halves, in order:

1. **Ask.** If the `impeccable` skill is installed, run `/impeccable init` and let it interview.
   It asks better than the bank does. Otherwise use `blueprint ask look --depth <depth> --batch 4`.
   Either way `register` and `platform` are settled first, and the brand only questions are
   dropped once the answer is `product`.
2. **Record what the interview cannot.** Every shrug becomes `blueprint open add` with the cost
   line from `questions/9-look.toml` and a `blocks` glob. Every visual choice becomes
   `blueprint decide` with its reason. This is the half that makes the pass worth running: an
   interview ends when the talking stops, and only a priced entry makes silence expensive.

The three rules:

1. Taste is not yours to supply. "You decide" produces a decision record with your reasoning,
   visible and reversible, and you say out loud that you made it.
2. An adjective that fits the competitor is not an answer. Push until it does not.
3. Nothing here is written into a spec, and `DESIGN.md` is not written here at all. It is
   generated from the code, once code exists.

Stop when the pass is exhausted or they are done. Then report: what is captured, what is open with
its price, and whether `blueprint check --gate look` passes.
