# tags-live-in-comments
status: accepted
superseded-by: 
date: 2026-07-26

## Context
Answers the open question fixture-tags-vs-real-tags. blueprint's own tests must contain fake @spec tags as fixture data to exercise the scanner, and the orphan gate read all six as claims about requirements that do not exist. A path carve-out for test files was the obvious fix and the wrong one: users' tests are exactly where real tags must live, since coverage is measured there.

## Decision
A tag counts only when its line begins, after whitespace, with a comment marker. Fixture strings therefore hold inert text, and no path is ever exempt from a gate.

## Because
Any path a gate cannot see is a place untraceable code can hide. One rule, language agnostic, still greppable by hand. The cost is that a tag trailing live code on the same line no longer counts, which is acceptable: a tag deserves its own line.

## Consequences

