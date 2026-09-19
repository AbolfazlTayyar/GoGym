-- Distinguishes how the coach engages with an athlete after a plan is built:
-- 'private' athletes are managed on an ongoing basis (surfaced on the coach's
-- home screen); 'public' athletes get a plan delivered via share link with no
-- further coach involvement. See docs/er-diagram.md and docs/product-direction.md.
ALTER TABLE athlete
    ADD COLUMN athlete_type TEXT NOT NULL DEFAULT 'private'
        CHECK (athlete_type IN ('private', 'public'));
