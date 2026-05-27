import axios, { type AxiosInstance, type AxiosError, type InternalAxiosRequestConfig } from 'axios'

// ============================================================
// Type Definitions
// ============================================================

/** 用户信息 */
export interface UserProfile {
  id: string
  username: string
  email: string
  avatar?: string
  phone?: string
  bio?: string
  createdAt: string
}

/** 登录凭证 */
export interface LoginCredentials {
  username: string
  password: string
}

/** 注册信息 */
export interface RegisterData {
  username: string
  email: string
  password: string
  confirmPassword: string
}

/** 认证响应 */
export interface AuthResponse {
  token: string
  user: UserProfile
}

/** 商品信息 */
export interface Product {
  id: string
  title: string
  description: string
  price: number
  originalPrice?: number
  images: string[]
  category: string
  condition: 'new' | 'like_new' | 'good' | 'fair' | 'poor'
  status: 'active' | 'sold' | 'inactive'
  seller: UserProfile
  createdAt: string
  updatedAt: string
}

/** 商品查询参数 */
export interface ProductQuery {
  page?: number
  pageSize?: number
  search?: string
  category?: string
  condition?: string
  minPrice?: number
  maxPrice?: number
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
}

/** 分页响应 */
export interface PaginatedResponse<T> {
  items: T[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}

/** 购物车项 */
export interface CartItem {
  id: string
  product: Product
  quantity: number
}

/** 订单 */
export interface Order {
  id: string
  orderNo: string
  items: OrderItem[]
  totalAmount: number
  status: OrderStatus
  shippingAddress?: Address
  createdAt: string
  updatedAt: string
}

/** 订单项 */
export interface OrderItem {
  id: string
  productId: string
  title: string
  image: string
  price: number
  quantity: number
}

/** 订单状态 */
export type OrderStatus = 'pending' | 'paid' | 'shipped' | 'completed' | 'cancelled'

/** 地址信息 */
export interface Address {
  name: string
  phone: string
  province: string
  city: string
  district: string
  detail: string
  zipCode?: string
}

/** API 错误响应 */
export interface ApiError {
  message: string
  code?: string
  details?: Record<string, string[]>
}

// ============================================================
// Axios Instance
// ============================================================

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api'

const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器：自动附加 JWT token
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem('tiaozao_token')
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error: AxiosError) => {
    return Promise.reject(error)
  }
)

// 响应拦截器：统一错误处理
apiClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError<ApiError>) => {
    if (error.response) {
      const { status, data } = error.response
      switch (status) {
        case 401:
          // token 过期，清除登录状态
          localStorage.removeItem('tiaozao_token')
          window.location.href = '/login'
          break
        case 403:
          console.error('权限不足:', data?.message)
          break
        case 404:
          console.error('资源不存在:', data?.message)
          break
        case 422:
          console.error('参数验证失败:', data?.details)
          break
        case 500:
          console.error('服务器错误:', data?.message)
          break
      }
    } else if (error.request) {
      console.error('网络错误，请检查连接')
    }
    return Promise.reject(error)
  }
)

// ============================================================
// API Functions
// ============================================================

/** 认证相关 API */
export const authApi = {
  /** 用户登录 */
  login: (credentials: LoginCredentials) =>
    apiClient.post<AuthResponse>('/auth/login', credentials).then((res) => res.data),

  /** 用户注册 */
  register: (data: RegisterData) =>
    apiClient.post<AuthResponse>('/auth/register', data).then((res) => res.data),
}

/** 用户相关 API */
export const userApi = {
  /** 获取当前用户信息 */
  getProfile: () =>
    apiClient.get<UserProfile>('/user/profile').then((res) => res.data),

  /** 更新用户信息 */
  updateProfile: (data: Partial<UserProfile>) =>
    apiClient.put<UserProfile>('/user/profile', data).then((res) => res.data),
}

/** 商品相关 API */
export const productApi = {
  /** 获取商品列表（支持分页和筛选） */
  getList: (query?: ProductQuery) =>
    apiClient.get<PaginatedResponse<Product>>('/products', { params: query }).then((res) => res.data),

  /** 获取商品详情 */
  getById: (id: string) =>
    apiClient.get<Product>(`/products/${id}`).then((res) => res.data),

  /** 创建商品 */
  create: (data: FormData) =>
    apiClient.post<Product>('/products', data, {
      headers: { 'Content-Type': 'multipart/form-data' },
    }).then((res) => res.data),

  /** 更新商品 */
  update: (id: string, data: FormData) =>
    apiClient.put<Product>(`/products/${id}`, data, {
      headers: { 'Content-Type': 'multipart/form-data' },
    }).then((res) => res.data),

  /** 删除商品 */
  delete: (id: string) =>
    apiClient.delete(`/products/${id}`),

  /** 获取我的商品列表 */
  getMyProducts: (query?: ProductQuery) =>
    apiClient.get<PaginatedResponse<Product>>('/user/products', { params: query }).then((res) => res.data),

  /** 获取推荐商品 */
  getFeatured: (limit = 8) =>
    apiClient.get<PaginatedResponse<Product>>('/products/featured', { params: { limit } }).then((res) => res.data),
}

/** 购物车相关 API */
export const cartApi = {
  /** 获取购物车列表 */
  getCart: () =>
    apiClient.get<CartItem[]>('/cart').then((res) => res.data),

  /** 添加商品到购物车 */
  addItem: (productId: string, quantity = 1) =>
    apiClient.post<CartItem>('/cart/items', { productId, quantity }).then((res) => res.data),

  /** 更新购物车商品数量 */
  updateItem: (itemId: string, quantity: number) =>
    apiClient.put<CartItem>(`/cart/items/${itemId}`, { quantity }).then((res) => res.data),

  /** 删除购物车商品 */
  removeItem: (itemId: string) =>
    apiClient.delete(`/cart/items/${itemId}`),

  /** 清空购物车 */
  clearCart: () =>
    apiClient.delete('/cart'),
}

/** 订单相关 API */
export const orderApi = {
  /** 创建订单（从购物车结算） */
  create: (address: Address) =>
    apiClient.post<Order>('/orders', { address }).then((res) => res.data),

  /** 获取订单列表 */
  getList: (params?: { page?: number; pageSize?: number; status?: OrderStatus }) =>
    apiClient.get<PaginatedResponse<Order>>('/orders', { params }).then((res) => res.data),

  /** 获取订单详情 */
  getById: (id: string) =>
    apiClient.get<Order>(`/orders/${id}`).then((res) => res.data),

  /** 取消订单 */
  cancel: (id: string) =>
    apiClient.put<Order>(`/orders/${id}/cancel`).then((res) => res.data),
}

/** 分类信息 */
export interface Category {
  id: string
  name: string
  icon?: string
  productCount?: number
}

/** 分类相关 API */
export const categoryApi = {
  /** 获取所有分类 */
  getList: () =>
    apiClient.get<Category[]>('/categories').then((res) => res.data),
}

export default apiClient
