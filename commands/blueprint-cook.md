---
description: Implement one requirement, test first, traced by slug
argument-hint: "[feature/requirement-slug, or empty for the next one]"
---

Implement exactly one requirement. Load the `blueprint` skill first.

$ARGUMENTS

1. Resolve the target. With no argument, `blueprint req next`. Read it with
   `blueprint req show <feature>/<slug>`. Do not open the spec file.

2. Refuse to start if any of these are true. Report and stop:
   - the requirement is `derived` and unconfirmed
   - `blueprint open list --blocking <feature>` returns anything
   - the fit line is missing

3. Write the failing test first, carrying `@spec <feature>/<slug>` in the file. The test asserts
   the fit criterion, not your interpretation of the EARS line. Run it. It must fail for the
   stated reason, not because of a typo or a missing import.

4. Write the smallest code that passes it. Tag the implementing symbol with the same
   `@spec <feature>/<slug>`. Follow the exemplar named in `blueprint/CONVENTIONS.md`.

5. Build nothing else. Not a helper you might want, not a related edge case, not a nicer
   abstraction. If you notice a real gap, `blueprint open add` it and keep going. Unrequested
   work fails review even when it is good work.

6. `blueprint check` before you report. Then say plainly: the slug, the test that covers it,
   and anything you noticed and deliberately did not do.
