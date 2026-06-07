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
	createFunc                  func(order *model.Order) error
	findByIDFunc                func(id uint) (*model.Order, error)
	findByOrderNoFunc           func(orderNo string) (*model.Order, error)
	updateFunc                  func(order *model.Order) error
	listByBuyerFunc             func(buyerID uint, status *model.OrderStatus, page, pageSize int) ([]model.Order, int64, error)
	listBySellerFunc            func(sellerID uint, status *model.OrderStatus, page, pageSize int) ([]model.Order, int64, error)
	listByStatusFunc            func(status model.OrderStatus, page, pageSize int) ([]model.Order, int64, error)
	decrementStockFunc          func(productID uint, quantity int) (bool, error)
	incrementStockFunc          func(productID uint, quantity int) error
	createStatusLogFunc         func(log *model.OrderStatusLog) error
	listStatusLogsFunc          func(orderID uint) ([]model.OrderStatusLog, error)
	transactionFunc             func(fc func(txRepo repository.OrderRepository) error) error
	findUserByIDFunc            func(id uint) (*model.User, error)
	updateUserBalanceFunc       func(userID uint, balanceDelta, frozenDelta int64) error
	createWalletTransactionFunc func(wt *model.WalletTransaction) error
	updateStatusConditionalFunc func(id uint, newStatus, expectedStatus model.OrderStatus) (bool, error)
}

func (m *mockOrderRepo) Create(order *model.Order) error { return m.createFunc(order) }
func (m *mockOrderRepo) FindByID(id uint) (*model.Order, error) { return m.findByIDFunc(id) }
func (m *mockOrderRepo) FindByOrderNo(orderNo string) (*model.Order, error) { return m.findByOrderNoFunc(orderNo) }
func (m *mockOrderRepo) Update(order *model.Order) error { return m.updateFunc(order) }
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
func (m *mockOrderRepo) UpdateUserBalance(userID uint, balanceDelta, frozenDelta int64) error {
	if m.updateUserBalanceFunc != nil {
		return m.updateUserBalanceFunc(userID, balanceDelta, frozenDelta)
	}
	return nil
}
func (m *mockOrderRepo) CreateWalletTransaction(wt *model.WalletTransaction) error {
	if m.createWalletTransactionFunc != nil {
		return m.createWalletTransactionFunc(wt)
	}
	return nil
}
func (m *mockOrderRepo) FindUserByID(id uint) (*model.User, error) {
	if m.findUserByIDFunc != nil {
		return m.findUserByIDFunc(id)
	}
	return nil, nil
}
func (m *mockOrderRepo) UpdateStatusConditional(id uint, newStatus, expectedStatus model.OrderStatus) (bool, error) {
	if m.updateStatusConditionalFunc != nil {
		return m.updateStatusConditionalFunc(id, newStatus, expectedStatus)
	}
	return true, nil
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
	listWalletTransactionsByTxFunc func(tx *gorm.DB, userID uint) ([]model.WalletTransaction, error)
	transactionFunc          func(fc func(tx *gorm.DB) error) error
}

func (m *mockUserRepo) Create(user *model.User) error { return m.createFunc(user) }
func (m *mockUserRepo) FindByID(id uint) (*model.User, error) { return m.findByIDFunc(id) }
func (m *mockUserRepo) FindByEmail(email string) (*model.User, error) { return m.findByEmailFunc(email) }
func (m *mockUserRepo) Update(user *model.User) error { return m.updateFunc(user) }
func (m *mockUserRepo) UpdateBalance(tx *gorm.DB, userID uint, balanceDelta, frozenDelta int64) error {
	return m.updateBalanceFunc(tx, userID, balanceDelta, frozenDelta)
}
func (m *mockUserRepo) CreateWalletTransaction(tx *gorm.DB, wt *model.WalletTransaction) error {
	return m.createWalletTxFunc(tx, wt)
}
func (m *mockUserRepo) ListWalletTransactions(userID uint, page, pageSize int) ([]model.WalletTransaction, int64, error) {
	return m.listWalletTransactionsFunc(userID, page, pageSize)
}
func (m *mockUserRepo) ListWalletTransactionsByTx(tx *gorm.DB, userID uint) ([]model.WalletTransaction, error) {
	return m.listWalletTransactionsByTxFunc(tx, userID)
}
func (m *mockUserRepo) Transaction(fc func(tx *gorm.DB) error) error {
	if m.transactionFunc != nil {
		return m.transactionFunc(fc)
	}
	return nil
}

// ---- Mock ProductRepository ----

type mockProductRepo2 struct {
	findByIDFunc func(id uint) (*model.Product, error)
}

