package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/Jarlene/tiaozao/backend/internal/model"
	"github.com/Jarlene/tiaozao/backend/internal/repository"
	"gorm.io/gorm"
)

// OrderService 订单业务逻辑
type OrderService struct {
	orderRepo   *repository.OrderRepository
	productRepo *repository.ProductRepository
	userRepo    *repository.UserRepository
}

// NewOrderService 创建订单服务实例
func NewOrderService(
	orderRepo *repository.OrderRepository,
	productRepo *repository.ProductRepository,
	userRepo *repository.UserRepository,
) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		userRepo:    userRepo,
	}
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	ProductID uint   `json:"product_id" binding:"required"`
	AddressID uint   `json:"address_id"`
	Remark    string `json:"remark"`
}

// CreateOrderResponse 创建订单响应
type CreateOrderResponse struct {
	OrderNo string `json:"order_no"`
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(buyerID uint, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	// 查询商品
	product, err := s.productRepo.FindByID(req.ProductID)
	if err != nil {
		return nil, errors.New("商品不存在")
	}

	// 不能购买自己的商品
	if product.UserID == buyerID {
		return nil, errors.New("不能购买自己的商品")
	}

	// 检查商品状态
	if product.Status != 1 {
		return nil, errors.New("商品已下架或已售出")
	}

	// 生成订单号
	orderNo := generateOrderNo()

	order := &model.Order{
		OrderNo:    orderNo,
		BuyerID:    buyerID,
		SellerID:   product.UserID,
		TotalPrice: product.Price,
		Status:     "pending",
		AddressID:  req.AddressID,
		Remark:     req.Remark,
		Items: []model.OrderItem{
			{
				ProductID: product.ID,
				Price:     product.Price,
			},
		},
	}

	if err := s.orderRepo.Create(order); err != nil {
		return nil, errors.New("创建订单失败")
	}

	// 更新商品状态为已售出
	product.Status = 2
	if err := s.productRepo.Update(product); err != nil {
		return nil, errors.New("更新商品状态失败")
	}

	return &CreateOrderResponse{OrderNo: orderNo}, nil
}

// GetOrderByID 获取订单详情
func (s *OrderService) GetOrderByID(orderID, userID uint) (*model.Order, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("订单不存在")
		}
		return nil, errors.New("获取订单信息失败")
	}

	// 只有买家或卖家可以查看
	if order.BuyerID != userID && order.SellerID != userID {
		return nil, errors.New("无权查看此订单")
	}

	return order, nil
}

// GetBuyerOrders 获取买家订单列表
func (s *OrderService) GetBuyerOrders(buyerID uint, page, pageSize int, status string) ([]model.Order, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.orderRepo.ListByBuyer(buyerID, page, pageSize, status)
}

// GetSellerOrders 获取卖家订单列表
func (s *OrderService) GetSellerOrders(sellerID uint, page, pageSize int, status string) ([]model.Order, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.orderRepo.ListBySeller(sellerID, page, pageSize, status)
}

// validTransitions 订单状态转换规则
var validTransitions = map[string][]string{
	"pending":   {"paid", "cancelled"},
	"paid":      {"shipped", "cancelled", "refunding"},
	"shipped":   {"received"},
	"received":  {"completed", "refunding"},
	"cancelled": {},
	"completed": {},
	"refunding": {"completed"},
}

// UpdateOrderStatus 更新订单状态
func (s *OrderService) UpdateOrderStatus(orderID, userID uint, newStatus string) (*model.Order, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, errors.New("订单不存在")
	}

	// 验证状态转换
	allowed, ok := validTransitions[order.Status]
	if !ok {
		return nil, errors.New("当前订单状态异常")
	}

	valid := false
	for _, s := range allowed {
		if s == newStatus {
			valid = true
			break
		}
	}
	if !valid {
		return nil, fmt.Errorf("订单不能从 %s 变更为 %s", order.Status, newStatus)
	}

	// 根据状态变更验证权限
	switch newStatus {
	case "paid", "completed":
		if order.BuyerID != userID {
			return nil, errors.New("只有买家可以执行此操作")
		}
	case "shipped":
		if order.SellerID != userID {
			return nil, errors.New("只有卖家可以执行此操作")
		}
	case "cancelled":
		if order.BuyerID != userID && order.SellerID != userID {
			return nil, errors.New("无权取消此订单")
		}
	case "received":
		if order.BuyerID != userID {
			return nil, errors.New("只有买家可以确认收货")
		}
	}

	order.Status = newStatus
	now := time.Now()
	_ = now // 后续扩展可记录时间线

	if err := s.orderRepo.Update(order); err != nil {
		return nil, errors.New("更新订单状态失败")
	}

	return order, nil
}
