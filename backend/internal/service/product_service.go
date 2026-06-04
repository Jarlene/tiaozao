package service

import (
	stdErrors "errors"
	"mime/multipart"
	"path"
	"strings"

	"flea-market/internal/model"
	"flea-market/internal/repository"
	"flea-market/pkg/errors"

	"gorm.io/gorm"
)

// 允许的图片 MIME 类型
var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

const (
	maxImageSize  = 5 * 1024 * 1024 // 5MB
	maxImageCount = 9
)

// FileStorage 文件存储接口，支持 MinIO 实现和测试 mock
type FileStorage interface {
	UploadFile(bucket string, file multipart.File, header *multipart.FileHeader) (string, error)
	DeleteFile(bucket string, objectKey string) error
	GetFileURL(bucket string, objectKey string) string
	ProductBucket() string
}

type ProductService struct {
	productRepo  repository.ProductRepository
	categoryRepo repository.CategoryRepository
	minio        FileStorage
}

func NewProductService(productRepo repository.ProductRepository, categoryRepo repository.CategoryRepository, minio FileStorage) *ProductService {
	return &ProductService{productRepo: productRepo, categoryRepo: categoryRepo, minio: minio}
}

type CreateProductReq struct {
	Title       string `json:"title" binding:"required,max=200"`
	Description string `json:"description"`
	Price       int64  `json:"price" binding:"required,min=0"`       // 单位：分
	CategoryID  *uint  `json:"category_id"`
	ImageIDs    []uint `json:"image_ids"` // 预上传的图片ID列表
}

type UpdateProductReq struct {
	Title       string `json:"title" binding:"required,max=200"`
	Description string `json:"description"`
	Price       int64  `json:"price" binding:"required,min=0"`
	CategoryID  *uint  `json:"category_id"`
}

type ProductListItem struct {
	ID        uint                 `json:"id"`
	Title     string               `json:"title"`
	Price     int64                `json:"price"`
	Status    model.ProductStatus  `json:"status"`
	UserID    uint                 `json:"user_id"`
	CreatedAt string               `json:"created_at"`
	Images    []ProductImageItem   `json:"images,omitempty"`
	User      *UserSimpleProfile   `json:"user,omitempty"`
}

type ProductDetail struct {
	ID          uint                `json:"id"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Price       int64               `json:"price"`
	Status      model.ProductStatus `json:"status"`
	UserID      uint                `json:"user_id"`
	CategoryID  *uint               `json:"category_id"`
	CreatedAt   string              `json:"created_at"`
	UpdatedAt   string              `json:"updated_at"`
	Images      []ProductImageItem  `json:"images,omitempty"`
	User        *UserSimpleProfile  `json:"user,omitempty"`
}

type UserSimpleProfile struct {
	ID        uint   `json:"id"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
	CreatedAt string `json:"created_at"`
}

type ProductImageItem struct {
	ID        uint   `json:"id"`
	URL       string `json:"url"`
	SortOrder int    `json:"sort_order"`
}

type PaginatedResult struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// Create 创建商品
func (s *ProductService) Create(userID uint, req *CreateProductReq) (*ProductDetail, int, error) {
	// 验证分类存在
	if req.CategoryID != nil {
		if _, err := s.categoryRepo.FindByID(*req.CategoryID); err != nil {
			if stdErrors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.ErrCategoryNotFound, nil
			}
			return nil, errors.ErrInternal, err
		}
	}

	product := &model.Product{
		Title:       req.Title,
		Description: req.Description,
		Price:       req.Price,
		Status:      model.ProductStatusActive,
		UserID:      userID,
		CategoryID:  req.CategoryID,
	}

	if err := s.productRepo.Create(product); err != nil {
		return nil, errors.ErrInternal, err
	}

	// 关联预上传的图片
	if len(req.ImageIDs) > 0 {
		for _, imgID := range req.ImageIDs {
			img, err := s.productRepo.FindImageByID(imgID)
			if err != nil {
				continue
			}
			// 只关联未归属的图片
			if img.ProductID == 0 {
				img.ProductID = product.ID
				s.productRepo.UpdateImage(img)
			}
		}
	}

	return s.productDetailFromModel(product), errors.Success, nil
}

// GetByID 获取商品详情
func (s *ProductService) GetByID(id uint) (*ProductDetail, int, error) {
	product, err := s.productRepo.FindByID(id)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrProductNotFound, nil
		}
		return nil, errors.ErrInternal, err
	}

	return s.productDetailFromModel(product), errors.Success, nil
}

