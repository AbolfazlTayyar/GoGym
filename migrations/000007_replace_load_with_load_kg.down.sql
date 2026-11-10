ALTER TABLE block_movement
    DROP COLUMN IF EXISTS load_kg;
ALTER TABLE block_movement
    ADD COLUMN IF NOT EXISTS load TEXT;
