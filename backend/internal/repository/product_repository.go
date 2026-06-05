package repository

import (
	"flea-market/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProductRepository interface {
	Create(product *model.Product) error
	FindByID(id uint) (*model.Product, error)
	Update(product *model.Product) error
	Delete(id uint) error

	// 列表查询
	List(page, pageSize int, status model.ProductStatus) ([]model.Product, int64, error)
	Search(keyword string, categoryID *uint, priceMin, priceMax int64, page, pageSize int) ([]model.Product, int64, error)

	// 用户商品
	ListByUser(userID uint, status *model.ProductStatus, page, pageSize int) ([]model.Product, int64, error)
	CountByUserAndStatus(userID uint) (activeCount, soldCount, inactiveCount int64, err error)

	// 图片
	CreateImage(image *model.ProductImage) error
	UpdateImage(image *model.ProductImage) error
	DeleteImage(id uint) error
	DeleteImagesByProduct(productID uint) error
	FindImageByID(id uint) (*model.ProductImage, error)
	FindImagesByIDs(ids []uint) ([]model.ProductImage, error)
	ListImagesByProduct(productID uint) ([]model.ProductImage, error)

	// 事务
	Transaction(fc func(txRepo ProductRepository) error) error
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(product *model.Product) error {
	return r.db.Create(product).Error
}

func (r *productRepository) FindByID(id uint) (*model.Product, error) {
	var product model.Product
	err := r.db.Preload("User").Preload("Images").First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) Update(product *model.Product) error {
	return r.db.Save(product).Error
}

func (r *productRepository) Delete(id uint) error {
	return r.db.Delete(&model.Product{}, id).Error
}

func (r *productRepository) List(page, pageSize int, status model.ProductStatus) ([]model.Product, int64, error) {
	var products []model.Product
	var total int64

	query := r.db.Model(&model.Product{}).Where("status = ?", status)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("User").
		Preload("Images").
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&products).Error
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *productRepository) Search(keyword string, categoryID *uint, priceMin, priceMax int64, page, pageSize int) ([]model.Product, int64, error) {
	var products []model.Product
	var total int64

	query := r.db.Model(&model.Product{}).Where("status = ?", model.ProductStatusActive)

	if keyword != "" {
		// 使用 PostgreSQL 全文搜索
		query = query.Where(
			"to_tsvector('simple', coalesce(title,'') || ' ' || coalesce(description,'')) @@ plainto_tsquery('simple', ?)",
			keyword,
		)
	}
	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}
	if priceMin > 0 {
		query = query.Where("price >= ?", priceMin)
	}
	if priceMax > 0 {
		query = query.Where("price <= ?", priceMax)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 有关键词时按相关性排序，否则按时间排序
	orderClause := "created_at DESC"
	if keyword != "" {
		orderClause = "ts_rank(to_tsvector('simple', coalesce(title,'') || ' ' || coalesce(description,'')), plainto_tsquery('simple', ?)) DESC, created_at DESC"
	}

	err := query.
		Preload("User").
		Preload("Images").
		Order(clause.Expr{SQL: orderClause, Vars: []interface{}{keyword}}).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&products).Error
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *productRepository) ListByUser(userID uint, status *model.ProductStatus, page, pageSize int) ([]model.Product, int64, error) {
	var products []model.Product
	var total int64

	query := r.db.Model(&model.Product{}).Where("user_id = ?", userID)
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Images").
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&products).Error
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *productRepository) CountByUserAndStatus(userID uint) (activeCount, soldCount, inactiveCount int64, err error) {
	type StatusCount struct {
		Status int
		Count  int64
	}
	var counts []StatusCount

	r.db.Model(&model.Product{}).
		Select("status, count(*) as count").
		Where("user_id = ?", userID).
		Group("status").
		Find(&counts)

	for _, c := range counts {
		switch model.ProductStatus(c.Status) {
		case model.ProductStatusActive:
			activeCount = c.Count
		case model.ProductStatusSold:
			soldCount = c.Count
		case model.ProductStatusInactive:
			inactiveCount = c.Count
		}
	}
	return
}

func (r *productRepository) CreateImage(image *model.ProductImage) error {
	return r.db.Create(image).Error
}

func (r *productRepository) UpdateImage(image *model.ProductImage) error {
	return r.db.Model(&model.ProductImage{}).Where("id = ?", image.ID).Updates(map[string]interface{}{
		"product_id": image.ProductID,
		"sort_order": image.SortOrder,
	}).Error
}

func (r *productRepository) DeleteImage(id uint) error {
	return r.db.Delete(&model.ProductImage{}, id).Error
}

func (r *productRepository) DeleteImagesByProduct(productID uint) error {
	return r.db.Where("product_id = ?", productID).Delete(&model.ProductImage{}).Error
}

func (r *productRepository) FindImageByID(id uint) (*model.ProductImage, error) {
	var image model.ProductImage
	err := r.db.First(&image, id).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}

func (r *productRepository) FindImagesByIDs(ids []uint) ([]model.ProductImage, error) {
	var images []model.ProductImage
	err := r.db.Where("id IN ?", ids).Find(&images).Error
	return images, err
}

func (r *productRepository) ListImagesByProduct(productID uint) ([]model.ProductImage, error) {
	var images []model.ProductImage
	err := r.db.Where("product_id = ?", productID).Order("sort_order ASC").Find(&images).Error
	return images, err
}

func (r *productRepository) Transaction(fc func(txRepo ProductRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		txRepo := &productRepository{db: tx}
		return fc(txRepo)
	})
}
