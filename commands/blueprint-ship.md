---
description: Turn a requirement into a mint unit and clear the completion floor
argument-hint: "[feature/requirement-slug]"
---

Hand one requirement to mint. blueprint says what is worth building; mint says whether it is
allowed to be called done. Do not re-implement either side.

$ARGUMENTS

1. `blueprint mint <feature>/<slug> --scope <globs>` prints the unit command. It does not run
   it. Name every code path this touches in `--scope`, read the printed command, then run it.

   The spec file is added to `--scope` deliberately: editing the spec afterwards invalidates the
   receipt. It is also `--protected`, so mint flags an edit to it during an attempt.

2. Initialise an attempt with honest provenance. Executor, vendor, model and locality are
   attributable claims, so state what actually ran, not what you wish had run.

3. Run the declared gates with `mint verify`, then record the independent reviews. Same vendor
   as the maker does not count as independent, and safety-tier work needs a different vendor
   outright.

4. `mint done` with the acceptance verdict. Treat only its receipt as proof. A green
   `blueprint check` means the code is traceable, which is not the same claim.

If the floor rejects it, that is evidence to sharpen the unit or to run the check that is
missing. It is never a reason to weaken a verifier.
