DROP TRIGGER IF EXISTS update_orders_modtime ON orders;
DROP TRIGGER IF EXISTS update_inventory_modtime ON inventory;
DROP FUNCTION IF EXISTS update_updated_at_column;

DROP TABLE IF EXISTS orders CASCADE;
DROP TABLE IF EXISTS inventory CASCADE;
DROP TABLE IF EXISTS products CASCADE;
DROP TABLE IF EXISTS users CASCADE;