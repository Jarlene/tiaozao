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

const AdminRole = 2 // 管理员角色

type OrderService struct {
	orderRepo   repository.OrderRepository
	productRepo repository.ProductRepository
	userRepo    repository.UserRepository
	fsm         *OrderFSM
	rdb         *redis.Client
}

func NewOrderService(orderRepo repository.OrderRepository, productRepo repository.ProductRepository, userRepo repository.UserRepository, rdb *redis.Client) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		userRepo:    userRepo,
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
	ID               uint                  `json:"id"`
	OrderNo          string                `json:"order_no"`
	ProductID        uint                  `json:"product_id"`
	BuyerID          uint                  `json:"buyer_id"`
	SellerID         uint                  `json:"seller_id"`
	Title            string                `json:"title"`
	Price            int64                 `json:"price"`
	Status           model.OrderStatus     `json:"status"`
	StatusName       string                `json:"status_name"`
	PreRefundStatus  model.OrderStatus     `json:"pre_refund_status,omitempty"`
	RefundReason     string                `json:"refund_reason,omitempty"`
	TrackingNumber   string                `json:"tracking_number,omitempty"`
	LogisticsCompany string                `json:"logistics_company,omitempty"`
	ShippingAddress  string                `json:"shipping_address"`
	BuyerNote        string                `json:"buyer_note,omitempty"`
	CreatedAt        string                `json:"created_at"`
	UpdatedAt        string                `json:"updated_at"`
	Product          *ProductListItem      `json:"product,omitempty"`
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
	Product    *ProductListItem  `json:"product,omitempty"`
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
	if code != errors.Success {
		return code, nil
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
// 执行模拟资金托管：从买家账户扣减余额，冻结为托管资金
func (s *OrderService) Pay(userID, orderID uint) (int, error) {
	order, code, err := s.getOrderForBuyer(orderID, userID)
	if err != nil {
		return code, err
	}
	if code != errors.Success {
		return code, nil
	}

	newStatus, err := s.fsm.Transition(order.Status, OrderEventPay)
	if err != nil {
		return errors.ErrBadRequest, nil
	}

	if err := s.orderRepo.Transaction(func(txRepo repository.OrderRepository) error {
		tx := txRepo.GetDB()

		// 1. 更新订单状态为已付款
		fromStatus := order.Status
		order.Status = newStatus
		if err := txRepo.Update(order); err != nil {
			return err
		}
		if err := s.logStatusChange(txRepo, order.ID, fromStatus, newStatus, string(OrderEventPay), userID, "buyer"); err != nil {
			return err
		}

		// 2. 在事务内读取买家最新余额，确保审计快照准确
		buyer, err := s.userRepo.FindByID(userID)
		if err != nil {
			return err
		}
		// 检查余额是否充足（防御性校验，UpdateBalance 内部也有原子校验）
		if buyer.Balance < order.Price {
			return repository.ErrBalanceInsufficient
		}

		// 3. 原子性扣减可用余额，增加冻结余额（资金进入托管）
		if err := s.userRepo.UpdateBalance(tx, userID, -order.Price, order.Price); err != nil {
			return err
		}

		// 4. 记录钱包交易流水
		wt := &model.WalletTransaction{
			UserID:        userID,
			Type:          model.WalletTxPay,
			Amount:        -order.Price,
			OrderID:       order.ID,
			BalanceBefore: buyer.Balance,
			BalanceAfter:  buyer.Balance - order.Price,
			FrozenBefore:  buyer.FrozenBalance,
			FrozenAfter:   buyer.FrozenBalance + order.Price,
			Description:   fmt.Sprintf("订单 %s 付款 %.2f 元", order.OrderNo, float64(order.Price)/100),
		}
		if err := s.userRepo.CreateWalletTransaction(tx, wt); err != nil {
			return err
		}

		// 5. 在同一事务中推进状态：Paid → PendingShipment（通知卖家）
		notifyStatus, fsmErr := s.fsm.Transition(newStatus, OrderEventNotifySeller)
		if fsmErr != nil {
			return fmt.Errorf("notify seller transition failed: %w", fsmErr)
		}
		fromNotify := order.Status
		order.Status = notifyStatus
		if err := txRepo.Update(order); err != nil {
			return err
		}
		if err := s.logStatusChange(txRepo, order.ID, fromNotify, notifyStatus, string(OrderEventNotifySeller), userID, "system"); err != nil {
			return err
		}

		return nil
	}); err != nil {
		if stdErrors.Is(err, repository.ErrBalanceInsufficient) {
			return errors.ErrInsufficientBalance, nil
		}
		return errors.ErrInternal, err
	}

	return errors.Success, nil
}

// Ship 卖家发货
func (s *OrderService) Ship(userID, orderID uint, req *ShipReq) (int, error) {
	order, code, err := s.getOrderForSeller(orderID, userID)
	if err != nil {
		return code, err
	}
	if code != errors.Success {
		return code, nil
	}

	newStatus, err := s.fsm.Transition(order.Status, OrderEventShip)
	if err != nil {
		return errors.ErrBadRequest, nil
	}

	// 保存物流信息
	if req != nil {
		order.TrackingNumber = req.TrackingNumber
		order.LogisticsCompany = req.LogisticsCompany
	}

	if err := s.transitionAndLog(order, newStatus, OrderEventShip, userID, "seller"); err != nil {
		return errors.ErrInternal, err
	}

	return errors.Success, nil
}

// ConfirmReceive 买家确认收货
// 流程：已发货(5) → 已收货(6) → 已完成(7) + 资金释放
// 从买家冻结余额释放托管资金到卖家可用余额
func (s *OrderService) ConfirmReceive(userID, orderID uint) (int, error) {
	order, code, err := s.getOrderForBuyer(orderID, userID)
	if err != nil {
		return code, err
	}
	if code != errors.Success {
		return code, nil
	}

	// 1. 校验状态机：Shipped → Received
	receivedStatus, err := s.fsm.Transition(order.Status, OrderEventConfirmReceive)
	if err != nil {
		return errors.ErrBadRequest, nil
	}

	// 2. 校验后续自动完成路径：Received → Completed
	completedStatus, err := s.fsm.Transition(receivedStatus, OrderEventComplete)
	if err != nil {
		return errors.ErrInternal, nil
	}

	if err := s.orderRepo.Transaction(func(txRepo repository.OrderRepository) error {
		tx := txRepo.GetDB()

		// 3a. 更新订单状态为已收货
		fromStatus := order.Status
		order.Status = receivedStatus
		if err := txRepo.Update(order); err != nil {
			return err
		}
		if err := s.logStatusChange(txRepo, order.ID, fromStatus, receivedStatus, string(OrderEventConfirmReceive), userID, "buyer"); err != nil {
			return err
		}

		// 3b. 资金释放：买家冻结余额扣减，卖家可用余额增加
		// UpdateBalance 内部有余额充足校验（frozen_balance + delta >= 0）
		if err := s.userRepo.UpdateBalance(tx, order.BuyerID, 0, -order.Price); err != nil {
			return err
		}
		if err := s.userRepo.UpdateBalance(tx, order.SellerID, order.Price, 0); err != nil {
			return err
		}

		// 3c. 在事务内读取卖家最新余额，确保审计快照准确
		seller, err := s.userRepo.FindByID(order.SellerID)
		if err != nil {
			return err
		}

		// 3d. 记录卖家收款流水（type=complete）
		wt := &model.WalletTransaction{
			UserID:        order.SellerID,
			Type:          model.WalletTxComplete,
			Amount:        order.Price,
			OrderID:       order.ID,
			BalanceBefore: seller.Balance,
			BalanceAfter:  seller.Balance + order.Price,
			FrozenBefore:  seller.FrozenBalance,
			FrozenAfter:   seller.FrozenBalance,
			Description:   fmt.Sprintf("订单 %s 收款 %.2f 元", order.OrderNo, float64(order.Price)/100),
		}
		if err := s.userRepo.CreateWalletTransaction(tx, wt); err != nil {
			return err
		}

		// 3e. 自动完成：已收货 → 已完成
		order.Status = completedStatus
		if err := txRepo.Update(order); err != nil {
			return err
		}
		if err := s.logStatusChange(txRepo, order.ID, receivedStatus, completedStatus, string(OrderEventComplete), userID, "system"); err != nil {
			return err
		}

		return nil
	}); err != nil {
		if stdErrors.Is(err, repository.ErrBalanceInsufficient) {
			return errors.ErrInsufficientBalance, nil
		}
		return errors.ErrInternal, err
	}

	return errors.Success, nil
}

// RequestRefundReq 退款申请请求
type RequestRefundReq struct {
	Reason string `json:"reason" binding:"required,max=500"`
}

// truncateEvent 截断事件描述到安全长度，防止数据库字段溢出
// OrderStatusLog.Event 字段为 size:500，留足余量
func truncateEvent(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen])
	}
	return s
}

