package chat

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// MessageType WebSocket 消息类型
const (
	MsgSend       = "send"        // 发送消息
	MsgReceived   = "received"    // 服务端确认收到
	MsgNewMessage = "new_message" // 新消息推送
	MsgTyping     = "typing"      // 输入状态
	MsgRead       = "read"        // 已读回执
	MsgOnline     = "online"      // 在线状态
	MsgError      = "error"       // 错误
)

// WSMessage WebSocket 消息结构
type WSMessage struct {
	Type           string `json:"type"`
	ConversationID uint   `json:"conversation_id,omitempty"`
	SenderID       uint   `json:"sender_id,omitempty"`
	Content        string `json:"content,omitempty"`
	MsgType        int    `json:"msg_type,omitempty"`
	MessageID      uint   `json:"message_id,omitempty"`
	UserID         uint   `json:"user_id,omitempty"`
	IsOnline       bool   `json:"is_online,omitempty"`
	Timestamp      int64  `json:"timestamp,omitempty"`
}

// Client 单个 WebSocket 连接
type Client struct {
	Conn           *websocket.Conn
	UserID         uint
	ConversationID uint
	Send           chan []byte
	Hub            *Hub
	mu             sync.Mutex
}

// Hub 连接管理器
type Hub struct {
	// conversations maps conversationID -> map[userID]*Client
	conversations map[uint]map[uint]*Client
	// userConns maps userID -> set of conversationIDs (for broadcasting to user across convos)
	userConns map[uint]map[uint]bool
	register  chan *Client
	unregister chan *Client
	broadcast chan *BroadcastMsg
	mu        sync.RWMutex
}

// BroadcastMsg 广播消息
type BroadcastMsg struct {
	ConversationID uint
	Message        []byte
	SkipUserID     uint // 0 means send to all
}

var (
	// MaxMessageSize 最大消息大小
	MaxMessageSize = int64(4096)
	// WriteWait 写超时
	WriteWait = 10 * time.Second
	// PongWait pong 超时
	PongWait = 60 * time.Second
	// PingPeriod ping 周期
	PingPeriod = (PongWait * 9) / 10
)

// NewHub 创建 Hub
func NewHub() *Hub {
	return &Hub{
		conversations: make(map[uint]map[uint]*Client),
		userConns:     make(map[uint]map[uint]bool),
		register:      make(chan *Client, 256),
		unregister:    make(chan *Client, 256),
		broadcast:     make(chan *BroadcastMsg, 256),
	}
}

// Run Hub 主循环
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			// 添加到会话
			if _, ok := h.conversations[client.ConversationID]; !ok {
				h.conversations[client.ConversationID] = make(map[uint]*Client)
			}
			// 关闭同一用户在同一个会话中的旧连接
			if old, exists := h.conversations[client.ConversationID][client.UserID]; exists {
				close(old.Send)
				delete(h.conversations[client.ConversationID], client.UserID)
			}
			h.conversations[client.ConversationID][client.UserID] = client

			// 添加到用户映射
			if _, ok := h.userConns[client.UserID]; !ok {
				h.userConns[client.UserID] = make(map[uint]bool)
			}
			h.userConns[client.UserID][client.ConversationID] = true

			h.mu.Unlock()

			// 广播在线状态
			h.broadcastOnlineStatus(client.ConversationID, client.UserID, true)

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.conversations[client.ConversationID]; ok {
				if _, exists := clients[client.UserID]; exists {
					delete(clients, client.UserID)
					if len(clients) == 0 {
						delete(h.conversations, client.ConversationID)
					}
				}
			}
			if convs, ok := h.userConns[client.UserID]; ok {
				delete(convs, client.ConversationID)
				if len(convs) == 0 {
					delete(h.userConns, client.UserID)
				}
			}
			h.mu.Unlock()

			// 广播离线状态
			h.broadcastOnlineStatus(client.ConversationID, client.UserID, false)

		case msg := <-h.broadcast:
			h.mu.RLock()
			clients, ok := h.conversations[msg.ConversationID]
			if !ok {
				h.mu.RUnlock()
				continue
			}
			for _, client := range clients {
				if client.UserID == msg.SkipUserID {
					continue
				}
				select {
				case client.Send <- msg.Message:
				default:
					// 发送缓冲区满，跳过
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastToConversation 向会话广播消息（可选跳过发送者）
func (h *Hub) BroadcastToConversation(convID, skipUserID uint, msg []byte) {
	h.broadcast <- &BroadcastMsg{
		ConversationID: convID,
		Message:        msg,
		SkipUserID:     skipUserID,
	}
}

// GetOnlineUsers 获取会话中在线用户
func (h *Hub) GetOnlineUsers(conversationID uint) []uint {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var users []uint
	if clients, ok := h.conversations[conversationID]; ok {
		for userID := range clients {
			users = append(users, userID)
		}
	}
	return users
}

// IsUserOnline 检查用户是否在线
func (h *Hub) IsUserOnline(userID uint) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	convs, ok := h.userConns[userID]
	return ok && len(convs) > 0
}

// broadcastOnlineStatus 广播在线状态变化
func (h *Hub) broadcastOnlineStatus(conversationID, userID uint, online bool) {
	msg := WSMessage{
		Type:     MsgOnline,
		UserID:   userID,
		IsOnline: online,
		Timestamp: time.Now().UnixMilli(),
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	h.BroadcastToConversation(conversationID, userID, data)
}

// ReadPump 读取消息循环
func (c *Client) ReadPump(handleMessage func(client *Client, msg WSMessage)) {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(MaxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(PongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(PongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			errMsg := WSMessage{Type: MsgError, Content: "invalid message format"}
			data, _ := json.Marshal(errMsg)
			c.Send <- data
			continue
		}

		handleMessage(c, wsMsg)
	}
}

// WritePump 写入消息循环
func (c *Client) WritePump() {
	ticker := time.NewTicker(PingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
