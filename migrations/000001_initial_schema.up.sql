-- 1. Create inventory_items table
CREATE TABLE IF NOT EXISTS inventory_items (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    stock_level NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    unit VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. Create products table
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    price NUMERIC(12, 2) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 3. Create recipe_items table
CREATE TABLE IF NOT EXISTS recipe_items (
    id SERIAL PRIMARY KEY,
    product_id INT REFERENCES products(id) ON DELETE CASCADE,
    inventory_item_id INT REFERENCES inventory_items(id) ON DELETE RESTRICT,
    quantity_used NUMERIC(10, 2) NOT NULL
);

-- 4. Create transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    total_amount NUMERIC(12, 2) NOT NULL,
    payment_method VARCHAR(50) NOT NULL,
    transaction_time TIMESTAMPTZ DEFAULT NOW()
);

-- 5. Create transaction_items table
CREATE TABLE IF NOT EXISTS transaction_items (
    id SERIAL PRIMARY KEY,
    transaction_id INT REFERENCES transactions(id) ON DELETE CASCADE,
    product_id INT REFERENCES products(id),
    quantity INT NOT NULL,
    subtotal NUMERIC(12, 2) NOT NULL
);

-- 6. Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 7. Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_recipe_items_product_id ON recipe_items(product_id);
CREATE INDEX IF NOT EXISTS idx_recipe_items_inventory_item_id ON recipe_items(inventory_item_id);
CREATE INDEX IF NOT EXISTS idx_transaction_items_transaction_id ON transaction_items(transaction_id);
CREATE INDEX IF NOT EXISTS idx_transaction_items_product_id ON transaction_items(product_id);

---
-- INSERTS / SEEDING DATA
---

-- 8. Insert initial data (Inventory Items)
INSERT INTO inventory_items (name, stock_level, unit) VALUES
('Roti Hotdog', 100, 'pcs'),
('Mentega', 1000, 'gram'),
('Selai Coklat', 1000, 'gram'),
('Susu Kental Manis', 500, 'ml'),
('Kantong Kresek', 500, 'pcs')
ON CONFLICT (name) DO NOTHING;

-- 9. Insert sample product
INSERT INTO products (name, price) VALUES
('Roti Coklat', 6000.00)
ON CONFLICT (name) DO NOTHING;

-- 10. Insert recipe for Roti Coklat (FIXED AND ROBUST)
WITH product_id_cte AS (
    SELECT id FROM products WHERE name = 'Roti Coklat'
)
INSERT INTO recipe_items (product_id, inventory_item_id, quantity_used)
SELECT
    (SELECT id FROM product_id_cte), 
    ii.id,                           
    CASE 
        WHEN ii.name = 'Roti Hotdog' THEN 1.00
        WHEN ii.name = 'Mentega' THEN 10.00
        WHEN ii.name = 'Selai Coklat' THEN 15.00
        WHEN ii.name = 'Susu Kental Manis' THEN 5.00
        WHEN ii.name = 'Kantong Kresek' THEN 1.00
    END
FROM inventory_items ii
WHERE ii.name IN ('Roti Hotdog', 'Mentega', 'Selai Coklat', 'Susu Kental Manis', 'Kantong Kresek')
  AND EXISTS (SELECT 1 FROM product_id_cte)
ON CONFLICT DO NOTHING;

-- 11. Create default admin user (password: admin123)
-- Note: The hash is for 'password', but we will use it as provided.
INSERT INTO users (username, password_hash, role) VALUES
('admin', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin')
ON CONFLICT (username) DO NOTHING;
