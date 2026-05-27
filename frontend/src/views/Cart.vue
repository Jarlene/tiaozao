<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useCartStore } from '@/stores/cart'
import { ElMessage, ElMessageBox } from 'element-plus'

const router = useRouter()
const cartStore = useCartStore()

const loading = ref(false)

onMounted(async () => {
  await loadCart()
})

/** 加载购物车 */
async function loadCart() {
  loading.value = true
  try {
    await cartStore.loadCart()
  } catch {
    // 错误已在 store 中处理
  } finally {
    loading.value = false
  }
}

/** 更新商品数量 */
async function handleQuantityChange(itemId: string, quantity: number) {
  if (quantity < 1) return
  try {
    await cartStore.updateQuantity(itemId, quantity)
  } catch {
    ElMessage.error('更新数量失败')
  }
}

/** 删除商品 */
async function handleRemove(itemId: string) {
  try {
    await cartStore.removeItem(itemId)
    ElMessage.success('已移除该商品')
  } catch {
    ElMessage.error('移除失败')
  }
}

/** 清空购物车 */
async function handleClear() {
  await ElMessageBox.confirm('确定要清空购物车吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  })
  try {
    await cartStore.clearCart()
    ElMessage.success('购物车已清空')
  } catch {
    ElMessage.error('清空失败')
  }
}

/** 去结算 */
function handleCheckout() {
  if (cartStore.isEmpty) {
    ElMessage.warning('购物车为空')
    return
  }
  router.push('/orders')
}

/** 格式化价格 */
function formatPrice(price: number) {
  return `¥${price.toFixed(2)}`
}
</script>

<template>
  <div class="cart-page">
    <div class="page-header">
      <h2 class="page-title">购物车</h2>
      <el-button
        v-if="!cartStore.isEmpty"
        text
        type="danger"
        @click="handleClear"
      >
        <el-icon><Delete /></el-icon>
        清空购物车
      </el-button>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-state">
      <el-skeleton :rows="4" animated />
    </div>

    <!-- 空购物车 -->
    <div v-else-if="cartStore.isEmpty" class="empty-cart">
      <el-empty description="购物车空空如也">
        <el-button type="primary" @click="router.push('/products')">去逛逛</el-button>
      </el-empty>
    </div>

    <!-- 购物车列表 -->
    <template v-else>
      <div class="cart-list">
        <div
          v-for="item in cartStore.items"
          :key="item.id"
          class="cart-item"
        >
          <!-- 商品图片 -->
          <div class="item-image" @click="router.push(`/products/${item.product.id}`)">
            <img
              :src="item.product.images[0] || '/placeholder.svg'"
              :alt="item.product.title"
            />
          </div>

          <!-- 商品信息 -->
          <div class="item-info" @click="router.push(`/products/${item.product.id}`)">
            <h4 class="item-title">{{ item.product.title }}</h4>
            <p class="item-desc">{{ item.product.description }}</p>
          </div>

          <!-- 单价 -->
          <div class="item-price">
            <span class="unit-price">{{ formatPrice(item.product.price) }}</span>
          </div>

          <!-- 数量 -->
          <div class="item-quantity">
            <el-input-number
              :model-value="item.quantity"
              :min="1"
              :max="99"
              size="small"
              @change="(val: number) => handleQuantityChange(item.id, val)"
            />
          </div>

          <!-- 小计 -->
          <div class="item-subtotal">
            <span class="subtotal-price">
              {{ formatPrice(item.product.price * item.quantity) }}
            </span>
          </div>

          <!-- 操作 -->
          <div class="item-actions">
            <el-button
              text
              type="danger"
              @click="handleRemove(item.id)"
            >
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
        </div>
      </div>

      <!-- 结算栏 -->
      <div class="checkout-bar">
        <div class="total-section">
          <span class="total-label">合计：</span>
          <span class="total-price">{{ formatPrice(cartStore.totalPrice) }}</span>
        </div>
        <el-button type="danger" size="large" @click="handleCheckout">
          去结算
        </el-button>
      </div>
    </template>
  </div>
</template>

<style scoped>
.cart-page {
  padding-bottom: 40px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-title {
  font-size: 22px;
  font-weight: 600;
  margin: 0;
  color: #333;
}

.loading-state {
  background: #fff;
  border-radius: 8px;
  padding: 40px;
}

.empty-cart {
  padding: 60px;
  background: #fff;
  border-radius: 12px;
}

.cart-list {
  background: #fff;
  border-radius: 12px;
  overflow: hidden;
}

.cart-item {
  display: flex;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #f0f0f0;
  gap: 16px;
}

.cart-item:last-child {
  border-bottom: none;
}

.item-image {
  width: 100px;
  height: 100px;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  flex-shrink: 0;
  background: #f5f5f5;
}

.item-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.item-info {
  flex: 1;
  min-width: 0;
  cursor: pointer;
}

.item-title {
  font-size: 15px;
  font-weight: 500;
  margin: 0 0 6px 0;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-desc {
  font-size: 13px;
  color: #999;
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-price {
  text-align: center;
  min-width: 80px;
}

.unit-price {
  font-size: 16px;
  color: #f56c6c;
  font-weight: 500;
}

.item-quantity {
  min-width: 120px;
  display: flex;
  justify-content: center;
}

.item-subtotal {
  text-align: center;
  min-width: 80px;
}

.subtotal-price {
  font-size: 16px;
  color: #f56c6c;
  font-weight: 600;
}

.item-actions {
  min-width: 40px;
  display: flex;
  justify-content: center;
}

.checkout-bar {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 24px;
  margin-top: 20px;
  padding: 20px 24px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.total-section {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.total-label {
  font-size: 16px;
  color: #666;
}

.total-price {
  font-size: 24px;
  font-weight: bold;
  color: #f56c6c;
}
</style>
