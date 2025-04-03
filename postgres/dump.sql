-- dump.sql
-- PostgreSQL database dump with idempotent operations

-- 1. Drop existing triggers (if any)
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS token_expiry_trigger ON refresh_tokens;

-- 2. Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_cron";

-- 3. Configure pg_cron (alternative syntax)
-- Removed the DO block for ALTER SYSTEM as it's not allowed in functions
-- This should be executed separately by your deployment system

-- 4. Create tables
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    token UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_revoked BOOLEAN DEFAULT FALSE
);

-- 5. Create indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens (user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_revoked ON refresh_tokens (expires_at)
    WHERE is_revoked = FALSE;

-- 6. Create or replace functions
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $func$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
RETURN NEW;
END;
$func$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION revoke_expired_tokens()
RETURNS BIGINT AS $func$
DECLARE
tokens_revoked BIGINT;
BEGIN
WITH updated AS (
UPDATE refresh_tokens
SET is_revoked = TRUE
WHERE expires_at < NOW()
  AND is_revoked = FALSE
    RETURNING 1
    )
SELECT COUNT(*) INTO tokens_revoked FROM updated;
RETURN tokens_revoked;
END;
$func$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION check_token_expiry()
RETURNS TRIGGER AS $func$
BEGIN
    IF NEW.expires_at < NOW() THEN
        NEW.is_revoked := TRUE;
END IF;
RETURN NEW;
END;
$func$ LANGUAGE plpgsql;

-- 7. Create triggers
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER token_expiry_trigger
    BEFORE INSERT OR UPDATE ON refresh_tokens
                         FOR EACH ROW
                         EXECUTE FUNCTION check_token_expiry();

-- 8. Configure pg_cron jobs (alternative syntax)
DO LANGUAGE plpgsql $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_cron') THEN
        -- Remove existing jobs to avoid duplicates
        EXECUTE 'DELETE FROM cron.job WHERE command IN (
            ''SELECT revoke_expired_tokens()'',
            ''DELETE FROM refresh_tokens WHERE is_revoked = TRUE''
        )';

        -- Add new jobs
        PERFORM cron.schedule(
            'revoke-expired-tokens',
            '*/30 * * * *',
            'SELECT revoke_expired_tokens()'
        );

        PERFORM cron.schedule(
            'cleanup-revoked-tokens',
            '0 3 * * *',
            'DELETE FROM refresh_tokens WHERE is_revoked = TRUE'
        );
END IF;
END
$$;