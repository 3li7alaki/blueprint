# Open question format v1

## Layout and identity

One question occupies `blueprint/open/<slug>.md`. Its id is `<slug>`, such as
`rounding-rule`. The slug matches `^[a-z0-9]+(-[a-z0-9]+)*$`. Exactly one path
segment is required below the open root. The id is not repeated inside the file.

Path validation is lexical and does not access files or resolve symbolic links.
The cleaned path must remain under the supplied open root and end with `.md`.
This format replaces entries in a shared `blueprint/OPEN.md` file. Migration and
command integration are outside the format contract.

## Encoding and header

Files use UTF-8. Normalize CRLF to LF before parsing. A leading UTF-8 byte order
mark produces `byte-order-mark` at line zero and stops parsing. Invalid UTF-8
produces `malformed-line` at line zero and stops parsing. Otherwise header line
numbers start at one.

The header contains one `key: value` per line, with lowercase keys and exactly
one space after the colon. Keys may appear in any order. No indentation,
comments, repeated keys or unknown keys are allowed. Empty values retain the
separator space and are checked by the relevant value rule.

The first empty line ends the header. Everything after that separator is the
body, kept verbatim after newline normalization, including trailing newlines.
The body receives no content validation. A trailing newline alone does not add
body content. There is no header continuation syntax.

## Keys

| Key | Required | Value |
|---|---|---|
| `question` | yes | Nonempty question text ending with `?`. |
| `cost` | yes | Nonempty text describing the cost of answering incorrectly. |
| `blocks` | yes | Comma-separated requirement ids or `<id-prefix>/*` entries. |
| `status` | yes | `open`, `deferred` or `resolved`. |
| `asked` | yes | Date parsed by Go `time.Parse` with layout `2006-01-02` (`YYYY-MM-DD`). |
| `owner` | no | Nonempty text when present. |
| `resolution` | only for `resolved` | Nonempty decision name, such as `decisions/rounding-half-even`. Forbidden for every other status. |

Values are preserved except block entries, which lose surrounding whitespace.
Blocks retain their order and duplicates. A requirement id contains at least two
slash-separated slug segments, as defined in [requirement-v1.md](requirement-v1.md).
A prefix contains one or more slash-separated slug segments followed by `/*`.
For example, `billing/*` and `billing/checkout/*` select every requirement under
their respective prefixes. No other wildcard syntax is valid.

Question and cost text cannot be empty or whitespace-only. A question with
trailing whitespace after `?` is invalid. Owner and resolution values cannot
be empty or whitespace-only when required. Resolution names receive no further
syntax or file-existence validation. The parser does not resolve block targets
or decide whether work can start. Open and deferred questions remain unresolved;
resolved questions record their decision.

## Validation rules

Rules are stable identifiers. Missing keys use line zero. Other value problems
name their header line. Duplicate keys name the later occurrence and retain the
first value. Independent problems may be reported together. Encoding errors stop
parsing, so a BOM with a valid id yields only its own problem. Invalid fixtures
isolate one problem each. Callers treat any problem as fatal.

A `resolution` key with a missing `status` reports both `missing-status` and
`resolution-without-resolved`. The missing status uses line zero; the resolution
problem names the resolution line.

### Header rules

1. `byte-order-mark`: the file starts with UTF-8 BOM bytes `EF BB BF`.
2. `duplicate-key`: a known header key occurs more than once.
3. `unknown-key`: a syntactically valid header key is not listed above.
4. `malformed-line`: a header line lacks the exact header syntax, or the file has invalid UTF-8.

### Required key rules

5. `missing-question`: `question` is absent or empty after trimming whitespace.
6. `missing-cost`: `cost` is absent or empty after trimming whitespace.
7. `missing-blocks`: the required `blocks` key is absent.
8. `missing-status`: the required `status` key is absent.
9. `missing-asked`: the required `asked` key is absent.

### Value rules

10. `bad-question`: a supplied nonempty question does not end with `?`.
11. `bad-blocks`: a supplied blocks list contains an empty entry, invalid requirement id or invalid prefix wildcard.
12. `bad-open-status`: a supplied status is not `open`, `deferred` or `resolved`, including an empty value.
13. `bad-asked`: a supplied date cannot be parsed with layout `2006-01-02`, including an empty value or impossible calendar date.
14. `empty-owner`: a supplied owner is empty after trimming whitespace.

### Resolution and identity rules

15. `resolved-without-resolution`: status is `resolved` and resolution is absent or empty after trimming whitespace. The problem names the status line.
16. `resolution-without-resolved`: a resolution key is present with any status other than `resolved`, even if its value is empty or the status key is missing.
17. `bad-id`: a supplied id or question path violates the identity rules. The problem uses line zero.

## Examples

`blueprint/open/retention-window.md`:

```text
question: How long must invoice drafts remain available?
cost: A wrong retention window requires data restoration or early deletion.
blocks: billing/*, review/loads-proposal
status: deferred
asked: 2026-11-02
owner: billing-team

The policy needs an answer before either blocked area starts.
```

`blueprint/open/rounding-rule.md`:

```text
question: Which rounding rule applies to invoice totals?
cost: Incorrect totals require ledger repairs.
blocks: billing/checkout/rounds-half-even
status: resolved
asked: 2026-11-02
resolution: decisions/rounding-half-even

The decision records the approved rounding policy and its rationale.
```

## Changelog

Version 1 defines the initial keys, validation rules and object identity.
