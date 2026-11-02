# Glossary term format v1

## Layout and identity

One term occupies `blueprint/terms/<namespace>/<term>.md`. Its id is
`<namespace>:<term>`, such as `money:amount`. Both parts match
`^[a-z0-9]+(-[a-z0-9]+)*$`. Exactly two path segments are required below the
terms root. The id is not repeated inside the file.

Path validation is lexical and does not access files or resolve symbolic links.
The cleaned path must remain under the supplied terms root and end with `.md`.
Requirements refer to terms as `[[namespace:term]]`.

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
| `definition` | yes | Nonempty text defining the term in one or more sentences. Sentence grammar is not validated. |
| `status` | yes | `existing` when the entity exists in code, or `planned` when it is not built yet. |
| `states` | no | Comma-separated lifecycle state slugs, such as `draft, active, closed`. Omit rather than supply an empty list. |

Values are preserved except list entries, which lose surrounding whitespace.
States retain their order and duplicates. The parser does not inspect code or
verify lifecycle transitions. A definition cannot be empty or whitespace-only.

## Validation rules

Rules are stable identifiers. Missing keys use line zero. Other value problems
name their header line. Duplicate keys name the later occurrence and retain the
first value. Independent problems may be reported together. Encoding errors stop
parsing, so a BOM with a valid id yields only its own problem. Invalid fixtures
isolate one problem each.

### Header rules

1. `byte-order-mark`: the file starts with UTF-8 BOM bytes `EF BB BF`.
2. `duplicate-key`: a known header key occurs more than once.
3. `unknown-key`: a syntactically valid header key is not listed above.
4. `malformed-line`: a header line lacks the exact header syntax, or the file has invalid UTF-8.

### Value rules

1. `missing-definition`: `definition` is absent or empty after trimming whitespace.
2. `missing-status`: the required `status` key is absent.
3. `bad-term-status`: a supplied status is neither `existing` nor `planned`, including an empty value.
4. `bad-states`: a supplied states list contains an empty entry or an entry outside the slug grammar.

### Identity rule

1. `bad-id`: a supplied id or term path violates the identity rules. The problem uses line zero.

## Examples

`blueprint/terms/billing/invoice.md`:

```text
definition: An invoice records the amount owed for a completed purchase.
status: existing
states: draft, issued, paid

Issuing an invoice fixes its amount and currency.
```

`blueprint/terms/review/proposal.md`:

```text
definition: A proposal is a change awaiting review. It records the reason for the change.
status: planned

Reviewers need its rationale before deciding whether to accept it.
```

## Changelog

Version 1 defines the initial keys, validation rules and object identity.
