import client from './client'
import type { ApiResponse } from './client'

export interface ProductImageItem {
  id: number
  url: string
  sort_order: number
}

export interface OrderProductItem {
  id: number
  title: string
  price: number
  images?: ProductImageItem[]
}

export interface OrderListItem {
  id: number
  order_no: string
  product_id: number
  title: string
  price: number
  status: number
  status_name: string
  created_at: string
  product?: OrderProductItem
}

export interface OrderDetail {
  id: number
  order_no: string
  product_id: number
  buyer_id: number
  seller_id: number
  title: string
  price: number
  status: number
  status_name: string
  shipping_address: string
  buyer_note: string
  refund_reason?: string
  tracking_number?: string
  logistics_company?: string
  created_at: string
  updated_at: string
  product?: OrderProductItem
}

export interface PaginatedResult<T> {
  items: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

// 订单状态常量
export const OrderStatus = {
  PENDING_PAYMENT: 1,   // 待付款
  CANCELLED: 2,          // 已取消
  PAID: 3,               // 已付款
  PENDING_SHIPMENT: 4,   // 待发货
  SHIPPED: 5,            // 已发货
  RECEIVED: 6,           // 已收货
  COMPLETED: 7,          // 已完成
  REFUNDING: 8,          // 退款中
  DISPUTE: 9,            // 纠纷
} as const

export const orderAPI = {
  /**
   * 获取买家的订单列表（我的订单）
   */
  listMine(params: { status?: number; page?: number; size?: number }) {
    return client
      .get<ApiResponse<PaginatedResult<OrderListItem>>>('/orders/mine', { params })
      .then((r) => r.data.data)
  },

  /**
   * 获取卖家的订单列表（售出的订单）
   */
  listSold(params: { status?: number; page?: number; size?: number }) {
    return client
      .get<ApiResponse<PaginatedResult<OrderListItem>>>('/orders/sold', { params })
      .then((r) => r.data.data)
  },

  /**
   * 管理员获取纠纷订单列表
   */
  listDisputes(params: { page?: number; size?: number }) {
    return client
      .get<ApiResponse<PaginatedResult<OrderDetail>>>('/orders/disputes', { params })
      .then((r) => r.data.data)
  },

  /**
   * 获取订单详情
   */
  getById(id: number) {
    return client.get<ApiResponse<OrderDetail>>(`/orders/${id}`).then((r) => r.data.data)
  },

  /**
   * 取消订单
   */
  cancel(id: number) {
    return client.post<ApiResponse<null>>(`/orders/${id}/cancel`).then((r) => r.data)
  },

  /**
   * 付款
   */
  pay(id: number) {
    return client.post<ApiResponse<null>>(`/orders/${id}/pay`).then((r) => r.data)
  },

  /**
   * 发货
   */
  ship(id: number, data?: { tracking_number?: string; logistics_company?: string }) {
    return client.post<ApiResponse<null>>(`/orders/${id}/ship`, data).then((r) => r.data)
  },

  /**
   * 确认收货
   */
  confirmReceive(id: number) {
    return client.post<ApiResponse<null>>(`/orders/${id}/confirm`).then((r) => r.data)
  },

  /**
   * 申请退款
   */
  requestRefund(id: number, data?: { reason: string; note?: string }) {
    return client.post<ApiResponse<null>>(`/orders/${id}/refund`, data).then((r) => r.data)
  },

  /**
   * 同意退款
   */
  approveRefund(id: number) {
    return client.post<ApiResponse<null>>(`/orders/${id}/refund/approve`).then((r) => r.data)
  },

  /**
   * 拒绝退款（进入纠纷）
   */
  rejectRefund(id: number) {
    return client.post<ApiResponse<null>>(`/orders/${id}/refund/reject`).then((r) => r.data)
  },

  /**
   * 发起纠纷
   */
  raiseDispute(id: number) {
    return client.post<ApiResponse<null>>(`/orders/${id}/dispute`).then((r) => r.data)
  },

  /**
   * 管理员仲裁
   */
  arbitrate(id: number, decision: string) {
    return client.post<ApiResponse<null>>(`/orders/${id}/arbitrate`, { decision }).then((r) => r.data)
  },

  /**
   * 获取订单状态变更日志
   */
  getStatusLogs(id: number) {
    return client.get<ApiResponse<any[]>>(`/orders/${id}/logs`).then((r) => r.data.data)
  },
}
