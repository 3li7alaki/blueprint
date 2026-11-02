# Requirement format v1

## Layout and identity

One requirement occupies `blueprint/spec/<segment>/.../<slug>.md`. Its id is the
path below `blueprint/spec/` without `.md`, such as
`billing/checkout/rounds-half-even`. An id has two or more segments joined by `/`.
Each segment matches `^[a-z0-9]+(-[a-z0-9]+)*$`. Identity comes from meaningful
slugs, never numeric counters. The id is not repeated inside the file.

Path validation is lexical and does not access files or resolve symbolic links.
The cleaned path must remain under the supplied spec root and end with `.md`.

## Encoding and header

Files use UTF-8. Normalize CRLF to LF before parsing. Invalid UTF-8 produces
`malformed-line` with line zero. All other header line numbers start at one.
A leading UTF-8 byte order mark produces only `byte-order-mark` at line zero
for an otherwise valid id. Parsing stops before header or value validation.

The header contains one `key: value` per line, with lowercase keys and exactly
one space after the colon. Keys may appear in any order. No indentation,
comments, repeated keys or unknown keys are allowed in the header. Empty values
retain the separator space and are checked by the relevant value rule.

The first empty line ends the header. Everything after that separator is the
body, kept verbatim after newline normalization, including trailing newlines.
The body is free text and receives no content validation. A trailing newline
alone does not add body content. There is no header continuation syntax.

## Keys

| Key | Required | Value |
|---|---|---|
| `ears` | yes | One EARS sentence ending with `.`, containing ` SHALL ` and at least one of `WHEN `, `WHILE `, `IF ` or `WHERE `. |
| `fit` | yes | `<kind>: <criterion>`, with kind `test`, `property` or `judged` and a criterion that is not empty or whitespace. |
| `reference` | only for `judged` | `<path> sha256:<digest>`, with a nonblank path and exactly 64 lowercase hexadecimal digest characters. Forbidden for other fit kinds. |
| `confidence` | yes | `stated` or `derived`. Derived requirements remain unimplementable until confirmed. |
| `evidence` | no | Free text, such as `cli/internal/example.go:12`; may be empty. |
| `scope` | yes | Comma-separated globs; trim surrounding spaces on each entry and reject empty entries. No further glob syntax validation occurs. |
| `needs` | no | Comma-separated requirement ids, each trimmed and valid. These are semantic dependencies. Omit rather than supply an empty list. |
| `after` | no | Comma-separated requirement ids, each trimmed and valid. These impose scheduling order only. Omit rather than supply an empty list. |
| `tier` | no | `chore`, `normal` or `consequential`; omission assigns no default. |
| `superseded-by` | no | Exactly one requirement id. |

Values are preserved except list entries, which lose surrounding whitespace.
The parsed fit separates its kind from its criterion. The parsed reference
separates its path from its digest and drops the `sha256:` prefix. Reference
validation checks syntax only; it does not read or hash the referenced file.

In `ears` and the fit criterion, `[[namespace:term]]` refers to a glossary term.
Both parts use the segment slug grammar. Every `[[` must begin a valid reference.
Extracted terms omit brackets, sort lexically and remove duplicates. References
in evidence, body and other keys receive no term validation or extraction.

## Validation rules

Rules are stable identifiers. Missing fields use line zero. Other value problems
name their header line. Duplicate keys name the later occurrence and retain the
first value. Independent problems may be reported together; invalid fixtures
isolate one problem each.

1. `missing-ears`: the required `ears` key is absent.
2. `missing-fit`: the required `fit` key is absent.
3. `missing-confidence`: the required `confidence` key is absent.
4. `missing-scope`: the required `scope` key is absent.
5. `bad-ears-shape`: the sentence lacks a final `.`, ` SHALL ` or a supported trigger or precondition keyword.
6. `bad-fit-kind`: the fit lacks the `: ` separator or its kind is not `test`, `property` or `judged`.
7. `empty-fit-criterion`: a fit with a `: ` separator has an empty or whitespace-only criterion.
8. `judged-without-reference`: a judged fit has no `reference` key. The problem names the fit line.
9. `reference-without-judged`: a reference appears without a judged fit.
10. `bad-reference`: a reference does not match the required path and digest syntax.
11. `bad-confidence`: confidence is neither `stated` nor `derived`.
12. `bad-tier`: a supplied tier is not `chore`, `normal` or `consequential`.
13. `bad-scope`: at least one scope entry is empty after trimming.
14. `bad-dependency-id`: at least one entry in `needs` or `after` is not a requirement id, including an empty entry.
15. `bad-superseded-by`: the supplied value is not one requirement id.
16. `bad-term-ref`: a `[[` in EARS or the fit criterion does not begin a valid glossary reference.
17. `duplicate-key`: a known header key occurs more than once.
18. `unknown-key`: a syntactically valid header key is not listed above.
19. `malformed-line`: a header line lacks the exact header syntax, or the file has invalid UTF-8.
20. `dependency-cycle`: the combined `needs` and `after` graph contains a directed cycle.
21. `unknown-dependency`: a dependency targets an id absent from the supplied requirement set.
22. `self-dependency`: a dependency targets its own requirement id.
23. `bad-id`: a supplied id or requirement path violates the identity rules.
24. `byte-order-mark`: the file starts with UTF-8 BOM bytes `EF BB BF`.

Graph problems use line zero. Both edge kinds participate equally in validation.
Repeated edges, including the same edge in both lists, count once. Self edges
produce only `self-dependency`, not a cycle. Each directed cycle appears once,
listing ids in edge order starting with the lexically smallest id and repeating
that id at the end. Distinct overlapping cycles are separate problems.
Unknown and self edges are reported by source id, then target id, both sorted.
Cycles follow, sorted by starting id and dependency traversal order. Results do
not depend on requirement input order. Supersession is not a dependency edge.

## Examples

`blueprint/spec/billing/checkout/rounds-half-even.md`:

```text
ears: WHEN checkout computes [[money:amount]], THE system SHALL round half to even.
fit: test: The amounts 2.5 and 3.5 produce 2 and 4 respectively.
confidence: stated
scope: cli/internal/billing/**, cli/internal/checkout/**

Rounding avoids directional bias across repeated calculations.
```

`blueprint/spec/review/presents-rationale.md`:

```text
ears: IF a reviewer opens [[review:proposal]], THE system SHALL present its rationale.
fit: judged: The explanation matches the reference and defines [[review:proposal]].
reference: reference/review.txt sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
confidence: derived
evidence: cli/internal/review/view.go:12
scope: cli/internal/review/**
needs: review/loads-proposal, review/defines-rationale
after: review/selects-proposal
tier: consequential
superseded-by: review/presents-complete-rationale

Observed behavior remains subject to confirmation.
The reference pins the intended explanation for a human judgment.
```

## Changelog

Version 1 defines the initial keys, validation rules and object identity.

Minor change: leading UTF-8 BOM rejection now has its own `byte-order-mark`
diagnostic at line zero instead of cascading header and missing-field problems.
