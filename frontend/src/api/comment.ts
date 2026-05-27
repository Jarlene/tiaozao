import axios from 'axios'
import type { CommentListResult, CreateCommentRequest, Comment } from '../types/comment'

const http = axios.create({
  baseURL: '/api',
})

// Inject X-User-Id header from localStorage for auth
http.interceptors.request.use((config) => {
  const userId = localStorage.getItem('user_id')
  if (userId) {
    config.headers['X-User-Id'] = userId
  }
  return config
})

export async function fetchComments(
  productId: string,
  page: number = 1,
  pageSize: number = 10,
): Promise<CommentListResult> {
  const res = await http.get(`/products/${productId}/comments`, {
    params: { page, page_size: pageSize },
  })
  return res.data
}

export async function createComment(
  productId: string,
  data: CreateCommentRequest,
): Promise<Comment> {
  const res = await http.post(`/products/${productId}/comments`, data)
  return res.data
}

export async function replyToComment(
  productId: string,
  commentId: number,
  data: CreateCommentRequest,
): Promise<Comment> {
  const res = await http.post(
    `/products/${productId}/comments/${commentId}/reply`,
    data,
  )
  return res.data
}
