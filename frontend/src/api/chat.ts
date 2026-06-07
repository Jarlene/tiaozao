import client from './client'
import type { ApiResponse } from './client'

export interface Conversation {
  id: number
  product_id: number
  buyer_id: number
  seller_id: number
  last_msg: string
  last_msg_id: number
  unread_buyer: number
  unread_seller: number
  created_at: string
  updated_at: string
}

export interface Message {
  id: number
  conversation_id: number
  sender_id: number
  content: string
  msg_type: number
  status: number
  created_at: string
}

export interface CreateConversationReq {
  product_id: number
  seller_id: number
}

export interface SendMessageReq {
  conversation_id: number
  content: string
  msg_type?: number
}

export const chatAPI = {
  async createConversation(data: CreateConversationReq) {
    const res = await client.post<ApiResponse<Conversation>>('/conversations', data)
    return res.data.data
  },

  async listConversations() {
    const res = await client.get<ApiResponse<Conversation[]>>('/conversations')
    return res.data.data
  },

  async getMessages(conversationId: number, cursor = 0, limit = 20) {
    const res = await client.get<ApiResponse<{ messages: Message[]; has_more: boolean }>>(
      `/conversations/${conversationId}/messages`,
      { params: { cursor, limit } }
    )
    return res.data.data
  },

  async markConversationRead(conversationId: number) {
    const res = await client.post<ApiResponse<{ status: string }>>(`/conversations/${conversationId}/read`)
    return res.data.data
  },

  async sendMessage(data: SendMessageReq) {
    const res = await client.post<ApiResponse<Message>>('/messages', data)
    return res.data.data
  },

  getWebSocketUrl(conversationId: number): string {
    const token = localStorage.getItem('access_token')
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    return `${protocol}//${host}/api/v1/ws?token=${token}&conversation_id=${conversationId}`
  }
}
