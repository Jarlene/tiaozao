<template>
  <n-space vertical>
    <n-h2>我的出售</n-h2>

    <!-- 状态筛选 -->
    <n-card :bordered="true" size="small">
      <n-space>
        <n-button
          v-for="tab in statusTabs"
          :key="tab.value"
          :type="currentStatus === tab.value ? 'primary' : 'default'"
          size="small"
          @click="changeStatus(tab.value)"
        >
          {{ tab.label }}
        </n-button>
      </n-space>
    </n-card>

    <!-- 订单列表 -->
    <n-spin :show="loading">
      <n-empty v-if="orders.length === 0 && !loading" description="暂无订单" style="padding: 60px 0" />

      <n-list v-else>
        <n-list-item v-for="order in orders" :key="order.id">
          <template #prefix>
            <n-image
              v-if="order.product?.images && order.product.images.length > 0"
              :src="order.product.images[0].url"
              style="width: 80px; height: 80px; object-fit: cover; border-radius: 4px; cursor: pointer"
              fallback-src="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 80 80'><rect fill='%23e0e0e0' width='80' height='80'/></svg>"
              @click="$router.push(`/products/${order.product_id}`)"
            />
            <div
              v-else
              style="width: 80px; height: 80px; background: #f0f0f0; border-radius: 4px; display: flex; align-items: center; justify-content: center; color: #999; font-size: 12px; cursor: pointer"
              @click="$router.push(`/products/${order.product_id}`)"
            >
              无图
            </div>
          </template>

          <!-- 订单信息 -->
          <n-space vertical :size="4" style="flex: 1">
            <n-space align="center">
              <n-text
                strong
                style="cursor: pointer"
                @click="$router.push(`/orders/${order.id}`)"
              >
                {{ order.title }}
              </n-text>
            </n-space>
            <n-text style="color: #f5222d; font-weight: bold">
              {{ formatPrice(order.price) }}
            </n-text>
            <n-space align="center" size="small">
              <n-tag :type="getStatusTagType(order.status)" size="tiny">
                {{ order.status_name }}
              </n-tag>
              <n-text depth="3" style="font-size: 12px">{{ order.created_at }}</n-text>
            </n-space>
          </n-space>

          <!-- 操作按钮 -->
          <template #suffix>
            <n-space vertical>
              <!-- 所有状态都显示查看详情 -->
              <n-button size="tiny" quaternary @click="$router.push(`/orders/${order.id}`)">
                详情
              </n-button>

              <!-- 待发货：发货 -->
              <template v-if="order.status === OrderStatus.PENDING_SHIPMENT">
                <n-button size="tiny" type="primary" @click="handleShip(order.id)">
                  立即发货
                </n-button>
              </template>

              <!-- 退款中：同意退款 / 拒绝退款 -->
              <template v-else-if="order.status === OrderStatus.REFUNDING">
                <n-button size="tiny" type="success" @click="handleApproveRefund(order.id)">
                  同意退款
                </n-button>
                <n-button size="tiny" type="warning" @click="handleRejectRefund(order.id)">
                  拒绝退款
                </n-button>
              </template>

              <!-- 纠纷中 -->
              <template v-else-if="order.status === OrderStatus.DISPUTE">
                <n-button size="tiny" disabled>
                  纠纷处理中
                </n-button>
              </template>

              <!-- 已完成 -->
              <template v-else-if="order.status === OrderStatus.COMPLETED">
                <n-button
                  size="tiny"
                  @click="$router.push(`/products/${order.product_id}`)"
                >
                  查看商品
                </n-button>
              </template>

              <!-- 待付款 / 已付款 / 已发货等状态只显示状态文本 -->
              <template v-else>
                <n-button size="tiny" disabled>
                  {{ getPassiveLabel(order.status) }}
                </n-button>
              </template>
            </n-space>
          </template>
        </n-list-item>
      </n-list>

      <!-- 分页 -->
      <n-space justify="center" style="margin-top: 16px" v-if="totalPages > 1">
        <n-pagination
          :page="currentPage"
          :page-size="pageSize"
          :item-count="total"
          @update:page="changePage"
        />
      </n-space>
    </n-spin>
  </n-space>
