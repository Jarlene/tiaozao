package model

import (
	"time"

	"gorm.io/gorm"
)

// Conversation 会话
type Conversation struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	ProductID uint           `gorm:"not null;index" json:"product_id"`
	BuyerID   uint           `gorm:"not null;index" json:"buyer_id"`
	SellerID  uint           `gorm:"not null;index" json:"seller_id"`
	LastMsgID uint           `gorm:"default:0" json:"last_msg_id"`
	LastMsg   string         `gorm:"size:500;default:''" json:"last_msg"`
	UnreadBuyer  int        `gorm:"default:0" json:"unread_buyer"`
	UnreadSeller int        `gorm:"default:0" json:"unread_seller"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Conversation) TableName() string {
	return "conversations"
}

// Message 消息
type Message struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	ConversationID uint      `gorm:"not null;index" json:"conversation_id"`
	SenderID       uint      `gorm:"not null;index" json:"sender_id"`
	Content        string    `gorm:"type:text;not null" json:"content"`
	MsgType        int       `gorm:"default:1;not null" json:"msg_type"` // 1=text, 2=image, 3=system
	Status         int       `gorm:"default:1;not null" json:"status"`   // 1=sent, 2=delivered, 3=read
	CreatedAt      time.Time `json:"created_at"`
}

func (Message) TableName() string {
	return "messages"
}
