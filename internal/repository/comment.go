package repository

import (
	"context"
	"time"

	"github.com/Jarlene/tiaozao/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CommentRepository struct {
	pool *pgxpool.Pool
}

func NewCommentRepository(pool *pgxpool.Pool) *CommentRepository {
	return &CommentRepository{pool: pool}
}

func (r *CommentRepository) Create(ctx context.Context, comment *model.Comment) error {
	now := time.Now()
	comment.CreatedAt = now
	comment.UpdatedAt = now

	return r.pool.QueryRow(ctx,
		`INSERT INTO comments (product_id, user_id, content, parent_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id`,
		comment.ProductID, comment.UserID, comment.Content, comment.ParentID, comment.CreatedAt, comment.UpdatedAt,
	).Scan(&comment.ID)
}

func (r *CommentRepository) GetByID(ctx context.Context, id uint) (*model.Comment, error) {
	c := &model.Comment{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, product_id, user_id, content, parent_id, created_at, updated_at
		 FROM comments WHERE id = $1`, id,
	).Scan(&c.ID, &c.ProductID, &c.UserID, &c.Content, &c.ParentID, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// ListByProduct returns top-level comments for a product with pagination,
// each including its nested replies.
func (r *CommentRepository) ListByProduct(ctx context.Context, productID uint, page, pageSize int) ([]*model.Comment, int, error) {
	// Count top-level comments
	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM comments WHERE product_id = $1 AND parent_id IS NULL`,
		productID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Fetch top-level comments
	offset := (page - 1) * pageSize
	rows, err := r.pool.Query(ctx,
		`SELECT id, product_id, user_id, content, parent_id, created_at, updated_at
		 FROM comments
		 WHERE product_id = $1 AND parent_id IS NULL
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		productID, pageSize, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	comments := make([]*model.Comment, 0)
	for rows.Next() {
		c := &model.Comment{}
		if err := rows.Scan(&c.ID, &c.ProductID, &c.UserID, &c.Content, &c.ParentID, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, err
		}
		// Fetch replies for each comment
		replies, err := r.ListReplies(ctx, c.ID)
		if err != nil && err != pgx.ErrNoRows {
			return nil, 0, err
		}
		c.Replies = replies
		comments = append(comments, c)
	}
	return comments, total, nil
}

func (r *CommentRepository) ListReplies(ctx context.Context, parentID uint) ([]*model.Comment, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, product_id, user_id, content, parent_id, created_at, updated_at
		 FROM comments
		 WHERE parent_id = $1
		 ORDER BY created_at ASC`,
		parentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	replies := make([]*model.Comment, 0)
	for rows.Next() {
		c := &model.Comment{}
		if err := rows.Scan(&c.ID, &c.ProductID, &c.UserID, &c.Content, &c.ParentID, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		replies = append(replies, c)
	}
	return replies, nil
}
