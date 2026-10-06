# Architecture Decision Records

One file per decision, [MADR](https://adr.github.io/madr/)-style. Process rules are in
[0001](0001-record-architecture-decisions.md). Copy any existing ADR as the template.

| # | Decision | Date | Status |
|---|---|---|---|
| [0001](0001-record-architecture-decisions.md) | Record architecture decisions | 2026-10-03 | accepted |
| [0002](0002-modular-monolith.md) | Modular monolith with layered modules | 2026-09-14 | accepted |
| [0003](0003-go-gin-gorm-postgres-stack.md) | Go + Gin + GORM + PostgreSQL | 2026-09-14 | accepted |
| [0004](0004-sql-migrations-as-schema-source.md) | Hand-written SQL migrations are the only schema source | 2026-09-16 | accepted |
| [0005](0005-integration-tests-on-real-postgres.md) | Unit tests plus integration tests on a real Postgres | 2026-09-16 | accepted |
| [0006](0006-ci-fail-fast-pinned-tools.md) | Fail-fast CI with pinned tool versions | 2026-09-17 | accepted |
| [0007](0007-swagger-dev-only-root-basepath.md) | swaggo docs, dev-only UI, root basePath | 2026-09-17 | accepted |
| [0008](0008-soft-delete.md) | Soft delete for coach-deletable tables | 2026-09-19 | accepted |
| [0009](0009-free-text-load.md) | Prescribed load is free text | 2026-09-19 | accepted |
| [0010](0010-coach-jwt-auth.md) | Coach-only JWT auth with in-process rate limiting | 2026-09-20 | accepted |
| [0011](0011-tenant-isolation-single-helper.md) | Tenant isolation through one scoping helper | 2026-09-20 | accepted |
| [0012](0012-no-organization-layer.md) | No organization layer above coach | 2026-09-20 | accepted |
| [0013](0013-response-envelope-explicit-helpers.md) | Response envelope written by explicit helpers | 2026-09-23 | accepted |
| [0014](0014-athlete-list-offset-pagination.md) | Offset pagination and private-by-default athlete list | 2026-10-02 | accepted |
| [0015](0015-child-ownership-via-parent-lookup.md) | Child tables prove ownership through the parent lookup | 2026-10-03 | accepted |
| [0016](0016-current-plan-latest-started.md) | The current plan is the latest one that has started | 2026-10-06 | accepted |
| [0017](0017-plan-day-cap-in-schema.md) | A plan has at most 7 days, enforced by the schema | 2026-10-06 | accepted |
