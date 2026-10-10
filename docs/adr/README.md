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
| [0018](0018-workout-logs-snapshot-prescription.md) | Workout logs snapshot the prescription, not the plan rows | 2026-10-07 | accepted |
| [0019](0019-unusable-body-reference-is-a-field-error.md) | An unusable id in the request body is a field error, not a 404 | 2026-10-07 | accepted |
| [0020](0020-movement-filter-vocabularies.md) | Movement muscle group and equipment are closed vocabularies | 2026-10-08 | accepted |
| [0021](0021-universal-rows-forbidden.md) | Changing a universal row is a 403; another coach's row stays a 404 | 2026-10-08 | accepted |
| [0022](0022-delete-movement-in-use.md) | Deleting a movement a plan uses is allowed, as a soft delete | 2026-10-08 | accepted |
| [0023](0023-migrations-in-compose-service.md) | Migrations run in a one-shot Compose service before the API starts | 2026-10-08 | accepted |
| [0024](0024-trusted-proxies-explicit.md) | Forwarded client IPs are trusted only from configured proxies | 2026-10-10 | accepted |
