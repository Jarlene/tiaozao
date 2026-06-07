package service

import (
	"testing"

	"flea-market/internal/model"
	"flea-market/internal/repository"
	errs "flea-market/pkg/errors"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ---- Mock OrderRepository ----

type mockOrderRepo struct {
	createFunc             func(order *model.Order) error
	findByIDFunc           func(id uint) (*model.Order, error)
	findByOrderNoFunc      func(orderNo string) (*model.Order, error)
	updateFunc             func(order *model.Order) error
	listByBuyerFunc        func(buyerID uint, status *model.OrderStatus, page, pageSize int) ([]model.Order, int64, error)
	listBySellerFunc       func(sellerID uint, status *model.OrderStatus, page, pageSize int) ([]model.Order, int64, error)
	listByStatusFunc       func(status model.OrderStatus, page, pageSize int) ([]model.Order, int64, error)
	decrementStockFunc     func(productID uint, quantity int) (bool, error)
	incrementStockFunc     func(productID uint, quantity int) error
	createStatusLogFunc    func(log *model.OrderStatusLog) error
	listStatusLogsFunc     func(orderID uint) ([]model.OrderStatusLog, error)
	transactionFunc        func(fc func(txRepo repository.OrderRepository) error) error
	getDBFunc              func() *gorm.DB
}

func (m *mockOrderRepo) Create(order *model.Order) error       { return m.createFunc(order) }
func (m *mockOrderRepo) FindByID(id uint) (*model.Order, error)  { return m.findByIDFunc(id) }
func (m *mockOrderRepo) FindByOrderNo(orderNo string) (*model.Order, error) { return m.findByOrderNoFunc(orderNo) }
func (m *mockOrderRepo) Update(order *model.Order) error        { return m.updateFunc(order) }
func (m *mockOrderRepo) ListByBuyer(buyerID uint, status *model.OrderStatus, page, pageSize int) ([]model.Order, int64, error) {
	return m.listByBuyerFunc(buyerID, status, page, pageSize)
}
func (m *mockOrderRepo) ListBySeller(sellerID uint, status *model.OrderStatus, page, pageSize int) ([]model.Order, int64, error) {
	return m.listBySellerFunc(sellerID, status, page, pageSize)
}
func (m *mockOrderRepo) ListByStatus(status model.OrderStatus, page, pageSize int) ([]model.Order, int64, error) {
	return m.listByStatusFunc(status, page, pageSize)
}
func (m *mockOrderRepo) DecrementProductStock(productID uint, quantity int) (bool, error) {
	return m.decrementStockFunc(productID, quantity)
}
func (m *mockOrderRepo) IncrementProductStock(productID uint, quantity int) error {
	return m.incrementStockFunc(productID, quantity)
}
func (m *mockOrderRepo) CreateStatusLog(log *model.OrderStatusLog) error { return m.createStatusLogFunc(log) }
func (m *mockOrderRepo) ListStatusLogs(orderID uint) ([]model.OrderStatusLog, error) {
	return m.listStatusLogsFunc(orderID)
}
func (m *mockOrderRepo) Transaction(fc func(txRepo repository.OrderRepository) error) error {
	return m.transactionFunc(fc)
}
func (m *mockOrderRepo) GetDB() *gorm.DB {
	if m.getDBFunc != nil {
		return m.getDBFunc()
	}
	return nil
}

// ---- Mock UserRepository ----

type mockUserRepo struct {
	createFunc               func(user *model.User) error
	findByIDFunc             func(id uint) (*model.User, error)
	findByEmailFunc          func(email string) (*model.User, error)
	updateFunc               func(user *model.User) error
	updateBalanceFunc        func(tx *gorm.DB, userID uint, balanceDelta, frozenDelta int64) error
	createWalletTxFunc       func(tx *gorm.DB, wt *model.WalletTransaction) error
	listWalletTransactionsFunc func(userID uint, page, pageSize int) ([]model.WalletTransaction, int64, error)
	transactionFunc          func(fc func(tx *gorm.DB) error) error
}

func (m *mockUserRepo) Create(user *model.User) error                     { return m.createFunc(user) }
func (m *mockUserRepo) FindByID(id uint) (*model.User, error)             { return m.findByIDFunc(id) }
func (m *mockUserRepo) FindByEmail(email string) (*model.User, error)     { return m.findByEmailFunc(email) }
func (m *mockUserRepo) Update(user *model.User) error                     { return m.updateFunc(user) }
func (m *mockUserRepo) UpdateBalance(tx *gorm.DB, userID uint, balanceDelta, frozenDelta int64) error {
	return m.updateBalanceFunc(tx, userID, balanceDelta, frozenDelta)
}
func (m *mockUserRepo) CreateWalletTransaction(tx *gorm.DB, wt *model.WalletTransaction) error {
	return m.createWalletTxFunc(tx, wt)
}
func (m *mockUserRepo) Transaction(fc func(tx *gorm.DB) error) error {
	if m.transactionFunc != nil {
		return m.transactionFunc(fc)
	}
	return fc(nil)
}
func (m *mockUserRepo) ListWalletTransactionsByTx(tx *gorm.DB, userID uint) ([]model.WalletTransaction, error) {
	return nil, nil
}
func (m *mockUserRepo) ListWalletTransactions(userID uint, page, pageSize int) ([]model.WalletTransaction, int64, error) {
	return m.listWalletTransactionsFunc(userID, page, pageSize)
}

// ---- Mock ProductRepository ----

type mockProductRepo2 struct {
	findByIDFunc func(id uint) (*model.Product, error)
}

func (m *mockProductRepo2) Create(product *model.Product) error      { panic("unexpected") }
func (m *mockProductRepo2) FindByID(id uint) (*model.Product, error) { return m.findByIDFunc(id) }
func (m *mockProductRepo2) Update(product *model.Product) error      { panic("unexpected") }
func (m *mockProductRepo2) Delete(id uint) error                     { panic("unexpected") }
func (m *mockProductRepo2) List(page, pageSize int, status model.ProductStatus) ([]model.Product, int64, error) {
	panic("unexpected")
}
func (m *mockProductRepo2) Search(keyword string, categoryID *uint, priceMin, priceMax int64, page, pageSize int) ([]model.Product, int64, error) {
	panic("unexpected")
}
func (m *mockProductRepo2) ListByUser(userID uint, status *model.ProductStatus, page, pageSize int) ([]model.Product, int64, error) {
	panic("unexpected")
}
func (m *mockProductRepo2) CountByUserAndStatus(userID uint) (activeCount, soldCount, inactiveCount int64, err error) {
	panic("unexpected")
}
func (m *mockProductRepo2) CreateImage(image *model.ProductImage) error      { panic("unexpected") }
func (m *mockProductRepo2) UpdateImage(image *model.ProductImage) error      { panic("unexpected") }
func (m *mockProductRepo2) DeleteImage(id uint) error                        { panic("unexpected") }
func (m *mockProductRepo2) FindImageByID(id uint) (*model.ProductImage, error) { panic("unexpected") }
func (m *mockProductRepo2) FindImagesByIDs(ids []uint) ([]model.ProductImage, error) { panic("unexpected") }
func (m *mockProductRepo2) DeleteImagesByProduct(productID uint) error       { panic("unexpected") }
func (m *mockProductRepo2) ListImagesByProduct(productID uint) ([]model.ProductImage, error) { panic("unexpected") }
func (m *mockProductRepo2) Transaction(fc func(txRepo repository.ProductRepository) error) error {
	panic("unexpected")
}

// ---- Test Helpers ----

func newTestOrderService(o mockOrderRepo, u mockUserRepo, p mockProductRepo2) *OrderService {
	return &OrderService{
		orderRepo:   &o,
		productRepo: &p,
		userRepo:    &u,
		fsm:         NewOrderFSM(),
		rdb:         redis.NewClient(&redis.Options{Addr: "localhost:6379"}),
	}
}

func makePendingOrder(id, buyerID, sellerID uint, price int64) *model.Order {
	return &model.Order{
		ID:      id,
		OrderNo: "2026010112000000001",
		ProductID: 1,
		BuyerID:   buyerID,
		SellerID:  sellerID,
		Title:     "测试商品",
		Price:     price,
		Status:    model.OrderStatusPendingPayment,
	}
}

// ---- Tests: Pay with wallet escrow ----

func TestPay_SufficientBalance(t *testing.T) {
	var walletTxCreated bool

	order := makePendingOrder(1, 100, 200, 3000) // price = 30.00元

	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return order, nil
			},
			transactionFunc: func(fc func(txRepo repository.OrderRepository) error) error {
				return fc(&mockOrderRepo{
					updateFunc: func(o *model.Order) error {
						// 第一次调用是 Pay 转换（OrderStatusPaid），第二次是 NotifySeller 转换（OrderStatusPendingShipment）
						return nil
					},
					createStatusLogFunc: func(log *model.OrderStatusLog) error {
						return nil
					},
					getDBFunc: func() *gorm.DB { return nil },
				})
			},
		},
		mockUserRepo{
			findByIDFunc: func(id uint) (*model.User, error) {
				if id != 100 {
					t.Fatalf("expected buyerID 100, got %d", id)
				}
				return &model.User{
					ID:            100,
					Balance:       10000, // 100.00元
					FrozenBalance: 0,
				}, nil
			},
			updateBalanceFunc: func(tx *gorm.DB, userID uint, balanceDelta, frozenDelta int64) error {
				if balanceDelta != -3000 {
					t.Fatalf("expected balanceDelta -3000, got %d", balanceDelta)
				}
				if frozenDelta != 3000 {
					t.Fatalf("expected frozenDelta 3000, got %d", frozenDelta)
				}
				return nil
			},
			createWalletTxFunc: func(tx *gorm.DB, wt *model.WalletTransaction) error {
				walletTxCreated = true
				if wt.Type != model.WalletTxPay {
					t.Fatalf("expected WalletTxPay, got %s", wt.Type)
				}
				if wt.Amount != -3000 {
					t.Fatalf("expected Amount -3000, got %d", wt.Amount)
				}
				if wt.BalanceBefore != 10000 || wt.BalanceAfter != 7000 {
					t.Fatalf("balance mismatch: before=%d after=%d", wt.BalanceBefore, wt.BalanceAfter)
				}
				return nil
			},
		},
		mockProductRepo2{},
	)

	code, err := svc.Pay(100, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.Success {
		t.Fatalf("expected Success(%d), got %d", errs.Success, code)
	}
	if !walletTxCreated {
		t.Fatal("expected wallet transaction to be created")
	}
}

