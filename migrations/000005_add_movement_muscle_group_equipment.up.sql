-- Primary muscle group and equipment: the two filters that make picking a movement in the plan
-- builder fast. Each is a closed set so a filter can't miss a row over spelling or language, and the
-- client can label the values in its own language. Nullable: a coach's own movement may leave them unset.
ALTER TABLE movement
    ADD COLUMN muscle_group TEXT
        CONSTRAINT movement_muscle_group_check
        CHECK (muscle_group IN ('chest', 'back', 'shoulders', 'arms', 'legs', 'core', 'full_body')),
    ADD COLUMN equipment TEXT
        CONSTRAINT movement_equipment_check
        CHECK (equipment IN ('bodyweight', 'barbell', 'dumbbell', 'kettlebell', 'machine', 'cable', 'band', 'other'));
