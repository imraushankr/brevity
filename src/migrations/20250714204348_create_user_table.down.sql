-- Brevity Migration: create_user_table
-- Generated: 2025-07-14T20:43:48Z
-- Direction: DOWN

-- Add your SQL below this line

DROP TRIGGER IF EXISTS update_users_updated_at;
DROP INDEX IF EXISTS idx_users_deleted_at;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_username;
DROP TABLE IF EXISTS users;