func TestPay_InsufficientBalance(t *testing.T) {
	order := makePendingOrder(1, 100, 200, 10000) // price = 100.00元

	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return order, nil
			},
			transactionFunc: func(fc func(txRepo repository.OrderRepository) error) error {
				// 余额检查已移入事务内部，事务会被启动但内部返回余额不足错误
				err := fc(&mockOrderRepo{
					updateFunc: func(o *model.Order) error { return nil },
					createStatusLogFunc: func(log *model.OrderStatusLog) error { return nil },
				})
				if err == nil {
					t.Fatal("expected insufficient balance error")
				}
				return err
			},
		},
		mockUserRepo{
			findByIDFunc: func(id uint) (*model.User, error) {
				return &model.User{
					ID:            100,
					Balance:       5000, // only 50.00元, not enough for 100.00元
					FrozenBalance: 0,
				}, nil
			},
		},
		mockProductRepo2{},
	)

	code, err := svc.Pay(100, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.ErrInsufficientBalance {
		t.Fatalf("expected ErrInsufficientBalance(%d), got %d", errs.ErrInsufficientBalance, code)
	}
}

func TestPay_OrderNotFound(t *testing.T) {
	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return nil, gorm.ErrRecordNotFound
			},
		},
		mockUserRepo{},
		mockProductRepo2{},
	)

	code, err := svc.Pay(100, 999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.ErrOrderNotFound {
		t.Fatalf("expected ErrOrderNotFound(%d), got %d", errs.ErrOrderNotFound, code)
	}
}

