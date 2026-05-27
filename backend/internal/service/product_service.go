package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/Jarlene/tiaozao/backend/internal/model"
	"github.com/Jarlene/tiaozao/backend/internal/repository"
	"gorm.io/gorm"
)

// ProductService 商品业务逻辑
type ProductService struct {
	repo       *repository.ProductRepository
	userRepo   *repository.UserRepository
}

// NewProductService 创建商品服务实例
func NewProductService(repo *repository.ProductRepository, userRepo *repository.UserRepository) *ProductService {
	return &ProductService{repo: repo, userRepo: userRepo}
}

// CreateProductRequest 创建商品请求
type CreateProductRequest struct {
	Title         string   `json:"title" binding:"required,max=200"`
	Description   string   `json:"description"`
	CategoryID    uint     `json:"category_id"`
	Price         float64  `json:"price" binding:"required,gt=0"`
	OriginalPrice float64  `json:"original_price"`
	Condition     string   `json:"condition" binding:"required,oneof=new like_new good fair poor"`
	Images        []string `json:"images"`
}

// UpdateProductRequest 更新商品请求
type UpdateProductRequest struct {
	Title         *string  `json:"title"`
	Description   *string  `json:"description"`
	CategoryID    *uint    `json:"category_id"`
	Price         *float64 `json:"price"`
	OriginalPrice *float64 `json:"original_price"`
	Condition     *string  `json:"condition"`
	Status        *int     `json:"status"`
}

// ProductListQuery 商品列表查询参数
type ProductListQuery struct {
	Page       int     `form:"page" binding:"omitempty,min=1"`
	PageSize   int     `form:"page_size" binding:"omitempty,min=1,max=100"`
	CategoryID uint    `form:"category_id"`
	Status     int     `form:"status"`
	MinPrice   float64 `form:"min_price"`
	MaxPrice   float64 `form:"max_price"`
	Condition  string  `form:"condition"`
	Keyword    string  `form:"keyword"`
	UserID     uint    `form:"user_id"`
}

// Create 创建商品
func (s *ProductService) Create(userID uint, req *CreateProductRequest) (*model.Product, error) {
	product := &model.Product{
		UserID:        userID,
		CategoryID:    req.CategoryID,
		Title:         req.Title,
		Description:   req.Description,
		Price:         req.Price,
		OriginalPrice: req.OriginalPrice,
		Condition:     req.Condition,
		Status:        1, // 默认上架
	}

	// 添加图片
	for i, url := range req.Images {
		product.Images = append(product.Images, model.ProductImage{
			URL:       url,
			SortOrder: i,
		})
	}

	if err := s.repo.Create(product); err != nil {
		return nil, errors.New("创建商品失败")
	}

	// 重新查询以获取完整关联数据
	return s.repo.FindByID(product.ID)
}

// GetByID 根据ID获取商品
func (s *ProductService) GetByID(id uint) (*model.Product, error) {
	product, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("商品不存在")
		}
		return nil, errors.New("获取商品信息失败")
	}
	return product, nil
}

// Update 更新商品
func (s *ProductService) Update(userID, productID uint, req *UpdateProductRequest) (*model.Product, error) {
	product, err := s.repo.FindByID(productID)
	if err != nil {
		return nil, errors.New("商品不存在")
	}

	// 验证所有权
	if product.UserID != userID {
		return nil, errors.New("无权操作此商品")
	}

	if req.Title != nil {
		product.Title = *req.Title
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.CategoryID != nil {
		product.CategoryID = *req.CategoryID
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.OriginalPrice != nil {
		product.OriginalPrice = *req.OriginalPrice
	}
	if req.Condition != nil {
		product.Condition = *req.Condition
	}
	if req.Status != nil {
		product.Status = *req.Status
	}

	if err := s.repo.Update(product); err != nil {
		return nil, errors.New("更新商品失败")
	}

	return s.repo.FindByID(product.ID)
}

// Delete 删除商品（实际用户只能下架，管理员可删除）
func (s *ProductService) Delete(userID, productID uint) error {
	product, err := s.repo.FindByID(productID)
	if err != nil {
		return errors.New("商品不存在")
	}

	if product.UserID != userID {
		return errors.New("无权操作此商品")
	}

	return s.repo.Delete(productID)
}

// List 商品列表
func (s *ProductService) List(query *ProductListQuery) ([]model.Product, int64, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 20
	}

	filters := make(map[string]interface{})
	if query.CategoryID > 0 {
		filters["category_id"] = query.CategoryID
	}
	if query.Status > 0 {
		filters["status"] = query.Status
	}
	if query.MinPrice > 0 {
		filters["min_price"] = query.MinPrice
	}
	if query.MaxPrice > 0 {
		filters["max_price"] = query.MaxPrice
	}
	if query.Condition != "" {
		filters["condition"] = query.Condition
	}
	if query.Keyword != "" {
		filters["keyword"] = query.Keyword
	}
	if query.UserID > 0 {
		filters["user_id"] = query.UserID
	}

	return s.repo.List(query.Page, query.PageSize, filters)
}

// GetUserProducts 获取用户的商品列表
func (s *ProductService) GetUserProducts(userID uint, page, pageSize int) ([]model.Product, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.GetUserProducts(userID, page, pageSize)
}

// UploadImage 上传图片，返回图片URL
func (s *ProductService) UploadImage(filePath string) (string, error) {
	// 目前返回本地路径，生产环境应接入MinIO或OSS
	imageURL := fmt.Sprintf("/uploads/%s", filePath)
	return imageURL, nil
}

// generateOrderNo 生成订单号
func generateOrderNo() string {
	now := time.Now()
	return fmt.Sprintf("TZ%s%08d", now.Format("20060102150405"), now.UnixMilli()%100000000)
}
