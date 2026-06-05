import client from './client'
import type { ApiResponse } from './client'
import type { PaginatedResult } from './products'

export interface ReviewItem {
  id: number
  product_id: number
  user_id: number
  rating: number
  content: string
  reply_content?: string
  replied_at?: string
  created_at: string
  user?: {
    id: number
    nickname: string
    avatar_url: string
    created_at: string
  }
}

export interface RatingStats {
  average: number
  total: number
  dist_1: number
  dist_2: number
  dist_3: number
  dist_4: number
  dist_5: number
}

export interface CreateReviewReq {
  rating: number
  content: string
}

export interface ReplyReviewReq {
  content: string
}

export interface CheckCanReviewResult {
  can_review: boolean
  code: number
}

export const reviewAPI = {
  create(productId: number, data: CreateReviewReq) {
    return client
      .post<ApiResponse<ReviewItem>>(`/products/${productId}/reviews`, data)
      .then((r) => r.data.data)
  },

  list(productId: number, page = 1, size = 20) {
    return client
      .get<ApiResponse<PaginatedResult<ReviewItem>>>(`/products/${productId}/reviews`, {
        params: { page, size },
      })
      .then((r) => r.data.data)
  },

  getStats(productId: number) {
    return client
      .get<ApiResponse<RatingStats>>(`/products/${productId}/reviews/stats`)
      .then((r) => r.data.data)
  },

  reply(reviewId: number, data: ReplyReviewReq) {
    return client
      .post<ApiResponse<ReviewItem>>(`/reviews/${reviewId}/reply`, data)
      .then((r) => r.data.data)
  },

  checkCanReview(productId: number) {
    return client
      .get<ApiResponse<CheckCanReviewResult>>(`/products/${productId}/reviews/check`)
      .then((r) => r.data.data)
  },
}
