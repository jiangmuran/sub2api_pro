-- Add never_suspend flag to accounts table
-- When enabled, account will not be marked as unavailable on upstream errors
-- Useful for stable, high-priority accounts that should always remain in rotation

-- Use IF NOT EXISTS to make this migration idempotent (safe for renumbering)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'accounts' AND column_name = 'never_suspend'
    ) THEN
        ALTER TABLE accounts ADD COLUMN never_suspend BOOLEAN DEFAULT false NOT NULL;
    END IF;
END $$;

-- Add index for filtering never_suspend accounts (idempotent)
CREATE INDEX IF NOT EXISTS idx_accounts_never_suspend ON accounts(never_suspend) WHERE never_suspend = true;

-- Add comment (always safe to re-run)
COMMENT ON COLUMN accounts.never_suspend IS 'If true, this account will never be suspended/marked unavailable on upstream errors. Use for stable, high-priority accounts.';
