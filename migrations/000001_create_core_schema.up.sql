-- Core schema per docs/er-diagram.md: two ownership chains rooted at coach.

CREATE TABLE coach (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name    TEXT NOT NULL,
    last_name     TEXT NOT NULL,
    phone         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE movement (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coach_id    UUID REFERENCES coach(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    category    TEXT,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON COLUMN movement.coach_id IS 'NULL = system-seeded, universal movement visible to every coach; non-null = a coach''s own custom movement.';

CREATE INDEX idx_movement_coach_id ON movement (coach_id);

CREATE TABLE athlete (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coach_id          UUID NOT NULL REFERENCES coach(id) ON DELETE CASCADE,
    first_name        TEXT NOT NULL,
    last_name         TEXT NOT NULL,
    phone             TEXT NOT NULL,
    experience_level  TEXT,
    injuries          TEXT,
    goal              TEXT,
    height            NUMERIC,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_athlete_coach_id ON athlete (coach_id);

CREATE TABLE athlete_measurement (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    athlete_id  UUID NOT NULL REFERENCES athlete(id) ON DELETE CASCADE,
    date        DATE NOT NULL,
    weight      NUMERIC,
    chest       NUMERIC,
    waist       NUMERIC,
    arm         NUMERIC,
    thigh       NUMERIC,
    hip         NUMERIC,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_athlete_measurement_athlete_id ON athlete_measurement (athlete_id);

CREATE TABLE plan (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    athlete_id  UUID NOT NULL REFERENCES athlete(id) ON DELETE CASCADE,
    start_date  DATE NOT NULL,
    title       TEXT NOT NULL,
    note        TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_plan_athlete_id ON plan (athlete_id);

CREATE TABLE day (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id      UUID NOT NULL REFERENCES plan(id) ON DELETE CASCADE,
    label        TEXT NOT NULL,
    order_index  INT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_day_plan_id ON day (plan_id);

CREATE TABLE block (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    day_id        UUID NOT NULL REFERENCES day(id) ON DELETE CASCADE,
    order_index   INT NOT NULL,
    sets          INT NOT NULL,
    rest_seconds  INT,
    notes         TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_block_day_id ON block (day_id);

CREATE TABLE block_movement (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    block_id          UUID NOT NULL REFERENCES block(id) ON DELETE CASCADE,
    movement_id       UUID NOT NULL REFERENCES movement(id),
    reps              INT,
    duration_seconds  INT,
    order_in_block    INT NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_block_movement_block_id ON block_movement (block_id);
CREATE INDEX idx_block_movement_movement_id ON block_movement (movement_id);
