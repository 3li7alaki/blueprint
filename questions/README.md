# The question bank

The grill is data, not improvisation. A model asked to "interview the founder thoroughly" will
freestyle, skip, and forget. A model handed this bank cannot.

One file per pass, loaded on demand so a grill session only carries the pass it is running.

## Entry shape

```toml
[[q]]
slug   = "delete-cascade"          # stable id; becomes the OPEN.md entry slug if unanswered
depth  = "standard"                # minimum depth at which this is asked
ask    = "When a {entity} is deleted, what happens to its {children}?"
exit   = "named one of: cascade | orphan | block | soft-delete"
cost   = "data model and every query; discovered late it is a migration under load"
blocks = ["*"]                     # what stays unstartable while this is unanswered
```

- `{braces}` are filled from the nouns pass. A question about an entity is asked once per entity.
- `exit` is the stop condition. Restate the question until the answer clears it. An answer that
  does not clear `exit` is not an answer.
- `cost` is copied verbatim into `OPEN.md` when the human defers. That line is the price tag.
- `blocks` is enforced by `blueprint check`. `"*"` means the whole project.

## Depth

| depth | asks | for |
|---|---|---|
| `quick` | `depth = quick` | prototypes, throwaway, internal tools |
| `standard` | `quick` + `standard` | anything real, default |
| `paranoid` | everything | money, health, auth, regulated, multi-tenant |

Depth filters questions. It never skips a pass.

## Rules for the asker

1. Batch small. Three to five questions, never a wall.
2. Never answer your own question. "You decide" produces a decision record with rationale,
   written down, not a silent choice.
3. Never accept an answer that fails `exit`. Restate it more concretely instead.
4. "I don't know" is a valid, complete answer. Write the `OPEN.md` entry, quote the cost, move on.
5. Never invent an answer into a spec. Ever. This is the one unforgivable failure.
