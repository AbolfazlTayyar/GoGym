-- The plan a copy was made from, for the builder's "کپی از ..." line. It can't be worked out later,
-- so it is stored now; deleting the source leaves the copy standing, with no source.
ALTER TABLE plan
    ADD COLUMN source_plan_id UUID REFERENCES plan(id) ON DELETE SET NULL;
