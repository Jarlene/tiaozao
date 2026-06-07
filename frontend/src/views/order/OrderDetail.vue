<template>
  <n-spin :show="loading">
    <n-card v-if="order" style="max-width: 800px; margin: 0 auto">
      <template #header>
        <n-space align="center" justify="space-between">
          <n-space align="center">
            <n-button quaternary size="small" @click="goBack">
              ← 返回
            </n-button>
            <n-h2 style="margin: 0; font-size: 20px">订单详情</n-h2>
          </n-space>
          <n-tag :type="statusTagType" size="medium">
            {{ order.status_name }}
          </n-tag>
        </n-space>
      </template>

      <!-- 商品信息 -->
      <n-card title="商品信息" size="small" style="margin-bottom: 16px">
        <n-space align="center">
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

          <n-space vertical :size="4" style="flex: 1">
            <n-text
              strong
              style="font-size: 16px; cursor: pointer; color: #18a058"
              @click="$router.push(`/products/${order.product_id}`)"
            >
              {{ order.title }}
            </n-text>
            <n-text style="color: #f5222d; font-weight: bold; font-size: 22px">
              {{ formatPrice(order.price) }}
            </n-text>
          </n-space>
        </n-space>
      </n-card>

      <!-- 订单信息 -->
      <n-card title="订单信息" size="small" style="margin-bottom: 16px">
        <n-descriptions label-placement="left" :column="1">
          <n-descriptions-item label="订单编号">
            <n-text>{{ order.order_no }}</n-text>
          </n-descriptions-item>
          <n-descriptions-item label="订单状态">
            <n-tag :type="statusTagType" size="small">{{ order.status_name }}</n-tag>
          </n-descriptions-item>
          <n-descriptions-item v-if="order.buyer_note" label="买家留言">
            <n-text>{{ order.buyer_note }}</n-text>
          </n-descriptions-item>
          <n-descriptions-item label="创建时间">
            {{ order.created_at }}
          </n-descriptions-item>
          <n-descriptions-item label="更新时间">
            {{ order.updated_at }}
          </n-descriptions-item>
        </n-descriptions>
      </n-card>

      <!-- 收货地址 -->
      <n-card title="收货地址" size="small" style="margin-bottom: 16px">
        <n-text>{{ order.shipping_address }}</n-text>
      </n-card>

      <!-- 物流信息（已发货/已收货/已完成/退款中时显示） -->
      <n-card v-if="order.tracking_number || order.logistics_company" title="物流信息" size="small" style="margin-bottom: 16px">
        <n-descriptions label-placement="left" :column="1">
          <n-descriptions-item v-if="order.logistics_company" label="物流公司">
            <n-text>{{ order.logistics_company }}</n-text>
          </n-descriptions-item>
          <n-descriptions-item v-if="order.tracking_number" label="快递单号">
            <n-text>{{ order.tracking_number }}</n-text>
          </n-descriptions-item>
        </n-descriptions>
      </n-card>

      <!-- 联系对方 -->
      <n-card size="small" style="margin-bottom: 16px">
        <n-space align="center" justify="space-between">
          <n-text>
            <template v-if="isBuyer">需要联系卖家？</template>
            <template v-else-if="isSeller">需要联系买家？</template>
          </n-text>
          <n-button size="small" type="primary" ghost :loading="contactLoading" @click="handleContact">
            <template #icon>
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
              </svg>
            </template>
            {{ isBuyer ? '联系卖家' : '联系买家' }}
          </n-button>
        </n-space>
      </n-card>

      <!-- 状态变更时间线 -->
      <n-card title="状态变更记录" size="small" style="margin-bottom: 16px">
        <n-spin :show="logsLoading">
          <n-empty v-if="!logsLoading && statusLogs.length === 0" description="暂无记录" />
          <n-timeline v-else>
            <n-timeline-item
              v-for="log in statusLogs"
              :key="log.id"
              :type="getTimelineItemType(log)"
              :time="log.created_at"
              :line-type="'default'"
            >
              <template #header>
                <n-tag size="tiny" :type="getStatusTagType(log.to_status)">
                  {{ getStatusName(log.to_status) }}
                </n-tag>
              </template>
              {{ formatLogEvent(log) }}
              <template #footer>
                <n-text depth="3" style="font-size: 12px">
                  {{ getOperatorLabel(log) }}
                </n-text>
              </template>
            </n-timeline-item>
          </n-timeline>
        </n-spin>
      </n-card>

      <!-- 操作按钮 -->
      <template #footer>
        <n-space justify="center">
          <n-button @click="goBack">返回列表</n-button>

          <!-- ===== 买家操作 ===== -->
          <template v-if="isBuyer">
            <!-- 待付款：付款 + 取消 -->
            <template v-if="order.status === OrderStatus.PENDING_PAYMENT">
              <n-button type="primary" :loading="actionLoading" @click="handlePay">
                立即付款
              </n-button>
              <n-button :loading="actionLoading" @click="handleCancel">
                取消订单
              </n-button>
            </template>

            <!-- 已发货：确认收货 + 申请退款 -->
            <template v-else-if="order.status === OrderStatus.SHIPPED">
              <n-button type="primary" :loading="actionLoading" @click="handleConfirmReceive">
                确认收货
              </n-button>
              <n-button type="warning" :loading="actionLoading" @click="handleRequestRefund">
                申请退款
              </n-button>
            </template>

            <!-- 已完成：查看商品 -->
            <template v-else-if="order.status === OrderStatus.COMPLETED">
              <n-button @click="$router.push(`/products/${order.product_id}`)">
                查看商品
              </n-button>
            </template>

            <!-- 已取消 -->
            <template v-else-if="order.status === OrderStatus.CANCELLED">
              <n-button @click="$router.push(`/products/${order.product_id}`)">
                重新购买
              </n-button>
            </template>
          </template>

          <!-- ===== 卖家操作 ===== -->
          <template v-else-if="isSeller">
            <!-- 待发货：发货 -->
            <template v-if="order.status === OrderStatus.PENDING_SHIPMENT">
              <n-button type="primary" :loading="actionLoading" @click="openShipModal">
                立即发货
              </n-button>
            </template>

            <!-- 退款中：同意退款 / 拒绝退款 -->
            <template v-else-if="order.status === OrderStatus.REFUNDING">
              <n-button type="success" :loading="actionLoading" @click="handleApproveRefund">
                同意退款
              </n-button>
              <n-button type="warning" :loading="actionLoading" @click="handleRejectRefund">
                拒绝退款
              </n-button>
            </template>

            <!-- 已完成 -->
            <template v-else-if="order.status === OrderStatus.COMPLETED">
              <n-button @click="$router.push(`/products/${order.product_id}`)">
                查看商品
              </n-button>
            </template>
          </template>
        </n-space>
      </template>
    </n-card>
  </n-spin>

  <!-- 发货弹窗 -->
  <n-modal v-model:show="showShipModal" preset="card" title="发货" style="width: 420px" :mask-closable="false">
    <n-space vertical>
      <n-form label-placement="top" label-width="auto">
        <n-form-item label="快递单号">
          <n-input v-model:value="trackingNumber" placeholder="请输入快递单号（选填）" clearable />
        </n-form-item>
        <n-form-item label="物流公司">
          <n-input v-model:value="logisticsCompany" placeholder="请输入物流公司名称（选填）" clearable />
        </n-form-item>
      </n-form>
      <n-space justify="end">
        <n-button @click="closeShipModal">取消</n-button>
        <n-button type="primary" :loading="actionLoading" @click="submitShip">确认发货</n-button>
      </n-space>
    </n-space>
  </n-modal>

  <!-- 退款弹窗 -->
  <n-modal v-model:show="showRefundModal" preset="card" title="申请退款" style="width: 420px" :mask-closable="false">
    <n-space vertical>
      <n-form label-placement="top" label-width="auto">
        <n-form-item label="退款原因" required>
          <n-select
            v-model:value="refundReason"
            :options="refundReasonOptions"
            placeholder="请选择退款原因"
            clearable
          />
        </n-form-item>
        <n-form-item label="备注说明">
          <n-input
            v-model:value="refundNote"
            type="textarea"
            :maxlength="500"
            placeholder="请输入补充说明（选填）"
            clearable
          />
        </n-form-item>
      </n-form>
      <n-space justify="end">
        <n-button @click="closeRefundModal">取消</n-button>
        <n-button type="warning" :loading="actionLoading" @click="submitRefund">提交退款申请</n-button>
      </n-space>
    </n-space>
  </n-modal>
