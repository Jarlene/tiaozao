import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { chatAPI } from '@/api/chat'
import type { Conversation, Message } from '@/api/chat'

export const useChatStore = defineStore('chat', () => {
  const conversations = ref<Conversation[]>([])
  const currentConversation = ref<Conversation | null>(null)
  const messages = ref<Message[]>([])
  const connected = ref(false)
  const ws = ref<WebSocket | null>(null)
  const typingUsers = ref<Set<number>>(new Set())
  const onlineUsers = ref<Set<number>>(new Set())
  const loading = ref(false)

  const totalUnread = computed(() => {
    return conversations.value.reduce((sum, c) => {
      return sum + c.unread_buyer + c.unread_seller
    }, 0)
  })

  async function fetchConversations() {
    loading.value = true
    try {
      conversations.value = await chatAPI.listConversations()
    } finally {
      loading.value = false
    }
  }

  async function fetchMessages(conversationId: number, page = 1) {
    const result = await chatAPI.getMessages(conversationId, page)
    if (page === 1) {
      messages.value = result.messages.reverse()
    } else {
      messages.value = [...result.messages.reverse(), ...messages.value]
    }
    return result
  }

  async function createConversation(productId: number, sellerId: number) {
    const conv = await chatAPI.createConversation({ product_id: productId, seller_id: sellerId })
    const exists = conversations.value.find(c => c.id === conv.id)
    if (!exists) {
      conversations.value.unshift(conv)
    }
    return conv
  }

  function connectWebSocket(conversationId: number) {
    if (ws.value) {
      ws.value.close()
    }

    const url = chatAPI.getWebSocketUrl(conversationId)
    const socket = new WebSocket(url)
    ws.value = socket

    socket.onopen = () => {
      connected.value = true
    }

    socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        handleWSMessage(data)
      } catch {
        // ignore invalid messages
      }
    }

    socket.onclose = () => {
      connected.value = false
      typingUsers.value = new Set()
      ws.value = null
    }

    socket.onerror = () => {
      connected.value = false
    }
  }

  function handleWSMessage(data: any) {
    switch (data.type) {
      case 'new_message':
        messages.value.push({
          id: data.message_id,
          conversation_id: data.conversation_id,
          sender_id: data.sender_id,
          content: data.content,
          msg_type: data.msg_type || 1,
          status: 2,
          created_at: new Date(data.timestamp).toISOString(),
        })
        break
      case 'typing':
        if (data.user_id) {
          typingUsers.value.add(data.user_id)
          setTimeout(() => {
            typingUsers.value.delete(data.user_id)
          }, 3000)
        }
        break
      case 'online':
        if (data.is_online) {
          onlineUsers.value.add(data.user_id)
        } else {
          onlineUsers.value.delete(data.user_id)
        }
        break
      case 'read':
        messages.value.forEach(msg => {
          if (msg.sender_id !== data.user_id && msg.status < 3) {
            msg.status = 3
          }
        })
        break
    }
  }

  function sendMessage(content: string, msgType = 1) {
    if (!ws.value || !connected.value) return

    ws.value.send(JSON.stringify({
      type: 'send',
      conversation_id: currentConversation.value?.id,
      content,
      msg_type: msgType,
    }))
  }

  function sendTyping() {
    if (!ws.value || !connected.value) return
    ws.value.send(JSON.stringify({
      type: 'typing',
      conversation_id: currentConversation.value?.id,
      timestamp: Date.now(),
    }))
  }

  function sendReadReceipt(conversationId: number) {
    if (!ws.value || !connected.value) return
    ws.value.send(JSON.stringify({
      type: 'read',
      conversation_id: conversationId,
    }))
  }

  function disconnectWebSocket() {
    if (ws.value) {
      ws.value.close()
      ws.value = null
    }
    connected.value = false
  }

  return {
    conversations,
    currentConversation,
    messages,
    connected,
    typingUsers,
    onlineUsers,
    loading,
    totalUnread,
    fetchConversations,
    fetchMessages,
    createConversation,
    connectWebSocket,
    sendMessage,
    sendTyping,
    sendReadReceipt,
    disconnectWebSocket,
  }
})
