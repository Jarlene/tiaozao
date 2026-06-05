package chat

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"flea-market/internal/config"
	"flea-market/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许跨域
	},
}

type WSHandler struct {
	hub         *Hub
	chatService *service.ChatService
	jwtCfg      *config.JWTConfig
}

func NewWSHandler(hub *Hub, chatService *service.ChatService, jwtCfg *config.JWTConfig) *WSHandler {
	return &WSHandler{hub: hub, chatService: chatService, jwtCfg: jwtCfg}
}

// HandleWebSocket 处理 WebSocket 连接
func (h *WSHandler) HandleWebSocket(c *gin.Context) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "missing token"})
		return
	}

	convIDStr := c.Query("conversation_id")
	if convIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "missing conversation_id"})
		return
	}

	convID, err := strconv.ParseUint(convIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid conversation_id"})
		return
	}

	// 解析 JWT
	claims := &struct {
		UserID uint `json:"user_id"`
		jwt.RegisteredClaims
	}{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(h.jwtCfg.Secret), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "invalid token"})
		return
	}

	// 升级为 WebSocket 连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		Conn:           conn,
		UserID:         claims.UserID,
		ConversationID: uint(convID),
		Send:           make(chan []byte, 256),
		Hub:            h.hub,
	}

	h.hub.register <- client

	go client.WritePump()
	go client.ReadPump(h.handleMessage)
}

// handleMessage 处理收到的 WebSocket 消息
func (h *WSHandler) handleMessage(client *Client, wsMsg WSMessage) {
	switch wsMsg.Type {
	case MsgSend:
		h.handleSendMessage(client, wsMsg)
	case MsgTyping:
		h.handleTyping(client, wsMsg)
	case MsgRead:
		h.handleReadReceipt(client, wsMsg)
	default:
		errMsg := WSMessage{Type: MsgError, Content: "unknown message type"}
		data, _ := json.Marshal(errMsg)
		client.Send <- data
	}
}

func (h *WSHandler) handleSendMessage(client *Client, wsMsg WSMessage) {
	msg, code, err := h.chatService.SendMessage(&service.SendMessageReq{
		ConversationID: wsMsg.ConversationID,
		SenderID:       client.UserID,
		Content:        wsMsg.Content,
		MsgType:        wsMsg.MsgType,
	})
	if err != nil || code != 0 {
		errMsg := WSMessage{Type: MsgError, Content: "failed to send message"}
		data, _ := json.Marshal(errMsg)
		client.Send <- data
		return
	}

	// 发送确认给发送者
	ack := WSMessage{
		Type:      MsgReceived,
		MessageID: msg.ID,
		Timestamp: time.Now().UnixMilli(),
	}
	ackData, _ := json.Marshal(ack)
	client.Send <- ackData

	// 广播给会话中其他人
	now := time.Now().UnixMilli()
	newMsg := WSMessage{
		Type:           MsgNewMessage,
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		Content:        msg.Content,
		MessageID:      msg.ID,
		MsgType:        msg.MsgType,
		Timestamp:      now,
	}
	broadcastData, _ := json.Marshal(newMsg)
	h.hub.BroadcastToConversation(wsMsg.ConversationID, client.UserID, broadcastData)
}

func (h *WSHandler) handleTyping(client *Client, wsMsg WSMessage) {
	typingMsg := WSMessage{
		Type:           MsgTyping,
		ConversationID: wsMsg.ConversationID,
		UserID:         client.UserID,
		Timestamp:      wsMsg.Timestamp,
	}
	data, _ := json.Marshal(typingMsg)
	h.hub.BroadcastToConversation(wsMsg.ConversationID, client.UserID, data)
}

func (h *WSHandler) handleReadReceipt(client *Client, wsMsg WSMessage) {
	readMsg := WSMessage{
		Type:           MsgRead,
		ConversationID: wsMsg.ConversationID,
		UserID:         client.UserID,
		MessageID:      wsMsg.MessageID,
	}
	data, _ := json.Marshal(readMsg)
	h.hub.BroadcastToConversation(wsMsg.ConversationID, client.UserID, data)
}
