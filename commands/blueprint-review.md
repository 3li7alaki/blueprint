---
description: Review a diff against the spec, adversarially, from a different vendor
argument-hint: "[base ref, defaults to the working diff]"
---

Review against the spec, not against intent. Intent is not evidence.

$ARGUMENTS

Read `blueprint/REVIEW.md` for this repo's lenses and run them in that order. Stop at the first
failure and report it. Do not accumulate a wishlist.

Take the diff, then for each touched requirement use `blueprint req show` and
`blueprint trace <feature>/<slug>`. Never read the spec file; the reviewer inheriting the
builder's framing is how a bad diff gets waved through.

The lens that earns its keep is scope creep: anything in this diff that no requirement asked
for. Agents are relentlessly helpful and helpful is a defect here. An unrequested improvement is
untested, unspecified, unowned code that someone maintains forever.

Independence matters more than thoroughness. A review by the same model that wrote the code is
not a review. Dispatch this to a different vendor. Per the routing rules that means codex or
opencode when the code came from Claude, and a Claude agent when it came from codex.

Record the outcome so it counts:

```sh
mint exec record-review <unit> <attempt> <lens> passed|failed \
  --executor <e> --vendor <v> --model <m> --locality <l> --execution-ref <ref>
```

Verdict is one line: `pass`, or `fail: <lens>: <what>`. No praise, no style opinions, no
"consider".
