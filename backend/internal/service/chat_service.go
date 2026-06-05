package service

import (
	"flea-market/internal/model"
	"flea-market/internal/repository"
	"flea-market/pkg/errors"
)

type ChatService struct {
	convRepo repository.ConversationRepository
	msgRepo  repository.MessageRepository
}

func NewChatService(convRepo repository.ConversationRepository, msgRepo repository.MessageRepository) *ChatService {
	return &ChatService{convRepo: convRepo, msgRepo: msgRepo}
}

// CreateConversationReq 创建会话请求
type CreateConversationReq struct {
	ProductID uint `json:"product_id" binding:"required"`
	BuyerID   uint `json:"buyer_id" binding:"required"`
	SellerID  uint `json:"seller_id" binding:"required"`
}

// ConversationResp 会话响应
type ConversationResp struct {
	ID           uint   `json:"id"`
	ProductID    uint   `json:"product_id"`
	BuyerID      uint   `json:"buyer_id"`
	SellerID     uint   `json:"seller_id"`
	LastMsg      string `json:"last_msg"`
	LastMsgID    uint   `json:"last_msg_id"`
	UnreadBuyer  int    `json:"unread_buyer"`
	UnreadSeller int    `json:"unread_seller"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// MessageResp 消息响应
type MessageResp struct {
	ID             uint   `json:"id"`
	ConversationID uint   `json:"conversation_id"`
	SenderID       uint   `json:"sender_id"`
	Content        string `json:"content"`
	MsgType        int    `json:"msg_type"`
	Status         int    `json:"status"`
	CreatedAt      string `json:"created_at"`
}

// SendMessageReq 发送消息请求
type SendMessageReq struct {
	ConversationID uint   `json:"conversation_id" binding:"required"`
	SenderID       uint   `json:"sender_id" binding:"required"`
	Content        string `json:"content" binding:"required"`
	MsgType        int    `json:"msg_type"`
}

func toConversationResp(conv *model.Conversation) *ConversationResp {
	return &ConversationResp{
		ID:           conv.ID,
		ProductID:    conv.ProductID,
		BuyerID:      conv.BuyerID,
		SellerID:     conv.SellerID,
		LastMsg:      conv.LastMsg,
		LastMsgID:    conv.LastMsgID,
		UnreadBuyer:  conv.UnreadBuyer,
		UnreadSeller: conv.UnreadSeller,
		CreatedAt:    conv.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    conv.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toMessageResp(msg *model.Message) *MessageResp {
	return &MessageResp{
		ID:             msg.ID,
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		Content:        msg.Content,
		MsgType:        msg.MsgType,
		Status:         msg.Status,
		CreatedAt:      msg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// CreateConversation 创建或获取已有会话
func (s *ChatService) CreateConversation(req *CreateConversationReq) (*ConversationResp, int, error) {
	existing, err := s.convRepo.FindByProductAndBuyer(req.ProductID, req.BuyerID)
	if err == nil && existing != nil {
		return toConversationResp(existing), errors.Success, nil
	}

	conv := &model.Conversation{
		ProductID:    req.ProductID,
		BuyerID:      req.BuyerID,
		SellerID:     req.SellerID,
		UnreadBuyer:  0,
		UnreadSeller: 0,
	}
	if err := s.convRepo.Create(conv); err != nil {
		return nil, errors.ErrInternal, err
	}
	return toConversationResp(conv), errors.Success, nil
}

// ListConversations 获取用户的会话列表
func (s *ChatService) ListConversations(userID uint) ([]*ConversationResp, int, error) {
	convs, err := s.convRepo.ListByUser(userID)
	if err != nil {
		return nil, errors.ErrInternal, err
	}
	resp := make([]*ConversationResp, len(convs))
	for i, conv := range convs {
		resp[i] = toConversationResp(&conv)
	}
	return resp, errors.Success, nil
}

// GetMessages 获取消息历史（分页）
func (s *ChatService) GetMessages(convID, userID uint, page, pageSize int) ([]*MessageResp, int, int, error) {
	conv, err := s.convRepo.FindByID(convID)
	if err != nil {
		return nil, 0, errors.ErrForbidden, nil
	}
	if conv.BuyerID != userID && conv.SellerID != userID {
		return nil, 0, errors.ErrForbidden, nil
	}

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	msgs, err := s.msgRepo.ListByConversation(convID, pageSize, offset)
	if err != nil {
		return nil, 0, errors.ErrInternal, err
	}

	resp := make([]*MessageResp, len(msgs))
	for i, msg := range msgs {
		resp[i] = toMessageResp(&msg)
	}

	// 标记为己读
	isBuyer := conv.BuyerID == userID
	_ = s.msgRepo.BatchUpdateStatus(convID, userID, 3) // 3=read
	_ = s.convRepo.ResetUnread(convID, isBuyer)

	return resp, len(msgs), errors.Success, nil
}

// SendMessage 保存消息
func (s *ChatService) SendMessage(req *SendMessageReq) (*MessageResp, int, error) {
	conv, err := s.convRepo.FindByID(req.ConversationID)
	if err != nil {
		return nil, errors.ErrForbidden, nil
	}
	if conv.BuyerID != req.SenderID && conv.SellerID != req.SenderID {
		return nil, errors.ErrForbidden, nil
	}

	if req.MsgType == 0 {
		req.MsgType = 1
	}

	msg := &model.Message{
		ConversationID: req.ConversationID,
		SenderID:       req.SenderID,
		Content:        req.Content,
		MsgType:        req.MsgType,
		Status:         1, // sent
	}
	if err := s.msgRepo.Create(msg); err != nil {
		return nil, errors.ErrInternal, err
	}

	// 更新会话的最后一条消息
	conv.LastMsg = req.Content
	conv.LastMsgID = msg.ID
	_ = s.convRepo.Update(conv)

	// 更新接收方的未读计数
	isBuyer := conv.BuyerID == req.SenderID
	_ = s.convRepo.UpdateUnread(req.ConversationID, isBuyer, 1)

	return toMessageResp(msg), errors.Success, nil
}
