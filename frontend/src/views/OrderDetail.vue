<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { Order } from '@/api'
import { orderApi } from '@/api'
import { ElMessage, ElMessageBox } from 'element-plus'

const route = useRoute()
const router = useRouter()

const order = ref<Order | null>(null)
const loading = ref(false)
const error = ref('')

/** 订单 ID */
const orderId = computed(() => route.params.id as string)

/** 订单状态类型映射 */
const statusTypeMap: Record<string, string> = {
  pending: 'warning',
  paid: 'primary',
  shipped: 'info',
  completed: 'success',
  cancelled: 'danger',
}

/** 订单状态文本映射 */
const statusLabelMap: Record<string, string> = {
  pending: '待付款',
  paid: '待发货',
  shipped: '待收货',
  completed: '已完成',
  cancelled: '已取消',
}

/** 订单状态时间线步骤 */
const statusSteps = [
  { title: '提交订单', status: 'process' },
  { title: '支付', status: 'wait' },
  { title: '卖家发货', status: 'wait' },
  { title: '确认收货', status: 'wait' },
]

onMounted(async () => {
  await loadOrder()
})

/** 加载订单详情 */
async function loadOrder() {
  loading.value = true
  error.value = ''
  try {
    order.value = await orderApi.getById(orderId.value)
    // 根据订单状态更新步骤
    updateSteps(order.value.status)
  } catch (err: any) {
    error.value = err?.response?.data?.message || '加载订单详情失败'
  } finally {
    loading.value = false
  }
}

/** 根据订单状态更新时间线步骤 */
function updateSteps(status: string) {
  const stepMap: Record<string, number> = {
    cancelled: -1,
    pending: 0,
    paid: 1,
    shipped: 2,
    completed: 3,
  }
  const currentStep = stepMap[status] ?? 0

  statusSteps.forEach((step, index) => {
    if (status === 'cancelled') {
      step.status = 'danger'
    } else if (index < currentStep) {
      step.status = 'success'
    } else if (index === currentStep) {
      step.status = 'process'
    } else {
      step.status = 'wait'
    }
  })
}

/** 取消订单 */
async function handleCancel() {
  if (!order.value) return

  await ElMessageBox.confirm('确定要取消此订单吗？', '提示', {
    confirmButtonText: '确定取消',
    cancelButtonText: '暂不取消',
    type: 'warning',
  })

  try {
    await orderApi.cancel(order.value.id)
    ElMessage.success('订单已取消')
    await loadOrder()
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.message || '取消失败')
  }
}

/** 格式化价格 */
function formatPrice(price: number) {
  return `¥${price.toFixed(2)}`
}

/** 格式化时间 */
function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}
</script>

