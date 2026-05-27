package model

import "time"

// Product 商品模型
type Product struct {
	ID            uint          `gorm:"primaryKey" json:"id"`
	UserID        uint          `gorm:"index;not null" json:"user_id"`
	CategoryID    uint          `gorm:"index" json:"category_id"`
	Title         string        `gorm:"size:200;not null" json:"title"`
	Description   string        `gorm:"type:text" json:"description"`
	Price         float64       `gorm:"type:decimal(10,2);not null" json:"price"`
	OriginalPrice float64       `gorm:"type:decimal(10,2)" json:"original_price"`
	Condition     string        `gorm:"size:20;not null" json:"condition"` // new, like_new, good, fair, poor
	Status        int           `gorm:"default:1" json:"status"`           // 1:active 2:sold 3:offline
	Images        []ProductImage `gorm:"foreignKey:ProductID" json:"images,omitempty"`
	User          User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Category      Category      `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// ProductImage 商品图片模型
type ProductImage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProductID uint      `gorm:"index;not null" json:"product_id"`
	URL       string    `gorm:"size:500;not null" json:"url"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}
