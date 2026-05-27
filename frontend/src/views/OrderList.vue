<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import type { Order, OrderStatus } from '@/api'
import { orderApi } from '@/api'

const router = useRouter()

const orders = ref<Order[]>([])
const loading = ref(false)
const error = ref('')
const total = ref(0)
const currentPage = ref(1)
const activeTab = ref('all')

/** 订单状态标签映射 */
const statusTabs = [
  { key: 'all', label: '全部' },
  { key: 'pending', label: '待付款' },
  { key: 'paid', label: '待发货' },
  { key: 'shipped', label: '待收货' },
  { key: 'completed', label: '已完成' },
  { key: 'cancelled', label: '已取消' },
]

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

onMounted(async () => {
  await loadOrders()
})

/** 加载订单列表 */
async function loadOrders() {
  loading.value = true
  error.value = ''
  try {
    const params: { page: number; pageSize: number; status?: OrderStatus } = {
      page: currentPage.value,
      pageSize: 10,
    }
    if (activeTab.value !== 'all') {
      params.status = activeTab.value as OrderStatus
    }
    const res = await orderApi.getList(params)
    orders.value = res.items
    total.value = res.total
  } catch (err: any) {
    error.value = err?.response?.data?.message || '加载订单列表失败'
  } finally {
    loading.value = false
  }
}

/** 切换状态标签 */
function handleTabChange(tab: string) {
  activeTab.value = tab
  currentPage.value = 1
  loadOrders()
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
  <div class="order-list-page">
    <h2 class="page-title">我的订单</h2>

    <!-- 状态标签页 -->
    <el-tabs
      :model-value="activeTab"
      class="order-tabs"
      @tab-click="(tab: any) => handleTabChange(tab.props.name)"
    >
      <el-tab-pane
        v-for="tab in statusTabs"
        :key="tab.key"
        :label="tab.label"
        :name="tab.key"
      >
        <!-- 加载状态 -->
        <div v-if="loading" class="loading-state">
          <el-skeleton :rows="4" animated />
        </div>

        <!-- 错误状态 -->
        <div v-else-if="error" class="error-state">
          <el-result icon="error" title="加载失败" :sub-title="error">
            <template #extra>
              <el-button type="primary" @click="loadOrders">重新加载</el-button>
            </template>
          </el-result>
        </div>

        <!-- 空状态 -->
        <div v-else-if="orders.length === 0" class="empty-state">
          <el-empty description="暂无订单" />
        </div>

        <!-- 订单列表 -->
        <div v-else class="order-cards">
          <el-card
            v-for="order in orders"
            :key="order.id"
            class="order-card"
            shadow="hover"
            @click="router.push(`/orders/${order.id}`)"
          >
            <div class="order-header">
              <div class="order-info">
                <span class="order-no">订单号：{{ order.orderNo }}</span>
                <span class="order-time">{{ formatDate(order.createdAt) }}</span>
              </div>
              <el-tag :type="statusTypeMap[order.status] || 'info'">
                {{ statusLabelMap[order.status] || order.status }}
              </el-tag>
            </div>

            <div class="order-body">
              <div
                v-for="item in order.items.slice(0, 3)"
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
                  <span class="item-meta">{{ formatPrice(item.price) }} x {{ item.quantity }}</span>
                </div>
              </div>
              <div v-if="order.items.length > 3" class="more-items">
                等 {{ order.items.length }} 件商品
              </div>
            </div>

            <div class="order-footer">
              <span class="order-total">
                合计：<em>{{ formatPrice(order.totalAmount) }}</em>
              </span>
            </div>
          </el-card>

          <!-- 分页 -->
          <div v-if="total > 10" class="pagination-wrapper">
            <el-pagination
              v-model:current-page="currentPage"
              :page-size="10"
              :total="total"
              layout="prev, pager, next"
              background
              @current-change="loadOrders"
            />
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<style scoped>
.order-list-page {
  padding-bottom: 40px;
}

.page-title {
  font-size: 22px;
  font-weight: 600;
  margin: 0 0 20px 0;
  color: #333;
}

.order-tabs {
  background: #fff;
  border-radius: 12px;
  padding: 16px 20px 0;
}

.loading-state {
  padding: 40px;
}

.error-state {
  padding: 20px;
}

.empty-state {
  padding: 60px;
}

.order-cards {
  padding: 16px 0;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.order-card {
  cursor: pointer;
  border-radius: 8px;
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
  margin-bottom: 12px;
}

.order-info {
  display: flex;
  gap: 16px;
  align-items: center;
}

.order-no {
  font-size: 14px;
  color: #666;
}

.order-time {
  font-size: 13px;
  color: #999;
}

.order-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 12px;
}

.order-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.item-image {
  width: 60px;
  height: 60px;
  border-radius: 6px;
  object-fit: cover;
  background: #f5f5f5;
}

.item-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.item-title {
  font-size: 14px;
  color: #333;
}

.item-meta {
  font-size: 13px;
  color: #999;
}

.more-items {
  font-size: 13px;
  color: #999;
  text-align: center;
}

.order-footer {
  display: flex;
  justify-content: flex-end;
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
}

.order-total {
  font-size: 14px;
  color: #666;
}

.order-total em {
  font-size: 18px;
  font-weight: bold;
  color: #f56c6c;
  font-style: normal;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}
</style>
