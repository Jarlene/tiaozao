<template>
  <n-space vertical>
    <n-h2>我的订单</n-h2>

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

              <!-- 待付款：立即付款 + 取消订单 -->
              <template v-if="order.status === OrderStatus.PENDING_PAYMENT">
                <n-button size="tiny" type="primary" @click="handlePay(order.id)">
                  立即付款
                </n-button>
                <n-button size="tiny" @click="handleCancel(order.id)">
                  取消订单
                </n-button>
              </template>

              <!-- 待发货 / 已付款：无买家操作 -->
              <template v-else-if="order.status === OrderStatus.PENDING_SHIPMENT || order.status === OrderStatus.PAID">
                <n-button size="tiny" disabled>
                  等待卖家发货
                </n-button>
              </template>

              <!-- 已发货：确认收货 + 申请退款 -->
              <template v-else-if="order.status === OrderStatus.SHIPPED">
                <n-button size="tiny" type="primary" @click="handleConfirmReceive(order.id)">
                  确认收货
                </n-button>
                <n-button size="tiny" type="warning" @click="handleRequestRefund(order.id)">
                  申请退款
                </n-button>
              </template>

              <!-- 退款中：无买家操作 -->
              <template v-else-if="order.status === OrderStatus.REFUNDING">
                <n-button size="tiny" disabled>
                  退款处理中
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
    
    <!-- 退款原因弹窗 -->
    <n-modal v-model:show="showRefundModal" preset="dialog" title="申请退款">
      <n-space vertical>
        <n-text>请填写退款原因：</n-text>
        <n-input
          v-model:value="refundReason"
          type="textarea"
          placeholder="请说明退款原因..."
          :maxlength="500"
          show-count
        />
      </n-space>
      <template #action>
        <n-space justify="end">
          <n-button @click="showRefundModal = false">取消</n-button>
          <n-button type="warning" @click="confirmRefund">申请退款</n-button>
        </n-space>
      </template>
    </n-modal>
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
const showRefundModal = ref(false)
const refundReason = ref('')
const refundOrderId = ref(0)
const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

// 筛选标签：匹配买家视角的常见筛选维度
const statusTabs = [
  { label: '全部', value: undefined },
  { label: '待付款', value: OrderStatus.PENDING_PAYMENT },
  { label: '待发货', value: OrderStatus.PENDING_SHIPMENT },
  { label: '待收货', value: OrderStatus.SHIPPED },
  { label: '已完成', value: OrderStatus.COMPLETED },
  { label: '退款中', value: OrderStatus.REFUNDING },
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

async function loadOrders() {
  loading.value = true
  try {
    const result = await orderAPI.listMine({
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

async function handlePay(orderId: number) {
  try {
    await orderAPI.pay(orderId)
    message.success('付款成功')
    loadOrders()
  } catch (err: any) {
    const msg = err?.response?.data?.message || '付款失败'
    message.error(msg)
  }
}

function handleCancel(orderId: number) {
  dialog.warning({
    title: '取消订单',
    content: '确定要取消该订单吗？取消后将恢复商品库存。',
    positiveText: '确定取消',
    negativeText: '再想想',
    onPositiveClick: async () => {
      try {
        await orderAPI.cancel(orderId)
        message.success('订单已取消')
        loadOrders()
      } catch (err: any) {
        const msg = err?.response?.data?.message || '取消失败'
        message.error(msg)
      }
    },
  })
}

async function handleConfirmReceive(orderId: number) {
  dialog.info({
    title: '确认收货',
    content: '请确认已收到商品。确认后将完成订单并释放资金给卖家。',
    positiveText: '确认收货',
    negativeText: '再等等',
    onPositiveClick: async () => {
      try {
        await orderAPI.confirmReceive(orderId)
        message.success('已确认收货')
        loadOrders()
      } catch (err: any) {
        const msg = err?.response?.data?.message || '操作失败'
        message.error(msg)
      }
    },
  })
}

function handleRequestRefund(orderId: number) {
  refundOrderId.value = orderId
  refundReason.value = ''
  showRefundModal.value = true
}

async function confirmRefund() {
  if (!refundReason.value.trim()) {
    message.warning('请填写退款原因')
    return
  }
  showRefundModal.value = false
  try {
    await orderAPI.requestRefund(refundOrderId.value, { reason: refundReason.value.trim() })
    message.success('退款申请已提交')
    loadOrders()
  } catch (err: any) {
    const msg = err?.response?.data?.message || '申请失败'
    message.error(msg)
  }
}

onMounted(() => {
  loadOrders()
})
</script>
