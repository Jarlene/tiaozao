package repository

import (
	stdErrors "errors"

	"flea-market/internal/model"

	"gorm.io/gorm"
)

type ReviewRepository interface {
	Create(review *model.Review) error
	FindByID(id uint) (*model.Review, error)
	FindByProductAndUser(productID, userID uint) (*model.Review, error)
	ListByProduct(productID uint, page, pageSize int) ([]model.Review, int64, error)
	GetRatingStats(productID uint) (*RatingStats, error)
	Update(review *model.Review) error
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

type reviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Create(review *model.Review) error {
	err := r.db.Create(review).Error
	if err != nil && stdErrors.Is(err, gorm.ErrDuplicatedKey) {
		return gorm.ErrDuplicatedKey
	}
	return err
}

func (r *reviewRepository) FindByID(id uint) (*model.Review, error) {
	var review model.Review
	err := r.db.Preload("User").First(&review, id).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *reviewRepository) FindByProductAndUser(productID, userID uint) (*model.Review, error) {
	var review model.Review
	// 使用 Unscoped() 包含软删除记录，防止重复评价绕过软删除检查
	err := r.db.Unscoped().Where("product_id = ? AND user_id = ?", productID, userID).First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *reviewRepository) ListByProduct(productID uint, page, pageSize int) ([]model.Review, int64, error) {
	var reviews []model.Review
	var total int64

	query := r.db.Model(&model.Review{}).Where("product_id = ?", productID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("User").
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&reviews).Error
	if err != nil {
		return nil, 0, err
	}

	return reviews, total, nil
}

func (r *reviewRepository) GetRatingStats(productID uint) (*RatingStats, error) {
	var stats RatingStats

	// 计算平均分和总评论数
	if err := r.db.Model(&model.Review{}).
		Select("COALESCE(AVG(rating), 0) as average, COUNT(*) as total").
		Where("product_id = ?", productID).
		Scan(&stats).Error; err != nil {
		return nil, err
	}

	// 获取各评分分布
	type distResult struct {
		Rating int
		Count  int64
	}
	var dists []distResult
	if err := r.db.Model(&model.Review{}).
		Select("rating, COUNT(*) as count").
		Where("product_id = ?", productID).
		Group("rating").
		Order("rating DESC").
		Scan(&dists).Error; err != nil {
		return nil, err
	}

	for _, d := range dists {
		switch d.Rating {
		case 1:
			stats.Dist1 = d.Count
		case 2:
			stats.Dist2 = d.Count
		case 3:
			stats.Dist3 = d.Count
		case 4:
			stats.Dist4 = d.Count
		case 5:
			stats.Dist5 = d.Count
		}
	}

	return &stats, nil
}

func (r *reviewRepository) Update(review *model.Review) error {
	return r.db.Save(review).Error
}
