#!/bin/sh
# SessionStart. Inject the live state of the spec, so no session begins by guessing.
#
# Kept deliberately small. This is prime context, and a wall of text here is the same
# failure as a bloated AGENTS.md: everything in it gets ignored, including the important line.
set -eu

command -v blueprint >/dev/null 2>&1 || exit 0
ROOT=$(git rev-parse --show-toplevel 2>/dev/null) || exit 0
[ -d "$ROOT/blueprint" ] || exit 0

BLOCKERS=$(blueprint open list --status OPEN 2>/dev/null | head -5)
NEXT=$(blueprint req next 2>/dev/null | head -1)
FAILED=$(blueprint check 2>/dev/null | grep -i '^fail' | head -5) || true

printf 'BLUEPRINT ACTIVE. Specs in blueprint/ are source of truth; code is derivative.\n'

if [ -n "$BLOCKERS" ]; then
  printf '\nOpen questions blocking work (do NOT answer these yourself, report them):\n%s\n' "$BLOCKERS"
fi

if [ -n "$FAILED" ]; then
  printf '\nFailing gates:\n%s\n' "$FAILED"
fi

if [ -n "$NEXT" ]; then
  printf '\nNext implementable requirement: %s\n' "$NEXT"
  printf 'Read it with: blueprint req show %s\n' "$NEXT"
fi

exit 0