</template>

<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMessage, useDialog } from 'naive-ui'
import { orderAPI, OrderStatus } from '@/api/orders'
import { chatAPI } from '@/api/chat'
import { formatPrice } from '@/utils/format'
import { useAuthStore } from '@/stores/auth'
import type { OrderDetail } from '@/api/orders'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const dialog = useDialog()
const authStore = useAuthStore()

const loading = ref(true)
const order = ref<OrderDetail | null>(null)

const logsLoading = ref(false)
const statusLogs = ref<any[]>([])
const actionLoading = ref(false)
const contactLoading = ref(false)

// 发货弹窗状态
const showShipModal = ref(false)
const trackingNumber = ref('')
const logisticsCompany = ref('')

// 退款弹窗状态
const showRefundModal = ref(false)
const refundReason = ref<string | null>(null)
const refundNote = ref('')
const refundReasonOptions = [
  { label: '商品与描述不符', value: '商品与描述不符' },
  { label: '质量问题', value: '质量问题' },
  { label: '未收到货', value: '未收到货' },
  { label: '卖家未发货', value: '卖家未发货' },
  { label: '其他原因', value: '其他原因' },
]

function openShipModal() {
  trackingNumber.value = order.value?.tracking_number || ''
  logisticsCompany.value = order.value?.logistics_company || ''
  showShipModal.value = true
}

