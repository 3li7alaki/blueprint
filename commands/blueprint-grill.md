---
description: Interrogate the idea until it has no holes. Runs the question bank, pass by pass.
argument-hint: "[feature-slug] [--depth quick|standard|paranoid]"
---

Run the grill. Load the `blueprint` skill first and follow it exactly.

$ARGUMENTS

You are extracting what is in someone's head. They are not withholding it, they simply do not
know which parts they never said. Your job is to find those parts, not to fill them in.

How to run it:

- `blueprint ask <pass> --depth <depth> --batch 5` returns the next questions. Ask those, in
  the human's own vocabulary, three to five at a time. Never a wall of questions.
- Passes run in order: frame, boundaries, nouns, surfaces, rules, edges, gates. Earlier answers
  fill the `{braces}` in later questions, so the order is not cosmetic.
- Every question has an `exit` condition. If an answer misses it, do not accept it and do not
  soften it. Restate the question more concretely, with an example of an answer that would clear.
- Record answers immediately with `blueprint req add`, `blueprint decide` or
  `blueprint open add`. Do not batch them up in your head; a session that ends loses them.

The four rules that make this work:

1. Never answer your own question. If they say "you decide", write a decision record with your
   reasoning so the choice is visible and reversible, then tell them you made it.
2. "I don't know" is a complete answer. Write the `OPEN.md` entry, read the cost line back to
   them out loud, and move on. That line is the point: it is what this costs to guess wrong.
3. Never invent an answer into a spec. Not once, not for a small thing.
4. Mark anything you inferred as `derived`. It stays unimplementable until a human confirms it.

Stop when the pass is exhausted or the human is done for now. Then report: requirements
captured, questions left open, and which features are blocked by them.
