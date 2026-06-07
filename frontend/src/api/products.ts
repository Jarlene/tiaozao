import client from './client'
import type { ApiResponse } from './client'

export interface ProductListItem {
  id: number
  title: string
  price: number
  status: number
  user_id: number
  created_at: string
  images: ProductImageItem[]
  user?: UserSimpleProfile
}

export interface ProductDetail {
  id: number
  title: string
  description: string
  price: number
  status: number
  user_id: number
  category_id: number | null
  created_at: string
  updated_at: string
  images: ProductImageItem[]
  user?: UserSimpleProfile
}

export interface ProductImageItem {
  id: number
  url: string
  sort_order: number
}

export interface UserSimpleProfile {
  id: number
  nickname: string
  avatar_url: string
  created_at: string
}

export interface PaginatedResult<T> {
  items: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface CategoryTreeItem {
  id: number
  name: string
  parent_id: number | null
  children?: CategoryTreeItem[]
}

export interface CreateProductReq {
  title: string
  description: string
  price: number
  category_id?: number | null
  image_ids?: number[]
}

export interface UpdateProductReq {
  title: string
  description: string
  price: number
  category_id?: number | null
}

export const productAPI = {
  list(page = 1, size = 20) {
    return client
      .get<ApiResponse<PaginatedResult<ProductListItem>>>('/products', { params: { page, size } })
      .then((r) => r.data.data)
  },

  search(params: {
    keyword?: string
    category_id?: number
    price_min?: number
    price_max?: number
    page?: number
    size?: number
  }) {
    return client
      .get<ApiResponse<PaginatedResult<ProductListItem>>>('/products/search', { params })
      .then((r) => r.data.data)
  },

  getById(id: number) {
    return client.get<ApiResponse<ProductDetail>>(`/products/${id}`).then((r) => r.data.data)
  },

  create(data: CreateProductReq) {
    return client.post<ApiResponse<ProductDetail>>('/products', data).then((r) => r.data.data)
  },

  update(id: number, data: UpdateProductReq) {
    return client.put<ApiResponse<ProductDetail>>(`/products/${id}`, data).then((r) => r.data.data)
  },

  delete(id: number) {
    return client.delete<ApiResponse<null>>(`/products/${id}`).then((r) => r.data)
  },

  listMine(params: { status?: number; page?: number; size?: number }) {
    return client
      .get<ApiResponse<PaginatedResult<ProductListItem>>>('/products/mine', { params })
      .then((r) => r.data.data)
  },

  getCounts() {
    return client
      .get<ApiResponse<Record<string, number>>>('/products/mine/counts')
      .then((r) => r.data.data)
  },

  updateStatus(id: number, status: number) {
    return client
      .put<ApiResponse<null>>(`/products/${id}/status`, { status })
      .then((r) => r.data)
  },

  uploadImage(file: File) {
    const formData = new FormData()
    formData.append('file', file)
    return client
      .post<ApiResponse<ProductImageItem>>('/images/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      .then((r) => r.data.data)
  },

  getImage(id: number) {
    return client.get<ApiResponse<ProductImageItem>>(`/images/${id}`).then((r) => r.data.data)
  },

  deleteImage(id: number) {
    return client.delete<ApiResponse<null>>(`/images/${id}`).then((r) => r.data)
  },
}

export const categoryAPI = {
  getTree() {
    return client.get<ApiResponse<CategoryTreeItem[]>>('/categories').then((r) => r.data.data)
  },

  getFlat() {
    return client.get<ApiResponse<CategoryTreeItem[]>>('/categories/flat').then((r) => r.data.data)
  },

  create(data: { name: string; parent_id?: number | null }) {
    return client.post<ApiResponse<CategoryTreeItem>>('/categories', data).then((r) => r.data.data)
  },

  update(id: number, data: { name: string }) {
    return client.put<ApiResponse<CategoryTreeItem>>(`/categories/${id}`, data).then((r) => r.data.data)
  },

  delete(id: number) {
    return client.delete<ApiResponse<null>>(`/categories/${id}`).then((r) => r.data)
  },
}
