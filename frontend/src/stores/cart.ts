import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { CartItem } from '@/api'
import { cartApi } from '@/api'

/**
 * 购物车状态管理
 * - 支持从 API 加载购物车数据
 * - 提供添加、删除、修改数量操作
 * - 计算总价和商品数量
 */
export const useCartStore = defineStore('cart', () => {
  const items = ref<CartItem[]>([])

  // 购物车商品总数
  const totalCount = computed(() =>
    items.value.reduce((sum, item) => sum + item.quantity, 0)
  )

  // 购物车商品总价
  const totalPrice = computed(() =>
    items.value.reduce((sum, item) => sum + item.product.price * item.quantity, 0)
  )

  // 购物车是否为空
  const isEmpty = computed(() => items.value.length === 0)

  /** 从 API 加载购物车数据 */
  async function loadCart() {
    try {
      items.value = await cartApi.getCart()
    } catch {
      // 未登录或网络错误，清空本地购物车
      items.value = []
    }
  }

  /** 添加商品到购物车 */
  async function addItem(productId: string, quantity = 1) {
    const cartItem = await cartApi.addItem(productId, quantity)
    // 检查是否已存在该商品，存在则替换，否则追加
    const index = items.value.findIndex((item) => item.product.id === productId)
    if (index >= 0) {
      items.value[index] = cartItem
    } else {
      items.value.push(cartItem)
    }
  }

  /** 更新商品数量 */
  async function updateQuantity(itemId: string, quantity: number) {
    const updated = await cartApi.updateItem(itemId, quantity)
    const index = items.value.findIndex((item) => item.id === itemId)
    if (index >= 0) {
      items.value[index] = updated
    }
  }

  /** 删除购物车商品 */
  async function removeItem(itemId: string) {
    await cartApi.removeItem(itemId)
    items.value = items.value.filter((item) => item.id !== itemId)
  }

  /** 清空购物车 */
  async function clearCart() {
    await cartApi.clearCart()
    items.value = []
  }

  return {
    items,
    totalCount,
    totalPrice,
    isEmpty,
    loadCart,
    addItem,
    updateQuantity,
    removeItem,
    clearCart,
  }
})