// Update 更新商品
func (s *ProductService) Update(userID, productID uint, req *UpdateProductReq) (*ProductDetail, int, error) {
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrProductNotFound, nil
		}
		return nil, errors.ErrInternal, err
	}

	// 权限校验
	if product.UserID != userID {
		return nil, errors.ErrForbidden, nil
	}

	if req.CategoryID != nil {
		if _, err := s.categoryRepo.FindByID(*req.CategoryID); err != nil {
			if stdErrors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.ErrCategoryNotFound, nil
			}
			return nil, errors.ErrInternal, err
		}
	}

	product.Title = req.Title
	product.Description = req.Description
	product.Price = req.Price
	product.CategoryID = req.CategoryID

	if err := s.productRepo.Update(product); err != nil {
		return nil, errors.ErrInternal, err
	}

	// 重新获取以加载关联
	updated, err := s.productRepo.FindByID(productID)
	if err != nil {
		return nil, errors.ErrInternal, err
	}

	return s.productDetailFromModel(updated), errors.Success, nil
}

// Delete 下架商品
func (s *ProductService) Delete(userID, productID uint) (int, error) {
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrProductNotFound, nil
		}
		return errors.ErrInternal, err
	}

	if product.UserID != userID {
		return errors.ErrForbidden, nil
	}

	// 清理关联图片（从 MinIO 删除）
	for _, img := range product.Images {
		if img.DeletedAt.Valid {
			continue
		}
		if err := s.minio.DeleteFile(s.minio.ProductBucket(), img.ObjectKey); err != nil {
			// 即使单张图片删除失败，继续清理其余图片
		}
	}

	product.Status = model.ProductStatusInactive
	if err := s.productRepo.Update(product); err != nil {
		return errors.ErrInternal, err
	}

	return errors.Success, nil
}

