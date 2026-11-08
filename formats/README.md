# Object file formats

Each object occupies one file. Each file has its own identity, so unrelated edits
stay separate. These version 0 formats share UTF-8 encoding, BOM rejection,
CRLF normalization and header syntax.

**Status: draft (v0).** blueprint does not write or read these files yet; its specs still use one file per feature. v0 may change incompatibly. v1 is cut, and frozen under the versioning rules below, when blueprint's own commands first write and read these files.

| Kind | Format | Path | Id |
|---|---|---|---|
| Requirement | [requirement-v0.md](requirement-v0.md) | `blueprint/spec/<segment>/.../<slug>.md` | At least two slash-separated slug segments. |
| Glossary term | [term-v0.md](term-v0.md) | `blueprint/terms/<namespace>/<term>.md` | `<namespace>:<term>`. |
| Open question | [open-v0.md](open-v0.md) | `blueprint/open/<slug>.md` | One slug. |

Adding an optional key is a minor change, recorded in that format document's
changelog. Clarifying the diagnostic for an already invalid file is also a minor
change. Removing, renaming or
retyping a key, or adding a required key, requires a new major version with new
format file names. Existing major format documents retain their contract.
The v0 line may break freely; these rules apply from v1 on.

Fixtures live under `fixtures/valid/` and `fixtures/invalid/`. Valid names are
`<kind>--<name>.md`, with prefixes `requirement--`, `term--` and `open--`.
Invalid names are `<kind>--<rule-slug>.md`; each breaks exactly its named rule.
Requirement fixtures use the id `fixture/<name>`, term fixtures use
`fixture:<name>` and open question fixtures use `<name>`. Graph and path rules
have tests rather than file fixtures. Object parsers validate syntax only;
resolving block targets and decision names is outside these formats.

[Fingerprint v0](fingerprint-v0.md) defines canonical bytes for terms and
requirements, semantic dependency resolution and judged reference verification.
Its [vectors](vectors/fingerprint.json) contain parsed inputs, canonical strings
and expected digests. Run `sh formats/vectors/check.sh` from the repo root for
an independent digest check using `jq` and `shasum -a 256`.
