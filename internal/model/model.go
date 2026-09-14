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

// ProductStatus is a product lifecycle state stored as an int enum.
//   - 1 = active (在售)
//   - 2 = sold (已售出)
//   - 3 = offline (已下架)
type ProductStatus int

const (
	ProductStatusActive  ProductStatus = 1 // active, listed for sale
	ProductStatusSold    ProductStatus = 2 // sold
	ProductStatusOffline ProductStatus = 3 // offline / delisted
)

// Valid reports whether s is one of the known product statuses.
func (s ProductStatus) Valid() bool {
	switch s {
	case ProductStatusActive, ProductStatusSold, ProductStatusOffline:
		return true
	}
	return false
}

type Product struct {
	ID            uint           `json:"id"`
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	Price         float64        `json:"price"`
	OriginalPrice *float64       `json:"original_price,omitempty"`
	Status        ProductStatus  `json:"status"`
	Stock         int            `json:"stock"`
	Sales         int            `json:"sales"`
	IsFeatured    bool           `json:"is_featured"`
	CategoryID    *uint          `json:"category_id"`
	Category      *Category      `json:"category,omitempty"`
	SellerID      uint           `json:"seller_id"`
	Rating        *RatingSummary `json:"rating,omitempty"`
	Images        []ProductImage `json:"images,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// RatingSummary is the aggregate rating for a product, computed from
// top-level reviews that carry a 1-5 star rating.
type RatingSummary struct {
	Average float64 `json:"average"`
	Count   int     `json:"count"`
}

type ProductImage struct {
	ID        uint   `json:"id"`
	ProductID uint   `json:"product_id"`
	URL       string `json:"url"`
	SortOrder int    `json:"sort_order"`
	IsCover   bool   `json:"is_cover"`
}

// Comment is a user review or reply on a product. Top-level reviews may
// carry a 1-5 star Rating; replies (ParentID != nil) are text-only.
type Comment struct {
	ID        uint       `json:"id"`
	ProductID uint       `json:"product_id"`
	UserID    uint       `json:"user_id"`
	UserNick  string     `json:"user_nick,omitempty"`
	Content   string     `json:"content"`
	Rating    *int       `json:"rating,omitempty"`
	ParentID  *uint      `json:"parent_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Replies   []*Comment `json:"replies,omitempty"`
}
