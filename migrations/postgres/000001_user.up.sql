-- Ensure the user_schema exists
CREATE SCHEMA IF NOT EXISTS user_schema;

-- Create users table
CREATE TABLE user_schema.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    verified_at TIMESTAMP WITH TIME ZONE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    failed_login_attempts INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for faster lookups
CREATE INDEX idx_users_username ON user_schema.users (username);
CREATE INDEX idx_users_email ON user_schema.users (email);
CREATE INDEX idx_users_status ON user_schema.users (status);

-- Create password_reset_tokens table
CREATE TABLE IF NOT EXISTS user_schema.password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    is_used BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES user_schema.users (id) ON DELETE CASCADE
);

-- Indexes for optimized lookups
CREATE INDEX IF NOT EXISTS idx_password_reset_user_id ON user_schema.password_reset_tokens (user_id);
CREATE INDEX IF NOT EXISTS idx_password_reset_token ON user_schema.password_reset_tokens (token);

-- Create partial index for faster lookup of valid tokens
CREATE INDEX IF NOT EXISTS idx_active_tokens
ON user_schema.password_reset_tokens (token)
WHERE is_used = false AND expires_at > NOW();

