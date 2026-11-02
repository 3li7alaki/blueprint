# Object file formats

Each object occupies one file. Requirements live at
`blueprint/spec/<segment>/.../<slug>.md`, with at least two path segments below
`blueprint/spec/`. Each file has its own identity, so unrelated edits stay separate.
Other object formats are outside this version.

`requirement-v1.md` defines the first requirement format. Adding an optional key is
a minor change, recorded in that document's changelog. Removing, renaming or
retyping a key, or adding a required key, requires a new major version with new
format file names. Existing major format documents retain their contract.

Fixtures live under `fixtures/valid/` and `fixtures/invalid/`. Valid names are
`requirement--<name>.md`. Invalid names are `requirement--<rule-slug>.md`; each
breaks exactly its named rule. Fixtures use the id `fixture/<name>`. Graph and
path rules have tests rather than file fixtures.