// RequestRefund 买家申请退款
// 记录退款原因和退款前状态，用于后续卖家审批时的资金处理
func (s *OrderService) RequestRefund(userID, orderID uint, req *RequestRefundReq) (int, error) {
	order, code, err := s.getOrderForBuyer(orderID, userID)
	if err != nil {
		return code, err
	}
	if code != errors.Success {
		return code, nil
	}

	// 仅已发货或已收货状态可申请退款
	if order.Status != model.OrderStatusShipped && order.Status != model.OrderStatusReceived {
		return errors.ErrBadRequest, nil
	}

	newStatus, err := s.fsm.Transition(order.Status, OrderEventRequestRefund)
	if err != nil {
		return errors.ErrBadRequest, nil
	}

	if err := s.orderRepo.Transaction(func(txRepo repository.OrderRepository) error {
		fromStatus := order.Status
		// 记录退款前状态和退款原因，用于后续资金处理
		order.PreRefundStatus = fromStatus
		order.Status = newStatus
		if req != nil && req.Reason != "" {
			order.RefundReason = req.Reason
		}
		if err := txRepo.Update(order); err != nil {
			return err
		}

		// 记录退款申请日志，附上退款原因（截断到安全长度）
		eventDesc := string(OrderEventRequestRefund)
		if req != nil && req.Reason != "" {
			reason := truncateEvent(req.Reason, 450) // 保留预留长度给前缀
			eventDesc = eventDesc + ":" + reason
		}
		return s.logStatusChange(txRepo, order.ID, fromStatus, newStatus, eventDesc, userID, "buyer")
	}); err != nil {
		return errors.ErrInternal, err
	}

	return errors.Success, nil
}

