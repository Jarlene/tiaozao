package model

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	ParentID  *uint          `json:"parent_id"` // nil = 根分类
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 树形结构（通过代码构建）
	Children []Category `gorm:"-" json:"children,omitempty"`
}

func (Category) TableName() string {
	return "categories"
}
