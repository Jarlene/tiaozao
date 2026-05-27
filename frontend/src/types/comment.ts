export interface Comment {
  id: number
  product_id: number
  user_id: string
  content: string
  parent_id: number | null
  created_at: string
  updated_at: string
  user_nick?: string
  replies?: Comment[]
}

export interface CommentListResult {
  comments: Comment[]
  total: number
  page: number
  page_size: number
}

export interface CreateCommentRequest {
  content: string
}
