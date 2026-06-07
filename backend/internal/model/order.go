package model

import (
	"time"

	"gorm.io/gorm"
)

// OrderStatus 订单状态
type OrderStatus int

const (
	OrderStatusPendingPayment  OrderStatus = 1  // 待付款
	OrderStatusCancelled       OrderStatus = 2  // 已取消
	OrderStatusPaid            OrderStatus = 3  // 已付款（资金托管）
	OrderStatusPendingShipment OrderStatus = 4  // 待发货
	OrderStatusShipped         OrderStatus = 5  // 已发货
	OrderStatusReceived        OrderStatus = 6  // 已收货（资金释放）
	OrderStatusCompleted       OrderStatus = 7  // 已完成
	OrderStatusRefunding       OrderStatus = 8  // 退款中
	OrderStatusDispute         OrderStatus = 9  // 纠纷处理
)

// OrderStatusNames 状态中文映射（用于操作日志）
var OrderStatusNames = map[OrderStatus]string{
	OrderStatusPendingPayment:  "待付款",
	OrderStatusCancelled:       "已取消",
	OrderStatusPaid:            "已付款",
	OrderStatusPendingShipment: "待发货",
	OrderStatusShipped:         "已发货",
	OrderStatusReceived:        "已收货",
	OrderStatusCompleted:       "已完成",
	OrderStatusRefunding:       "退款中",
	OrderStatusDispute:         "纠纷处理",
}

type Order struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	OrderNo   string         `gorm:"size:64;uniqueIndex;not null" json:"order_no"`
	ProductID uint           `gorm:"not null;index" json:"product_id"`
	BuyerID   uint           `gorm:"not null;index" json:"buyer_id"`
	SellerID  uint           `gorm:"not null;index" json:"seller_id"`
	Title            string         `gorm:"size:200;not null" json:"title"`              // 快照：下单时商品标题
	Price            int64          `gorm:"not null" json:"price"`                       // 快照：下单时商品价格（分）
	Status           OrderStatus    `gorm:"default:1;not null;index" json:"status"`
	ShippingAddress  string         `gorm:"type:text;not null" json:"shipping_address"`  // 收货地址
	BuyerNote        string         `gorm:"type:text" json:"buyer_note,omitempty"`       // 买家留言`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Buyer   *User    `gorm:"foreignKey:BuyerID" json:"buyer,omitempty"`
	Seller  *User    `gorm:"foreignKey:SellerID" json:"seller,omitempty"`
}

func (Order) TableName() string {
	return "orders"
}

// OrderStatusLog 订单状态变更日志
type OrderStatusLog struct {
	ID           uint        `gorm:"primarykey" json:"id"`
	OrderID      uint        `gorm:"not null;index" json:"order_id"`
	FromStatus   OrderStatus `gorm:"not null" json:"from_status"`
	ToStatus     OrderStatus `gorm:"not null" json:"to_status"`
	Event        string      `gorm:"size:50;not null" json:"event"`
	OperatorID   uint        `gorm:"not null" json:"operator_id"`
	OperatorType string      `gorm:"size:20;not null" json:"operator_type"` // buyer, seller, admin, system
	CreatedAt    time.Time   `json:"created_at"`
}

func (OrderStatusLog) TableName() string {
	return "order_status_logs"
}