func TestPay_NotBuyer(t *testing.T) {
	order := makePendingOrder(1, 100, 200, 3000)

	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return order, nil
			},
		},
		mockUserRepo{},
		mockProductRepo2{},
	)

	// user 999 is not the buyer (buyer is 100)
	code, err := svc.Pay(999, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.ErrForbidden {
		t.Fatalf("expected ErrForbidden(%d), got %d", errs.ErrForbidden, code)
	}
}

func TestPay_AlreadyPaid(t *testing.T) {
	order := makePendingOrder(1, 100, 200, 3000)
	order.Status = model.OrderStatusPaid // already paid

	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return order, nil
			},
		},
		mockUserRepo{},
		mockProductRepo2{},
	)

	code, err := svc.Pay(100, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.ErrBadRequest {
		t.Fatalf("expected ErrBadRequest(%d), got %d", errs.ErrBadRequest, code)
	}
}

// ---- Tests: Wallet Service ----

func TestWalletService_GetWalletInfo(t *testing.T) {
	svc := NewWalletService(&mockUserRepo{
		findByIDFunc: func(id uint) (*model.User, error) {
			return &model.User{
				ID:            1,
				Balance:       50000,
				FrozenBalance: 10000,
			}, nil
		},
	})

	info, code, err := svc.GetWalletInfo(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.Success {
		t.Fatalf("expected success, got %d", code)
	}
	if info.Balance != 50000 {
		t.Fatalf("expected balance 50000, got %d", info.Balance)
	}
	if info.FrozenBalance != 10000 {
		t.Fatalf("expected frozen_balance 10000, got %d", info.FrozenBalance)
	}
	if info.TotalBalance != 60000 {
		t.Fatalf("expected total_balance 60000, got %d", info.TotalBalance)
	}
}

func TestWalletService_TopUp(t *testing.T) {
	var balanceUpdated bool
	var txCreated bool

	svc := NewWalletService(&mockUserRepo{
		findByIDFunc: func(id uint) (*model.User, error) {
			return &model.User{
				ID:            1,
				Balance:       5000,
				FrozenBalance: 0,
			}, nil
		},
		updateBalanceFunc: func(tx *gorm.DB, userID uint, balanceDelta, frozenDelta int64) error {
			balanceUpdated = true
			if balanceDelta != 10000 {
				t.Fatalf("expected balanceDelta 10000, got %d", balanceDelta)
			}
			if frozenDelta != 0 {
				t.Fatalf("expected frozenDelta 0, got %d", frozenDelta)
			}
			return nil
		},
		createWalletTxFunc: func(tx *gorm.DB, wt *model.WalletTransaction) error {
			txCreated = true
			if wt.Type != model.WalletTxTopUp {
				t.Fatalf("expected WalletTxTopUp, got %s", wt.Type)
			}
			if wt.Amount != 10000 {
				t.Fatalf("expected Amount 10000, got %d", wt.Amount)
			}
			return nil
		},
	})

	info, code, err := svc.TopUp(1, &TopUpReq{Amount: 10000})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.Success {
		t.Fatalf("expected success, got %d", code)
	}
	if !balanceUpdated {
		t.Fatal("expected balance to be updated")
	}
	if !txCreated {
		t.Fatal("expected wallet transaction to be created")
	}
	_ = info
}

// ---- Tests: ConfirmReceive with fund release ----

func makeShippedOrder(id, buyerID, sellerID uint, price int64) *model.Order {
	return &model.Order{
		ID:        id,
		OrderNo:   "2026010112000000001",
		ProductID: 1,
		BuyerID:   buyerID,
		SellerID:  sellerID,
		Title:     "测试商品",
		Price:     price,
		Status:    model.OrderStatusShipped,
	}
}

func TestConfirmReceive_Success(t *testing.T) {
	var walletTxCreated bool
	var balanceUpdates []struct {
		userID       uint
		balanceDelta int64
		frozenDelta  int64
	}

	order := makeShippedOrder(1, 100, 200, 3000) // price = 30.00元
	var updateCallCount int
	var logCallCount int

	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return order, nil
			},
			transactionFunc: func(fc func(txRepo repository.OrderRepository) error) error {
				return fc(&mockOrderRepo{
					updateFunc: func(o *model.Order) error {
						updateCallCount++
						return nil
					},
					createStatusLogFunc: func(log *model.OrderStatusLog) error {
						logCallCount++
						return nil
					},
					getDBFunc: func() *gorm.DB { return nil },
				})
			},
		},
		mockUserRepo{
			findByIDFunc: func(id uint) (*model.User, error) {
				if id == 200 {
					return &model.User{
						ID:            200,
						Balance:       5000,
						FrozenBalance: 0,
					}, nil
				}
				return &model.User{
					ID:            100,
					Balance:       7000,
					FrozenBalance: 3000,
				}, nil
			},
			updateBalanceFunc: func(tx *gorm.DB, userID uint, balanceDelta, frozenDelta int64) error {
				balanceUpdates = append(balanceUpdates, struct {
					userID       uint
					balanceDelta int64
					frozenDelta  int64
				}{userID, balanceDelta, frozenDelta})
				return nil
			},
			createWalletTxFunc: func(tx *gorm.DB, wt *model.WalletTransaction) error {
				walletTxCreated = true
				if wt.Type != model.WalletTxComplete {
					t.Fatalf("expected WalletTxComplete, got %s", wt.Type)
				}
				if wt.Amount != 3000 {
					t.Fatalf("expected Amount 3000, got %d", wt.Amount)
				}
				if wt.UserID != 200 {
					t.Fatalf("expected sellerID 200, got %d", wt.UserID)
				}
				if wt.BalanceBefore != 5000 || wt.BalanceAfter != 8000 {
					t.Fatalf("balance mismatch: before=%d after=%d", wt.BalanceBefore, wt.BalanceAfter)
				}
				return nil
			},
		},
		mockProductRepo2{},
	)

	code, err := svc.ConfirmReceive(100, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.Success {
		t.Fatalf("expected Success(%d), got %d", errs.Success, code)
	}
	if !walletTxCreated {
		t.Fatal("expected wallet transaction to be created")
	}
	if len(balanceUpdates) != 2 {
		t.Fatalf("expected 2 balance updates (buyer frozen release + seller credit), got %d", len(balanceUpdates))
	}
	// First update should be buyer frozen release (0, -price)
	if balanceUpdates[0].userID != 100 || balanceUpdates[0].balanceDelta != 0 || balanceUpdates[0].frozenDelta != -3000 {
		t.Fatalf("first update should be buyer frozen release: got user=%d bal=%d frozen=%d",
			balanceUpdates[0].userID, balanceUpdates[0].balanceDelta, balanceUpdates[0].frozenDelta)
	}
	// Second update should be seller credit (+price, 0)
	if balanceUpdates[1].userID != 200 || balanceUpdates[1].balanceDelta != 3000 || balanceUpdates[1].frozenDelta != 0 {
		t.Fatalf("second update should be seller credit: got user=%d bal=%d frozen=%d",
			balanceUpdates[1].userID, balanceUpdates[1].balanceDelta, balanceUpdates[1].frozenDelta)
	}
	// Should have 2 status log entries (Received + Completed)
	if logCallCount != 2 {
		t.Fatalf("expected 2 status log entries, got %d", logCallCount)
	}
}

