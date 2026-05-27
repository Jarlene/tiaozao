package model

import "time"

// Review 商品评价模型
type Review struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProductID uint      `gorm:"index;not null" json:"product_id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	OrderID   uint      `gorm:"index" json:"order_id"`
	Rating    int       `gorm:"not null" json:"rating"` // 1-5
	Content   string    `gorm:"type:text" json:"content"`
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Product   Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
