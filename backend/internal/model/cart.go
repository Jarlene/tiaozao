package model

import "time"

// Cart 购物车模型
type Cart struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex:idx_user_product;not null" json:"user_id"`
	ProductID uint      `gorm:"uniqueIndex:idx_user_product;not null" json:"product_id"`
	Quantity  int       `gorm:"default:1" json:"quantity"`
	Product   Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
