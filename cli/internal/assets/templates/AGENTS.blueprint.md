## Blueprint

Specs in `blueprint/` are source of truth. Code is derivative.

Before code:
1. Read `blueprint/spec/<feature>.md`. No spec → stop, run `/blueprint grill <feature>`.
2. Check `blueprint/OPEN.md`. An entry blocking this feature → stop and report. Never invent
   the answer.
3. Follow `blueprint/CONVENTIONS.md` and copy the exemplar file it names.

While coding:
- Tag implementing symbols `@spec <feature>/<requirement-slug>`. Tag the covering test too.
- One test per requirement slug. Test first.
- Build only what a requirement asks. Unrequested work fails review.

Before done:
- `blueprint check` passes.
- Never edit a spec to match code. Spec changes go through `/blueprint amend`.

Never write an em dash or an en dash, anywhere, in any file.
