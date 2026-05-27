package model

import "time"

type Category struct {
	ID        uint       `json:"id"`
	Name      string     `json:"name"`
	ParentID  *uint      `json:"parent_id"`
	SortOrder int        `json:"sort_order"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Children  []Category `json:"children,omitempty"`
}

type ProductStatus string

const (
	ProductStatusDraft    ProductStatus = "draft"
	ProductStatusActive   ProductStatus = "active"
	ProductStatusInactive ProductStatus = "inactive"
)

type Product struct {
	ID            uint           `json:"id"`
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	Price         float64        `json:"price"`
	OriginalPrice *float64       `json:"original_price,omitempty"`
	Status        ProductStatus  `json:"status"`
	CategoryID    *uint          `json:"category_id"`
	Category      *Category      `json:"category,omitempty"`
	SellerID      uint           `json:"seller_id"`
	Images        []ProductImage `json:"images,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type ProductImage struct {
	ID        uint   `json:"id"`
	ProductID uint   `json:"product_id"`
	URL       string `json:"url"`
	SortOrder int    `json:"sort_order"`
	IsCover   bool   `json:"is_cover"`
}
