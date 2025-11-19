-- Drop tables in reverse order of creation to avoid foreign key constraint violations
DROP TABLE IF EXISTS transaction_items;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS recipe_items;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS inventory_items;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS operational_costs;
