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
		`INSERT INTO comments (product_id, user_id, content, rating, parent_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id`,
		comment.ProductID, comment.UserID, comment.Content, comment.Rating, comment.ParentID,
		comment.CreatedAt, comment.UpdatedAt,
	).Scan(&comment.ID)
}

func (r *CommentRepository) GetByID(ctx context.Context, id uint) (*model.Comment, error) {
	c := &model.Comment{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, product_id, user_id, content, rating, parent_id, created_at, updated_at
		 FROM comments WHERE id = $1`, id,
	).Scan(&c.ID, &c.ProductID, &c.UserID, &c.Content, &c.Rating, &c.ParentID, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// ListByProduct returns top-level comments for a product with pagination,
// each including its nested replies, plus the aggregate rating summary and
// the total number of top-level comments.
func (r *CommentRepository) ListByProduct(ctx context.Context, productID uint, page, pageSize int) ([]*model.Comment, *model.RatingSummary, int, error) {
	// Aggregate rating over top-level reviews that carry a rating.
	res := &model.RatingSummary{}
	row := r.pool.QueryRow(ctx,
		`SELECT COALESCE(AVG(rating), 0), COUNT(*) FILTER (WHERE rating IS NOT NULL)
		 FROM comments
		 WHERE product_id = $1 AND parent_id IS NULL`,
		productID,
	)
	var avg float64
	if err := row.Scan(&avg, &res.Count); err != nil {
		return nil, nil, 0, err
	}
	res.Average = round1(avg)

	// Count top-level comments
	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM comments WHERE product_id = $1 AND parent_id IS NULL`,
		productID,
	).Scan(&total)
	if err != nil {
		return nil, nil, 0, err
	}

	// Fetch top-level comments
	offset := (page - 1) * pageSize
	rows, err := r.pool.Query(ctx,
		`SELECT id, product_id, user_id, content, rating, parent_id, created_at, updated_at
		 FROM comments
		 WHERE product_id = $1 AND parent_id IS NULL
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		productID, pageSize, offset,
	)
	if err != nil {
		return nil, nil, 0, err
	}
	defer rows.Close()

	comments := make([]*model.Comment, 0)
	for rows.Next() {
		c := &model.Comment{}
		if err := rows.Scan(&c.ID, &c.ProductID, &c.UserID, &c.Content, &c.Rating, &c.ParentID, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, nil, 0, err
		}
		replies, err := r.ListReplies(ctx, c.ID)
		if err != nil && err != pgx.ErrNoRows {
			return nil, nil, 0, err
		}
		c.Replies = replies
		comments = append(comments, c)
	}
	return comments, res, total, nil
}

func (r *CommentRepository) ListReplies(ctx context.Context, parentID uint) ([]*model.Comment, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, product_id, user_id, content, rating, parent_id, created_at, updated_at
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
		if err := rows.Scan(&c.ID, &c.ProductID, &c.UserID, &c.Content, &c.Rating, &c.ParentID, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		replies = append(replies, c)
	}
	return replies, nil
}

// GetProductRating returns the aggregate rating for a product, or nil when
// the product has no rated reviews.
func (r *CommentRepository) GetProductRating(ctx context.Context, productID uint) (*model.RatingSummary, error) {
	res := &model.RatingSummary{}
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(AVG(rating), 0), COUNT(*) FILTER (WHERE rating IS NOT NULL)
		 FROM comments
		 WHERE product_id = $1 AND parent_id IS NULL AND rating IS NOT NULL`,
		productID,
	).Scan(&res.Average, &res.Count)
	if err != nil {
		return nil, err
	}
	if res.Count == 0 {
		return nil, nil
	}
	res.Average = round1(res.Average)
	return res, nil
}

// round1 rounds f to one decimal place (used for average ratings).
func round1(f float64) float64 {
	return float64(int(f*10+0.5)) / 10
}