func TestConfirmReceive_OrderNotFound(t *testing.T) {
	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return nil, gorm.ErrRecordNotFound
			},
		},
		mockUserRepo{},
		mockProductRepo2{},
	)

	code, err := svc.ConfirmReceive(100, 999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.ErrOrderNotFound {
		t.Fatalf("expected ErrOrderNotFound(%d), got %d", errs.ErrOrderNotFound, code)
	}
}

func TestConfirmReceive_NotBuyer(t *testing.T) {
	order := makeShippedOrder(1, 100, 200, 3000)

	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return order, nil
			},
		},
		mockUserRepo{},
		mockProductRepo2{},
	)

	// user 999 is not the buyer (buyer is 100)
	code, err := svc.ConfirmReceive(999, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.ErrForbidden {
		t.Fatalf("expected ErrForbidden(%d), got %d", errs.ErrForbidden, code)
	}
}

func TestConfirmReceive_WrongStatus(t *testing.T) {
	order := makeShippedOrder(1, 100, 200, 3000)
	order.Status = model.OrderStatusPendingPayment // not shipped

	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return order, nil
			},
		},
		mockUserRepo{},
		mockProductRepo2{},
	)

	code, err := svc.ConfirmReceive(100, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.ErrBadRequest {
		t.Fatalf("expected ErrBadRequest(%d), got %d", errs.ErrBadRequest, code)
	}
}

// ---- Tests: RequestRefund ----

func makeRefundingOrder(id, buyerID, sellerID uint, price int64, preRefundStatus model.OrderStatus) *model.Order {
	return &model.Order{
		ID:              id,
		OrderNo:         "2026010112000000001",
		ProductID:       1,
		BuyerID:         buyerID,
		SellerID:        sellerID,
		Title:           "测试商品",
		Price:           price,
		Status:          model.OrderStatusRefunding,
		PreRefundStatus: preRefundStatus,
	}
}

