---
description: Install blueprint into this repo: AGENTS.md block, CLAUDE.md, blueprint/, hooks
---

Set this repo up for spec-anchored development.

1. Run `blueprint init`. It writes `blueprint/` from the templates, appends the blueprint block
   to `AGENTS.md` (creating it if absent), and writes `CLAUDE.md` as a single `@AGENTS.md` line.
   It never overwrites an existing spec.

2. Fill the parts only a human can answer, by asking rather than assuming:
   - the gate command (`AGENTS.md`): the one command that runs this repo's tests
   - `blueprint/CONVENTIONS.md`: for each recurring thing, the path to the exemplar file that
     already does it right. If no exemplar exists, the convention does not exist yet. Leave it.
   - `blueprint/PROJECT.md`: only if the project is already understood. Otherwise leave it for
     the grill, which is what the grill is for.

3. Report what is now enforced and what is still empty. Do not fill an empty section with a
   plausible guess; that is the failure this tool exists to prevent.

Then suggest `/blueprint-grill` if `blueprint/PROJECT.md` is still a template.
