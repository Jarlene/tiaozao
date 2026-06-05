package handler

import (
	"strconv"

	"flea-market/internal/service"
	"flea-market/pkg/errors"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatService *service.ChatService
}

func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

// CreateConversation 创建会话
func (h *ChatHandler) CreateConversation(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)

	var req service.CreateConversationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	req.BuyerID = uint(userID)

	conv, code, err := h.chatService.CreateConversation(&req)
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if code != errors.Success {
		Error(c, 400, code)
		return
	}

	Success(c, conv)
}

// ListConversations 获取会话列表
func (h *ChatHandler) ListConversations(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)

	convs, code, err := h.chatService.ListConversations(uint(userID))
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if code != errors.Success {
		Error(c, 400, code)
		return
	}

	Success(c, convs)
}

// GetMessages 获取消息历史
func (h *ChatHandler) GetMessages(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)

	convID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	msgs, total, code, err := h.chatService.GetMessages(uint(convID), uint(userID), page, pageSize)
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if code != errors.Success {
		httpStatus := 400
		if code == errors.ErrForbidden {
			httpStatus = 403
		}
		Error(c, httpStatus, code)
		return
	}

	Success(c, gin.H{
		"messages":  msgs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// SendMessage 发送消息
func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)

	var req service.SendMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	req.SenderID = uint(userID)

	msg, code, err := h.chatService.SendMessage(&req)
	if err != nil {
		Error(c, 500, errors.ErrInternal)
		return
	}
	if code != errors.Success {
		httpStatus := 400
		if code == errors.ErrForbidden {
			httpStatus = 403
		}
		Error(c, httpStatus, code)
		return
	}

	Success(c, msg)
}
