package model

import "time"

// WalletTxType 钱包交易类型
type WalletTxType string

const (
	WalletTxPay        WalletTxType = "pay"             // 付款（买家余额减少，冻结增加）
	WalletTxComplete   WalletTxType = "complete"        // 完成订单（卖家收到款项）
	WalletTxRefund     WalletTxType = "refund"          // 退款（冻结解冻，退回买家）
	WalletTxTopUp      WalletTxType = "top_up"          // 充值
	WalletTxWithdraw   WalletTxType = "withdraw"        // 提现
)

// WalletTransaction 钱包交易流水
// 记录每一笔资金变动的完整上下文，支持审计和余额对账
type WalletTransaction struct {
	ID            uint         `gorm:"primarykey" json:"id"`
	UserID        uint         `gorm:"not null;index" json:"user_id"`
	Type          WalletTxType `gorm:"size:20;not null;index" json:"type"`
	Amount        int64        `gorm:"not null" json:"amount"`                         // 变动金额（正数=增加，负数=减少）
	OrderID       uint         `gorm:"not null;index" json:"order_id"`                 // 关联订单ID
	BalanceBefore int64        `gorm:"not null" json:"balance_before"`                 // 变动前可用余额
	BalanceAfter  int64        `gorm:"not null" json:"balance_after"`                  // 变动后可用余额
	FrozenBefore  int64        `gorm:"default:0;not null" json:"frozen_before"`        // 变动前冻结余额
	FrozenAfter   int64        `gorm:"default:0;not null" json:"frozen_after"`         // 变动后冻结余额
	Description   string       `gorm:"size:200" json:"description,omitempty"`          // 交易描述
	CreatedAt     time.Time    `json:"created_at"`
}

func (WalletTransaction) TableName() string {
	return "wallet_transactions"
}