</template>

<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useMessage, useDialog } from 'naive-ui'
import { orderAPI, OrderStatus } from '@/api/orders'
import { formatPrice } from '@/utils/format'
import type { OrderListItem } from '@/api/orders'

const message = useMessage()
const dialog = useDialog()

const loading = ref(false)
const orders = ref<OrderListItem[]>([])
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const currentStatus = ref<number | undefined>(undefined)
const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

// 筛选标签：卖家视角
const statusTabs = [
  { label: '全部', value: undefined },
  { label: '待发货', value: OrderStatus.PENDING_SHIPMENT },
  { label: '已发货', value: OrderStatus.SHIPPED },
  { label: '已完成', value: OrderStatus.COMPLETED },
  { label: '退款中', value: OrderStatus.REFUNDING },
  { label: '纠纷中', value: OrderStatus.DISPUTE },
]

// 状态标签颜色映射
function getStatusTagType(status: number): 'default' | 'primary' | 'info' | 'success' | 'warning' | 'error' {
  switch (status) {
    case OrderStatus.PENDING_PAYMENT: return 'warning'
    case OrderStatus.PAID:
    case OrderStatus.PENDING_SHIPMENT: return 'info'
    case OrderStatus.SHIPPED: return 'primary'
    case OrderStatus.RECEIVED:
    case OrderStatus.COMPLETED: return 'success'
    case OrderStatus.REFUNDING:
    case OrderStatus.DISPUTE: return 'error'
    case OrderStatus.CANCELLED: return 'default'
    default: return 'default'
  }
}

function getPassiveLabel(status: number): string {
  switch (status) {
    case OrderStatus.PENDING_PAYMENT: return '等待买家付款'
    case OrderStatus.PAID: return '买家已付款'
    case OrderStatus.SHIPPED: return '等待买家确认'
    case OrderStatus.RECEIVED: return '买家已收货'
    case OrderStatus.CANCELLED: return '已取消'
    default: return ''
  }
}

async function loadOrders() {
  loading.value = true
  try {
    const result = await orderAPI.listSold({
      status: currentStatus.value,
      page: currentPage.value,
      size: pageSize.value,
    })
    orders.value = result.items
    total.value = result.total
  } catch {
    orders.value = []
    message.error('加载订单列表失败')
  } finally {
    loading.value = false
  }
}

function changeStatus(status: number | undefined) {
  currentStatus.value = status
  currentPage.value = 1
  loadOrders()
}

function changePage(page: number) {
  currentPage.value = page
  loadOrders()
}

async function handleShip(orderId: number) {
  dialog.info({
    title: '确认发货',
    content: '确定要标记该订单为已发货吗？',
    positiveText: '确认发货',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await orderAPI.ship(orderId)
        message.success('已标记为发货')
        loadOrders()
      } catch (err: any) {
        const msg = err?.response?.data?.message || '操作失败'
        message.error(msg)
      }
    },
  })
}

function handleApproveRefund(orderId: number) {
  dialog.warning({
    title: '同意退款',
    content: '确定要同意该退款申请吗？退款将根据订单状态自动处理资金。',
    positiveText: '同意退款',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await orderAPI.approveRefund(orderId)
        message.success('已同意退款')
        loadOrders()
      } catch (err: any) {
        const msg = err?.response?.data?.message || '操作失败'
        message.error(msg)
      }
    },
  })
}

function handleRejectRefund(orderId: number) {
  dialog.warning({
    title: '拒绝退款',
    content: '拒绝退款后订单将进入纠纷处理，需要管理员介入仲裁。确定拒绝吗？',
    positiveText: '拒绝退款',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await orderAPI.rejectRefund(orderId)
        message.success('已拒绝退款，订单进入纠纷处理')
        loadOrders()
      } catch (err: any) {
        const msg = err?.response?.data?.message || '操作失败'
        message.error(msg)
      }
    },
  })
}

onMounted(() => {
  loadOrders()
})
</script>
