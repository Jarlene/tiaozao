package repository

import (
	"flea-market/internal/model"

	"gorm.io/gorm"
)

type ConversationRepository interface {
	Create(conv *model.Conversation) error
	FindByID(id uint) (*model.Conversation, error)
	FindByProductAndBuyer(productID, buyerID uint) (*model.Conversation, error)
	ListByUser(userID uint) ([]model.Conversation, error)
	Update(conv *model.Conversation) error
	UpdateUnread(convID uint, isBuyer bool, delta int) error
	ResetUnread(convID uint, isBuyer bool) error
}

type conversationRepository struct {
	db *gorm.DB
}

func NewConversationRepository(db *gorm.DB) ConversationRepository {
	return &conversationRepository{db: db}
}

func (r *conversationRepository) Create(conv *model.Conversation) error {
	return r.db.Create(conv).Error
}

func (r *conversationRepository) FindByID(id uint) (*model.Conversation, error) {
	var conv model.Conversation
	err := r.db.First(&conv, id).Error
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *conversationRepository) FindByProductAndBuyer(productID, buyerID uint) (*model.Conversation, error) {
	var conv model.Conversation
	err := r.db.Where("product_id = ? AND buyer_id = ?", productID, buyerID).First(&conv).Error
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *conversationRepository) ListByUser(userID uint) ([]model.Conversation, error) {
	var convs []model.Conversation
	err := r.db.Where("buyer_id = ? OR seller_id = ?", userID, userID).
		Order("updated_at DESC").Find(&convs).Error
	if err != nil {
		return nil, err
	}
	return convs, nil
}

func (r *conversationRepository) Update(conv *model.Conversation) error {
	return r.db.Save(conv).Error
}

func (r *conversationRepository) UpdateUnread(convID uint, isBuyer bool, delta int) error {
	field := "unread_seller"
	if isBuyer {
		field = "unread_buyer"
	}
	return r.db.Model(&model.Conversation{}).Where("id = ?", convID).
		UpdateColumn(field, gorm.Expr(field+" + ?", delta)).Error
}

func (r *conversationRepository) ResetUnread(convID uint, isBuyer bool) error {
	field := "unread_seller"
	if isBuyer {
		field = "unread_buyer"
	}
	return r.db.Model(&model.Conversation{}).Where("id = ?", convID).
		UpdateColumn(field, 0).Error
}

// MessageRepository 消息数据访问接口
type MessageRepository interface {
	Create(msg *model.Message) error
	FindByID(id uint) (*model.Message, error)
	ListByConversation(convID uint, limit, offset int) ([]model.Message, error)
	BatchUpdateStatus(convID uint, senderID uint, status int) error
	GetLastMessage(convID uint) (*model.Message, error)
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(msg *model.Message) error {
	return r.db.Create(msg).Error
}

func (r *messageRepository) FindByID(id uint) (*model.Message, error) {
	var msg model.Message
	err := r.db.First(&msg, id).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *messageRepository) ListByConversation(convID uint, limit, offset int) ([]model.Message, error) {
	var msgs []model.Message
	err := r.db.Where("conversation_id = ?", convID).
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&msgs).Error
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

func (r *messageRepository) BatchUpdateStatus(convID uint, senderID uint, status int) error {
	return r.db.Model(&model.Message{}).
		Where("conversation_id = ? AND sender_id != ? AND status < ?", convID, senderID, status).
		UpdateColumn("status", status).Error
}

func (r *messageRepository) GetLastMessage(convID uint) (*model.Message, error) {
	var msg model.Message
	err := r.db.Where("conversation_id = ?", convID).
		Order("created_at DESC").First(&msg).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}
