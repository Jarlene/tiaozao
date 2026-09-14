-- 002_comments.sql
-- Comments/reviews with 1-5 star ratings and one-level nested replies.
-- Top-level rows are product reviews; rows with parent_id set are replies
-- to a review (text-only, rating IS NULL).

CREATE TABLE IF NOT EXISTS comments (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL DEFAULT 1,
    content TEXT NOT NULL,
    rating SMALLINT CHECK (rating IS NULL OR (rating >= 1 AND rating <= 5)),
    parent_id BIGINT REFERENCES comments(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_comments_product ON comments(product_id, parent_id);
CREATE INDEX IF NOT EXISTS idx_comments_created ON comments(created_at DESC);
