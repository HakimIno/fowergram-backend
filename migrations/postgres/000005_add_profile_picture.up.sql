-- Add profile_picture column to users table
ALTER TABLE users
ADD COLUMN IF NOT EXISTS profile_picture TEXT DEFAULT '';

COMMENT ON COLUMN users.profile_picture IS 'URL to the user''s profile picture'; 