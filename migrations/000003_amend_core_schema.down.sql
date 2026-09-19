DROP INDEX IF EXISTS idx_movement_deleted_at;
DROP INDEX IF EXISTS idx_athlete_deleted_at;

ALTER TABLE movement
    DROP COLUMN deleted_at;

ALTER TABLE athlete
    DROP COLUMN deleted_at;

ALTER TABLE movement
    DROP COLUMN media_url;

ALTER TABLE block_movement
    DROP COLUMN load;
