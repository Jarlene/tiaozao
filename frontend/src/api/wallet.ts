import client from './client'
import type { ApiResponse } from './client'

export interface WalletInfo {
  balance: number
  frozen_balance: number
  total_balance: number
}

export interface TransactionItem {
  id: number
  type: string
  type_name: string
  amount: number
  order_id: number
  description?: string
  created_at: string
}

export interface PaginatedResult<T> {
  items: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export const walletAPI = {
  /**
   * 获取钱包信息
   */
  getWallet() {
    return client.get<ApiResponse<WalletInfo>>('/wallet').then((r) => r.data.data)
  },

  /**
   * 充值
   */
  topUp(amount: number) {
    return client.post<ApiResponse<WalletInfo>>('/wallet/topup', { amount }).then((r) => r.data.data)
  },

  /**
   * 获取交易流水
   */
  listTransactions(params: { page?: number; size?: number }) {
    return client
      .get<ApiResponse<PaginatedResult<TransactionItem>>>('/wallet/transactions', { params })
      .then((r) => r.data.data)
  },
}
