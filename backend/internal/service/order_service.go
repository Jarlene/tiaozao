package service

import (
	"context"
	stdErrors "errors"
	"fmt"
	"time"

	"flea-market/internal/model"
	"flea-market/internal/repository"
	"flea-market/pkg/errors"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type OrderService struct {
	orderRepo   repository.OrderRepository
	productRepo repository.ProductRepository
	fsm         *OrderFSM
	rdb         *redis.Client
}

func NewOrderService(orderRepo repository.OrderRepository, productRepo repository.ProductRepository, rdb *redis.Client) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		fsm:         NewOrderFSM(),
		rdb:         rdb,
	}
}

type CreateOrderReq struct {
	ProductID       uint   `json:"product_id" binding:"required"`
	ShippingAddress string `json:"shipping_address" binding:"required,max=500"`
	BuyerNote       string `json:"buyer_note" binding:"max=200"`
}

type OrderDetail struct {
	ID              uint                  `json:"id"`
	OrderNo         string                `json:"order_no"`
	ProductID       uint                  `json:"product_id"`
	BuyerID         uint                  `json:"buyer_id"`
	SellerID        uint                  `json:"seller_id"`
	Title           string                `json:"title"`
	Price           int64                 `json:"price"`
	Status          model.OrderStatus     `json:"status"`
	StatusName      string                `json:"status_name"`
	ShippingAddress string                `json:"shipping_address"`
	BuyerNote       string                `json:"buyer_note,omitempty"`
	CreatedAt       string                `json:"created_at"`
	UpdatedAt       string                `json:"updated_at"`
	Product         *ProductListItem      `json:"product,omitempty"`
}

type OrderListItem struct {
	ID         uint              `json:"id"`
	OrderNo    string            `json:"order_no"`
	ProductID  uint              `json:"product_id"`
	Title      string            `json:"title"`
	Price      int64             `json:"price"`
	Status     model.OrderStatus `json:"status"`
	StatusName string            `json:"status_name"`
	CreatedAt  string            `json:"created_at"`
}

// CreateOrder 买家创建订单（使用 Redis 分布式锁防止并发冲突）
func (s *OrderService) CreateOrder(buyerID uint, req *CreateOrderReq) (*OrderDetail, int, error) {
	lockKey := fmt.Sprintf("order:product:%d:lock", req.ProductID)
	ctx := context.Background()

	// 获取 Redis 分布式锁（5秒超时，防止死锁）
	locked, err := s.rdb.SetNX(ctx, lockKey, "1", 5*time.Second).Result()
	if err != nil {
		return nil, errors.ErrInternal, err
	}
	if !locked {
		return nil, errors.ErrBadRequest, fmt.Errorf("操作太频繁，请稍后重试")
	}
	defer s.rdb.Del(ctx, lockKey)

	product, err := s.productRepo.FindByID(req.ProductID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrProductNotFound, nil
		}
		return nil, errors.ErrInternal, err
	}

	// 不能买自己的商品
	if product.UserID == buyerID {
		return nil, errors.ErrForbidden, nil
	}

	// 商品必须在售
	if product.Status != model.ProductStatusActive {
		return nil, errors.ErrBadRequest, nil
	}

	orderNo := fmt.Sprintf("%s%08d", time.Now().Format("20060102150405"), product.ID)
	if len(orderNo) > 64 {
		orderNo = orderNo[:64]
	}

	order := &model.Order{
		OrderNo:         orderNo,
		ProductID:       product.ID,
		BuyerID:         buyerID,
		SellerID:        product.UserID,
		Title:           product.Title,
		Price:           product.Price,
		Status:          model.OrderStatusPendingPayment,
		ShippingAddress: req.ShippingAddress,
		BuyerNote:       req.BuyerNote,
	}

	err = s.orderRepo.Transaction(func(txRepo repository.OrderRepository) error {
		// 原子性扣减库存（UPDATE ... SET stock = stock - 1 WHERE stock >= 1）
		ok, err := txRepo.DecrementProductStock(product.ID, 1)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("insufficient stock")
		}

		if err := txRepo.Create(order); err != nil {
			return err
		}

		return txRepo.CreateStatusLog(&model.OrderStatusLog{
			OrderID:      order.ID,
			FromStatus:   model.OrderStatusPendingPayment,
			ToStatus:     model.OrderStatusPendingPayment,
			Event:        "create",
			OperatorID:   buyerID,
			OperatorType: "buyer",
		})
	})
	if err != nil {
		if err.Error() == "insufficient stock" {
			return nil, errors.ErrInsufficientStock, nil
		}
		return nil, errors.ErrInternal, err
	}

	return s.detailFromModel(order), errors.Success, nil
}

