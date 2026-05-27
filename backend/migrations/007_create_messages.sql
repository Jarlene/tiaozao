-- 站内信/聊天消息表
CREATE TABLE IF NOT EXISTS messages (
    id BIGSERIAL PRIMARY KEY,
    sender_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    receiver_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id BIGINT REFERENCES products(id) ON DELETE SET NULL,
    content TEXT NOT NULL,
    message_type SMALLINT NOT NULL DEFAULT 1,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    read_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_messages_sender_id ON messages(sender_id);
CREATE INDEX idx_messages_receiver_id ON messages(receiver_id);
CREATE INDEX idx_messages_product_id ON messages(product_id);
CREATE INDEX idx_messages_created_at ON messages(created_at DESC);
CREATE INDEX idx_messages_conversation ON messages(sender_id, receiver_id, created_at DESC);
CREATE INDEX idx_messages_unread ON messages(receiver_id, is_read) WHERE is_read = FALSE;

COMMENT ON TABLE messages IS '站内信/聊天消息表';
COMMENT ON COLUMN messages.sender_id IS '发送者用户ID';
COMMENT ON COLUMN messages.receiver_id IS '接收者用户ID';
COMMENT ON COLUMN messages.product_id IS '关联商品ID，私聊关于某个商品时可关联';
COMMENT ON COLUMN messages.content IS '消息内容';
COMMENT ON COLUMN messages.message_type IS '消息类型: 1-文本 2-图片 3-系统通知';
COMMENT ON COLUMN messages.is_read IS '是否已读';
COMMENT ON COLUMN messages.read_at IS '读取时间';

-- DOWN: DROP TABLE IF EXISTS messages CASCADE;
