-- Ensure the budgeting_schema exists
CREATE SCHEMA IF NOT EXISTS budgeting_schema;

-- Create items table
CREATE TABLE budgeting_schema.items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    category VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user_schema.users (id) ON DELETE CASCADE
);

-- Create indexes for items
CREATE INDEX idx_items_user_id ON budgeting_schema.items (user_id);
CREATE INDEX idx_items_category ON budgeting_schema.items (category);

-- Create transactions table
CREATE TABLE budgeting_schema.transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    item_id UUID,
    type VARCHAR(20) NOT NULL CHECK (type IN ('income', 'expense')),
    amount DECIMAL(10, 2) NOT NULL,
    category VARCHAR(50) NOT NULL,
    description TEXT,
    transaction_date TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user_schema.users (id) ON DELETE CASCADE,
    FOREIGN KEY (item_id) REFERENCES budgeting_schema.items (id) ON DELETE SET NULL
);

-- Create indexes for transactions
CREATE INDEX idx_transactions_user_id ON budgeting_schema.transactions (user_id);
CREATE INDEX idx_transactions_item_id ON budgeting_schema.transactions (item_id);
CREATE INDEX idx_transactions_type ON budgeting_schema.transactions (type);
CREATE INDEX idx_transactions_category ON budgeting_schema.transactions (category);
CREATE INDEX idx_transactions_date ON budgeting_schema.transactions (transaction_date);
CREATE INDEX idx_transactions_user_date ON budgeting_schema.transactions (user_id, transaction_date);

