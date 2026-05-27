-- 订单表
CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    order_no VARCHAR(50) NOT NULL UNIQUE,
    buyer_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    seller_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    total_amount DECIMAL(12, 2) NOT NULL,
    discount_amount DECIMAL(12, 2) NOT NULL DEFAULT 0,
    pay_amount DECIMAL(12, 2) NOT NULL,
    status SMALLINT NOT NULL DEFAULT 1,
    payment_method SMALLINT,
    paid_at TIMESTAMP WITH TIME ZONE,
    shipped_at TIMESTAMP WITH TIME ZONE,
    received_at TIMESTAMP WITH TIME ZONE,
    closed_at TIMESTAMP WITH TIME ZONE,
    remark VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 订单商品表
CREATE TABLE IF NOT EXISTS order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_title VARCHAR(200) NOT NULL,
    product_image VARCHAR(500),
    price DECIMAL(10, 2) NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    subtotal DECIMAL(12, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_buyer_id ON orders(buyer_id);
CREATE INDEX idx_orders_seller_id ON orders(seller_id);
CREATE INDEX idx_orders_order_no ON orders(order_no);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);
CREATE INDEX idx_orders_paid_at ON orders(paid_at) WHERE paid_at IS NOT NULL;
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_product_id ON order_items(product_id);

COMMENT ON TABLE orders IS '订单表';
COMMENT ON COLUMN orders.order_no IS '订单编号';
COMMENT ON COLUMN orders.buyer_id IS '买家用户ID';
COMMENT ON COLUMN orders.seller_id IS '卖家用户ID';
COMMENT ON COLUMN orders.total_amount IS '订单总金额';
COMMENT ON COLUMN orders.discount_amount IS '优惠金额';
COMMENT ON COLUMN orders.pay_amount IS '实付金额';
COMMENT ON COLUMN orders.status IS '状态: 1-待付款 2-待发货 3-待收货 4-已完成 5-已关闭 6-退款中';
COMMENT ON COLUMN orders.payment_method IS '支付方式: 1-微信 2-支付宝 3-银行卡';
COMMENT ON TABLE order_items IS '订单商品表';
COMMENT ON COLUMN order_items.product_title IS '商品快照标题';
COMMENT ON COLUMN order_items.product_image IS '商品快照图片';
COMMENT ON COLUMN order_items.price IS '购买时的单价';
COMMENT ON COLUMN order_items.quantity IS '购买数量';
COMMENT ON COLUMN order_items.subtotal IS '小计金额';

-- DOWN: DROP TABLE IF EXISTS order_items CASCADE; DROP TABLE IF EXISTS orders CASCADE;
