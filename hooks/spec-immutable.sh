#!/bin/sh
# PreToolUse on Edit|Write|MultiEdit.
#
# Code follows spec. A spec edited to match the code it produced is the implementation
# grading its own homework, which is the exact failure blueprint exists to prevent.
#
# The only legal mutation path is `blueprint amend`, which records the previous text and a
# reason. This hook blocks every other write into blueprint/spec/ and blueprint/decisions/.
# `/blueprint amend` sets BLUEPRINT_AMEND=1 for its own edits.
set -eu

[ "${BLUEPRINT_AMEND:-0}" = "1" ] && exit 0

PATH_IN=$(head -c 8192 | sed -n 's/.*"file_path"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')
[ -n "$PATH_IN" ] || exit 0

case "$PATH_IN" in
  */blueprint/spec/*)
    cat >&2 <<'MSG'
BLOCKED: blueprint/spec/ is not directly editable.

A spec is the acceptance criteria for the code you are writing. Editing it to match your
implementation destroys the only independent record of what was asked for.

  new requirement      blueprint req add <feature>/<slug> --ears "..." --fit "..."
  change an existing   blueprint amend <feature>/<slug> --ears "..." --reason "..."
  cannot answer it     blueprint open add <slug> --question "..." --cost "..." --blocks "..."
MSG
    exit 2 ;;
  */blueprint/decisions/*)
    cat >&2 <<'MSG'
BLOCKED: decision records are immutable.

A decision that changes is a new decision. Write it, then link them:

  blueprint decide <new-slug> --context "..." --decision "..." --because "..."
  blueprint supersede <old-slug> <new-slug>
MSG
    exit 2 ;;
esac

exit 0
