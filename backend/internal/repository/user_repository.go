package repository

import (
	"errors"

	"flea-market/internal/model"

	"gorm.io/gorm"
)

// 余额不足错误，供上层识别并转换为业务错误码
var ErrBalanceInsufficient = errors.New("insufficient balance")

type UserRepository interface {
	Create(user *model.User) error
	FindByID(id uint) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	Update(user *model.User) error

	// 钱包操作
	UpdateBalance(tx *gorm.DB, userID uint, balanceDelta, frozenDelta int64) error
	CreateWalletTransaction(tx *gorm.DB, wt *model.WalletTransaction) error
	ListWalletTransactions(userID uint, page, pageSize int) ([]model.WalletTransaction, int64, error)
	ListWalletTransactionsByTx(tx *gorm.DB, userID uint) ([]model.WalletTransaction, error)

	// 事务支持
	Transaction(fc func(tx *gorm.DB) error) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// UpdateBalance 原子性更新用户余额
// balanceDelta: 可用余额变动（正数增加，负数减少）
// frozenDelta: 冻结余额变动（正数增加，负数减少）
// tx 可选：为 nil 时使用仓库默认数据库连接
// 使用原生 SQL 确保原子性和乐观锁效果
// 返回 ErrInsufficientBalance 当余额不足时更新影响 0 行
func (r *userRepository) UpdateBalance(tx *gorm.DB, userID uint, balanceDelta, frozenDelta int64) error {
	db := tx
	if db == nil {
		db = r.db
	}
	result := db.Model(&model.User{}).
		Where("id = ? AND balance + ? >= 0 AND frozen_balance + ? >= 0", userID, balanceDelta, frozenDelta).
		Updates(map[string]interface{}{
			"balance":        gorm.Expr("balance + ?", balanceDelta),
			"frozen_balance": gorm.Expr("frozen_balance + ?", frozenDelta),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrBalanceInsufficient
	}
	return nil
}

func (r *userRepository) CreateWalletTransaction(tx *gorm.DB, wt *model.WalletTransaction) error {
	db := tx
	if db == nil {
		db = r.db
	}
	return db.Create(wt).Error
}

func (r *userRepository) Transaction(fc func(tx *gorm.DB) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return fc(tx)
	})
}

func (r *userRepository) ListWalletTransactions(userID uint, page, pageSize int) ([]model.WalletTransaction, int64, error) {
	var transactions []model.WalletTransaction
	var total int64

	query := r.db.Model(&model.WalletTransaction{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&transactions).Error
	if err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

func (r *userRepository) ListWalletTransactionsByTx(tx *gorm.DB, userID uint) ([]model.WalletTransaction, error) {
	var transactions []model.WalletTransaction
	err := tx.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&transactions).Error
	return transactions, err
}
