-- Add never_suspend flag to accounts table
-- When enabled, account will not be marked as unavailable on upstream errors
-- Useful for stable, high-priority accounts that should always remain in rotation

ALTER TABLE accounts ADD COLUMN never_suspend BOOLEAN DEFAULT false NOT NULL;

-- Add index for filtering never_suspend accounts
CREATE INDEX idx_accounts_never_suspend ON accounts(never_suspend) WHERE never_suspend = true;

COMMENT ON COLUMN accounts.never_suspend IS 'If true, this account will never be suspended/marked unavailable on upstream errors. Use for stable, high-priority accounts.';
