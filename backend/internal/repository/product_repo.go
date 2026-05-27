package repository

import (
	"github.com/Jarlene/tiaozao/backend/internal/model"
	"gorm.io/gorm"
)

// ProductRepository 商品数据仓库
type ProductRepository struct {
	db *gorm.DB
}

// NewProductRepository 创建商品仓库实例
func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Create 创建商品
func (r *ProductRepository) Create(product *model.Product) error {
	return r.db.Create(product).Error
}

// FindByID 根据ID查找商品，预加载关联数据
func (r *ProductRepository) FindByID(id uint) (*model.Product, error) {
	var product model.Product
	err := r.db.Preload("Images").
		Preload("User").
		Preload("Category").
		First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// Update 更新商品
func (r *ProductRepository) Update(product *model.Product) error {
	return r.db.Save(product).Error
}

// Delete 删除商品
func (r *ProductRepository) Delete(id uint) error {
	// 级联删除图片
	r.db.Where("product_id = ?", id).Delete(&model.ProductImage{})
	return r.db.Delete(&model.Product{}, id).Error
}

// List 商品列表，支持分页、分类、状态、价格范围过滤
func (r *ProductRepository) List(page, pageSize int, filters map[string]interface{}) ([]model.Product, int64, error) {
	var products []model.Product
	var total int64

	query := r.db.Model(&model.Product{}).Preload("Images").Preload("Category")

	// 按分类过滤
	if categoryID, ok := filters["category_id"]; ok {
		query = query.Where("category_id = ?", categoryID)
	}

	// 按状态过滤
	if status, ok := filters["status"]; ok {
		query = query.Where("status = ?", status)
	} else {
		// 默认只展示上架商品
		query = query.Where("status = ?", 1)
	}

	// 按卖家过滤
	if userID, ok := filters["user_id"]; ok {
		query = query.Where("user_id = ?", userID)
	}

	// 按价格范围过滤
	if minPrice, ok := filters["min_price"]; ok {
		query = query.Where("price >= ?", minPrice)
	}
	if maxPrice, ok := filters["max_price"]; ok {
		query = query.Where("price <= ?", maxPrice)
	}

	// 按新旧程度过滤
	if condition, ok := filters["condition"]; ok {
		query = query.Where("condition = ?", condition)
	}

	// 按关键词搜索
	if keyword, ok := filters["keyword"]; ok {
		like := "%" + keyword.(string) + "%"
		query = query.Where("title ILIKE ? OR description ILIKE ?", like, like)
	}

	query.Count(&total)
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&products).Error
	return products, total, err
}

// GetUserProducts 获取用户的商品列表
func (r *ProductRepository) GetUserProducts(userID uint, page, pageSize int) ([]model.Product, int64, error) {
	var products []model.Product
	var total int64

	r.db.Model(&model.Product{}).Where("user_id = ?", userID).Count(&total)
	offset := (page - 1) * pageSize
	err := r.db.Where("user_id = ?", userID).
		Preload("Images").
		Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&products).Error
	return products, total, err
}

// DB 暴露原始DB用于事务
func (r *ProductRepository) DB() *gorm.DB {
	return r.db
}
