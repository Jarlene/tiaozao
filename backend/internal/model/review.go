package model

import (
	"time"

	"gorm.io/gorm"
)

type Review struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	ProductID    uint           `gorm:"not null;uniqueIndex:idx_review_product_user" json:"product_id"`
	UserID       uint           `gorm:"not null;uniqueIndex:idx_review_product_user" json:"user_id"`
	Rating       int            `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"`
	Content      string         `gorm:"type:varchar(1000);not null" json:"content"`
	ReplyContent string         `gorm:"type:varchar(1000)" json:"reply_content,omitempty"`
	RepliedAt    *time.Time     `json:"replied_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Review) TableName() string { return "reviews" }