// ApproveRefund 卖家同意退款
// 根据退款前状态决定资金处理方式：
//   - 从已发货退款：买家冻结余额 → 买家可用余额（解冻）
//   - 从已收货退款：卖家可用余额 → 买家可用余额（退款到账）
func (s *OrderService) ApproveRefund(sellerID, orderID uint) (int, error) {
	order, code, err := s.getOrderForSeller(orderID, sellerID)
	if err != nil {
		return code, err
	}
	if code != errors.Success {
		return code, nil
	}

	// 必须是退款中状态
	if order.Status != model.OrderStatusRefunding {
		return errors.ErrBadRequest, nil
	}

	// 使用 refund_success 事件完成 FSM 转换（语义区分在日志中体现）
	newStatus, err := s.fsm.Transition(order.Status, OrderEventRefundSuccess)
	if err != nil {
		return errors.ErrBadRequest, nil
	}

	price := order.Price

	if err := s.orderRepo.Transaction(func(txRepo repository.OrderRepository) error {
		tx := txRepo.GetDB()

		// 根据退款前状态执行不同的资金操作
		switch order.PreRefundStatus {
		case model.OrderStatusShipped:
			// 资金仍在买家冻结余额中 → 解冻并退回买家可用余额
			// 在事务内读取余额确保审计快照准确
			buyer, err := s.userRepo.FindByID(order.BuyerID)
			if err != nil {
				return err
			}

			if err := s.userRepo.UpdateBalance(tx, order.BuyerID, price, -price); err != nil {
				return err
			}

			// 记录买家退款流水
			wt := &model.WalletTransaction{
				UserID:        order.BuyerID,
				Type:          model.WalletTxRefund,
				Amount:        price,
				OrderID:       order.ID,
				BalanceBefore: buyer.Balance,
				BalanceAfter:  buyer.Balance + price,
				FrozenBefore:  buyer.FrozenBalance,
				FrozenAfter:   buyer.FrozenBalance - price,
				Description:   fmt.Sprintf("订单 %s 退款 %.2f 元（解冻退回）", order.OrderNo, float64(price)/100),
			}
			if err := s.userRepo.CreateWalletTransaction(tx, wt); err != nil {
				return err
			}

		case model.OrderStatusReceived:
			// 资金已释放给卖家 → 从卖家可用余额扣除，退回买家
			// 在事务内读取余额确保审计快照准确
			buyer, err := s.userRepo.FindByID(order.BuyerID)
			if err != nil {
				return err
			}
			seller, err := s.userRepo.FindByID(order.SellerID)
			if err != nil {
				return err
			}

			// 扣减卖家余额
			if err := s.userRepo.UpdateBalance(tx, order.SellerID, -price, 0); err != nil {
				return err
			}
			// 增加买家余额
			if err := s.userRepo.UpdateBalance(tx, order.BuyerID, price, 0); err != nil {
				return err
			}

			// 记录卖家扣款流水
			wtSeller := &model.WalletTransaction{
				UserID:        order.SellerID,
				Type:          model.WalletTxRefund,
				Amount:        -price,
				OrderID:       order.ID,
				BalanceBefore: seller.Balance,
				BalanceAfter:  seller.Balance - price,
				FrozenBefore:  seller.FrozenBalance,
				FrozenAfter:   seller.FrozenBalance,
				Description:   fmt.Sprintf("订单 %s 退款 %.2f 元（卖家扣除）", order.OrderNo, float64(price)/100),
			}
			if err := s.userRepo.CreateWalletTransaction(tx, wtSeller); err != nil {
				return err
			}

			// 记录买家收款流水
			wtBuyer := &model.WalletTransaction{
				UserID:        order.BuyerID,
				Type:          model.WalletTxRefund,
				Amount:        price,
				OrderID:       order.ID,
				BalanceBefore: buyer.Balance,
				BalanceAfter:  buyer.Balance + price,
				FrozenBefore:  buyer.FrozenBalance,
				FrozenAfter:   buyer.FrozenBalance,
				Description:   fmt.Sprintf("订单 %s 退款 %.2f 元（买家到账）", order.OrderNo, float64(price)/100),
			}
			if err := s.userRepo.CreateWalletTransaction(tx, wtBuyer); err != nil {
				return err
			}

		default:
			// 未知退款前状态（如数据异常），拒绝操作
			return fmt.Errorf("invalid pre-refund status: %d", order.PreRefundStatus)
		}

		// 更新订单状态为已完成
		fromStatus := order.Status
		order.Status = newStatus
		if err := txRepo.Update(order); err != nil {
			return err
		}
		if err := s.logStatusChange(txRepo, order.ID, fromStatus, newStatus, string(OrderEventApproveRefund), sellerID, "seller"); err != nil {
			return err
		}

		return nil
	}); err != nil {
		if stdErrors.Is(err, repository.ErrBalanceInsufficient) {
			return errors.ErrInsufficientBalance, nil
		}
		return errors.ErrInternal, err
	}

	return errors.Success, nil
}