<template>
  <div class="order-detail-page">
    <!-- 面包屑 -->
    <div class="breadcrumb">
      <el-breadcrumb>
        <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
        <el-breadcrumb-item :to="{ path: '/orders' }">我的订单</el-breadcrumb-item>
        <el-breadcrumb-item>订单详情</el-breadcrumb-item>
      </el-breadcrumb>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-state">
      <el-skeleton :rows="6" animated />
    </div>

    <!-- 错误状态 -->
    <div v-else-if="error" class="error-state">
      <el-result icon="error" title="加载失败" :sub-title="error">
        <template #extra>
          <el-button type="primary" @click="loadOrder">重新加载</el-button>
          <el-button @click="router.push('/orders')">返回订单列表</el-button>
        </template>
      </el-result>
    </div>

    <!-- 订单详情 -->
    <template v-else-if="order">
      <!-- 订单状态头部 -->
      <div class="order-status-header">
        <el-result
          :icon="order.status === 'cancelled' ? 'warning' : 'success'"
          :title="statusLabelMap[order.status] || order.status"
          :sub-title="`订单号：${order.orderNo}`"
        >
          <template #extra>
            <el-button @click="router.push('/orders')">返回列表</el-button>
            <el-button
              v-if="order.status === 'pending'"
              type="danger"
              @click="handleCancel"
            >
              取消订单
            </el-button>
          </template>
        </el-result>
      </div>

      <!-- 订单进度 -->
      <el-card class="detail-card">
        <template #header>
          <span>订单进度</span>
        </template>
        <el-steps :active="statusSteps.findIndex(s => s.status === 'process' || s.status === 'danger') + 1" align-center>
          <el-step
            v-for="(step, index) in statusSteps"
            :key="index"
            :title="step.title"
            :status="step.status as any"
          />
        </el-steps>
      </el-card>

      <!-- 商品列表 -->
      <el-card class="detail-card">
        <template #header>
          <span>商品信息</span>
        </template>
        <div class="order-items">
          <div
            v-for="item in order.items"
            :key="item.id"
            class="order-item"
          >
            <img
              :src="item.image || '/placeholder.svg'"
              :alt="item.title"
              class="item-image"
            />
            <div class="item-info">
              <span class="item-title">{{ item.title }}</span>
              <span class="item-price">{{ formatPrice(item.price) }}</span>
            </div>
            <span class="item-quantity">x{{ item.quantity }}</span>
            <span class="item-subtotal">{{ formatPrice(item.price * item.quantity) }}</span>
          </div>
        </div>

        <div class="order-total-line">
          <span>
            共 {{ order.items.length }} 件商品
            <span class="total-amount">合计：<em>{{ formatPrice(order.totalAmount) }}</em></span>
          </span>
        </div>
      </el-card>

      <!-- 收货地址 -->
      <el-card v-if="order.shippingAddress" class="detail-card">
        <template #header>
          <span>收货信息</span>
        </template>
        <div class="address-info">
          <p>
            <span class="address-label">收货人：</span>
            {{ order.shippingAddress.name }}
          </p>
          <p>
            <span class="address-label">联系电话：</span>
            {{ order.shippingAddress.phone }}
          </p>
          <p>
            <span class="address-label">收货地址：</span>
            {{ order.shippingAddress.province }}
            {{ order.shippingAddress.city }}
            {{ order.shippingAddress.district }}
            {{ order.shippingAddress.detail }}
          </p>
        </div>
      </el-card>

      <!-- 订单信息 -->
      <el-card class="detail-card">
        <template #header>
          <span>订单信息</span>
        </template>
        <div class="order-meta">
          <p><span class="meta-label">订单编号：</span>{{ order.orderNo }}</p>
          <p><span class="meta-label">创建时间：</span>{{ formatDate(order.createdAt) }}</p>
          <p><span class="meta-label">更新时间：</span>{{ formatDate(order.updatedAt) }}</p>
        </div>
      </el-card>
    </template>
  </div>
</template>

<style scoped>
.order-detail-page {
  padding-bottom: 40px;
}

.breadcrumb {
  margin-bottom: 20px;
}

.loading-state {
  background: #fff;
  border-radius: 8px;
  padding: 40px;
}

.error-state {
  padding: 40px;
}

.order-status-header {
  background: #fff;
  border-radius: 12px;
  margin-bottom: 20px;
}

.detail-card {
  margin-bottom: 16px;
  border-radius: 8px;
}

.order-items {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.order-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 0;
  border-bottom: 1px solid #f5f5f5;
}

.order-item:last-child {
  border-bottom: none;
}

.item-image {
  width: 80px;
  height: 80px;
  border-radius: 6px;
  object-fit: cover;
  background: #f5f5f5;
}

.item-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.item-title {
  font-size: 15px;
  font-weight: 500;
  color: #333;
}

.item-price {
  font-size: 14px;
  color: #999;
}

.item-quantity {
  font-size: 14px;
  color: #666;
}

.item-subtotal {
  font-size: 15px;
  font-weight: 500;
  color: #f56c6c;
  min-width: 80px;
  text-align: right;
}

.order-total-line {
  text-align: right;
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
  font-size: 14px;
  color: #666;
}

.total-amount {
  margin-left: 16px;
}

.total-amount em {
  font-size: 20px;
  font-weight: bold;
  color: #f56c6c;
  font-style: normal;
  margin-left: 4px;
}

.address-info p {
  margin: 8px 0;
  font-size: 14px;
  color: #333;
}

.address-label {
  color: #999;
  display: inline-block;
  min-width: 80px;
}

.order-meta p {
  margin: 8px 0;
  font-size: 14px;
  color: #333;
}

.meta-label {
  color: #999;
  display: inline-block;
  min-width: 80px;
}
</style>
