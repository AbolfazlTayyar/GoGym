# 0005 — Unit tests plus integration tests on a real Postgres

- Status: accepted
- Date: 2026-09-16
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

Most of the risk sits at the query layer: tenant scoping, soft-delete scopes, ordering, and
constraint errors. Mocks of `*gorm.DB` would only test that the code calls the mock the way the
mock expects.

## Decision Drivers

- Test the paths that would corrupt or leak athlete/plan data, not chase a coverage number.
- A fast loop for service logic that doesn't need Docker.

## Considered Options

1. Unit tests with a mocked DB only
2. SQLite in-memory for integration tests
3. testify unit tests + `testcontainers-go` Postgres with real migrations applied

## Decision Outcome

Chosen option: **3**. `make test-unit` runs with `-short` and needs no Docker. `make test` starts
a throwaway Postgres, applies `migrations/`, and runs everything. `internal/testutil` hands
modules a ready `*gorm.DB`.

### Consequences

- Good: tests run against the same SQL, constraints and error translation as prod.
- Bad: the full suite needs Docker and is slower, so CI splits it into a separate job (see [0006](0006-ci-fail-fast-pinned-tools.md)).