// RejectRefund 卖家拒绝退款
// 订单转入纠纷处理状态
func (s *OrderService) RejectRefund(sellerID, orderID uint) (int, error) {
	order, code, err := s.getOrderForSeller(orderID, sellerID)
	if err != nil {
		return code, err
	}
	if code != errors.Success {
		return code, nil
	}

	// 必须是退款中状态
	if order.Status != model.OrderStatusRefunding {
		return errors.ErrBadRequest, nil
	}

	// 使用 raise_dispute 事件完成 FSM 转换（语义区分在日志中体现）
	newStatus, err := s.fsm.Transition(order.Status, OrderEventRaiseDispute)
	if err != nil {
		return errors.ErrBadRequest, nil
	}

	if err := s.transitionAndLog(order, newStatus, OrderEventRejectRefund, sellerID, "seller"); err != nil {
		return errors.ErrInternal, err
	}

	return errors.Success, nil
}

// CompleteRefund 退款成功（系统/管理员操作）
// 需要管理员权限，根据 PreRefundStatus 执行资金处理：
//   - 从已发货退款：买家冻结余额 → 买家可用余额（解冻）
//   - 从已收货退款：卖家可用余额 → 买家可用余额（退款到账）
func (s *OrderService) CompleteRefund(operatorID, orderID uint) (int, error) {
	// 验证管理员身份
	admin, err := s.userRepo.FindByID(operatorID)
	if err != nil {
		return errors.ErrInternal, err
	}
	if admin.Role != AdminRole {
		return errors.ErrForbiddenNotAdmin, nil
	}

	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrOrderNotFound, nil
		}
		return errors.ErrInternal, err
	}

	// 必须是退款中状态
	if order.Status != model.OrderStatusRefunding {
		return errors.ErrBadRequest, nil
	}

	newStatus, err := s.fsm.Transition(order.Status, OrderEventRefundSuccess)
	if err != nil {
		return errors.ErrBadRequest, nil
	}

	price := order.Price

	if err := s.orderRepo.Transaction(func(txRepo repository.OrderRepository) error {
		tx := txRepo.GetDB()

		// 根据退款前状态执行不同的资金操作
		switch order.PreRefundStatus {
		case model.OrderStatusShipped:
			// 资金仍在买家冻结余额中 → 解冻并退回买家可用余额
			buyer, err := s.userRepo.FindByID(order.BuyerID)
			if err != nil {
				return err
			}

			if err := s.userRepo.UpdateBalance(tx, order.BuyerID, price, -price); err != nil {
				return err
			}

			wt := &model.WalletTransaction{
				UserID:        order.BuyerID,
				Type:          model.WalletTxRefund,
				Amount:        price,
				OrderID:       order.ID,
				BalanceBefore: buyer.Balance,
				BalanceAfter:  buyer.Balance + price,
				FrozenBefore:  buyer.FrozenBalance,
				FrozenAfter:   buyer.FrozenBalance - price,
				Description:   fmt.Sprintf("管理员退款：订单 %s 退款 %.2f 元（解冻退回）", order.OrderNo, float64(price)/100),
			}
			if err := s.userRepo.CreateWalletTransaction(tx, wt); err != nil {
				return err
			}

		case model.OrderStatusReceived:
			// 资金已释放给卖家 → 从卖家可用余额扣除，退回买家
			buyer, err := s.userRepo.FindByID(order.BuyerID)
			if err != nil {
				return err
			}
			seller, err := s.userRepo.FindByID(order.SellerID)
			if err != nil {
				return err
			}

			if err := s.userRepo.UpdateBalance(tx, order.SellerID, -price, 0); err != nil {
				return err
			}
			if err := s.userRepo.UpdateBalance(tx, order.BuyerID, price, 0); err != nil {
				return err
			}

			wtSeller := &model.WalletTransaction{
				UserID:        order.SellerID,
				Type:          model.WalletTxRefund,
				Amount:        -price,
				OrderID:       order.ID,
				BalanceBefore: seller.Balance,
				BalanceAfter:  seller.Balance - price,
				FrozenBefore:  seller.FrozenBalance,
				FrozenAfter:   seller.FrozenBalance,
				Description:   fmt.Sprintf("管理员退款：订单 %s 退款 %.2f 元（卖家扣除）", order.OrderNo, float64(price)/100),
			}
			if err := s.userRepo.CreateWalletTransaction(tx, wtSeller); err != nil {
				return err
			}

			wtBuyer := &model.WalletTransaction{
				UserID:        order.BuyerID,
				Type:          model.WalletTxRefund,
				Amount:        price,
				OrderID:       order.ID,
				BalanceBefore: buyer.Balance,
				BalanceAfter:  buyer.Balance + price,
				FrozenBefore:  buyer.FrozenBalance,
				FrozenAfter:   buyer.FrozenBalance,
				Description:   fmt.Sprintf("管理员退款：订单 %s 退款 %.2f 元（买家到账）", order.OrderNo, float64(price)/100),
			}
			if err := s.userRepo.CreateWalletTransaction(tx, wtBuyer); err != nil {
				return err
			}

		default:
			return fmt.Errorf("invalid pre-refund status: %d", order.PreRefundStatus)
		}

		// 更新订单状态为已完成
		fromStatus := order.Status
		order.Status = newStatus
		if err := txRepo.Update(order); err != nil {
			return err
		}
		if err := s.logStatusChange(txRepo, order.ID, fromStatus, newStatus, string(OrderEventRefundSuccess), operatorID, "admin"); err != nil {
			return err
		}

		return nil
	}); err != nil {
		if stdErrors.Is(err, repository.ErrBalanceInsufficient) {
			return errors.ErrInsufficientBalance, nil
		}
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
	if code != errors.Success {
		return code, nil
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

// ArbitrateReq 仲裁请求
type ArbitrateReq struct {
	Decision string `json:"decision" binding:"required"` // "buyer"=支持买家, "seller"=支持卖家
}

// ShipReq 发货请求
type ShipReq struct {
	TrackingNumber   string `json:"tracking_number" binding:"max=100"`    // 快递单号
	LogisticsCompany string `json:"logistics_company" binding:"max=50"`   // 物流公司名称
}

// Arbitrate 管理员仲裁
// 根据仲裁决定处理资金：
//   - 支持买家：根据退款前状态，将资金退还买家
//   - 支持卖家：根据退款前状态，将资金释放给卖家
func (s *OrderService) Arbitrate(adminID, orderID uint, req *ArbitrateReq) (int, error) {
	// 验证管理员身份
	admin, err := s.userRepo.FindByID(adminID)
	if err != nil {
		return errors.ErrInternal, err
	}
	if admin.Role != AdminRole {
		return errors.ErrForbiddenNotAdmin, nil
	}

	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.ErrOrderNotFound, nil
		}
		return errors.ErrInternal, err
	}

	// 必须是纠纷处理状态
	if order.Status != model.OrderStatusDispute {
		return errors.ErrBadRequest, nil
	}

	// 验证仲裁决定
	if req.Decision != "buyer" && req.Decision != "seller" {
		return errors.ErrBadRequest, nil
	}

	newStatus, err := s.fsm.Transition(order.Status, OrderEventArbitrate)
	if err != nil {
		return errors.ErrBadRequest, nil
	}

	price := order.Price

	if err := s.orderRepo.Transaction(func(txRepo repository.OrderRepository) error {
		tx := txRepo.GetDB()

		if req.Decision == "buyer" {
			// 管理员裁定支持买家 → 退款给买家
			switch order.PreRefundStatus {
			case model.OrderStatusShipped:
				// 资金仍在买家冻结余额中 → 解冻并退回买家可用余额
				// 在事务内读取余额确保审计快照准确
				buyer, err := s.userRepo.FindByID(order.BuyerID)
				if err != nil {
					return err
				}

				if err := s.userRepo.UpdateBalance(tx, order.BuyerID, price, -price); err != nil {
					return err
				}

				wt := &model.WalletTransaction{
					UserID:        order.BuyerID,
					Type:          model.WalletTxRefund,
					Amount:        price,
					OrderID:       order.ID,
					BalanceBefore: buyer.Balance,
					BalanceAfter:  buyer.Balance + price,
					FrozenBefore:  buyer.FrozenBalance,
					FrozenAfter:   buyer.FrozenBalance - price,
					Description:   fmt.Sprintf("仲裁支持买家：订单 %s 退款 %.2f 元（解冻退回）", order.OrderNo, float64(price)/100),
				}
				if err := s.userRepo.CreateWalletTransaction(tx, wt); err != nil {
					return err
				}

			case model.OrderStatusReceived:
				// 资金已释放给卖家 → 从卖家可用余额扣除，退回买家
				// 在事务内读取余额确保审计快照准确
				buyer, err := s.userRepo.FindByID(order.BuyerID)
				if err != nil {
					return err
				}
				seller, err := s.userRepo.FindByID(order.SellerID)
				if err != nil {
					return err
				}

				// 扣减卖家余额
				if err := s.userRepo.UpdateBalance(tx, order.SellerID, -price, 0); err != nil {
					return err
				}
				// 增加买家余额
				if err := s.userRepo.UpdateBalance(tx, order.BuyerID, price, 0); err != nil {
					return err
				}

				// 记录卖家扣款流水
				wtSeller := &model.WalletTransaction{
					UserID:        order.SellerID,
					Type:          model.WalletTxRefund,
					Amount:        -price,
					OrderID:       order.ID,
					BalanceBefore: seller.Balance,
					BalanceAfter:  seller.Balance - price,
					FrozenBefore:  seller.FrozenBalance,
					FrozenAfter:   seller.FrozenBalance,
					Description:   fmt.Sprintf("仲裁支持买家：订单 %s 退款 %.2f 元（卖家扣除）", order.OrderNo, float64(price)/100),
				}
				if err := s.userRepo.CreateWalletTransaction(tx, wtSeller); err != nil {
					return err
				}

				// 记录买家收款流水
				wtBuyer := &model.WalletTransaction{
					UserID:        order.BuyerID,
					Type:          model.WalletTxRefund,
					Amount:        price,
					OrderID:       order.ID,
					BalanceBefore: buyer.Balance,
					BalanceAfter:  buyer.Balance + price,
					FrozenBefore:  buyer.FrozenBalance,
					FrozenAfter:   buyer.FrozenBalance,
					Description:   fmt.Sprintf("仲裁支持买家：订单 %s 退款 %.2f 元（买家到账）", order.OrderNo, float64(price)/100),
				}
				if err := s.userRepo.CreateWalletTransaction(tx, wtBuyer); err != nil {
					return err
				}

			default:
				return fmt.Errorf("invalid pre-refund status for arbitration: %d", order.PreRefundStatus)
			}
		} else {
			// 管理员裁定支持卖家 → 资金释放给卖家
			switch order.PreRefundStatus {
			case model.OrderStatusShipped:
				// 资金仍在买家冻结余额中 → 释放给卖家
				// 在事务内读取余额确保审计快照准确
				seller, err := s.userRepo.FindByID(order.SellerID)
				if err != nil {
					return err
				}

				// 买家冻结余额扣减
				if err := s.userRepo.UpdateBalance(tx, order.BuyerID, 0, -price); err != nil {
					return err
				}
				// 卖家可用余额增加
				if err := s.userRepo.UpdateBalance(tx, order.SellerID, price, 0); err != nil {
					return err
				}

				// 记录卖家收款流水
				wt := &model.WalletTransaction{
					UserID:        order.SellerID,
					Type:          model.WalletTxComplete,
					Amount:        price,
					OrderID:       order.ID,
					BalanceBefore: seller.Balance,
					BalanceAfter:  seller.Balance + price,
					FrozenBefore:  seller.FrozenBalance,
					FrozenAfter:   seller.FrozenBalance,
					Description:   fmt.Sprintf("仲裁支持卖家：订单 %s 资金释放 %.2f 元", order.OrderNo, float64(price)/100),
				}
				if err := s.userRepo.CreateWalletTransaction(tx, wt); err != nil {
					return err
				}

			case model.OrderStatusReceived:
				// 资金已释放给卖家，无需额外资金操作，直接完成订单
				// 仅记录仲裁日志

			default:
				return fmt.Errorf("invalid pre-refund status for arbitration: %d", order.PreRefundStatus)
			}
		}

		// 更新订单状态为已完成
		fromStatus := order.Status
		order.Status = newStatus
		if err := txRepo.Update(order); err != nil {
			return err
		}

		eventDesc := string(OrderEventArbitrate) + ":" + req.Decision
		return s.logStatusChange(txRepo, order.ID, fromStatus, newStatus, eventDesc, adminID, "admin")
	}); err != nil {
		if stdErrors.Is(err, repository.ErrBalanceInsufficient) {
			return errors.ErrInsufficientBalance, nil
		}
		return errors.ErrInternal, err
	}

	return errors.Success, nil
}

