-- Drop indexes
DROP INDEX IF EXISTS user_schema.idx_active_tokens;
DROP INDEX IF EXISTS user_schema.idx_password_reset_token;
DROP INDEX IF EXISTS user_schema.idx_password_reset_user_id;
DROP INDEX IF EXISTS user_schema.idx_users_status;
DROP INDEX IF EXISTS user_schema.idx_users_email;
DROP INDEX IF EXISTS user_schema.idx_users_username;

-- Drop tables
DROP TABLE IF EXISTS user_schema.password_reset_tokens;
DROP TABLE IF EXISTS user_schema.users;

-- Drop schema
DROP SCHEMA IF EXISTS user_schema;

