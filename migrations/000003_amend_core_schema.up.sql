-- Amends the schema created in 000001_create_core_schema:
-- 1. block_movement.load: prescribed load, free text (not numeric kg).
--    Coaches prescribe load in varied notations that don't reduce to a single
--    unit — kg, "%1RM", RPE, or "bodyweight" — so a numeric-only column would
--    force lossy conversion at entry time. Revisit once plan-builder UX or
--    reporting needs a structured/numeric value.
-- 2. movement.media_url: link to demo video/image. Adding the column does
--    NOT commit us to building an upload flow in v1; it implies object
--    storage (S3/MinIO/local) will be needed once that flow is built.
-- 3. Soft delete on the coach-owned, UI-deletable tables (athlete, movement)
--    via GORM's deleted_at convention, so GORM's default scope excludes
--    deleted rows automatically.

ALTER TABLE block_movement
    ADD COLUMN load TEXT;

ALTER TABLE movement
    ADD COLUMN media_url TEXT;

ALTER TABLE athlete
    ADD COLUMN deleted_at TIMESTAMPTZ;

ALTER TABLE movement
    ADD COLUMN deleted_at TIMESTAMPTZ;

CREATE INDEX idx_athlete_deleted_at ON athlete (deleted_at);
CREATE INDEX idx_movement_deleted_at ON movement (deleted_at);
