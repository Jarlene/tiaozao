package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	Email         string         `gorm:"uniqueIndex;size:255;not null" json:"email"`
	PasswordHash  string         `gorm:"size:255;not null" json:"-"`
	Nickname      string         `gorm:"size:100;not null;default:''" json:"nickname"`
	AvatarURL     string         `gorm:"size:500" json:"avatar_url"`
	Role          int            `gorm:"default:1;not null" json:"role"`                       // 1=user, 2=admin
	Status        int            `gorm:"default:1;not null" json:"status"`                     // 1=active, 0=disabled
	Balance       int64          `gorm:"default:0;not null" json:"balance"`                    // 可用余额（分）
	FrozenBalance int64          `gorm:"default:0;not null" json:"frozen_balance"`             // 冻结/托管余额（分）
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}
