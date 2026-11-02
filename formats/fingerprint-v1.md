# Fingerprint format v1

A fingerprint identifies a term's definition and states, or a requirement's
meaning including its glossary terms and semantic dependencies. SHA-256 hashes
the canonical UTF-8 bytes. The result is `sha256:` followed by 64 lowercase
hexadecimal characters.

Inputs are successfully parsed [terms](term-v1.md) and
[requirements](requirement-v1.md), with unique requirement ids. Use the exact
parsed bytes: CRLF has already become LF, and list entries have already lost
surrounding whitespace. Apply no Unicode normalization or whitespace collapsing.
For example, `a  b` and `a b` produce different fingerprints.

## Term canonical bytes

The following notation is normative. Each `\n` denotes one LF byte, not two
literal characters. Replace angle-bracket placeholders with their values.
Every line, including the last, ends with exactly one LF byte.

```text
blueprint-term-fingerprint-v1\n
definition\n
<definition>\n
states\n
<states joined by "," with no spaces, in file order; empty line if none>\n
```

State order and duplicates are significant. Term id, `status` and body are
excluded. Moving a term from `planned` to `existing` does not change its meaning.

## Requirement canonical bytes

The same notation applies. `contract` is followed by the literal version `1`.

```text
blueprint-requirement-fingerprint-v1\n
contract\n
1\n
ears\n
<ears>\n
fit\n
<fit kind>: <fit criterion>\n
reference\n
<reference sha256 hex, or empty line when none>\n
term <id> <term fingerprint>\n
needs <id> <requirement fingerprint>\n
```

The `term` line repeats once per distinct term referenced in EARS or the fit
criterion, sorted by id bytewise. The `needs` line repeats once per distinct
needs target, sorted by id bytewise. Omit these lines when their lists are empty.
Each embedded fingerprint includes its `sha256:` prefix. The reference digest
contains only the 64 hexadecimal characters, without that prefix.

The requirement's own id, reference path, `after`, `scope`, `tier`, `confidence`,
`evidence`, `superseded-by` and body are excluded. Sorting applies to the appended
term and needs lines only; the EARS and fit text remain byte for byte as parsed.

## Resolution and problems

`TermFingerprint` hashes one parsed term. `Fingerprints` resolves referenced
terms from the supplied map and recursively resolves only `needs` edges from
the supplied requirement set. Completed requirements, including failed ones,
are memoized. Shared dependencies are computed once. Scheduling edges never
participate, even when they are unknown or form a cycle.

`unknown-term` reports an absent glossary term. `unknown-dependency` reports an
absent needs target. `dependency-cycle` reports a needs traversal returning to
an active requirement, including a self edge. Problems use line zero and name
the offending ids. Traversal sorts root ids, term ids and needs targets bytewise
and deduplicates repeated references, making results independent of input order.
Each cycle diagnostic lists its active path and repeats its first id at the end.

A requirement with a problem has no map entry. Neither does any requirement
that transitively needs it. Unaffected requirements still receive fingerprints.
Callers must treat problems as fatal before using the result as a complete set.
This operation does not replace `CheckGraph`, which validates scheduling edges
as well as semantic dependencies under [requirement v1](requirement-v1.md).

## Judged reference verification

`VerifyReference(repoRoot, r)` checks only judged requirements. Other fit kinds
return no problems and perform no filesystem access. A reference path must be
relative to the repository root. Convert the root to an absolute path, clean
the joined path and check containment with path components, not a string prefix.
An absolute reference or a cleaned path outside the root produces
`reference-outside-root` before attempting to read it.

Resolve both root and reference with `filepath.EvalSymlinks`, then check
containment again. A resolved reference outside the resolved root produces
`reference-outside-root`. An unavailable root or reference, a dangling symlink,
or any resolve/read failure produces `reference-missing`. A directory cannot
serve as a reference file. Symlinks that resolve inside the root are allowed,
including when the supplied root is itself a symlink.

Hash the reference's raw file bytes without newline or Unicode normalization.
A digest different from the parsed reference digest produces
`reference-digest-mismatch`. Reference problems use line zero and identify the
requirement. Callers treat any problem as fatal. These checks assume the
filesystem does not change between resolution and reading.

## Test vectors

[fingerprint.json](vectors/fingerprint.json) is an array. Each vector names its
target, includes source text with ids for parsing, the exact canonical string
and its expected fingerprint. A `term` object selects a term vector. Otherwise,
`requirements`, `terms` and `target` supply a requirement graph and its target id.
Optional `same_as` and `different_from` name other vectors for comparison.

The Go tests reproduce every canonical string and digest from parsed inputs.
From the repository root, `sh formats/vectors/check.sh` independently hashes
each canonical string using `printf '%s'`, `jq` and `shasum -a 256`. It preserves
the final LF and exits 1 naming each vector whose digest differs.
