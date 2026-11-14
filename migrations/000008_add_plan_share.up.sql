-- A plan's share link (/p/<token>): one live link per plan, replaced on every share and readable
-- without login until it expires. The CHECK keeps a token from ever existing without an expiry.
ALTER TABLE plan
    ADD COLUMN share_token TEXT UNIQUE,
    ADD COLUMN share_expires_at TIMESTAMPTZ,
    ADD CONSTRAINT plan_share_token_has_expiry CHECK ((share_token IS NULL) = (share_expires_at IS NULL));
