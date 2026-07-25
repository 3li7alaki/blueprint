# Open questions

Every unresolved question lives here, priced. Nothing here is ever answered by inference.

An entry with `status: OPEN` or `status: DEFERRED` blocks everything its `blocks:` line matches.
Blocked work is not startable. Report it, do not route around it.

Resolving an entry means a human answered it. Move the answer into a spec or a decision record,
then run `blueprint open resolve <slug>`. This file only holds live unknowns.

Written by `blueprint open add`. One key per line, no inline separators.

<!-- template, copy per question

## <question-slug>
status: OPEN
pass: frame
asked: 2026-07-25
owner: founder
question: <one line, verbatim as asked>
cost: <what has to be rebuilt when the guess turns out wrong>
blocks: <feature-slug>/*, <feature-slug>/<requirement-slug>

-->

## fixture-tags-vs-real-tags
status: OPEN
pass: gates
asked: 2026-07-26
owner: 3li7alaki
question: How should a scanner tell a real @spec tag from one inside test fixture data, in a repo whose tests must contain fake tags to exercise the scanner?
cost: either blueprint cannot pass its own orphan gate, or a real carve-out lets users hide untraceable code behind a test path
blocks: traceability/*
