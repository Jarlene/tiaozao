package service

import (
	"context"
	"testing"

	"github.com/Jarlene/tiaozao/internal/model"
	"github.com/jackc/pgx/v5"
)

func TestValidateRating(t *testing.T) {
	cases := []struct {
		rating int
		valid  bool
	}{
		{0, false},
		{1, true},
		{3, true},
		{5, true},
		{6, false},
		{-1, false},
	}
	for _, c := range cases {
		if got := validateRating(c.rating); got != c.valid {
			t.Errorf("validateRating(%d) = %v, want %v", c.rating, got, c.valid)
		}
	}
}

func TestCreateCommentValidation(t *testing.T) {
	svc := NewCommentService(nil)

	// Empty content -> invalid input
	if _, err := svc.CreateComment(context.Background(), 1, 1, "", 0); err != ErrInvalidInput {
		t.Errorf("empty content: got %v, want ErrInvalidInput", err)
	}

	// Out-of-range rating -> invalid input
	if _, err := svc.CreateComment(context.Background(), 1, 1, "good", 6); err != ErrInvalidInput {
		t.Errorf("rating=6: got %v, want ErrInvalidInput", err)
	}
	if _, err := svc.CreateComment(context.Background(), 1, 1, "good", -2); err != ErrInvalidInput {
		t.Errorf("rating=-2: got %v, want ErrInvalidInput", err)
	}
}

func TestReplyToCommentEmptyContent(t *testing.T) {
	svc := NewCommentService(nil)
	if _, err := svc.ReplyToComment(context.Background(), 1, 1, ""); err != ErrInvalidInput {
		t.Errorf("empty reply: got %v, want ErrInvalidInput", err)
	}
}

// testCommentRepo is an in-memory repository stub used to test service
// business rules without a live database.
type testCommentRepo struct {
	comments []*model.Comment
	nextID   uint
}

func newTestCommentRepo() *testCommentRepo {
	return &testCommentRepo{nextID: 1}
}

func (r *testCommentRepo) Create(_ context.Context, c *model.Comment) error {
	c.ID = r.nextID
	r.nextID++
	r.comments = append(r.comments, c)
	return nil
}

func (r *testCommentRepo) GetByID(_ context.Context, id uint) (*model.Comment, error) {
	for _, c := range r.comments {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, pgx.ErrNoRows
}

func (r *testCommentRepo) ListByProduct(_ context.Context, productID uint, page, pageSize int) ([]*model.Comment, *model.RatingSummary, int, error) {
	var comments []*model.Comment
	rated := 0
	sum := 0
	for _, c := range r.comments {
		if c.ProductID == productID && c.ParentID == nil {
			comments = append(comments, c)
			if c.Rating != nil {
				rated++
				sum += *c.Rating
			}
		}
	}
	res := &model.RatingSummary{Count: rated}
	if rated > 0 {
		res.Average = float64(int(float64(sum)/float64(rated)*10+0.5)) / 10
	}
	return comments, res, len(comments), nil
}

func (r *testCommentRepo) ListReplies(_ context.Context, parentID uint) ([]*model.Comment, error) {
	return nil, nil
}

func (r *testCommentRepo) GetProductRating(_ context.Context, productID uint) (*model.RatingSummary, error) {
	_, res, _, err := r.ListByProduct(context.Background(), productID, 1, 10)
	return res, err
}

func TestCreateCommentWithRating(t *testing.T) {
	svc := &CommentService{repo: &testCommentRepo{}}
	ctx := context.Background()

	c, err := svc.CreateComment(ctx, 10, 7, "很棒", 5)
	if err != nil {
		t.Fatalf("CreateComment: %v", err)
	}
	if c.Rating == nil || *c.Rating != 5 {
		t.Fatalf("expected rating 5, got %v", c.Rating)
	}
	if c.UserID != 7 || c.ProductID != 10 {
		t.Fatalf("unexpected ids: %+v", c)
	}

	// No-rating comment allowed
	c2, err := svc.CreateComment(ctx, 10, 8, "无评分", 0)
	if err != nil {
		t.Fatalf("CreateComment(no rating): %v", err)
	}
	if c2.Rating != nil {
		t.Fatalf("expected nil rating, got %d", *c2.Rating)
	}
}

func TestReplyToCommentRejectsNested(t *testing.T) {
	repo := newTestCommentRepo()
	svc := &CommentService{repo: repo}
	ctx := context.Background()

	parent, err := svc.CreateComment(ctx, 10, 7, "一级评论", 4)
	if err != nil {
		t.Fatalf("CreateComment: %v", err)
	}
	reply, err := svc.ReplyToComment(ctx, parent.ID, 9, "回复")
	if err != nil {
		t.Fatalf("ReplyToComment: %v", err)
	}
	if reply.ParentID == nil || *reply.ParentID != parent.ID {
		t.Fatalf("reply parent = %v, want %d", reply.ParentID, parent.ID)
	}

	// Replying to a reply should fail (1-level nesting limit)
	if _, err := svc.ReplyToComment(ctx, reply.ID, 9, "再回"); err != ErrInvalidInput {
		t.Errorf("nested reply: got %v, want ErrInvalidInput", err)
	}
}

func TestReplyToCommentNotFound(t *testing.T) {
	repo := newTestCommentRepo()
	svc := &CommentService{repo: repo}
	if _, err := svc.ReplyToComment(context.Background(), 999, 1, "x"); err != ErrNotFound {
		t.Errorf("missing comment: got %v, want ErrNotFound", err)
	}
}

func TestListCommentsReturnsRatingSummary(t *testing.T) {
	repo := newTestCommentRepo()
	svc := &CommentService{repo: repo}
	ctx := context.Background()

	svc.CreateComment(ctx, 5, 1, "a", 5)
	svc.CreateComment(ctx, 5, 2, "b", 4)
	svc.CreateComment(ctx, 5, 3, "c", 3)
	svc.CreateComment(ctx, 5, 4, "d", 0) // no rating, excluded from average

	result, err := svc.ListComments(ctx, 5, 1, 10)
	if err != nil {
		t.Fatalf("ListComments: %v", err)
	}
	if result.Total != 4 {
		t.Errorf("total = %d, want 4", result.Total)
	}
	if result.Rating == nil {
		t.Fatal("rating summary is nil")
	}
	if result.Rating.Count != 3 {
		t.Errorf("rating count = %d, want 3", result.Rating.Count)
	}
	// (5+4+3)/3 = 4.0
	if result.Rating.Average != 4.0 {
		t.Errorf("rating average = %v, want 4.0", result.Rating.Average)
	}
}

func TestCommentListResultDefaults(t *testing.T) {
	svc := &CommentService{repo: &testCommentRepo{}}
	result, err := svc.ListComments(context.Background(), 1, 0, 0)
	if err != nil {
		t.Fatalf("ListComments: %v", err)
	}
	if result.Page != 1 || result.PageSize != 10 {
		t.Errorf("defaults not applied: page=%d size=%d", result.Page, result.PageSize)
	}
}