// List 获取商品列表（在售）
func (s *ProductService) List(page, pageSize int) (*PaginatedResult, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	products, total, err := s.productRepo.List(page, pageSize, model.ProductStatusActive)
	if err != nil {
		return nil, errors.ErrInternal, err
	}

	items := make([]ProductListItem, len(products))
	for i, p := range products {
		items[i] = s.listItemFromModel(&p)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &PaginatedResult{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, errors.Success, nil
}

// Search 搜索商品
func (s *ProductService) Search(keyword string, categoryID *uint, priceMin, priceMax int64, page, pageSize int) (*PaginatedResult, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	products, total, err := s.productRepo.Search(keyword, categoryID, priceMin, priceMax, page, pageSize)
	if err != nil {
		return nil, errors.ErrInternal, err
	}

	items := make([]ProductListItem, len(products))
	for i, p := range products {
		items[i] = s.listItemFromModel(&p)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &PaginatedResult{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, errors.Success, nil
}

// ListMyProducts 获取用户的商品列表
func (s *ProductService) ListMyProducts(userID uint, status *model.ProductStatus, page, pageSize int) (*PaginatedResult, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	products, total, err := s.productRepo.ListByUser(userID, status, page, pageSize)
	if err != nil {
		return nil, errors.ErrInternal, err
	}

	items := make([]ProductListItem, len(products))
	for i, p := range products {
		items[i] = s.listItemFromModel(&p)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &PaginatedResult{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, errors.Success, nil
}

// GetMyCounts 获取用户各状态商品数量
func (s *ProductService) GetMyCounts(userID uint) (map[string]int64, int, error) {
	active, sold, inactive, err := s.productRepo.CountByUserAndStatus(userID)
	if err != nil {
		return nil, errors.ErrInternal, err
	}
	return map[string]int64{
		"active":   active,
		"sold":     sold,
		"inactive": inactive,
	}, errors.Success, nil
}

// UpdateStatus 更新商品状态
func (s *ProductService) UpdateStatus(userID, productID uint, status model.ProductStatus) (int, error) {
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrProductNotFound, nil
		}
		return errors.ErrInternal, err
	}

	if product.UserID != userID {
		return errors.ErrForbidden, nil
	}

	product.Status = status
	if err := s.productRepo.Update(product); err != nil {
		return errors.ErrInternal, err
	}

	return errors.Success, nil
}

// UploadImage 上传图片到 MinIO
func (s *ProductService) UploadImage(userID uint, file multipart.File, header *multipart.FileHeader) (*ProductImageItem, int, error) {
	// 验证文件类型
	contentType := header.Header.Get("Content-Type")
	ext, ok := allowedImageTypes[contentType]
	if !ok {
		// 也检查扩展名
		ext = strings.ToLower(path.Ext(header.Filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
			return nil, errors.ErrFileFormat, nil
		}
		_ = ext
	}

	// 验证文件大小（5MB）
	if header.Size > maxImageSize {
		return nil, errors.ErrFileTooLarge, nil
	}

	// 上传到 MinIO
	objectKey, err := s.minio.UploadFile(s.minio.ProductBucket(), file, header)
	if err != nil {
		return nil, errors.ErrInternal, err
	}

	// 保存图片记录
	image := &model.ProductImage{
		ObjectKey: objectKey,
		SortOrder: 0,
		UserID:    userID,
	}
	if err := s.productRepo.CreateImage(image); err != nil {
		return nil, errors.ErrInternal, err
	}

	return &ProductImageItem{
		ID:        image.ID,
		URL:       s.minio.GetFileURL(s.minio.ProductBucket(), objectKey),
		SortOrder: 0,
	}, errors.Success, nil
}

// GetImage 获取图片信息
func (s *ProductService) GetImage(imageID uint) (*ProductImageItem, int, error) {
	image, err := s.productRepo.FindImageByID(imageID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrImageNotFound, nil
		}
		return nil, errors.ErrInternal, err
	}

	return &ProductImageItem{
		ID:        image.ID,
		URL:       s.minio.GetFileURL(s.minio.ProductBucket(), image.ObjectKey),
		SortOrder: image.SortOrder,
	}, errors.Success, nil
}

// DeleteImage 删除图片
func (s *ProductService) DeleteImage(userID, imageID uint) (int, error) {
	image, err := s.productRepo.FindImageByID(imageID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrImageNotFound, nil
		}
		return errors.ErrInternal, err
	}

	// 校验图片归属
	if image.UserID != userID {
		return errors.ErrForbidden, nil
	}

	// 从 MinIO 删除
	if err := s.minio.DeleteFile(s.minio.ProductBucket(), image.ObjectKey); err != nil {
		// 即使 MinIO 删除失败，也继续删除数据库记录
	}

	if err := s.productRepo.DeleteImage(imageID); err != nil {
		return errors.ErrInternal, err
	}

	return errors.Success, nil
}

// 辅助：从 Product 模型生成详情
func (s *ProductService) productDetailFromModel(p *model.Product) *ProductDetail {
	detail := &ProductDetail{
		ID:          p.ID,
		Title:       p.Title,
		Description: p.Description,
		Price:       p.Price,
		Status:      p.Status,
		UserID:      p.UserID,
		CategoryID:  p.CategoryID,
		CreatedAt:   p.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if p.User.ID > 0 {
		detail.User = &UserSimpleProfile{
			ID:        p.User.ID,
			Nickname:  p.User.Nickname,
			AvatarURL: p.User.AvatarURL,
			CreatedAt: p.User.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	if len(p.Images) > 0 {
		images := make([]ProductImageItem, 0, len(p.Images))
		for _, img := range p.Images {
			if img.DeletedAt.Valid {
				continue
			}
			images = append(images, ProductImageItem{
				ID:        img.ID,
				URL:       s.minio.GetFileURL(s.minio.ProductBucket(), img.ObjectKey),
				SortOrder: img.SortOrder,
			})
		}
		detail.Images = images
	}

	return detail
}

// 辅助：从 Product 模型生成列表项
func (s *ProductService) listItemFromModel(p *model.Product) ProductListItem {
	item := ProductListItem{
		ID:        p.ID,
		Title:     p.Title,
		Price:     p.Price,
		Status:    p.Status,
		UserID:    p.UserID,
		CreatedAt: p.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if p.User.ID > 0 {
		item.User = &UserSimpleProfile{
			ID:        p.User.ID,
			Nickname:  p.User.Nickname,
			AvatarURL: p.User.AvatarURL,
			CreatedAt: p.User.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	if len(p.Images) > 0 {
		images := make([]ProductImageItem, 0, len(p.Images))
		for _, img := range p.Images {
			if img.DeletedAt.Valid {
				continue
			}
			images = append(images, ProductImageItem{
				ID:        img.ID,
				URL:       s.minio.GetFileURL(s.minio.ProductBucket(), img.ObjectKey),
				SortOrder: img.SortOrder,
			})
		}
		if len(images) > 0 {
			item.Images = images[:1] // 列表只展示第一张
		}
	}

	return item
}
