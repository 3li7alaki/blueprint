---
description: Convert existing specs into blueprint format, one document at a time
argument-hint: "[path to a spec doc or a directory of them]"
---

Load the `adopt` skill and follow it exactly.

$ARGUMENTS

With no argument, look for existing specs (`docs/spec*`, `docs/planning`, `PRD*`, `REQUIREMENTS*`)
and report what you found with a rough criteria count. Then ask which document to take first. One
at a time, never the whole directory in one pass: a bulk conversion is unreviewable, and an
unreviewed spec is exactly the failure blueprint exists to prevent.

Before converting, check whether adoption is even the right move:

- The document holds acceptance criteria a human wrote. Adopt it.
- The document is a design note, an architecture sketch or research. That is a decision record,
  not a spec. Use `blueprint decide`.
- The document describes what the code does, written after the fact. That is not stated, it is
  derived. Use `/blueprint-harvest` instead.

Then work through the skill's mapping: criteria to requirements, blocked items and unmitigated
risks to `OPEN.md`, rationale and rejected options to decision records. Every requirement
`stated`, every requirement carrying `--evidence` back to the source line and its original id.

Rewrite the code citations in the same pass. No compatibility mapping.

Finish by naming what should now be deleted. If the team is not ready to delete it, record that
hesitation with `blueprint open add`, because two live specs drift and the drift is invisible.
