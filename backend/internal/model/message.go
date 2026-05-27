package model

import "time"

// Message 站内信模型
type Message struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	FromID    uint      `gorm:"index;not null" json:"from_id"`
	ToID      uint      `gorm:"index;not null" json:"to_id"`
	ProductID *uint     `gorm:"index" json:"product_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	IsRead    bool      `gorm:"default:false" json:"is_read"`
	FromUser  User      `gorm:"foreignKey:FromID" json:"from_user,omitempty"`
	ToUser    User      `gorm:"foreignKey:ToID" json:"to_user,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
