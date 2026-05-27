-- 购物车表
CREATE TABLE IF NOT EXISTS carts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity INT NOT NULL DEFAULT 1,
    selected BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, product_id)
);

CREATE INDEX idx_carts_user_id ON carts(user_id);
CREATE INDEX idx_carts_product_id ON carts(product_id);
CREATE INDEX idx_carts_selected ON carts(user_id, selected) WHERE selected = TRUE;

COMMENT ON TABLE carts IS '购物车表';
COMMENT ON COLUMN carts.quantity IS '商品数量';
COMMENT ON COLUMN carts.selected IS '是否选中，用于批量结算';

-- DOWN: DROP TABLE IF EXISTS carts CASCADE;
