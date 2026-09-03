DROP INDEX IF EXISTS idx_password_reset_user;
DROP TABLE IF EXISTS password_reset_token;
DROP INDEX IF EXISTS idx_consent_log_user;
DROP TABLE IF EXISTS consent_log;
DROP INDEX IF EXISTS idx_user_profiles_email_active;
ALTER TABLE user_profiles
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS email_verified_at,
    DROP COLUMN IF EXISTS password_hash;
