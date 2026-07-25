#!/bin/sh
# Stop. "Done" is a claim, and a claim needs evidence.
#
# This blocks the end of a turn while a gate is red. It is the cheapest possible version of
# mint's floor: mint proves a unit is complete, this one only proves nothing is obviously
# untraceable, uncovered or blocked. Both, or neither, is fine. This is not a substitute.
set -eu

command -v blueprint >/dev/null 2>&1 || exit 0
ROOT=$(git rev-parse --show-toplevel 2>/dev/null) || exit 0
[ -d "$ROOT/blueprint" ] || exit 0

# A stop-hook that blocks its own re-entry would loop forever.
GATE="/tmp/blueprint-done-gate-${CLAUDE_CODE_SESSION_ID:-default}"
if [ -f "$GATE" ] && [ "$(find "$GATE" -mmin -1 2>/dev/null)" ]; then
  exit 0
fi

OUT=$(blueprint check 2>&1) && exit 0

touch "$GATE"
printf 'BLOCKED: blueprint gates are failing, so this is not done.\n\n%s\n\n' "$OUT" >&2
printf 'Fix the gate, or record why it cannot be fixed:\n' >&2
printf '  blueprint open add <slug> --question "..." --cost "..." --blocks "..."\n' >&2
exit 2
