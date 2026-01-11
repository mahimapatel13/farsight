-- Drop indexes
DROP INDEX IF EXISTS budgeting_schema.idx_transactions_user_date;
DROP INDEX IF EXISTS budgeting_schema.idx_transactions_date;
DROP INDEX IF EXISTS budgeting_schema.idx_transactions_category;
DROP INDEX IF EXISTS budgeting_schema.idx_transactions_type;
DROP INDEX IF EXISTS budgeting_schema.idx_transactions_item_id;
DROP INDEX IF EXISTS budgeting_schema.idx_transactions_user_id;
DROP INDEX IF EXISTS budgeting_schema.idx_items_category;
DROP INDEX IF EXISTS budgeting_schema.idx_items_user_id;

-- Drop tables
DROP TABLE IF EXISTS budgeting_schema.transactions;
DROP TABLE IF EXISTS budgeting_schema.items;

-- Drop schema
DROP SCHEMA IF EXISTS budgeting_schema;

