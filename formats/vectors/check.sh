#!/bin/sh
set -eu

vectors=$(dirname "$0")/fingerprint.json
count=$(jq 'length' "$vectors")
# An empty vector list must fail, or the check passes vacuously.
[ "$count" -gt 0 ] || { echo "no vectors in $vectors" >&2; exit 1; }
index=0
failed=0
while [ "$index" -lt "$count" ]; do
    name=$(jq -r --argjson i "$index" '.[$i].name' "$vectors")
    expected=$(jq -r --argjson i "$index" '.[$i].fingerprint' "$vectors")
    # The sentinel keeps command substitution from stripping canonical trailing LFs.
    canonical=$(jq -j --argjson i "$index" '.[$i].canonical' "$vectors"; printf '.')
    actual=$(printf '%s' "${canonical%.}" | shasum -a 256)
    actual=sha256:${actual%% *}
    if [ "$actual" != "$expected" ]; then
        printf '%s: fingerprint mismatch: got %s, want %s\n' "$name" "$actual" "$expected" >&2
        failed=1
    fi
    index=$((index + 1))
done
exit "$failed"
