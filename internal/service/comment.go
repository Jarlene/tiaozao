package service

import (
	"context"
	"errors"
	"strconv"

	"github.com/Jarlene/tiaozao/internal/model"
	"github.com/Jarlene/tiaozao/internal/repository"
	"github.com/jackc/pgx/v5"
)

type CommentService struct {
	repo *repository.CommentRepository
}

func NewCommentService(repo *repository.CommentRepository) *CommentService {
	return &CommentService{repo: repo}
}

// CommentListResult is the paginated response for listing comments.
type CommentListResult struct {
	Comments []*model.Comment `json:"comments"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

func (s *CommentService) CreateComment(ctx context.Context, productID uint, userID, content string) (*model.Comment, error) {
	if content == "" {
		return nil, errors.New("内容不能为空")
	}

	comment := &model.Comment{
		ProductID: productID,
		UserID:    userID,
		Content:   content,
	}

	if err := s.repo.Create(ctx, comment); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, comment.ID)
}

func (s *CommentService) ReplyToComment(ctx context.Context, commentID uint, userID, content string) (*model.Comment, error) {
	if content == "" {
		return nil, errors.New("回复内容不能为空")
	}

	parent, err := s.repo.GetByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("评论不存在")
		}
		return nil, err
	}

	// Prevent nested replies beyond 1 level
	if parent.ParentID != nil {
		return nil, errors.New("不支持多级嵌套回复")
	}

	reply := &model.Comment{
		ProductID: parent.ProductID,
		UserID:    userID,
		Content:   content,
		ParentID:  &commentID,
	}

	if err := s.repo.Create(ctx, reply); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, reply.ID)
}

func (s *CommentService) ListComments(ctx context.Context, productID uint, page, pageSize int) (*CommentListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	comments, total, err := s.repo.ListByProduct(ctx, productID, page, pageSize)
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
	}, nil
}

// parseInt is a helper for parsing uint from string (e.g., chi URL params).
func parseInt(s string) (uint, error) {
	id, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, errors.New("无效的ID")
	}
	return uint(id), nil
}
