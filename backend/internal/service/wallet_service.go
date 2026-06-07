package service

import (
	"fmt"
	"time"

	"flea-market/internal/model"
	"flea-market/internal/repository"
	"flea-market/pkg/errors"

	"gorm.io/gorm"
)

// WalletService 钱包服务
// 提供余额查询、充值、交易流水查询等功能
type WalletService struct {
	userRepo repository.UserRepository
}

func NewWalletService(userRepo repository.UserRepository) *WalletService {
	return &WalletService{userRepo: userRepo}
}

// WalletInfo 钱包信息
type WalletInfo struct {
	Balance       int64 `json:"balance"`
	FrozenBalance int64 `json:"frozen_balance"`
	TotalBalance  int64 `json:"total_balance"` // 可用 + 冻结
}

// GetWalletInfo 获取用户钱包信息
func (s *WalletService) GetWalletInfo(userID uint) (*WalletInfo, int, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.ErrInternal, err
	}

	return &WalletInfo{
		Balance:       user.Balance,
		FrozenBalance: user.FrozenBalance,
		TotalBalance:  user.Balance + user.FrozenBalance,
	}, errors.Success, nil
}

// TopUpReq 充值请求
type TopUpReq struct {
	Amount int64 `json:"amount" binding:"required,min=1,max=10000000"` // 上限 100,000.00 元（单位：分）
}

// TopUp 充值（直接增加可用余额）
func (s *WalletService) TopUp(userID uint, req *TopUpReq) (*WalletInfo, int, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.ErrInternal, err
	}

	// Execute balance update and transaction record atomically
	if err := s.userRepo.Transaction(func(tx *gorm.DB) error {
		// Read latest balance inside transaction for audit accuracy
		user, err = s.userRepo.FindByID(userID)
		if err != nil {
			return err
		}

		if err := s.userRepo.UpdateBalance(tx, userID, req.Amount, 0); err != nil {
			return err
		}

		// Record top-up transaction with accurate balance snapshot
		wt := &model.WalletTransaction{
			UserID:        userID,
			Type:          model.WalletTxTopUp,
			Amount:        req.Amount,
			OrderID:       0,
			BalanceBefore: user.Balance,
			BalanceAfter:  user.Balance + req.Amount,
			FrozenBefore:  user.FrozenBalance,
			FrozenAfter:   user.FrozenBalance,
			Description:   fmt.Sprintf("Top up %.2f", float64(req.Amount)/100),
		}
		return s.userRepo.CreateWalletTransaction(tx, wt)
	}); err != nil {
		return nil, errors.ErrInternal, err
	}

	// 更新后的用户信息
	updated, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.ErrInternal, err
	}

	return &WalletInfo{
		Balance:       updated.Balance,
		FrozenBalance: updated.FrozenBalance,
		TotalBalance:  updated.Balance + updated.FrozenBalance,
	}, errors.Success, nil
}

// TransactionItem 交易流水条目
type TransactionItem struct {
	ID          uint   `json:"id"`
	Type        string `json:"type"`
	TypeName    string `json:"type_name"`
	Amount      int64  `json:"amount"`
	OrderID     uint   `json:"order_id"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at"`
}

// WalletTxTypeNames 交易类型中文映射
var WalletTxTypeNames = map[model.WalletTxType]string{
	model.WalletTxPay:      "付款",
	model.WalletTxComplete: "收款",
	model.WalletTxRefund:   "退款",
	model.WalletTxTopUp:    "充值",
	model.WalletTxWithdraw: "提现",
}

// ListTransactions 获取交易流水
func (s *WalletService) ListTransactions(userID uint, page, pageSize int) ([]TransactionItem, int64, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	transactions, total, err := s.userRepo.ListWalletTransactions(userID, page, pageSize)
	if err != nil {
		return nil, 0, errors.ErrInternal, err
	}

	items := make([]TransactionItem, len(transactions))
	for i, t := range transactions {
		typeName := WalletTxTypeNames[t.Type]
		if typeName == "" {
			typeName = string(t.Type)
		}
		items[i] = TransactionItem{
			ID:          t.ID,
			Type:        string(t.Type),
			TypeName:    typeName,
			Amount:      t.Amount,
			OrderID:     t.OrderID,
			Description: t.Description,
			CreatedAt:   t.CreatedAt.Format(time.RFC3339),
		}
	}

	return items, total, errors.Success, nil
}
