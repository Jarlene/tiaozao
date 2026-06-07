package service

import (
	"flea-market/internal/model"
	"flea-market/internal/repository"
	"flea-market/pkg/errors"
)

type ChatService struct {
	convRepo    repository.ConversationRepository
	msgRepo     repository.MessageRepository
	userRepo    repository.UserRepository
	productRepo repository.ProductRepository
}

func NewChatService(convRepo repository.ConversationRepository, msgRepo repository.MessageRepository, userRepo repository.UserRepository, productRepo repository.ProductRepository) *ChatService {
	return &ChatService{convRepo: convRepo, msgRepo: msgRepo, userRepo: userRepo, productRepo: productRepo}
}

// CreateConversationReq 创建会话请求
type CreateConversationReq struct {
	ProductID uint `json:"product_id" binding:"required"`
	BuyerID   uint `json:"buyer_id" binding:"required"`
	SellerID  uint `json:"seller_id" binding:"required"`
}

// ProductSimpleInfo 商品简要信息
type ProductSimpleInfo struct {
	ID       uint   `json:"id"`
	Title    string `json:"title"`
	Price    int64  `json:"price"`
	ImageURL string `json:"image_url"`
}

// ConversationResp 会话响应
type ConversationResp struct {
	ID           uint               `json:"id"`
	ProductID    uint               `json:"product_id"`
	BuyerID      uint               `json:"buyer_id"`
	SellerID     uint               `json:"seller_id"`
	LastMsg      string             `json:"last_msg"`
	LastMsgID    uint               `json:"last_msg_id"`
	UnreadBuyer  int                `json:"unread_buyer"`
	UnreadSeller int                `json:"unread_seller"`
	CreatedAt    string             `json:"created_at"`
	UpdatedAt    string             `json:"updated_at"`
	Buyer        *UserSimpleProfile `json:"buyer,omitempty"`
	Seller       *UserSimpleProfile `json:"seller,omitempty"`
	Product      *ProductSimpleInfo `json:"product,omitempty"`
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

func (s *ChatService) enrichConversation(conv *ConversationResp) {
	// 补充买家、卖家信息
	if buyer, err := s.userRepo.FindByID(conv.BuyerID); err == nil {
		conv.Buyer = &UserSimpleProfile{
			ID:        buyer.ID,
			Nickname:  buyer.Nickname,
			AvatarURL: buyer.AvatarURL,
		}
	}
	if seller, err := s.userRepo.FindByID(conv.SellerID); err == nil {
		conv.Seller = &UserSimpleProfile{
			ID:        seller.ID,
			Nickname:  seller.Nickname,
			AvatarURL: seller.AvatarURL,
		}
	}

	// 补充商品信息
	if product, err := s.productRepo.FindByID(conv.ProductID); err == nil {
		info := &ProductSimpleInfo{
			ID:    product.ID,
			Title: product.Title,
			Price: product.Price,
		}
		if len(product.Images) > 0 {
			info.ImageURL = product.Images[0].URL
		}
		conv.Product = info
	}
}

func (s *ChatService) enrichConversations(convs []*ConversationResp) {
	for _, conv := range convs {
		s.enrichConversation(conv)
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
		resp := toConversationResp(existing)
		s.enrichConversation(resp)
		return resp, errors.Success, nil
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
	resp := toConversationResp(conv)
	s.enrichConversation(resp)
	return resp, errors.Success, nil
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
	s.enrichConversations(resp)
	return resp, errors.Success, nil
}

// GetMessages 获取消息历史（游标分页）
// cursorID=0 时获取最新消息，否则获取比 cursorID 更早的消息
func (s *ChatService) GetMessages(convID, userID uint, cursorID uint, limit int) ([]*MessageResp, bool, int, error) {
	conv, err := s.convRepo.FindByID(convID)
	if err != nil {
		return nil, false, errors.ErrForbidden, nil
	}
	if conv.BuyerID != userID && conv.SellerID != userID {
		return nil, false, errors.ErrForbidden, nil
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	msgs, err := s.msgRepo.ListByConversationCursor(convID, cursorID, limit+1)
	if err != nil {
		return nil, false, errors.ErrInternal, err
	}

	hasMore := len(msgs) > limit
	if hasMore {
		msgs = msgs[:limit]
	}

	resp := make([]*MessageResp, len(msgs))
	for i, msg := range msgs {
		resp[i] = toMessageResp(&msg)
	}

	// 标记为己读
	if cursorID == 0 {
		isBuyer := conv.BuyerID == userID
		_ = s.msgRepo.BatchUpdateStatusAll(convID, userID, 3) // 3=read
		_ = s.convRepo.ResetUnread(convID, isBuyer)
	}

	return resp, hasMore, errors.Success, nil
}

// MarkConversationRead 标记会话已读
func (s *ChatService) MarkConversationRead(convID, userID uint) (int, error) {
	conv, err := s.convRepo.FindByID(convID)
	if err != nil {
		return errors.ErrForbidden, nil
	}
	if conv.BuyerID != userID && conv.SellerID != userID {
		return errors.ErrForbidden, nil
	}

	isBuyer := conv.BuyerID == userID
	if err := s.msgRepo.BatchUpdateStatusAll(convID, userID, 3); err != nil {
		return errors.ErrInternal, err
	}
	if err := s.convRepo.ResetUnread(convID, isBuyer); err != nil {
		return errors.ErrInternal, err
	}

	return errors.Success, nil
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