// Cancel 买家取消订单（仅待付款状态，恢复库存）
func (s *OrderService) Cancel(userID, orderID uint) (int, error) {
	order, code, err := s.getOrderForBuyer(orderID, userID)
	if err != nil {
		return code, err
	}

	newStatus, err := s.fsm.Transition(order.Status, OrderEventCancel)
	if err != nil {
		return errors.ErrBadRequest, nil
	}

	if err := s.orderRepo.Transaction(func(txRepo repository.OrderRepository) error {
		originalStatus := order.Status
		order.Status = newStatus
		if err := txRepo.Update(order); err != nil {
			return err
		}
		// 取消订单，恢复库存
		if err := txRepo.IncrementProductStock(order.ProductID, 1); err != nil {
			return err
		}
		return txRepo.CreateStatusLog(&model.OrderStatusLog{
			OrderID:      order.ID,
			FromStatus:   originalStatus,
			ToStatus:     newStatus,
			Event:        string(OrderEventCancel),
			OperatorID:   userID,
			OperatorType: "buyer",
		})
	}); err != nil {
		return errors.ErrInternal, err
	}

	return errors.Success, nil
}

// Pay 买家付款（待付款 → 已付款 → 待发货）
// 付款成功后自动触发"通知卖家"转换
func (s *OrderService) Pay(userID, orderID uint) (int, error) {
	order, code, err := s.getOrderForBuyer(orderID, userID)
	if err != nil {
		return code, err
	}

	newStatus, err := s.fsm.Transition(order.Status, OrderEventPay)
	if err != nil {
		return errors.ErrBadRequest, nil
	}

	if err := s.transitionAndLog(order, newStatus, OrderEventPay, userID, "buyer"); err != nil {
		return errors.ErrInternal, err
	}

	// 自动通知卖家：已付款 → 待发货
	notifyStatus, _ := s.fsm.Transition(newStatus, OrderEventNotifySeller)
	if err := s.transitionAndLog(order, notifyStatus, OrderEventNotifySeller, userID, "system"); err != nil {
		return errors.ErrInternal, err
	}

	return errors.Success, nil
}

// Ship 卖家发货
func (s *OrderService) Ship(userID, orderID uint) (int, error) {
	order, code, err := s.getOrderForSeller(orderID, userID)
	if err != nil {
		return code, err
	}

	newStatus, err := s.fsm.Transition(order.Status, OrderEventShip)
	if err != nil {
		return errors.ErrBadRequest, nil
	}

	if err := s.transitionAndLog(order, newStatus, OrderEventShip, userID, "seller"); err != nil {
		return errors.ErrInternal, err
	}

	return errors.Success, nil
}

// ConfirmReceive 买家确认收货
func (s *OrderService) ConfirmReceive(userID, orderID uint) (int, error) {
	order, code, err := s.getOrderForBuyer(orderID, userID)
	if err != nil {
		return code, err
	}

	newStatus, err := s.fsm.Transition(order.Status, OrderEventConfirmReceive)
	if err != nil {
		return errors.ErrBadRequest, nil
	}

	if err := s.transitionAndLog(order, newStatus, OrderEventConfirmReceive, userID, "buyer"); err != nil {
		return errors.ErrInternal, err
	}

	return errors.Success, nil
}

// RequestRefund 买家申请退款
func (s *OrderService) RequestRefund(userID, orderID uint) (int, error) {
	order, code, err := s.getOrderForBuyer(orderID, userID)
	if err != nil {
		return code, err
	}

	newStatus, err := s.fsm.Transition(order.Status, OrderEventRequestRefund)
	if err != nil {
		return errors.ErrBadRequest, nil
	}

	if err := s.transitionAndLog(order, newStatus, OrderEventRequestRefund, userID, "buyer"); err != nil {
		return errors.ErrInternal, err
	}

	return errors.Success, nil
}

// CompleteRefund 退款成功（系统/管理员操作）
func (s *OrderService) CompleteRefund(operatorID, orderID uint) (int, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrOrderNotFound, nil
		}
		return errors.ErrInternal, err
	}

	newStatus, err := s.fsm.Transition(order.Status, OrderEventRefundSuccess)
	if err != nil {
		return errors.ErrBadRequest, nil
	}

	if err := s.transitionAndLog(order, newStatus, OrderEventRefundSuccess, operatorID, "admin"); err != nil {
		return errors.ErrInternal, err
	}

	return errors.Success, nil
}

// RaiseDispute 买家发起纠纷
func (s *OrderService) RaiseDispute(userID, orderID uint) (int, error) {
	order, code, err := s.getOrderForBuyer(orderID, userID)
	if err != nil {
		return code, err
	}

	newStatus, err := s.fsm.Transition(order.Status, OrderEventRaiseDispute)
	if err != nil {
		return errors.ErrBadRequest, nil
	}

	if err := s.transitionAndLog(order, newStatus, OrderEventRaiseDispute, userID, "buyer"); err != nil {
		return errors.ErrInternal, err
	}

	return errors.Success, nil
}

// Arbitrate 管理员仲裁
func (s *OrderService) Arbitrate(adminID, orderID uint) (int, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrOrderNotFound, nil
		}
		return errors.ErrInternal, err
	}

	newStatus, err := s.fsm.Transition(order.Status, OrderEventArbitrate)
	if err != nil {
		return errors.ErrBadRequest, nil
	}

	if err := s.transitionAndLog(order, newStatus, OrderEventArbitrate, adminID, "admin"); err != nil {
		return errors.ErrInternal, err
	}

	return errors.Success, nil
}

