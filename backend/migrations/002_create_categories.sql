-- 商品分类表
CREATE TABLE IF NOT EXISTS categories (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    parent_id BIGINT REFERENCES categories(id) ON DELETE SET NULL,
    level SMALLINT NOT NULL DEFAULT 1,
    sort_order INT NOT NULL DEFAULT 0,
    icon VARCHAR(255),
    status SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_categories_parent_id ON categories(parent_id);
CREATE INDEX idx_categories_level ON categories(level);
CREATE INDEX idx_categories_status ON categories(status);
CREATE INDEX idx_categories_sort_order ON categories(sort_order);

COMMENT ON TABLE categories IS '商品分类表';
COMMENT ON COLUMN categories.parent_id IS '父分类ID，支持多级分类';
COMMENT ON COLUMN categories.level IS '分类层级: 1-一级分类 2-二级分类 3-三级分类';
COMMENT ON COLUMN categories.sort_order IS '排序权重，数值越小越靠前';
COMMENT ON COLUMN categories.status IS '状态: 1-启用 2-禁用';

-- DOWN: DROP TABLE IF EXISTS categories CASCADE;
