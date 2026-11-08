# amend-lint-ears-has-shall-and-full-stop
status: accepted
superseded-by: 
date: 2026-11-08

## Context
Old EARS: WHEN req add, req correct or amend receives an EARS line without SHALL or without a final full stop, THE system SHALL refuse the write and name the rule ears-shape.

## Decision
New EARS: WHEN req add, req correct or amend receives an EARS line that lacks the response keyword or a final full stop, THE system SHALL refuse the write and name the rule ears-shape.

## Because
The first wording used the response keyword twice and broke one-shall, the rule next to it. Meaning unchanged.

## Consequences

