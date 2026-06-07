package model

import (
	"time"

	"gorm.io/gorm"
)

// ProductStatus 商品状态
type ProductStatus int

const (
	ProductStatusActive   ProductStatus = 1 // 在售
	ProductStatusSold     ProductStatus = 2 // 已售
	ProductStatusInactive ProductStatus = 0 // 下架
)

type Product struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Title       string         `gorm:"size:200;not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Price       int64          `gorm:"not null" json:"price"` // 单位：分，避免浮点数精度问题
	Status      ProductStatus  `gorm:"default:1;not null" json:"status"`
	UserID      uint           `gorm:"not null;index" json:"user_id"`
	CategoryID  *uint          `json:"category_id"`
	Stock       int            `gorm:"default:1;not null" json:"stock"` // 库存量
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联（通过 Preload 加载）
	User   User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Images []ProductImage `gorm:"foreignKey:ProductID" json:"images,omitempty"`
}

func (Product) TableName() string {
	return "products"
}

type ProductImage struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	ProductID uint           `gorm:"not null;index" json:"product_id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	ObjectKey string         `gorm:"size:500;not null" json:"object_key"` // MinIO object key
	URL       string         `gorm:"-" json:"url"`                        // 运行时生成，不存库
	SortOrder int            `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProductImage) TableName() string {
	return "product_images"
}
