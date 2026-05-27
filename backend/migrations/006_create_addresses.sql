-- 收货地址表
CREATE TABLE IF NOT EXISTS addresses (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    receiver_name VARCHAR(50) NOT NULL,
    receiver_phone VARCHAR(20) NOT NULL,
    province VARCHAR(50) NOT NULL,
    city VARCHAR(50) NOT NULL,
    district VARCHAR(50) NOT NULL,
    detail_address VARCHAR(200) NOT NULL,
    zip_code VARCHAR(10),
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    label VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_addresses_user_id ON addresses(user_id);
CREATE INDEX idx_addresses_is_default ON addresses(user_id, is_default) WHERE is_default = TRUE;

COMMENT ON TABLE addresses IS '收货地址表';
COMMENT ON COLUMN addresses.receiver_name IS '收件人姓名';
COMMENT ON COLUMN addresses.receiver_phone IS '收件人电话';
COMMENT ON COLUMN addresses.province IS '省';
COMMENT ON COLUMN addresses.city IS '市';
COMMENT ON COLUMN addresses.district IS '区/县';
COMMENT ON COLUMN addresses.detail_address IS '详细地址';
COMMENT ON COLUMN addresses.zip_code IS '邮政编码';
COMMENT ON COLUMN addresses.is_default IS '是否默认地址';
COMMENT ON COLUMN addresses.label IS '地址标签，如：家、公司、学校';

-- DOWN: DROP TABLE IF EXISTS addresses CASCADE;