function closeShipModal() {
  showShipModal.value = false
}

async function submitShip() {
  if (!order.value) return
  actionLoading.value = true
  try {
    await orderAPI.ship(order.value.id, {
      tracking_number: trackingNumber.value || undefined,
      logistics_company: logisticsCompany.value || undefined,
    })
    message.success('已标记为发货')
    showShipModal.value = false
    await loadOrder()
    await loadStatusLogs()
  } catch (err: any) {
    const msg = err?.response?.data?.message || '操作失败'
    message.error(msg)
  } finally {
    actionLoading.value = false
  }
}

const isBuyer = computed(() => {
  if (!order.value || !authStore.user) return false
  return order.value.buyer_id === authStore.user.id
})

const isSeller = computed(() => {
  if (!order.value || !authStore.user) return false
  return order.value.seller_id === authStore.user.id
})

const statusTagType = computed(() => {
  if (!order.value) return 'default'
  return getStatusTagType(order.value.status)
})

function getStatusTagType(status: number): 'default' | 'primary' | 'info' | 'success' | 'warning' | 'error' {
  switch (status) {
    case OrderStatus.PENDING_PAYMENT: return 'warning'
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

function getStatusName(status: number): string {
  const names: Record<number, string> = {
    [OrderStatus.PENDING_PAYMENT]: '待付款',
    [OrderStatus.CANCELLED]: '已取消',
    [OrderStatus.PAID]: '已付款',
    [OrderStatus.PENDING_SHIPMENT]: '待发货',
    [OrderStatus.SHIPPED]: '已发货',
    [OrderStatus.RECEIVED]: '已收货',
    [OrderStatus.COMPLETED]: '已完成',
    [OrderStatus.REFUNDING]: '退款中',
    [OrderStatus.DISPUTE]: '纠纷处理',
  }
  return names[status] || `状态(${status})`
}

function getTimelineItemType(log: any): 'default' | 'success' | 'info' | 'warning' | 'error' {
  switch (log.to_status) {
    case OrderStatus.COMPLETED: return 'success'
    case OrderStatus.CANCELLED:
    case OrderStatus.DISPUTE: return 'error'
    case OrderStatus.SHIPPED: return 'info'
    case OrderStatus.REFUNDING: return 'warning'
    default: return 'default'
  }
}

function formatLogEvent(log: any): string {
  const event = log.event || ''
  // 退款申请包含原因：request_refund:商品质量问题
  if (event.startsWith('request_refund:')) {
    const reason = event.substring('request_refund:'.length)
    return `申请退款 - 原因：${reason}`
  }
  // 仲裁包含决定：arbitrate:buyer
  if (event.startsWith('arbitrate:')) {
    const parts = event.split(':')
    const decision = parts.length > 1 ? parts[1] : ''
    if (decision === 'buyer') return '仲裁结果：支持买家'
    if (decision === 'seller') return '仲裁结果：支持卖家'
    return `仲裁结果：${decision}`
  }
  // 常规事件名转中文
  const eventNames: Record<string, string> = {
    create: '创建订单',
    cancel: '取消订单',
    pay: '付款',
    ship: '发货',
    confirm_receive: '确认收货',
    complete: '自动完成',
    request_refund: '申请退款',
    approve_refund: '同意退款',
    reject_refund: '拒绝退款',
    refund_success: '退款成功',
    raise_dispute: '发起纠纷',
    arbitrate: '仲裁',
    notify_seller: '通知卖家',
  }
  return eventNames[event] || event
}

function getOperatorLabel(log: any): string {
  const types: Record<string, string> = {
    buyer: '买家',
    seller: '卖家',
    admin: '管理员',
    system: '系统',
  }
  return types[log.operator_type] || log.operator_type
}

function goBack() {
  // 如果是卖家视角，返回出售列表；否则返回我的订单
  if (isSeller.value && !isBuyer.value) {
    router.push('/orders/sold')
  } else {
    router.push('/orders')
  }
}

async function loadOrder() {
  loading.value = true
  try {
    const id = Number(route.params.id)
    if (!id) {
      message.error('无效的订单ID')
      router.push('/orders')
      return
    }
    order.value = await orderAPI.getById(id)
  } catch (err: any) {
    const msg = err?.response?.data?.message || '加载订单详情失败'
    message.error(msg)
    router.push('/orders')
  } finally {
    loading.value = false
  }
}

async function loadStatusLogs() {
  logsLoading.value = true
  try {
    const id = Number(route.params.id)
    statusLogs.value = await orderAPI.getStatusLogs(id)
  } catch {
    // 日志加载失败不影响主页面展示
    statusLogs.value = []
  } finally {
    logsLoading.value = false
  }
}

// ========== 联系对方 ==========

async function handleContact() {
  if (!order.value) return
  contactLoading.value = true
  try {
    // 创建或获取与对方的会话，买家视角传 seller_id，卖家视角传 buyer_id
    const targetID = isBuyer.value ? order.value.seller_id : order.value.buyer_id
    const conv = await chatAPI.createConversation({
      product_id: order.value.product_id,
      seller_id: isBuyer.value ? order.value.seller_id : targetID,
    })
    router.push(`/chat/${conv.id}`)
  } catch (err: any) {
    const msg = err?.response?.data?.message || '跳转到聊天失败'
    message.error(msg)
  } finally {
    contactLoading.value = false
  }
}

// ========== 买家操作 ==========

async function handlePay() {
  if (!order.value) return
  actionLoading.value = true
  try {
    await orderAPI.pay(order.value.id)
    message.success('付款成功')
    await loadOrder()
    await loadStatusLogs()
  } catch (err: any) {
    const msg = err?.response?.data?.message || '付款失败'
    message.error(msg)
  } finally {
    actionLoading.value = false
  }
}

function handleCancel() {
  if (!order.value) return
  dialog.warning({
    title: '取消订单',
    content: '确定要取消该订单吗？取消后将恢复商品库存。',
    positiveText: '确定取消',
    negativeText: '再想想',
    onPositiveClick: async () => {
      actionLoading.value = true
      try {
        await orderAPI.cancel(order.value!.id)
        message.success('订单已取消')
        await loadOrder()
        await loadStatusLogs()
      } catch (err: any) {
        const msg = err?.response?.data?.message || '取消失败'
        message.error(msg)
      } finally {
        actionLoading.value = false
      }
    },
  })
}

async function handleConfirmReceive() {
  if (!order.value) return
  dialog.info({
    title: '确认收货',
    content: '请确认已收到商品。确认后将完成订单并释放资金给卖家。',
    positiveText: '确认收货',
    negativeText: '再等等',
    onPositiveClick: async () => {
      actionLoading.value = true
      try {
        await orderAPI.confirmReceive(order.value!.id)
        message.success('已确认收货')
        await loadOrder()
        await loadStatusLogs()
      } catch (err: any) {
        const msg = err?.response?.data?.message || '操作失败'
        message.error(msg)
      } finally {
        actionLoading.value = false
      }
    },
  })
}

function handleRequestRefund() {
  if (!order.value) return
  refundReason.value = null
  refundNote.value = ''
  showRefundModal.value = true
}

function closeRefundModal() {
  showRefundModal.value = false
  refundReason.value = null
  refundNote.value = ''
}

async function submitRefund() {
  if (!order.value) return
  if (!refundReason.value) {
    message.warning('请选择退款原因')
    return
  }
  actionLoading.value = true
  try {
    await orderAPI.requestRefund(order.value!.id, {
      reason: refundReason.value,
      note: refundNote.value || undefined,
    })
    message.success('退款申请已提交')
    showRefundModal.value = false
    await loadOrder()
    await loadStatusLogs()
  } catch (err: any) {
    const msg = err?.response?.data?.message || '申请失败'
    message.error(msg)
  } finally {
    actionLoading.value = false
  }
}

// ========== 卖家操作 ==========

async function handleApproveRefund() {
  if (!order.value) return
  dialog.warning({
    title: '同意退款',
    content: '确定要同意该退款申请吗？退款将根据订单状态自动处理资金。',
    positiveText: '同意退款',
    negativeText: '取消',
    onPositiveClick: async () => {
      actionLoading.value = true
      try {
        await orderAPI.approveRefund(order.value!.id)
        message.success('已同意退款')
        await loadOrder()
        await loadStatusLogs()
      } catch (err: any) {
        const msg = err?.response?.data?.message || '操作失败'
        message.error(msg)
      } finally {
        actionLoading.value = false
      }
    },
  })
}

function handleRejectRefund() {
  if (!order.value) return
  dialog.warning({
    title: '拒绝退款',
    content: '拒绝退款后订单将进入纠纷处理，需要管理员介入仲裁。确定拒绝吗？',
    positiveText: '拒绝退款',
    negativeText: '取消',
    onPositiveClick: async () => {
      actionLoading.value = true
      try {
        await orderAPI.rejectRefund(order.value!.id)
        message.success('已拒绝退款')
        await loadOrder()
        await loadStatusLogs()
      } catch (err: any) {
        const msg = err?.response?.data?.message || '操作失败'
        message.error(msg)
      } finally {
        actionLoading.value = false
      }
    },
  })
}

onMounted(async () => {
  await loadOrder()
  if (order.value) {
    await loadStatusLogs()
  }
})
</script>