// GetByID 获取订单详情
func (s *OrderService) GetByID(userID, orderID uint) (*OrderDetail, int, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrOrderNotFound, nil
		}
		return nil, errors.ErrInternal, err
	}

	// 仅订单参与方（买家/卖家）可查看
	if order.BuyerID != userID && order.SellerID != userID {
		return nil, errors.ErrForbidden, nil
	}

	return s.detailFromModel(order), errors.Success, nil
}

// ListByBuyer 买家的订单列表
func (s *OrderService) ListByBuyer(userID uint, status *model.OrderStatus, page, pageSize int) (*PaginatedResult, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	orders, total, err := s.orderRepo.ListByBuyer(userID, status, page, pageSize)
	if err != nil {
		return nil, errors.ErrInternal, err
	}

	items := make([]OrderListItem, len(orders))
	for i, o := range orders {
		items[i] = s.listItemFromModel(&o)
	}

	return &PaginatedResult{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: calcTotalPages(total, pageSize),
	}, errors.Success, nil
}

// ListBySeller 卖家的订单列表
func (s *OrderService) ListBySeller(userID uint, status *model.OrderStatus, page, pageSize int) (*PaginatedResult, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	orders, total, err := s.orderRepo.ListBySeller(userID, status, page, pageSize)
	if err != nil {
		return nil, errors.ErrInternal, err
	}

	items := make([]OrderListItem, len(orders))
	for i, o := range orders {
		items[i] = s.listItemFromModel(&o)
	}

	return &PaginatedResult{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: calcTotalPages(total, pageSize),
	}, errors.Success, nil
}

// GetStatusLogs 获取订单状态变更日志
func (s *OrderService) GetStatusLogs(userID, orderID uint) ([]model.OrderStatusLog, int, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrOrderNotFound, nil
		}
		return nil, errors.ErrInternal, err
	}

	if order.BuyerID != userID && order.SellerID != userID {
		return nil, errors.ErrForbidden, nil
	}

	logs, err := s.orderRepo.ListStatusLogs(orderID)
	if err != nil {
		return nil, errors.ErrInternal, err
	}

	return logs, errors.Success, nil
}

// 内部辅助方法

func (s *OrderService) getOrderForBuyer(orderID, buyerID uint) (*model.Order, int, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrOrderNotFound, nil
		}
		return nil, errors.ErrInternal, err
	}
	if order.BuyerID != buyerID {
		return nil, errors.ErrForbidden, nil
	}
	return order, errors.Success, nil
}

func (s *OrderService) getOrderForSeller(orderID, sellerID uint) (*model.Order, int, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrOrderNotFound, nil
		}
		return nil, errors.ErrInternal, err
	}
	if order.SellerID != sellerID {
		return nil, errors.ErrForbidden, nil
	}
	return order, errors.Success, nil
}

func (s *OrderService) transitionAndLog(order *model.Order, target model.OrderStatus, event OrderEvent, operatorID uint, operatorType string) error {
	return s.orderRepo.Transaction(func(txRepo repository.OrderRepository) error {
		from := order.Status
		order.Status = target
		if err := txRepo.Update(order); err != nil {
			return err
		}
		return s.logStatusChange(txRepo, order.ID, from, target, string(event), operatorID, operatorType)
	})
}

func (s *OrderService) logStatusChange(txRepo repository.OrderRepository, orderID uint, from, to model.OrderStatus, event string, operatorID uint, operatorType string) error {
	return txRepo.CreateStatusLog(&model.OrderStatusLog{
		OrderID:      orderID,
		FromStatus:   from,
		ToStatus:     to,
		Event:        event,
		OperatorID:   operatorID,
		OperatorType: operatorType,
	})
}

func (s *OrderService) detailFromModel(o *model.Order) *OrderDetail {
	detail := &OrderDetail{
		ID:              o.ID,
		OrderNo:         o.OrderNo,
		ProductID:       o.ProductID,
		BuyerID:         o.BuyerID,
		SellerID:        o.SellerID,
		Title:           o.Title,
		Price:           o.Price,
		Status:          o.Status,
		StatusName:      model.OrderStatusNames[o.Status],
		ShippingAddress: o.ShippingAddress,
		BuyerNote:       o.BuyerNote,
		CreatedAt:       o.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       o.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if o.Product != nil && o.Product.ID > 0 {
		detail.Product = &ProductListItem{
			ID:    o.Product.ID,
			Title: o.Product.Title,
			Price: o.Product.Price,
		}
	}

	return detail
}

func (s *OrderService) listItemFromModel(o *model.Order) OrderListItem {
	return OrderListItem{
		ID:         o.ID,
		OrderNo:    o.OrderNo,
		ProductID:  o.ProductID,
		Title:      o.Title,
		Price:      o.Price,
		Status:     o.Status,
		StatusName: model.OrderStatusNames[o.Status],
		CreatedAt:  o.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
