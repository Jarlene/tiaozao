package service

import (
	"context"
	"errors"

	"github.com/Jarlene/tiaozao/internal/model"
	"github.com/jackc/pgx/v5"
)

// CommentStore is the persistence contract the comment service depends on.
// Keeping it an interface makes the layering explicit and the service
// testable without a live database.
type CommentStore interface {
	Create(ctx context.Context, comment *model.Comment) error
	GetByID(ctx context.Context, id uint) (*model.Comment, error)
	ListByProduct(ctx context.Context, productID uint, page, pageSize int) ([]*model.Comment, *model.RatingSummary, int, error)
	ListReplies(ctx context.Context, parentID uint) ([]*model.Comment, error)
	GetProductRating(ctx context.Context, productID uint) (*model.RatingSummary, error)
}

// CommentService manages product reviews and one-level nested replies.
type CommentService struct {
	repo CommentStore
}

func NewCommentService(repo CommentStore) *CommentService {
	return &CommentService{repo: repo}
}

// CommentListResult is the paginated response for listing comments,
// including the aggregate rating summary.
type CommentListResult struct {
	Comments []*model.Comment     `json:"comments"`
	Total    int                  `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
	Rating   *model.RatingSummary `json:"rating,omitempty"`
}

// MinRating and MaxRating bound the acceptable product rating values.
const (
	MinRating = 1
	MaxRating = 5
)

// validateRating reports whether r is a valid 1-5 star rating.
func validateRating(r int) bool {
	return r >= MinRating && r <= MaxRating
}

// CreateComment adds a top-level review. rating (0 means "no rating") is
// optional and validated against the 1-5 range when provided.
func (s *CommentService) CreateComment(ctx context.Context, productID uint, userID uint, content string, rating int) (*model.Comment, error) {
	if content == "" {
		return nil, ErrInvalidInput
	}
	if rating < 0 || rating > MaxRating {
		return nil, ErrInvalidInput
	}

	var ratingPtr *int
	if rating > 0 {
		ratingPtr = &rating
	}

	comment := &model.Comment{
		ProductID: productID,
		UserID:    userID,
		Content:   content,
		Rating:    ratingPtr,
	}

	if err := s.repo.Create(ctx, comment); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, comment.ID)
}

// ReplyToComment adds a one-level nested reply (text-only, no rating).
func (s *CommentService) ReplyToComment(ctx context.Context, commentID uint, userID uint, content string) (*model.Comment, error) {
	if content == "" {
		return nil, ErrInvalidInput
	}

	parent, err := s.repo.GetByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Prevent nested replies beyond 1 level
	if parent.ParentID != nil {
		return nil, ErrInvalidInput
	}

	parentID := commentID
	reply := &model.Comment{
		ProductID: parent.ProductID,
		UserID:    userID,
		Content:   content,
		Rating:    nil,
		ParentID:  &parentID,
	}

	if err := s.repo.Create(ctx, reply); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, reply.ID)
}

// ListComments returns paginated top-level comments for a product plus the
// aggregate rating summary.
func (s *CommentService) ListComments(ctx context.Context, productID uint, page, pageSize int) (*CommentListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	comments, rating, total, err := s.repo.ListByProduct(ctx, productID, page, pageSize)
	if err != nil {
		return nil, err
	}

	if comments == nil {
		comments = []*model.Comment{}
	}

	return &CommentListResult{
		Comments: comments,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Rating:   rating,
	}, nil
}
