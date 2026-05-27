-- 商品评价表
CREATE TABLE IF NOT EXISTS reviews (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_id BIGINT REFERENCES orders(id) ON DELETE SET NULL,
    rating SMALLINT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    content TEXT,
    images JSONB,
    is_anonymous BOOLEAN NOT NULL DEFAULT FALSE,
    status SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, order_id)
);

CREATE INDEX idx_reviews_product_id ON reviews(product_id);
CREATE INDEX idx_reviews_user_id ON reviews(user_id);
CREATE INDEX idx_reviews_order_id ON reviews(order_id);
CREATE INDEX idx_reviews_rating ON reviews(product_id, rating);
CREATE INDEX idx_reviews_status ON reviews(status);
CREATE INDEX idx_reviews_created_at ON reviews(created_at DESC);

COMMENT ON TABLE reviews IS '商品评价表';
COMMENT ON COLUMN reviews.rating IS '评分: 1-5星';
COMMENT ON COLUMN reviews.content IS '评价内容';
COMMENT ON COLUMN reviews.images IS '评价图片列表，JSON数组格式';
COMMENT ON COLUMN reviews.is_anonymous IS '是否匿名评价';
COMMENT ON COLUMN reviews.status IS '状态: 1-正常 2-隐藏 3-已删除';

-- DOWN: DROP TABLE IF EXISTS reviews CASCADE;
