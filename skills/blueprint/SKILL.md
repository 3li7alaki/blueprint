---
name: blueprint
description: Read and write a repo's spec through the blueprint CLI instead of opening the files. Use when a repo has a blueprint/ directory and you need a requirement, a surface, the open questions, the next thing to build, or the review lenses; and whenever you would otherwise read blueprint/spec/*.md, blueprint/OPEN.md or blueprint/PROJECT.md directly. Also use for running the grill, adding requirements, recording unknowns, and handing a requirement to mint.
---

# blueprint

Never open `blueprint/*.md` with a file reader. Every one of those files is written to be
parsed, and the CLI returns the four lines you need instead of the hundred and fifty you do not.
Reading the file wastes context and, worse, invites you to skim past the constraint that
mattered.

```
blueprint req show auth-magic-link/link-expires-in-15-min     # 4 lines
Read blueprint/spec/auth-magic-link.md                        # 150 lines
```

If `blueprint` is not on PATH, install it: `curl -fsSL
https://raw.githubusercontent.com/3li7alaki/blueprint/main/install.sh | sh`. Do not fall back to
reading the files. A missing binary is a setup problem, not a licence to skim.

## Before writing any code

```sh
blueprint req next                          # what is actually startable right now
blueprint req show <feature>/<slug>         # confidence, EARS line, fit line
blueprint open list --blocking <feature>    # must be empty before you start
```

Rules that are not negotiable:

- A `derived` requirement is not implementable. Ask the human, then `blueprint req confirm`.
- A non-empty `open list --blocking` means stop and report. Never infer the missing answer.
- Tag implementing code and its test with `@spec <feature>/<slug>`. Coverage is a literal
  string match, so the tag is the whole mechanism.

## Reading a spec

Take the section, never the file:

```sh
blueprint spec list
blueprint spec show <feature> --section surfaces
blueprint spec show <feature> --section edges
blueprint req list --feature <feature> --uncovered
blueprint trace <feature>/<slug>            # who already implements this
```

## Running the grill

```sh
blueprint ask <pass> --depth standard --batch 5
```

Passes, in order: `frame`, `boundaries`, `nouns`, `surfaces`, `rules`, `edges`, `gates`. Never
skip one, never reorder. Depth filters questions inside a pass, it never skips a pass.

Ask the returned questions in small batches, in the human's own terms. For each answer:

- clears the question's `exit` condition, and it is a behaviour: `blueprint req add ... --confidence stated`
- clears `exit`, and it is a choice between options: `blueprint decide <slug> --context ... --decision ... --because ...`
- does not clear `exit`: restate the question more concretely. An answer that misses the exit
  condition is not an answer.
- "I don't know" or "you decide": `blueprint open add <slug> --question "..." --cost "<the bank's cost line, verbatim>" --blocks "..."`

The one unforgivable failure is inventing an answer into a spec. If nobody said it, it goes in
`OPEN.md` with its price, or it does not exist.

## Writing

Every mutation goes through the CLI. Editing `blueprint/` by hand is blocked by a hook, and the
block is correct.

```sh
blueprint spec new <feature>
blueprint req add <feature>/<slug> --ears "WHEN ..., THE system SHALL ..." --fit "..." --confidence stated
blueprint amend <feature>/<slug> --ears "..." --reason "..."      # only legal edit
blueprint open add <slug> --question "..." --cost "..." --blocks "<feature>/*"
blueprint open resolve <slug>                                     # only after a human answered
```

EARS grammar, always: `WHILE <precondition>, WHEN <trigger>, THE system SHALL <response>.`
No `shall` without a trigger or a precondition. Every requirement carries a fit criterion that
a machine can check. A requirement without one is a mood.

## Finishing

```sh
blueprint check                       # all gates, exit 1 on any failure
blueprint check --gate coverage
blueprint mint <feature>/<slug>       # prints the mint unit command, does not run it
```

`blueprint check` passing means the code is traceable. It does not mean the work is done. mint
issues that verdict, from independent evidence. Do not claim done on a green `check` alone.

## Never

- Never write an em dash or an en dash, in any file, in any repo. The `dash` gate fails on one.
- Never edit `blueprint/spec/` or `blueprint/decisions/` directly.
- Never build something no requirement asked for. Helpful is a defect.
- Never answer your own question to unblock yourself.
