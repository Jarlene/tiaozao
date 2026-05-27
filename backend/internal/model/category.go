package model

import "time"

// Category 商品分类模型
type Category struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Name      string     `gorm:"size:50;not null" json:"name"`
	ParentID  *uint      `json:"parent_id"`
	Parent    *Category  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children  []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	SortOrder int        `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