func TestRequestRefund_FromShipped(t *testing.T) {
	order := makeShippedOrder(1, 100, 200, 3000) // price = 30.00元
	var preRefundStatusSaved bool

	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return order, nil
			},
			transactionFunc: func(fc func(txRepo repository.OrderRepository) error) error {
				return fc(&mockOrderRepo{
					updateFunc: func(o *model.Order) error {
						if o.PreRefundStatus == model.OrderStatusShipped {
							preRefundStatusSaved = true
						}
						return nil
					},
					createStatusLogFunc: func(log *model.OrderStatusLog) error {
						return nil
					},
				})
			},
		},
		mockUserRepo{},
		mockProductRepo2{},
	)

	code, err := svc.RequestRefund(100, 1, &RequestRefundReq{Reason: "商品与描述不符"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.Success {
		t.Fatalf("expected Success(%d), got %d", errs.Success, code)
	}
	if !preRefundStatusSaved {
		t.Fatal("expected PreRefundStatus to be saved")
	}
}

func TestRequestRefund_FromReceived(t *testing.T) {
	order := &model.Order{
		ID:        1,
		OrderNo:   "2026010112000000001",
		ProductID: 1,
		BuyerID:   100,
		SellerID:  200,
		Title:     "测试商品",
		Price:     3000,
		Status:    model.OrderStatusReceived,
	}

	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return order, nil
			},
			transactionFunc: func(fc func(txRepo repository.OrderRepository) error) error {
				return fc(&mockOrderRepo{
					updateFunc: func(o *model.Order) error {
						if o.PreRefundStatus != model.OrderStatusReceived {
							t.Fatalf("expected PreRefundStatus Received(%d), got %d", model.OrderStatusReceived, o.PreRefundStatus)
						}
						return nil
					},
					createStatusLogFunc: func(log *model.OrderStatusLog) error {
						return nil
					},
				})
			},
		},
		mockUserRepo{},
		mockProductRepo2{},
	)

	code, err := svc.RequestRefund(100, 1, &RequestRefundReq{Reason: "商品损坏"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.Success {
		t.Fatalf("expected Success(%d), got %d", errs.Success, code)
	}
}

func TestRequestRefund_WrongStatus(t *testing.T) {
	order := makePendingOrder(1, 100, 200, 3000) // PendingPayment — cannot refund

	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return order, nil
			},
		},
		mockUserRepo{},
		mockProductRepo2{},
	)

	code, err := svc.RequestRefund(100, 1, &RequestRefundReq{Reason: "测试"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.ErrBadRequest {
		t.Fatalf("expected ErrBadRequest(%d), got %d", errs.ErrBadRequest, code)
	}
}

func TestRequestRefund_NotBuyer(t *testing.T) {
	order := makeShippedOrder(1, 100, 200, 3000)

	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return order, nil
			},
		},
		mockUserRepo{},
		mockProductRepo2{},
	)

	// user 999 is not the buyer (buyer is 100)
	code, err := svc.RequestRefund(999, 1, &RequestRefundReq{Reason: "测试"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.ErrForbidden {
		t.Fatalf("expected ErrForbidden(%d), got %d", errs.ErrForbidden, code)
	}
}

// ---- Tests: ApproveRefund ----

func TestApproveRefund_FromShipped_FrozenRelease(t *testing.T) {
	var walletTxCreated bool
	var balanceUpdates []struct {
		userID       uint
		balanceDelta int64
		frozenDelta  int64
	}

	order := makeRefundingOrder(1, 100, 200, 3000, model.OrderStatusShipped)
	var updateCallCount int
	var logCallCount int

	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return order, nil
			},
			transactionFunc: func(fc func(txRepo repository.OrderRepository) error) error {
				return fc(&mockOrderRepo{
					updateFunc: func(o *model.Order) error {
						updateCallCount++
						return nil
					},
					createStatusLogFunc: func(log *model.OrderStatusLog) error {
						logCallCount++
						return nil
					},
					getDBFunc: func() *gorm.DB { return nil },
				})
			},
		},
		mockUserRepo{
			findByIDFunc: func(id uint) (*model.User, error) {
				// buyer has frozen balance
				if id == 100 {
					return &model.User{
						ID:            100,
						Balance:       7000,
						FrozenBalance: 3000,
					}, nil
				}
				return nil, gorm.ErrRecordNotFound
			},
			updateBalanceFunc: func(tx *gorm.DB, userID uint, balanceDelta, frozenDelta int64) error {
				balanceUpdates = append(balanceUpdates, struct {
					userID       uint
					balanceDelta int64
					frozenDelta  int64
				}{userID, balanceDelta, frozenDelta})
				return nil
			},
			createWalletTxFunc: func(tx *gorm.DB, wt *model.WalletTransaction) error {
				walletTxCreated = true
				if wt.Type != model.WalletTxRefund {
					t.Fatalf("expected WalletTxRefund, got %s", wt.Type)
				}
				if wt.UserID != 100 {
					t.Fatalf("expected buyerID 100, got %d", wt.UserID)
				}
				if wt.Amount != 3000 {
					t.Fatalf("expected Amount 3000, got %d", wt.Amount)
				}
				if wt.BalanceBefore != 7000 || wt.BalanceAfter != 10000 {
					t.Fatalf("balance mismatch: before=%d after=%d", wt.BalanceBefore, wt.BalanceAfter)
				}
				if wt.FrozenBefore != 3000 || wt.FrozenAfter != 0 {
					t.Fatalf("frozen balance mismatch: before=%d after=%d", wt.FrozenBefore, wt.FrozenAfter)
				}
				return nil
			},
		},
		mockProductRepo2{},
	)

	code, err := svc.ApproveRefund(200, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.Success {
		t.Fatalf("expected Success(%d), got %d", errs.Success, code)
	}
	if !walletTxCreated {
		t.Fatal("expected wallet transaction to be created")
	}
	if len(balanceUpdates) != 1 {
		t.Fatalf("expected 1 balance update (buyer frozen release), got %d", len(balanceUpdates))
	}
	// Should unfreeze and return to buyer available balance
	if balanceUpdates[0].userID != 100 || balanceUpdates[0].balanceDelta != 3000 || balanceUpdates[0].frozenDelta != -3000 {
		t.Fatalf("expected buyer(100) balDelta=3000 frozenDelta=-3000, got user=%d bal=%d frozen=%d",
			balanceUpdates[0].userID, balanceUpdates[0].balanceDelta, balanceUpdates[0].frozenDelta)
	}
	if updateCallCount != 1 {
		t.Fatalf("expected 1 order update, got %d", updateCallCount)
	}
	if logCallCount != 1 {
		t.Fatalf("expected 1 status log entry, got %d", logCallCount)
	}
}

func TestApproveRefund_FromReceived_SellerDebit(t *testing.T) {
	var walletTxCount int
	var balanceUpdates []struct {
		userID       uint
		balanceDelta int64
		frozenDelta  int64
	}

	order := makeRefundingOrder(1, 100, 200, 3000, model.OrderStatusReceived)
	var updateCallCount int
	var logCallCount int

	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return order, nil
			},
			transactionFunc: func(fc func(txRepo repository.OrderRepository) error) error {
				return fc(&mockOrderRepo{
					updateFunc: func(o *model.Order) error {
						updateCallCount++
						return nil
					},
					createStatusLogFunc: func(log *model.OrderStatusLog) error {
						logCallCount++
						return nil
					},
					getDBFunc: func() *gorm.DB { return nil },
				})
			},
		},
		mockUserRepo{
			findByIDFunc: func(id uint) (*model.User, error) {
				if id == 100 {
					return &model.User{
						ID:            100,
						Balance:       7000,
						FrozenBalance: 0,
					}, nil
				}
				if id == 200 {
					return &model.User{
						ID:            200,
						Balance:       10000,
						FrozenBalance: 0,
					}, nil
				}
				return nil, gorm.ErrRecordNotFound
			},
			updateBalanceFunc: func(tx *gorm.DB, userID uint, balanceDelta, frozenDelta int64) error {
				balanceUpdates = append(balanceUpdates, struct {
					userID       uint
					balanceDelta int64
					frozenDelta  int64
				}{userID, balanceDelta, frozenDelta})
				return nil
			},
			createWalletTxFunc: func(tx *gorm.DB, wt *model.WalletTransaction) error {
				walletTxCount++
				if wt.Type != model.WalletTxRefund {
					t.Fatalf("expected WalletTxRefund, got %s", wt.Type)
				}
				return nil
			},
		},
		mockProductRepo2{},
	)

	code, err := svc.ApproveRefund(200, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.Success {
		t.Fatalf("expected Success(%d), got %d", errs.Success, code)
	}
	if walletTxCount != 2 {
		t.Fatalf("expected 2 wallet transactions (seller debit + buyer credit), got %d", walletTxCount)
	}
	if len(balanceUpdates) != 2 {
		t.Fatalf("expected 2 balance updates, got %d", len(balanceUpdates))
	}
	// First update: seller debit
	if balanceUpdates[0].userID != 200 || balanceUpdates[0].balanceDelta != -3000 || balanceUpdates[0].frozenDelta != 0 {
		t.Fatalf("first update should be seller(200) debit -3000: got user=%d bal=%d frozen=%d",
			balanceUpdates[0].userID, balanceUpdates[0].balanceDelta, balanceUpdates[0].frozenDelta)
	}
	// Second update: buyer credit
	if balanceUpdates[1].userID != 100 || balanceUpdates[1].balanceDelta != 3000 || balanceUpdates[1].frozenDelta != 0 {
		t.Fatalf("second update should be buyer(100) credit 3000: got user=%d bal=%d frozen=%d",
			balanceUpdates[1].userID, balanceUpdates[1].balanceDelta, balanceUpdates[1].frozenDelta)
	}
	if updateCallCount != 1 {
		t.Fatalf("expected 1 order update, got %d", updateCallCount)
	}
	if logCallCount != 1 {
		t.Fatalf("expected 1 status log entry, got %d", logCallCount)
	}
}

