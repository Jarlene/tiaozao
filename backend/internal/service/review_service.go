package service

import (
	stdErrors "errors"
	"html"
	"time"
	"unicode/utf8"

	"flea-market/internal/model"
	"flea-market/internal/repository"
	"flea-market/pkg/errors"

	"gorm.io/gorm"
)

type ReviewService struct {
	reviewRepo  repository.ReviewRepository
	productRepo repository.ProductRepository
	userRepo    repository.UserRepository
}

func NewReviewService(reviewRepo repository.ReviewRepository, productRepo repository.ProductRepository, userRepo repository.UserRepository) *ReviewService {
	return &ReviewService{reviewRepo: reviewRepo, productRepo: productRepo, userRepo: userRepo}
}

type CreateReviewReq struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Content string `json:"content" binding:"required,min=1,max=1000"`
}

type ReplyReviewReq struct {
	Content string `json:"content" binding:"required,min=1,max=1000"`
}

type ReviewItem struct {
	ID           uint   `json:"id"`
	ProductID    uint   `json:"product_id"`
	UserID       uint   `json:"user_id"`
	Rating       int    `json:"rating"`
	Content      string `json:"content"`
	ReplyContent string `json:"reply_content,omitempty"`
	RepliedAt    string `json:"replied_at,omitempty"`
	CreatedAt    string `json:"created_at"`
	User         *UserSimpleProfile `json:"user,omitempty"`
}

type RatingStats struct {
	Average float64 `json:"average"`
	Total   int64   `json:"total"`
	Dist1   int64   `json:"dist_1"`
	Dist2   int64   `json:"dist_2"`
	Dist3   int64   `json:"dist_3"`
	Dist4   int64   `json:"dist_4"`
	Dist5   int64   `json:"dist_5"`
}

// CheckCanReview 检查用户是否可以评价商品
func (s *ReviewService) CheckCanReview(userID, productID uint) (bool, int, error) {
	// 校验商品存在
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.ErrProductNotFound, nil
		}
		return false, errors.ErrInternal, err
	}

	// 不能评价自己的商品
	if product.UserID == userID {
		return false, errors.ErrCannotReviewOwn, nil
	}

	// 检查是否已评价过
	existing, _ := s.reviewRepo.FindByProductAndUser(productID, userID)
	if existing != nil {
		return false, errors.ErrAlreadyReviewed, nil
	}

	return true, errors.Success, nil
}

// CreateReview 创建评论
func (s *ReviewService) CreateReview(userID, productID uint, req *CreateReviewReq) (*ReviewItem, int, error) {
	// 校验评分范围
	if req.Rating < 1 || req.Rating > 5 {
		return nil, errors.ErrInvalidRating, nil
	}

	// 按字符数校验内容长度，与前端的 maxlength 语义一致
	if utf8.RuneCountInString(req.Content) > 1000 {
		return nil, errors.ErrReviewContentTooLong, nil
	}

	// 校验商品存在
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrProductNotFound, nil
		}
		return nil, errors.ErrInternal, err
	}

	// 不能评价自己的商品
	if product.UserID == userID {
		return nil, errors.ErrCannotReviewOwn, nil
	}

	// 检查是否已评价过
	existing, _ := s.reviewRepo.FindByProductAndUser(productID, userID)
	if existing != nil {
		return nil, errors.ErrAlreadyReviewed, nil
	}

	// XSS 防护：转义 HTML 标签
	req.Content = html.EscapeString(req.Content)

	review := &model.Review{
		ProductID: productID,
		UserID:    userID,
		Rating:    req.Rating,
		Content:   req.Content,
	}

	if err := s.reviewRepo.Create(review); err != nil {
		if stdErrors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.ErrAlreadyReviewed, nil
		}
		return nil, errors.ErrInternal, err
	}

	// 重新获取以加载关联
	created, err := s.reviewRepo.FindByID(review.ID)
	if err != nil {
		return nil, errors.ErrInternal, err
	}

	return s.reviewItemFromModel(created), errors.Success, nil
}

// ListReviews 获取商品评论列表
func (s *ReviewService) ListReviews(productID uint, page, pageSize int) (*PaginatedResult, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	reviews, total, err := s.reviewRepo.ListByProduct(productID, page, pageSize)
	if err != nil {
		return nil, errors.ErrInternal, err
	}

	items := make([]ReviewItem, len(reviews))
	for i, r := range reviews {
		items[i] = *s.reviewItemFromModel(&r)
	}

	return &PaginatedResult{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: calcTotalPages(int(total), pageSize),
	}, errors.Success, nil
}

// GetRatingStats 获取评分统计
func (s *ReviewService) GetRatingStats(productID uint) (*RatingStats, int, error) {
	repoStats, err := s.reviewRepo.GetRatingStats(productID)
	if err != nil {
		return nil, errors.ErrInternal, err
	}

	return &RatingStats{
		Average: repoStats.Average,
		Total:   repoStats.Total,
		Dist1:   repoStats.Dist1,
		Dist2:   repoStats.Dist2,
		Dist3:   repoStats.Dist3,
		Dist4:   repoStats.Dist4,
		Dist5:   repoStats.Dist5,
	}, errors.Success, nil
}

// ReplyReview 卖家回复评论
func (s *ReviewService) ReplyReview(userID, reviewID uint, req *ReplyReviewReq) (*ReviewItem, int, error) {
	// 按字符数校验内容长度，与前端的 maxlength 语义一致
	if utf8.RuneCountInString(req.Content) > 1000 {
		return nil, errors.ErrReviewContentTooLong, nil
	}

	// 查找评论
	review, err := s.reviewRepo.FindByID(reviewID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrReviewNotFound, nil
		}
		return nil, errors.ErrInternal, err
	}

	// 校验当前用户是否为商品卖家
	product, err := s.productRepo.FindByID(review.ProductID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrProductNotFound, nil
		}
		return nil, errors.ErrInternal, err
	}

	if product.UserID != userID {
		return nil, errors.ErrForbidden, nil
	}

	// 检查是否已回复过
	if review.ReplyContent != "" {
		return nil, errors.ErrAlreadyReplied, nil
	}

	now := time.Now()
	review.ReplyContent = html.EscapeString(req.Content)
	review.RepliedAt = &now

	if err := s.reviewRepo.Update(review); err != nil {
		return nil, errors.ErrInternal, err
	}

	updated, err := s.reviewRepo.FindByID(reviewID)
	if err != nil {
		return nil, errors.ErrInternal, err
	}

	return s.reviewItemFromModel(updated), errors.Success, nil
}

// 辅助：从 Review 模型生成响应项
func (s *ReviewService) reviewItemFromModel(r *model.Review) *ReviewItem {
	item := &ReviewItem{
		ID:        r.ID,
		ProductID: r.ProductID,
		UserID:    r.UserID,
		Rating:    r.Rating,
		Content:   r.Content,
		CreatedAt: r.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if r.ReplyContent != "" {
		item.ReplyContent = r.ReplyContent
		if r.RepliedAt != nil {
			item.RepliedAt = r.RepliedAt.Format("2006-01-02 15:04:05")
		}
	}

	if r.User != nil && r.User.ID > 0 {
		item.User = &UserSimpleProfile{
			ID:        r.User.ID,
			Nickname:  r.User.Nickname,
			AvatarURL: r.User.AvatarURL,
			CreatedAt: r.User.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return item
}
