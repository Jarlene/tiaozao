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

// User 用户模型（简要）
type User struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

// Order status constants (state machine)
const (
	OrderStatusPending   = "pending"   // 待付款
	OrderStatusPaid      = "paid"      // 待发货（已付款）
	OrderStatusShipped   = "shipped"   // 待收货（已发货）
	OrderStatusReceived  = "received"  // 已收货（等待完成）
	OrderStatusCompleted = "completed" // 已完成
	OrderStatusCancelled = "cancelled" // 已取消
	OrderStatusRefunding = "refunding" // 退款中
	OrderStatusRefunded  = "refunded"  // 已退款
)

// AllowedTransitions defines valid state transitions
var AllowedTransitions = map[string][]string{
	OrderStatusPending:   {OrderStatusPaid, OrderStatusCancelled},
	OrderStatusPaid:      {OrderStatusShipped, OrderStatusRefunding},
	OrderStatusShipped:   {OrderStatusReceived, OrderStatusRefunding},
	OrderStatusReceived:  {OrderStatusCompleted, OrderStatusRefunding},
	OrderStatusCompleted: {},
	OrderStatusCancelled: {},
	OrderStatusRefunding: {OrderStatusRefunded},
	OrderStatusRefunded:  {},
}

// CanTransition checks if a status transition is valid
func CanTransition(from, to string) bool {
	for _, allowed := range AllowedTransitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}

// TransitionDisplayName returns the Chinese display name for a status
func TransitionDisplayName(status string) string {
	names := map[string]string{
		OrderStatusPending:   "待付款",
		OrderStatusPaid:      "待发货",
		OrderStatusShipped:   "待收货",
		OrderStatusReceived:  "已完成",
		OrderStatusCompleted: "已完成",
		OrderStatusCancelled: "已取消",
		OrderStatusRefunding: "退款中",
		OrderStatusRefunded:  "已退款",
	}
	if name, ok := names[status]; ok {
		return name
	}
	return status
}

// Order 订单模型
type Order struct {
	ID             uint       `json:"id"`
	OrderNo        string     `json:"order_no"`
	BuyerID        uint       `json:"buyer_id"`
	SellerID       uint       `json:"seller_id"`
	TotalAmount    float64    `json:"total_amount"`
	DiscountAmount float64    `json:"discount_amount"`
	PayAmount      float64    `json:"pay_amount"`
	Status         string     `json:"status"`
	AddressID      uint       `json:"address_id"`
	Remark         string     `json:"remark"`
	PaidAt         *time.Time `json:"paid_at,omitempty"`
	ShippedAt      *time.Time `json:"shipped_at,omitempty"`
	ReceivedAt     *time.Time `json:"received_at,omitempty"`
	ClosedAt       *time.Time `json:"closed_at,omitempty"`
	Items          []OrderItem  `json:"items,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

// OrderItem 订单明细
type OrderItem struct {
	ID           uint    `json:"id"`
	OrderID      uint    `json:"order_id"`
	ProductID    uint    `json:"product_id"`
	ProductTitle string  `json:"product_title"`
	ProductImage string  `json:"product_image,omitempty"`
	Price        float64 `json:"price"`
	Quantity     int     `json:"quantity"`
	Subtotal     float64 `json:"subtotal"`
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	AddressID uint              `json:"address_id"`
	Items     []CreateOrderItem `json:"items"`
}

// CreateOrderItem 创建订单明细请求
type CreateOrderItem struct {
	ProductID uint `json:"product_id"`
	Quantity  int  `json:"quantity"`
}

// UpdateOrderStatusRequest 更新订单状态请求
type UpdateOrderStatusRequest struct {
	Action string `json:"action"` // pay, ship, receive, cancel, refund
}
