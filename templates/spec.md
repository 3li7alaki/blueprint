# <feature-slug>

<!-- budget: 150 lines. Over budget means split the feature. -->

status: `drafting` | `ready` | `building` | `shipped`
depth: `quick` | `standard` | `paranoid`

## Intent

<!-- One paragraph. What changes for a user when this exists. -->

## Surfaces

<!-- Every page, screen, endpoint or job this feature adds or touches.
     Each one needs all five states or it is not specified. -->

### `<surface-slug>`
- who: `<role-slug>`, `<role-slug>`
- data: what it reads, what it writes
- empty:
- loading:
- error:
- denied:

## Requirements

<!-- One `###` heading per requirement. The heading IS the slug.
     Slug is words, immutable once implemented, referenced from code and tests as
     `@spec <feature-slug>/<requirement-slug>`. -->

### `<requirement-slug>`
`stated` | `derived`
WHILE <precondition>, WHEN <trigger>, THE system SHALL <response>.
fit: <mechanical pass/fail, an observation and not an opinion>

## Edges

<!-- Adversarial sweep. Each answered edge that changes behaviour becomes a requirement above;
     each unanswered one becomes an OPEN.md entry. Nothing stays here unresolved. -->

| edge | answer |
|---|---|
| two actors at once | |
| empty / first run | |
| very large input | |
| parent deleted | |
| third party down | |
| hostile actor | |
| timezone / DST | |

## Out of scope

<!-- What this feature explicitly does NOT do, so nobody helpfully adds it. -->

## Depends on

<!-- Other feature slugs, decision slugs. -->
