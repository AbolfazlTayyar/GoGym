# GoGym — ER Diagram v1

Derived from the entity list in [spec.md](spec.md). Two chains hanging off `Coach`:

- `Coach → Athlete → AthleteMeasurement`
- `Athlete → Plan → Day → Block → BlockMovement → Movement`

`Movement` is a coach-owned library, not per-athlete, so `BlockMovement` is the join between a `Block` and the shared `Movement` catalog.

```mermaid
erDiagram
    COACH ||--o{ ATHLETE : manages
    COACH ||--o{ MOVEMENT : owns
    ATHLETE ||--o{ ATHLETE_MEASUREMENT : logs
    ATHLETE ||--o{ PLAN : has
    PLAN ||--o{ DAY : contains
    DAY ||--o{ BLOCK : contains
    BLOCK ||--o{ BLOCK_MOVEMENT : contains
    MOVEMENT ||--o{ BLOCK_MOVEMENT : "used in"

    COACH {
        uuid id PK
        text first_name
        text last_name
        text phone UK
        text password_hash
        timestamptz created_at
        timestamptz updated_at
    }

    ATHLETE {
        uuid id PK
        uuid coach_id FK
        text first_name
        text last_name
        text phone
        text experience_level "nullable"
        text injuries
        text goal
        numeric height
        timestamptz created_at
        timestamptz updated_at
    }

    ATHLETE_MEASUREMENT {
        uuid id PK
        uuid athlete_id FK
        date date
        numeric weight
        numeric chest
        numeric waist
        numeric arm
        numeric thigh
        numeric hip
        timestamptz created_at
        timestamptz updated_at
    }

    PLAN {
        uuid id PK
        uuid athlete_id FK
        date start_date
        text title
        text note
        timestamptz created_at
        timestamptz updated_at
    }

    DAY {
        uuid id PK
        uuid plan_id FK
        text label
        int order_index
        timestamptz created_at
    }

    BLOCK {
        uuid id PK
        uuid day_id FK
        int order_index
        int sets
        int rest_seconds
        text notes
        timestamptz created_at
    }

    BLOCK_MOVEMENT {
        uuid id PK
        uuid block_id FK
        uuid movement_id FK
        int reps
        int duration_seconds
        int order_in_block
        timestamptz created_at
    }

    MOVEMENT {
        uuid id PK
        uuid coach_id FK "nullable — NULL = system-seeded"
        text name
        text category
        text description
        timestamptz created_at
        timestamptz updated_at
    }
```

## Notes / decisions made while modeling

- **coach_id placement:** per the multi-coach-readiness note in the spec, `coach_id` lives on `Athlete` and `Movement` directly (matching the original entity definitions). `Plan`, `Day`, `Block`, and `BlockMovement` reach the coach transitively through `Athlete`/`Plan`, so they are not denormalized with a redundant `coach_id` column — revisit if query patterns need it.
- **Movement is coach-scoped, not athlete-scoped:** `BlockMovement` is the many-to-many join between `Block` and the shared `Movement` library, carrying the per-use fields (`reps`, `duration_seconds`, `order_in_block`).
- **`Movement.coach_id` is nullable:** `NULL` marks a universal, system-seeded movement (visible to every coach, read-only — coaches cannot edit or delete these); a non-null value is a coach's own custom movement, which they fully own. A coach's effective library is `coach_id = :coach_id OR coach_id IS NULL`.
- **Timestamps:** every table gets `created_at`. `updated_at` is added only to tables edited in place after creation (`Coach`, `Athlete`, `AthleteMeasurement`, `Plan`, `Movement`); `Day`, `Block`, and `BlockMovement` are typically deleted and recreated by the plan builder rather than edited, so they carry `created_at` only.
- **`Athlete.experience_level` is nullable:** a coach may add an athlete before assessing their experience level.