func TestApproveRefund_NotSeller(t *testing.T) {
	order := makeRefundingOrder(1, 100, 200, 3000, model.OrderStatusShipped)

	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return order, nil
			},
		},
		mockUserRepo{},
		mockProductRepo2{},
	)

	// user 999 is not the seller (seller is 200)
	code, err := svc.ApproveRefund(999, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.ErrForbidden {
		t.Fatalf("expected ErrForbidden(%d), got %d", errs.ErrForbidden, code)
	}
}

func TestApproveRefund_WrongStatus(t *testing.T) {
	order := makePendingOrder(1, 100, 200, 3000) // PendingPayment, not Refunding

	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return order, nil
			},
		},
		mockUserRepo{},
		mockProductRepo2{},
	)

	code, err := svc.ApproveRefund(200, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.ErrBadRequest {
		t.Fatalf("expected ErrBadRequest(%d), got %d", errs.ErrBadRequest, code)
	}
}

// ---- Tests: RejectRefund ----

func TestRejectRefund_Success(t *testing.T) {
	order := makeRefundingOrder(1, 100, 200, 3000, model.OrderStatusShipped)
	var updateCallCount int
	var logCallCount int

	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return order, nil
			},
			transactionFunc: func(fc func(txRepo repository.OrderRepository) error) error {
				return fc(&mockOrderRepo{
					updateFunc: func(o *model.Order) error {
						updateCallCount++
						if o.Status != model.OrderStatusDispute {
							t.Fatalf("expected status Dispute(%d), got %d", model.OrderStatusDispute, o.Status)
						}
						return nil
					},
					createStatusLogFunc: func(log *model.OrderStatusLog) error {
						logCallCount++
						return nil
					},
				})
			},
		},
		mockUserRepo{},
		mockProductRepo2{},
	)

	code, err := svc.RejectRefund(200, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.Success {
		t.Fatalf("expected Success(%d), got %d", errs.Success, code)
	}
	if updateCallCount != 1 {
		t.Fatalf("expected 1 order update, got %d", updateCallCount)
	}
	if logCallCount != 1 {
		t.Fatalf("expected 1 status log entry, got %d", logCallCount)
	}
}

func TestRejectRefund_NotSeller(t *testing.T) {
	order := makeRefundingOrder(1, 100, 200, 3000, model.OrderStatusShipped)

	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return order, nil
			},
		},
		mockUserRepo{},
		mockProductRepo2{},
	)

	code, err := svc.RejectRefund(999, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.ErrForbidden {
		t.Fatalf("expected ErrForbidden(%d), got %d", errs.ErrForbidden, code)
	}
}

func TestRejectRefund_WrongStatus(t *testing.T) {
	order := makeShippedOrder(1, 100, 200, 3000) // Shipped, not Refunding

	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return order, nil
			},
		},
		mockUserRepo{},
		mockProductRepo2{},
	)

	code, err := svc.RejectRefund(200, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.ErrBadRequest {
		t.Fatalf("expected ErrBadRequest(%d), got %d", errs.ErrBadRequest, code)
	}
}

// ---- Tests: Arbitrate ----

func makeDisputeOrder(id, buyerID, sellerID uint, price int64, preRefundStatus model.OrderStatus) *model.Order {
	order := makeRefundingOrder(id, buyerID, sellerID, price, preRefundStatus)
	order.Status = model.OrderStatusDispute
	return order
}

