package model

import "time"

// Favorite 收藏模型
type Favorite struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex:idx_user_product;not null" json:"user_id"`
	ProductID uint      `gorm:"uniqueIndex:idx_user_product;not null" json:"product_id"`
	Product   Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
