-- OpenAI OAuth status tracking

ALTER TABLE IF EXISTS accounts
    ADD COLUMN IF NOT EXISTS oauth_status VARCHAR(30) NOT NULL DEFAULT 'active',
    ADD COLUMN IF NOT EXISTS oauth_refresh_attempts INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS oauth_next_refresh_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS oauth_last_refresh_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS oauth_last_error TEXT NULL;

CREATE INDEX IF NOT EXISTS idx_accounts_oauth_status ON accounts (oauth_status);
CREATE INDEX IF NOT EXISTS idx_accounts_oauth_next_refresh_at ON accounts (oauth_next_refresh_at);