func TestArbitrate_SupportBuyer_FromShipped(t *testing.T) {
	var walletTxCreated bool
	var balanceUpdates []struct {
		userID       uint
		balanceDelta int64
		frozenDelta  int64
	}

	order := makeDisputeOrder(1, 100, 200, 3000, model.OrderStatusShipped)
	var updateCallCount int
	var logCallCount int

	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return order, nil
			},
			transactionFunc: func(fc func(txRepo repository.OrderRepository) error) error {
				return fc(&mockOrderRepo{
					updateFunc: func(o *model.Order) error {
						updateCallCount++
						return nil
					},
					createStatusLogFunc: func(log *model.OrderStatusLog) error {
						logCallCount++
						if log.Event != "arbitrate:buyer" {
							t.Fatalf("expected event 'arbitrate:buyer', got '%s'", log.Event)
						}
						if log.OperatorType != "admin" {
							t.Fatalf("expected operator_type 'admin', got '%s'", log.OperatorType)
						}
						return nil
					},
					getDBFunc: func() *gorm.DB { return nil },
				})
			},
		},
		mockUserRepo{
			findByIDFunc: func(id uint) (*model.User, error) {
				if id == 100 {
					return &model.User{
						ID:            100,
						Balance:       7000,
						FrozenBalance: 3000,
					}, nil
				}
				if id == 999 {
					return &model.User{Role: 2}, nil
				}
				return nil, gorm.ErrRecordNotFound
			},
			updateBalanceFunc: func(tx *gorm.DB, userID uint, balanceDelta, frozenDelta int64) error {
				balanceUpdates = append(balanceUpdates, struct {
					userID       uint
					balanceDelta int64
					frozenDelta  int64
				}{userID, balanceDelta, frozenDelta})
				return nil
			},
			createWalletTxFunc: func(tx *gorm.DB, wt *model.WalletTransaction) error {
				walletTxCreated = true
				if wt.Type != model.WalletTxRefund {
					t.Fatalf("expected WalletTxRefund, got %s", wt.Type)
				}
				if wt.UserID != 100 {
					t.Fatalf("expected buyerID 100, got %d", wt.UserID)
				}
				if wt.Amount != 3000 {
					t.Fatalf("expected Amount 3000, got %d", wt.Amount)
				}
				return nil
			},
		},
		mockProductRepo2{},
	)

	code, err := svc.Arbitrate(999, 1, &ArbitrateReq{Decision: "buyer"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.Success {
		t.Fatalf("expected Success(%d), got %d", errs.Success, code)
	}
	if !walletTxCreated {
		t.Fatal("expected wallet transaction to be created")
	}
	if len(balanceUpdates) != 1 {
		t.Fatalf("expected 1 balance update (buyer unfreeze), got %d", len(balanceUpdates))
	}
	if balanceUpdates[0].userID != 100 || balanceUpdates[0].balanceDelta != 3000 || balanceUpdates[0].frozenDelta != -3000 {
		t.Fatalf("expected buyer(100) balDelta=3000 frozenDelta=-3000, got user=%d bal=%d frozen=%d",
			balanceUpdates[0].userID, balanceUpdates[0].balanceDelta, balanceUpdates[0].frozenDelta)
	}
	if updateCallCount != 1 {
		t.Fatalf("expected 1 order update, got %d", updateCallCount)
	}
	if logCallCount != 1 {
		t.Fatalf("expected 1 status log entry, got %d", logCallCount)
	}
}

func TestArbitrate_SupportBuyer_FromReceived(t *testing.T) {
	var walletTxCount int
	var balanceUpdates []struct {
		userID       uint
		balanceDelta int64
		frozenDelta  int64
	}

	order := makeDisputeOrder(1, 100, 200, 3000, model.OrderStatusReceived)

	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return order, nil
			},
			transactionFunc: func(fc func(txRepo repository.OrderRepository) error) error {
				return fc(&mockOrderRepo{
					updateFunc: func(o *model.Order) error {
						return nil
					},
					createStatusLogFunc: func(log *model.OrderStatusLog) error {
						return nil
					},
					getDBFunc: func() *gorm.DB { return nil },
				})
			},
		},
		mockUserRepo{
			findByIDFunc: func(id uint) (*model.User, error) {
				if id == 100 {
					return &model.User{
						ID:            100,
						Balance:       7000,
						FrozenBalance: 0,
					}, nil
				}
				if id == 200 {
					return &model.User{
						ID:            200,
						Balance:       10000,
						FrozenBalance: 0,
					}, nil
				}
				if id == 999 {
					return &model.User{Role: 2}, nil
				}
				return nil, gorm.ErrRecordNotFound
			},
			updateBalanceFunc: func(tx *gorm.DB, userID uint, balanceDelta, frozenDelta int64) error {
				balanceUpdates = append(balanceUpdates, struct {
					userID       uint
					balanceDelta int64
					frozenDelta  int64
				}{userID, balanceDelta, frozenDelta})
				return nil
			},
			createWalletTxFunc: func(tx *gorm.DB, wt *model.WalletTransaction) error {
				walletTxCount++
				if wt.Type != model.WalletTxRefund {
					t.Fatalf("expected WalletTxRefund, got %s", wt.Type)
				}
				return nil
			},
		},
		mockProductRepo2{},
	)

	code, err := svc.Arbitrate(999, 1, &ArbitrateReq{Decision: "buyer"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.Success {
		t.Fatalf("expected Success(%d), got %d", errs.Success, code)
	}
	if walletTxCount != 2 {
		t.Fatalf("expected 2 wallet transactions (seller debit + buyer credit), got %d", walletTxCount)
	}
	if len(balanceUpdates) != 2 {
		t.Fatalf("expected 2 balance updates, got %d", len(balanceUpdates))
	}
	// First update: seller debit -3000
	if balanceUpdates[0].userID != 200 || balanceUpdates[0].balanceDelta != -3000 || balanceUpdates[0].frozenDelta != 0 {
		t.Fatalf("first update should be seller(200) debit -3000: got user=%d bal=%d frozen=%d",
			balanceUpdates[0].userID, balanceUpdates[0].balanceDelta, balanceUpdates[0].frozenDelta)
	}
	// Second update: buyer credit +3000
	if balanceUpdates[1].userID != 100 || balanceUpdates[1].balanceDelta != 3000 || balanceUpdates[1].frozenDelta != 0 {
		t.Fatalf("second update should be buyer(100) credit 3000: got user=%d bal=%d frozen=%d",
			balanceUpdates[1].userID, balanceUpdates[1].balanceDelta, balanceUpdates[1].frozenDelta)
	}
}

func TestArbitrate_SupportSeller_FromShipped(t *testing.T) {
	var walletTxCreated bool
	var balanceUpdates []struct {
		userID       uint
		balanceDelta int64
		frozenDelta  int64
	}

	order := makeDisputeOrder(1, 100, 200, 3000, model.OrderStatusShipped)

	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return order, nil
			},
			transactionFunc: func(fc func(txRepo repository.OrderRepository) error) error {
				return fc(&mockOrderRepo{
					updateFunc: func(o *model.Order) error {
						return nil
					},
					createStatusLogFunc: func(log *model.OrderStatusLog) error {
						return nil
					},
					getDBFunc: func() *gorm.DB { return nil },
				})
			},
		},
		mockUserRepo{
			findByIDFunc: func(id uint) (*model.User, error) {
				if id == 200 {
					return &model.User{
						ID:            200,
						Balance:       5000,
						FrozenBalance: 0,
					}, nil
				}
				if id == 999 {
					return &model.User{Role: 2}, nil
				}
				return nil, gorm.ErrRecordNotFound
			},
			updateBalanceFunc: func(tx *gorm.DB, userID uint, balanceDelta, frozenDelta int64) error {
				balanceUpdates = append(balanceUpdates, struct {
					userID       uint
					balanceDelta int64
					frozenDelta  int64
				}{userID, balanceDelta, frozenDelta})
				return nil
			},
			createWalletTxFunc: func(tx *gorm.DB, wt *model.WalletTransaction) error {
				walletTxCreated = true
				if wt.Type != model.WalletTxComplete {
					t.Fatalf("expected WalletTxComplete, got %s", wt.Type)
				}
				if wt.UserID != 200 {
					t.Fatalf("expected sellerID 200, got %d", wt.UserID)
				}
				if wt.Amount != 3000 {
					t.Fatalf("expected Amount 3000, got %d", wt.Amount)
				}
				return nil
			},
		},
		mockProductRepo2{},
	)

	code, err := svc.Arbitrate(999, 1, &ArbitrateReq{Decision: "seller"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.Success {
		t.Fatalf("expected Success(%d), got %d", errs.Success, code)
	}
	if !walletTxCreated {
		t.Fatal("expected wallet transaction to be created")
	}
	if len(balanceUpdates) != 2 {
		t.Fatalf("expected 2 balance updates (buyer frozen release + seller credit), got %d", len(balanceUpdates))
	}
	// First update: buyer frozen release (0, -price)
	if balanceUpdates[0].userID != 100 || balanceUpdates[0].balanceDelta != 0 || balanceUpdates[0].frozenDelta != -3000 {
		t.Fatalf("first update should be buyer(100) frozen release: got user=%d bal=%d frozen=%d",
			balanceUpdates[0].userID, balanceUpdates[0].balanceDelta, balanceUpdates[0].frozenDelta)
	}
	// Second update: seller credit (+price, 0)
	if balanceUpdates[1].userID != 200 || balanceUpdates[1].balanceDelta != 3000 || balanceUpdates[1].frozenDelta != 0 {
		t.Fatalf("second update should be seller(200) credit 3000: got user=%d bal=%d frozen=%d",
			balanceUpdates[1].userID, balanceUpdates[1].balanceDelta, balanceUpdates[1].frozenDelta)
	}
}

func TestArbitrate_SupportSeller_FromReceived(t *testing.T) {
	order := makeDisputeOrder(1, 100, 200, 3000, model.OrderStatusReceived)
	var updateCallCount int

	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return order, nil
			},
			transactionFunc: func(fc func(txRepo repository.OrderRepository) error) error {
				return fc(&mockOrderRepo{
					updateFunc: func(o *model.Order) error {
						updateCallCount++
						return nil
					},
					createStatusLogFunc: func(log *model.OrderStatusLog) error {
						return nil
					},
					getDBFunc: func() *gorm.DB { return nil },
				})
			},
		},
		mockUserRepo{
			// Seller already has the money, no wallet operations needed
			findByIDFunc: func(id uint) (*model.User, error) {
				if id == 999 {
					return &model.User{Role: 2}, nil
				}
				return nil, gorm.ErrRecordNotFound
			},
		},
		mockProductRepo2{},
	)

	code, err := svc.Arbitrate(999, 1, &ArbitrateReq{Decision: "seller"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.Success {
		t.Fatalf("expected Success(%d), got %d", errs.Success, code)
	}
	if updateCallCount != 1 {
		t.Fatalf("expected 1 order update, got %d", updateCallCount)
	}
}

func TestArbitrate_InvalidDecision(t *testing.T) {
	order := makeDisputeOrder(1, 100, 200, 3000, model.OrderStatusShipped)

	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return order, nil
			},
		},
			mockUserRepo{
				findByIDFunc: func(id uint) (*model.User, error) {
					return &model.User{Role: 2}, nil
				},
			},
			mockProductRepo2{},
		)

	code, err := svc.Arbitrate(999, 1, &ArbitrateReq{Decision: "invalid"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.ErrBadRequest {
		t.Fatalf("expected ErrBadRequest(%d), got %d", errs.ErrBadRequest, code)
	}
}

func TestArbitrate_OrderNotFound(t *testing.T) {
	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return nil, gorm.ErrRecordNotFound
			},
		},
			mockUserRepo{
				findByIDFunc: func(id uint) (*model.User, error) {
					return &model.User{Role: 2}, nil
				},
			},
			mockProductRepo2{},
		)

	code, err := svc.Arbitrate(999, 999, &ArbitrateReq{Decision: "buyer"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.ErrOrderNotFound {
		t.Fatalf("expected ErrOrderNotFound(%d), got %d", errs.ErrOrderNotFound, code)
	}
}

func TestArbitrate_WrongStatus(t *testing.T) {
	order := makePendingOrder(1, 100, 200, 3000) // PendingPayment, not Dispute

	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return order, nil
			},
		},
			mockUserRepo{
				findByIDFunc: func(id uint) (*model.User, error) {
					return &model.User{Role: 2}, nil
				},
			},
			mockProductRepo2{},
		)

	code, err := svc.Arbitrate(999, 1, &ArbitrateReq{Decision: "buyer"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != errs.ErrBadRequest {
		t.Fatalf("expected ErrBadRequest(%d), got %d", errs.ErrBadRequest, code)
	}
}