func (m *mockProductRepo2) Create(product *model.Product) error { panic("unexpected") }
func (m *mockProductRepo2) FindByID(id uint) (*model.Product, error) { return m.findByIDFunc(id) }
func (m *mockProductRepo2) Update(product *model.Product) error { panic("unexpected") }
func (m *mockProductRepo2) Delete(id uint) error { panic("unexpected") }
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
func (m *mockProductRepo2) CreateImage(image *model.ProductImage) error { panic("unexpected") }
func (m *mockProductRepo2) UpdateImage(image *model.ProductImage) error { panic("unexpected") }
func (m *mockProductRepo2) DeleteImage(id uint) error { panic("unexpected") }
func (m *mockProductRepo2) FindImageByID(id uint) (*model.ProductImage, error) { panic("unexpected") }
func (m *mockProductRepo2) FindImagesByIDs(ids []uint) ([]model.ProductImage, error) { panic("unexpected") }
func (m *mockProductRepo2) DeleteImagesByProduct(productID uint) error { panic("unexpected") }
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
		ID:       id,
		OrderNo:  "2026010112000000001",
		ProductID: 1,
		BuyerID:  buyerID,
		SellerID: sellerID,
		Title:    "测试商品",
		Price:    price,
		Status:   model.OrderStatusPendingPayment,
	}
}

// ---- Tests: Pay with wallet escrow ----

func TestPay_SufficientBalance(t *testing.T) {
	var walletTxCreated bool

	order := makePendingOrder(1, 100, 200, 3000)

	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return order, nil
			},
			transactionFunc: func(fc func(txRepo repository.OrderRepository) error) error {
				return fc(&mockOrderRepo{
					findUserByIDFunc: func(id uint) (*model.User, error) {
						if id != 100 {
							t.Fatalf("expected buyerID 100, got %d", id)
						}
						return &model.User{
							ID:            100,
							Balance:       10000,
							FrozenBalance: 0,
						}, nil
					},
					updateStatusConditionalFunc: func(id uint, newStatus, expectedStatus model.OrderStatus) (bool, error) {
						return true, nil
					},
					createStatusLogFunc: func(log *model.OrderStatusLog) error {
						return nil
					},
					updateUserBalanceFunc: func(userID uint, balanceDelta, frozenDelta int64) error {
						if balanceDelta != -3000 {
							t.Fatalf("expected balanceDelta -3000, got %d", balanceDelta)
						}
						if frozenDelta != 3000 {
							t.Fatalf("expected frozenDelta 3000, got %d", frozenDelta)
						}
						return nil
					},
					createWalletTransactionFunc: func(wt *model.WalletTransaction) error {
						walletTxCreated = true
						if wt.Type != model.WalletTxPay {
							t.Fatalf("expected WalletTxPay, got %s", wt.Type)
						}
						if wt.Amount != -3000 {
							t.Fatalf("expected Amount -3000, got %d", wt.Amount)
						}
						return nil
					},
				})
			},
		},
		mockUserRepo{
			findByIDFunc: func(id uint) (*model.User, error) {
				return &model.User{
					ID:            100,
					Balance:       10000,
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
	if code != errs.Success {
		t.Fatalf("expected Success(%d), got %d", errs.Success, code)
	}
	if !walletTxCreated {
		t.Fatal("expected wallet transaction to be created")
	}
}

func TestPay_InsufficientBalance(t *testing.T) {
	order := makePendingOrder(1, 100, 200, 10000)

	svc := newTestOrderService(
		mockOrderRepo{
			findByIDFunc: func(id uint) (*model.Order, error) {
				return order, nil
			},
			transactionFunc: func(fc func(txRepo repository.OrderRepository) error) error {
				return fc(&mockOrderRepo{
					findUserByIDFunc: func(id uint) (*model.User, error) {
						return &model.User{
							ID:            100,
							Balance:       5000,
							FrozenBalance: 0,
						}, nil
					},
					updateStatusConditionalFunc: func(id uint, newStatus, expectedStatus model.OrderStatus) (bool, error) {
						return true, nil
					},
					createStatusLogFunc: func(log *model.OrderStatusLog) error {
						return nil
					},
					updateUserBalanceFunc: func(userID uint, balanceDelta, frozenDelta int64) error {
						return repository.ErrBalanceInsufficient
					},
				})
			},
		},
		mockUserRepo{
			findByIDFunc: func(id uint) (*model.User, error) {
				return &model.User{
					ID:            100,
					Balance:       5000,
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
	order.Status = model.OrderStatusPaid

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
			// BalanceBefore should reflect the value read inside the transaction
			if wt.BalanceBefore != 5000 {
				t.Fatalf("expected BalanceBefore 5000, got %d", wt.BalanceBefore)
			}
			return nil
		},
		transactionFunc: func(fc func(tx *gorm.DB) error) error {
			// The call inside TopUp now reads user balance inside the transaction
			return fc(nil)
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
