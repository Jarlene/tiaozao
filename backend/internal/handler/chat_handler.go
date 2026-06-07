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

// GetMessages 获取消息历史（游标分页）
func (h *ChatHandler) GetMessages(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)

	convID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	cursorID, _ := strconv.ParseUint(c.DefaultQuery("cursor", "0"), 10, 32)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	msgs, hasMore, code, err := h.chatService.GetMessages(uint(convID), uint(userID), uint(cursorID), limit)
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
		"messages": msgs,
		"has_more": hasMore,
	})
}

// MarkRead 标记会话已读
func (h *ChatHandler) MarkRead(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.GetString("user_id"), 10, 32)

	convID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		Error(c, 400, errors.ErrBadRequest)
		return
	}

	code, err := h.chatService.MarkConversationRead(uint(convID), uint(userID))
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

	Success(c, gin.H{"status": "ok"})
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
