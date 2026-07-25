---
description: Extract a spec from an existing codebase, one area at a time
argument-hint: "[path or glob, e.g. src/checkout]"
---

Load the `harvest` skill and follow it exactly.

$ARGUMENTS

With no argument, run `blueprint map` and report which areas are unmapped, then ask which one to
take. Do not pick for them, and do not offer to do the whole repo. Scope is the single decision
that determines whether this succeeds or gets abandoned in week two.

Order of work:

1. `blueprint harvest scope '<glob>'`, then `blueprint inventory <path>`.
2. Read tests first, then schema, then routes, then guards, then config.
3. Write `derived` requirements with `--evidence "file:line"`. Tag the code as you go.
4. Record everything the code cannot answer with `blueprint open add`, priced.
5. `blueprint ask --confirm --path <path>` and take the human through it in small batches.

Report at the end: requirements harvested, loose ends found, questions left open, and the
`map` percentage before and after. The loose ends are the valuable part, so lead with those,
not with the count of requirements.

Never confirm a derived requirement yourself. That is an invented requirement with extra steps.
