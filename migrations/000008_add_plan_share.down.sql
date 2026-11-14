ALTER TABLE plan
    DROP CONSTRAINT IF EXISTS plan_share_token_has_expiry,
    DROP COLUMN IF EXISTS share_expires_at,
    DROP COLUMN IF EXISTS share_token;
