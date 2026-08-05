# look

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


### product-is-stated
`stated`
WHEN the look pass records an answer, THE system SHALL keep it in PRODUCT.md at the repository root and refuse a hand edit to that file.
fit: look show returns one section, and a Write to PRODUCT.md is blocked by the pre-write hook


### surfaces-need-a-product
`stated`
WHILE any spec declares a surface, WHEN PRODUCT.md is missing or a required section is empty, THE look gate SHALL fail and name the section.
fit: a spec with one surface and no PRODUCT.md exits 1, and the same repo with no surfaces passes


### tokens-stay-home
`stated`
WHILE CONVENTIONS.md declares a token home, WHEN a component file carries a raw colour or font stack, THE tokens gate SHALL fail and name the line.
fit: a component with a hex literal fails, the same line carrying blueprint:allow-raw passes, and a repo with no Tokens section passes

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
