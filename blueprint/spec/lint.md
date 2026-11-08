# lint

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


### ears-has-shall-and-full-stop
`stated`
WHEN req add, req correct or amend receives an EARS line that lacks the response keyword or a final full stop, THE system SHALL refuse the write and name the rule ears-shape.
fit: req add with the EARS line `THE system accepts cards.` exits 1, prints `ears-shape`, and leaves the spec file byte for byte unchanged
evidence: lint plan approved by Ali, 2026-11-08


### one-shall-per-requirement
`stated`
WHEN req add, req correct or amend receives an EARS line that repeats the response keyword, THE system SHALL refuse the write and name the rule one-shall.
fit: req add with an EARS line holding 2 SHALL clauses exits 1 and prints `one-shall`
evidence: lint plan approved by Ali, 2026-11-08


### no-vague-terms
`stated`
WHEN a write receives an EARS or fit line containing a whole word or phrase from the vague term list, THE system SHALL refuse the write and name the rule vague-term with the term found.
fit: req add with the fit `responds fast` exits 1 and prints `vague-term: fast`; the fit `breakfast is served at 8` is accepted
evidence: lint plan approved by Ali, 2026-11-08


### fit-is-observable
`stated`
WHEN req add receives a fit line with no number, no quoted or backticked literal and no path, THE system SHALL refuse the write and name the rule fit-not-observable.
fit: req add with the fit `returns success` exits 1 and prints `fit-not-observable`; the fit `returns 201` is accepted
evidence: lint plan approved by Ali, 2026-11-08


### read-accepts-what-write-accepts
`stated`
WHEN a spec file holds an EARS line that passes ears-shape, THE parser SHALL read that requirement without error.
fit: a spec holding `IF a card is declined, THEN THE system SHALL show the decline code.` and one holding `THE system SHALL log every refund.` both parse, and `blueprint check --gate shape` exits 0
evidence: go fixes approved by Ali, 2026-11-08

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
