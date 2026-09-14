# GoGym — ER Diagram v1

Derived from the entity list in [mvp-spec.md](mvp-spec.md). Two chains hanging off `Coach`:

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
    }

    ATHLETE {
        uuid id PK
        uuid coach_id FK
        text first_name
        text last_name
        text phone
        text experience_level
        text injuries
        text goal
        numeric height
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
    }

    PLAN {
        uuid id PK
        uuid athlete_id FK
        date start_date
        text title
        text note
    }

    DAY {
        uuid id PK
        uuid plan_id FK
        text label
        int order_index
    }

    BLOCK {
        uuid id PK
        uuid day_id FK
        int order_index
        int sets
        int rest_seconds
        text notes
    }

    BLOCK_MOVEMENT {
        uuid id PK
        uuid block_id FK
        uuid movement_id FK
        int reps
        int duration_seconds
        int order_in_block
    }

    MOVEMENT {
        uuid id PK
        uuid coach_id FK "nullable — NULL = system-seeded"
        text name
        text category
        text description
    }
```

## Notes / decisions made while modeling

- **coach_id placement:** per the multi-coach-readiness note in the spec, `coach_id` lives on `Athlete` and `Movement` directly (matching the original entity definitions). `Plan`, `Day`, `Block`, and `BlockMovement` reach the coach transitively through `Athlete`/`Plan`, so they are not denormalized with a redundant `coach_id` column — revisit if query patterns need it.
- **Movement is coach-scoped, not athlete-scoped:** `BlockMovement` is the many-to-many join between `Block` and the shared `Movement` library, carrying the per-use fields (`reps`, `duration_seconds`, `order_in_block`).
- **`Movement.coach_id` is nullable:** `NULL` marks a universal, system-seeded movement (visible to every coach, read-only — coaches cannot edit or delete these); a non-null value is a coach's own custom movement, which they fully own. A coach's effective library is `coach_id = :coach_id OR coach_id IS NULL`.
