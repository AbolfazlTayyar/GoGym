# 0003 — Go + Gin + GORM + PostgreSQL

- Status: accepted
- Date: 2026-09-14
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

One option had to be picked per concern (router, database, DB access) before scaffolding, and then kept.

## Decision Drivers

- Learn backend Go with a mainstream, well-documented toolset.
- Relational data with deep parent/child chains (athlete → plan → day → block → movement).
- Move fast on CRUD without writing every query by hand.

## Considered Options

- Router: `net/http` (Go 1.22+ routing), `chi`, **Gin**
- Database: **PostgreSQL**
- DB access: `sqlc`, **GORM**

## Decision Outcome

Chosen: **Gin, PostgreSQL, GORM**, with UUID primary keys (`gen_random_uuid()`) and zerolog for
structured JSON request logs to stdout. GORM is opened with `TranslateError: true` so repositories
match portable sentinels (`gorm.ErrDuplicatedKey`) instead of Postgres error codes.

### Consequences

- Good: fast CRUD, soft-delete scopes for free (see [0008](0008-soft-delete.md)), large ecosystem.
- Bad: GORM hides SQL. N+1 queries and unexpected statements have to be caught by reading the SQL log.
- Bad: GORM can migrate the schema itself, which is turned off on purpose (see [0004](0004-sql-migrations-as-schema-source.md)).
