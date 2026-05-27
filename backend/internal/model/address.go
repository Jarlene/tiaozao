package model

import "time"

// Address 收货地址模型
type Address struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Name      string    `gorm:"size:50;not null" json:"name"`
	Phone     string    `gorm:"size:20;not null" json:"phone"`
	Province  string    `gorm:"size:50" json:"province"`
	City      string    `gorm:"size:50" json:"city"`
	District  string    `gorm:"size:50" json:"district"`
	Detail    string    `gorm:"size:500;not null" json:"detail"`
	IsDefault bool      `gorm:"default:false" json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
