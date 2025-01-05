ALTER TABLE users 
DROP COLUMN notification_enabled,
DROP COLUMN theme,
DROP COLUMN language,
DROP COLUMN timezone;

DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_username; 