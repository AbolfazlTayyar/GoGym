# 0004 — Hand-written SQL migrations are the only schema source

- Status: accepted
- Date: 2026-09-16
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

The schema can be defined by GORM structs (`AutoMigrate`) or by versioned SQL files. Defining it
in both places lets the two drift apart without anyone noticing.

## Decision Drivers

- Reviewable, reversible schema changes, run the same way in dev, CI and prod.
- One place to read the real schema.

## Considered Options

1. GORM `AutoMigrate` from structs
2. `golang-migrate` with sequential SQL up/down pairs; structs only mirror the SQL
3. `goose`

## Decision Outcome

Chosen option: **2**. Migrations live in `migrations/`, `AutoMigrate` is never called, and GORM
tags must match the SQL exactly. `docs/er-diagram.md` is maintained by hand and updated in the
same commit as each migration.

### Consequences

- Good: every schema change is explicit SQL with a tested down path.
- Bad: each column is written twice (SQL + struct tag) plus once in the ER diagram, so a mismatch
  only shows up at runtime or in an integration test.
