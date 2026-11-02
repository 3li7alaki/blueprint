# Object file formats

Each object occupies one file. Each file has its own identity, so unrelated edits
stay separate. These version 1 formats share UTF-8 encoding, BOM rejection,
CRLF normalization and header syntax.

| Kind | Format | Path | Id |
|---|---|---|---|
| Requirement | [requirement-v1.md](requirement-v1.md) | `blueprint/spec/<segment>/.../<slug>.md` | At least two slash-separated slug segments. |
| Glossary term | [term-v1.md](term-v1.md) | `blueprint/terms/<namespace>/<term>.md` | `<namespace>:<term>`. |
| Open question | [open-v1.md](open-v1.md) | `blueprint/open/<slug>.md` | One slug. |

Adding an optional key is a minor change, recorded in that format document's
changelog. Clarifying the diagnostic for an already invalid file is also a minor
change. Removing, renaming or
retyping a key, or adding a required key, requires a new major version with new
format file names. Existing major format documents retain their contract.

Fixtures live under `fixtures/valid/` and `fixtures/invalid/`. Valid names are
`<kind>--<name>.md`, with prefixes `requirement--`, `term--` and `open--`.
Invalid names are `<kind>--<rule-slug>.md`; each breaks exactly its named rule.
Requirement fixtures use the id `fixture/<name>`, term fixtures use
`fixture:<name>` and open question fixtures use `<name>`. Graph and path rules
have tests rather than file fixtures. Object parsers validate syntax only;
resolving term references, block targets and decision names is outside these
formats. Fingerprints are outside version 1.

The requirement BOM fixture lives at
`fixtures/invalid/encoding/requirement--byte-order-mark.md` and has a dedicated
test. This extends encoding coverage while preserving the original requirement
fixture inventory and its unchanged tests.
