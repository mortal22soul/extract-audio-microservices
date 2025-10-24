-- Migration rollback: Drop user sessions table
-- Version: 002

DROP INDEX IF EXISTS idx_sessions_token_hash;
DROP INDEX IF EXISTS idx_sessions_expires;
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP TABLE IF EXISTS user_sessions;