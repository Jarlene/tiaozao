package model

import "time"

// Order 订单模型
type Order struct {
	ID         uint        `gorm:"primaryKey" json:"id"`
	OrderNo    string      `gorm:"uniqueIndex;size:50;not null" json:"order_no"`
	BuyerID    uint        `gorm:"index;not null" json:"buyer_id"`
	SellerID   uint        `gorm:"index;not null" json:"seller_id"`
	TotalPrice float64     `gorm:"type:decimal(10,2);not null" json:"total_price"`
	Status     string      `gorm:"size:20;default:pending" json:"status"` // pending, paid, shipped, received, completed, cancelled, refunding
	AddressID  uint        `json:"address_id"`
	Remark     string      `gorm:"size:500" json:"remark"`
	Items      []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	Buyer      User        `gorm:"foreignKey:BuyerID" json:"buyer,omitempty"`
	Seller     User        `gorm:"foreignKey:SellerID" json:"seller,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// OrderItem 订单明细模型
type OrderItem struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	OrderID   uint    `gorm:"index;not null" json:"order_id"`
	ProductID uint    `gorm:"not null" json:"product_id"`
	Price     float64 `gorm:"type:decimal(10,2);not null" json:"price"`
	Product   Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}