// ListDisputes 管理员获取所有纠纷订单列表
// 返回 DisputeDetail 结构体，含买家/卖家名称
func (s *OrderService) ListDisputes(adminID uint, page, pageSize int) (*PaginatedResult, int, error) {
	// 验证管理员身份
	admin, err := s.userRepo.FindByID(adminID)
	if err != nil {
		return nil, errors.ErrInternal, err
	}
	if admin.Role != AdminRole {
		return nil, errors.ErrForbiddenNotAdmin, nil
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	orders, total, err := s.orderRepo.ListByStatus(model.OrderStatusDispute, page, pageSize)
	if err != nil {
		return nil, errors.ErrInternal, err
	}

	items := make([]DisputeDetail, len(orders))
	for i, o := range orders {
		detail := s.detailFromModel(&o)
		disputeDetail := DisputeDetail{
			OrderDetail:  *detail,
			RefundReason: o.RefundReason,
		}
		if o.Buyer != nil {
			disputeDetail.BuyerName = o.Buyer.Nickname
		}
		if o.Seller != nil {
			disputeDetail.SellerName = o.Seller.Nickname
		}
		items[i] = disputeDetail
	}

	return &PaginatedResult{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: calcTotalPages(total, pageSize),
	}, errors.Success, nil
}

// DisputeDetail 纠纷订单详情（含买家卖家信息）
type DisputeDetail struct {
	OrderDetail
	BuyerName    string `json:"buyer_name"`
	SellerName   string `json:"seller_name"`
	RefundReason string `json:"refund_reason,omitempty"`
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

// validOrderStatus 检查订单状态值是否合法
func validOrderStatus(s model.OrderStatus) bool {
	switch s {
	case model.OrderStatusPendingPayment,
		model.OrderStatusCancelled,
		model.OrderStatusPaid,
		model.OrderStatusPendingShipment,
		model.OrderStatusShipped,
		model.OrderStatusReceived,
		model.OrderStatusCompleted,
		model.OrderStatusRefunding,
		model.OrderStatusDispute:
		return true
	}
	return false
}

// ListByBuyer 买家的订单列表
func (s *OrderService) ListByBuyer(userID uint, status *model.OrderStatus, page, pageSize int) (*PaginatedResult, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	// 校验状态参数
	if status != nil && !validOrderStatus(*status) {
		return nil, errors.ErrBadRequest, nil
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

	// 校验状态参数
	if status != nil && !validOrderStatus(*status) {
		return nil, errors.ErrBadRequest, nil
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
		ID:               o.ID,
		OrderNo:          o.OrderNo,
		ProductID:        o.ProductID,
		BuyerID:          o.BuyerID,
		SellerID:         o.SellerID,
		Title:            o.Title,
		Price:            o.Price,
		Status:           o.Status,
		StatusName:       model.OrderStatusNames[o.Status],
		PreRefundStatus:  o.PreRefundStatus,
		RefundReason:     o.RefundReason,
		TrackingNumber:   o.TrackingNumber,
		LogisticsCompany: o.LogisticsCompany,
		ShippingAddress:  o.ShippingAddress,
		BuyerNote:        o.BuyerNote,
		CreatedAt:        o.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:        o.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	detail.Product = buildProductListItem(o)

	return detail
}

// buildProductListItem 构建商品信息项（复用代码，减少重复）
func buildProductListItem(o *model.Order) *ProductListItem {
	if o.Product == nil || o.Product.ID == 0 {
		return nil
	}
	images := make([]ProductImageItem, len(o.Product.Images))
	for j, img := range o.Product.Images {
		images[j] = ProductImageItem{
			ID:        img.ID,
			URL:       img.URL,
			SortOrder: img.SortOrder,
		}
	}
	return &ProductListItem{
		ID:     o.Product.ID,
		Title:  o.Product.Title,
		Price:  o.Product.Price,
		Images: images,
	}
}

func (s *OrderService) listItemFromModel(o *model.Order) OrderListItem {
	item := OrderListItem{
		ID:         o.ID,
		OrderNo:    o.OrderNo,
		ProductID:  o.ProductID,
		Title:      o.Title,
		Price:      o.Price,
		Status:     o.Status,
		StatusName: model.OrderStatusNames[o.Status],
		CreatedAt:  o.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	item.Product = buildProductListItem(o)

	return item
}
