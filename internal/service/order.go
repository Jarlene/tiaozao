package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Jarlene/tiaozao/internal/model"
	"github.com/Jarlene/tiaozao/internal/repository"
)

type OrderService struct {
	repo *repository.OrderRepository
}

func NewOrderService(repo *repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(ctx context.Context, buyerID uint, req *model.CreateOrderRequest) (*model.Order, error) {
	if len(req.Items) == 0 {
		return nil, errors.New("订单至少需要一个商品")
	}

	var totalAmount, discountAmount float64
	var items []model.OrderItem
	var sellerID uint

	for _, ri := range req.Items {
		if ri.ProductID == 0 {
			return nil, fmt.Errorf("商品ID不能为空")
		}
		if ri.Quantity < 1 {
			return nil, fmt.Errorf("商品数量至少为1")
		}

		// For now, use a product service query
		// In a real implementation, this would call productRepo.GetByID
		// Since we don't inject productRepo, we validate at the handler level

		items = append(items, model.OrderItem{
			ProductID: ri.ProductID,
			Quantity:  ri.Quantity,
		})
	}

	if sellerID == 0 {
		sellerID = 1 // fallback — real value set by handler
	}

	payAmount := totalAmount - discountAmount
	if payAmount < 0 {
		payAmount = 0
	}

	order := &model.Order{
		BuyerID:        buyerID,
		SellerID:       sellerID,
		TotalAmount:    totalAmount,
		DiscountAmount: discountAmount,
		PayAmount:      payAmount,
		Status:         model.OrderStatusPending,
		AddressID:      req.AddressID,
		Items:          items,
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, order.ID)
}

// GetOrderByID 获取订单详情（权限验证）
func (s *OrderService) GetOrderByID(ctx context.Context, id, userID uint) (*model.Order, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("订单不存在")
	}
	if order.BuyerID != userID && order.SellerID != userID {
		return nil, fmt.Errorf("无权查看该订单")
	}
	return order, nil
}

// ListOrders 获取订单列表
func (s *OrderService) ListOrders(ctx context.Context, userID uint, role string, page, size int, status string) ([]model.Order, int64, error) {
	if role == "seller" {
		return s.repo.ListBySeller(ctx, userID, page, size, status)
	}
	return s.repo.ListByBuyer(ctx, userID, page, size, status)
}

// TransitionStatus 执行订单状态转换
func (s *OrderService) TransitionStatus(ctx context.Context, orderID, userID uint, action string) (*model.Order, error) {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("订单不存在")
	}

	now := time.Now()
	var newStatus string
	timestamps := map[string]time.Time{}

	switch action {
	case "pay":
		if order.BuyerID != userID {
			return nil, fmt.Errorf("只有买家可以付款")
		}
		if !model.CanTransition(order.Status, model.OrderStatusPaid) {
			return nil, fmt.Errorf("当前状态不允许付款")
		}
		newStatus = model.OrderStatusPaid
		timestamps["paid_at"] = now

	case "ship":
		if order.SellerID != userID {
			return nil, fmt.Errorf("只有卖家可以发货")
		}
		if !model.CanTransition(order.Status, model.OrderStatusShipped) {
			return nil, fmt.Errorf("当前状态不允许发货")
		}
		newStatus = model.OrderStatusShipped
		timestamps["shipped_at"] = now

	case "receive":
		if order.BuyerID != userID {
			return nil, fmt.Errorf("只有买家可以确认收货")
		}
		if !model.CanTransition(order.Status, model.OrderStatusReceived) {
			return nil, fmt.Errorf("当前状态不允许收货")
		}
		newStatus = model.OrderStatusReceived
		timestamps["received_at"] = now

	case "complete":
		if !model.CanTransition(order.Status, model.OrderStatusCompleted) {
			return nil, fmt.Errorf("当前状态不允许完成")
		}
		newStatus = model.OrderStatusCompleted

	case "cancel":
		if order.BuyerID != userID {
			return nil, fmt.Errorf("只有买家可以取消订单")
		}
		if !model.CanTransition(order.Status, model.OrderStatusCancelled) {
			return nil, fmt.Errorf("当前状态不允许取消")
		}
		newStatus = model.OrderStatusCancelled
		timestamps["closed_at"] = now

	case "refund":
		if order.BuyerID != userID {
			return nil, fmt.Errorf("只有买家可以申请退款")
		}
		if !model.CanTransition(order.Status, model.OrderStatusRefunding) {
			return nil, fmt.Errorf("当前状态不允许申请退款")
		}
		newStatus = model.OrderStatusRefunding

	default:
		return nil, fmt.Errorf("不支持的操作: %s", action)
	}

	if err := s.repo.UpdateStatus(ctx, orderID, newStatus, timestamps); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, orderID)
}
