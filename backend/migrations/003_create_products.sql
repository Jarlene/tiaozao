-- 商品表
CREATE TABLE IF NOT EXISTS products (
    id BIGSERIAL PRIMARY KEY,
    seller_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id BIGINT REFERENCES categories(id) ON DELETE SET NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    original_price DECIMAL(10, 2),
    stock INT NOT NULL DEFAULT 1,
    sales INT NOT NULL DEFAULT 0,
    status SMALLINT NOT NULL DEFAULT 1,
    is_featured BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 商品图片表
CREATE TABLE IF NOT EXISTS product_images (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    url VARCHAR(500) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    is_cover BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_seller_id ON products(seller_id);
CREATE INDEX idx_products_category_id ON products(category_id);
CREATE INDEX idx_products_status ON products(status);
CREATE INDEX idx_products_created_at ON products(created_at DESC);
CREATE INDEX idx_products_price ON products(price);
CREATE INDEX idx_products_sales ON products(sales DESC);
CREATE INDEX idx_product_images_product_id ON product_images(product_id);
CREATE INDEX idx_product_images_cover ON product_images(product_id, is_cover) WHERE is_cover = TRUE;

COMMENT ON TABLE products IS '商品表';
COMMENT ON COLUMN products.seller_id IS '卖家用户ID';
COMMENT ON COLUMN products.category_id IS '商品分类ID';
COMMENT ON COLUMN products.price IS '售价';
COMMENT ON COLUMN products.original_price IS '原价';
COMMENT ON COLUMN products.stock IS '库存数量';
COMMENT ON COLUMN products.sales IS '销量';
COMMENT ON COLUMN products.status IS '状态: 1-在售 2-下架 3-已删除';
COMMENT ON COLUMN products.is_featured IS '是否精选推荐';
COMMENT ON TABLE product_images IS '商品图片表';
COMMENT ON COLUMN product_images.sort_order IS '排序权重，数值越小越靠前';
COMMENT ON COLUMN product_images.is_cover IS '是否为封面图';

-- DOWN: DROP TABLE IF EXISTS product_images CASCADE; DROP TABLE IF EXISTS products CASCADE;
