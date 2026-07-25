# traceability

<!-- budget: 150 lines. Over budget means split the feature.
     Written by `blueprint spec new`. Add to it with `blueprint req add`, never by hand. -->

status: drafting
depth: standard

## Intent

<!-- One paragraph. What changes for a user when this exists. -->

## Surfaces

<!-- Every page, screen, endpoint or job this feature adds or touches. A surface is only
     specified when all five states are written down, so each entry looks like:

     ### checkout-review
     - who: buyer
     - data: reads cart and saved cards, writes an order
     - empty: cart is empty, so show the catalogue link
     - loading: skeleton rows for up to 400ms, then a spinner
     - error: keep the form filled, show what failed, offer retry
     - denied: 404, a signed-out visitor must not learn the order exists

     Blank values are a shape failure. If you do not know one, that is an OPEN.md entry. -->

## Requirements

<!-- One `###` heading per requirement, added by `blueprint req add`. The heading IS the slug:
     words, never numbers, immutable once implemented, referenced from code and tests as
     `@spec <feature-slug>/<requirement-slug>`. Each entry is exactly three lines:

     ### link-expires-in-15-min
     `stated`
     WHILE a magic link is unused, WHEN it is older than 15 minutes, THE system SHALL reject it.
     fit: a request with a 16-minute-old link returns 410 and no session is created.

     `stated` means a human said it. `derived` means you inferred it, and it stays
     unimplementable until a human runs `blueprint req confirm`. -->


### slugs-are-words
`stated`
WHEN a requirement is created, THE system SHALL identify it by a word slug and never by a number.
fit: req add rejects a slug matching ^[0-9]+$ and accepts link-expires-in-15-min


### coverage-is-literal
`stated`
WHEN a requirement has no test file containing its qualified slug, THE coverage gate SHALL fail and name it.
fit: a spec with one requirement and no test makes check --gate coverage exit 1 naming that slug


### gates-see-untracked
`stated`
WHEN a file is new and not yet committed, THE gates SHALL still scan it.
fit: an untracked file carrying an unknown @spec tag makes the orphan gate fail

## Edges

<!-- Adversarial sweep. Every edge that changes behaviour becomes a requirement above with its
     own slug. Every edge nobody can answer becomes an OPEN.md entry with its cost. Nothing
     stays here unresolved. -->

| edge | answer |
|---|---|

## Out of scope

<!-- What this feature explicitly does NOT do, so nobody helpfully adds it. -->

## Depends on

<!-- Other feature slugs, decision slugs. -->
