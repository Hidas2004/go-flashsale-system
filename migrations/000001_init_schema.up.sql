-- 1. Kích hoạt Extension để tạo UUID (BẮT BUỘC)
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- FUNC: Tự động cập nhật thời gian updated_at
--OR REPLACE giúp chạy lại code này nhiều lần không lỗi
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 2. Bảng Users
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 3. Bảng Products
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    image_url VARCHAR(500),
    
    -- Flash Sale Logic
    is_flash_sale BOOLEAN DEFAULT FALSE,
    flash_sale_price DECIMAL(10,2) CHECK (flash_sale_price >= 0),
    flash_sale_start TIMESTAMP WITH TIME ZONE,
    flash_sale_end TIMESTAMP WITH TIME ZONE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Index cho việc tìm kiếm sản phẩm Flash Sale nhanh chóng
CREATE INDEX idx_products_flash_sale ON products(is_flash_sale, flash_sale_start, flash_sale_end);

-- 4. Bảng Inventory (Tách riêng để tối ưu Write)
CREATE TABLE IF NOT EXISTS inventory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    stock INT NOT NULL DEFAULT 0 CHECK (stock >= 0), 
    reserved_stock INT NOT NULL DEFAULT 0 CHECK (reserved_stock >= 0),
    sold INT NOT NULL DEFAULT 0,
    version INT NOT NULL DEFAULT 1,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Đảm bảo 1 sản phẩm chỉ có 1 dòng inventory
    CONSTRAINT uq_inventory_product_id UNIQUE (product_id)
);

-- Index này giúp join bảng Products và Inventory nhanh hơn
CREATE INDEX idx_inventory_product_id ON inventory(product_id);

CREATE TRIGGER update_inventory_modtime
    BEFORE UPDATE ON inventory
    FOR EACH ROW
    EXECUTE PROCEDURE update_updated_at_column();


-- 5. Bảng Orders
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    product_id UUID NOT NULL REFERENCES products(id),
    quantity INT NOT NULL CHECK (quantity > 0), 
    total_price DECIMAL(10,2) NOT NULL CHECK (total_price >= 0),
status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'confirmed', 'failed', 'cancelled')),    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 6. bảng Inventory lưu lịch sử biến dộng kho((ai đổi, đổi bao nhiêu, lúc nào))
CREATE TABLE IF NOT EXISTS inventory_audit (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    inventory_id UUID NOT NULL REFERENCES inventory(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,

    old_stock INT NOT NULL, --so lượng trước khi đổi
    new_stock INT NOT NULL, --so luong sau khi đổi
    change_amount INT NOT NULL, --thay đổi bao nhiêu

    action_type VARCHAR(50) NOT NULL,-- 'ORDER', 'RESTOCK', 'CANCEL'
    note TEXT,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()

);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at);

CREATE TRIGGER update_orders_modtime
    BEFORE UPDATE ON orders
    FOR EACH ROW
    EXECUTE PROCEDURE update_updated_at_column();
