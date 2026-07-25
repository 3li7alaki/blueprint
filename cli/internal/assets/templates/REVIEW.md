# Review lenses

<!-- budget: 60 lines.

     This file never explains how to build. It exists so the reviewer does not inherit the
     builder's assumptions. If a line here also appears in AGENTS.md or CONVENTIONS.md,
     delete it here and reference the slug instead. -->

Review a diff against the spec, not against intent. Intent is not evidence.

## Order

Run these in order. Stop at the first failure and report it. Do not accumulate a wishlist.

1. **Trace.** Every changed non-trivial file carries a `@spec <feature>/<slug>` that exists in
   `blueprint/spec/`. Untagged code is orphan code.
2. **Coverage.** Every requirement slug touched by this diff appears in a test file. A test
   that cannot fail is not coverage.
3. **Fit.** For each touched requirement, read its `fit:` line and check the diff actually
   satisfies *that observation*, not something adjacent that feels equivalent.
4. **Scope creep.** Anything in the diff that no requirement asked for. This is the lens that
   catches agents being helpful. Helpful is a defect.
5. **Edges.** Walk the feature's `## Edges` table. Each row that this diff can reach must be
   handled or explicitly out of scope.
6. **Conventions.** Compare against the exemplar file named in `CONVENTIONS.md`, not against
   taste.
7. **Spec integrity.** The diff must not modify `blueprint/spec/**`. A spec change riding along
   with an implementation is the code rewriting its own acceptance criteria.

## Verdict

One of: `pass`, `fail: <lens>: <one line>`. No praise, no style opinions, no "consider".

## Independence

A review by the same executor and vendor that wrote the code is not a review. Run a different
vendor. Record it with `mint exec record-review`; the floor already enforces the rest.
