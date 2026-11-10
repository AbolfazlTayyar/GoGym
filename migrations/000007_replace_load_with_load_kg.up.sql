-- block_movement.load (free text) becomes load_kg: the weight in kg as a number, so it can be summed and
-- charted. Nothing is carried over: no plan held a load when this ran.
ALTER TABLE block_movement
    DROP COLUMN load;
ALTER TABLE block_movement
    ADD COLUMN load_kg NUMERIC(5,2);
