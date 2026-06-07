<template>
  <n-space vertical>
    <n-h2>纠纷管理（管理员）</n-h2>
    <n-text depth="3">管理所有纠纷订单，查看买卖双方信息并做出仲裁裁决</n-text>

    <!-- 统计卡片 -->
    <n-grid :cols="3" :x-gap="12">
      <n-gi>
        <n-card size="small" :bordered="true">
          <n-statistic label="纠纷总数" :value="total" />
        </n-card>
      </n-gi>
      <n-gi>
        <n-card size="small" :bordered="true">
          <n-statistic label="等待仲裁" :value="total" />
        </n-card>
      </n-gi>
      <n-gi>
        <n-card size="small" :bordered="true">
          <n-statistic label="页数" :value="totalPages" />
        </n-card>
      </n-gi>
    </n-grid>

    <!-- 纠纷订单列表 -->
    <n-spin :show="loading">
      <n-empty v-if="disputes.length === 0 && !loading" description="暂无纠纷订单" style="padding: 60px 0" />

      <n-card
        v-for="order in disputes"
        :key="order.id"
        :bordered="true"
        style="margin-bottom: 12px"
        size="small"
      >
        <n-space vertical :size="12">
          <!-- 顶部：商品主图和基础信息 -->
          <n-space align="center" justify="space-between">
            <n-space align="center">
              <n-image
                v-if="order.product?.images && order.product.images.length > 0"
                :src="order.product.images[0].url"
                style="width: 72px; height: 72px; object-fit: cover; border-radius: 4px"
                fallback-src="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 72 72'><rect fill='%23e0e0e0' width='72' height='72'/></svg>"
              />
              <div
                v-else
                style="width: 72px; height: 72px; background: #f0f0f0; border-radius: 4px; display: flex; align-items: center; justify-content: center; color: #999; font-size: 10px"
              >
                无图
              </div>

              <n-space vertical :size="4">
                <n-text strong style="font-size: 15px">{{ order.title }}</n-text>
                <n-text style="color: #f5222d; font-weight: bold; font-size: 18px">
                  {{ formatPrice(order.price) }}
                </n-text>
                <n-space align="center" size="small">
                  <n-tag type="error" size="tiny">{{ order.status_name }}</n-tag>
                  <n-text depth="3" style="font-size: 12px">{{ order.created_at }}</n-text>
                </n-space>
              </n-space>
            </n-space>

            <n-space>
              <n-button size="small" quaternary @click="$router.push(`/orders/${order.id}`)">
                查看订单详情
              </n-button>
            </n-space>
          </n-space>

          <!-- 中部：买卖双方信息 -->
          <n-grid :cols="2" :x-gap="12">
            <n-gi>
              <n-card size="small" title="买家信息" :bordered="true">
                <n-descriptions label-placement="left" :column="1" size="small">
                  <n-descriptions-item label="买家ID">{{ order.buyer_id }}</n-descriptions-item>
                </n-descriptions>
              </n-card>
            </n-gi>
            <n-gi>
              <n-card size="small" title="卖家信息" :bordered="true">
                <n-descriptions label-placement="left" :column="1" size="small">
                  <n-descriptions-item label="卖家ID">{{ order.seller_id }}</n-descriptions-item>
                </n-descriptions>
              </n-card>
            </n-gi>
          </n-grid>

          <!-- 底部：仲裁操作 -->
          <n-space justify="center">
            <n-button
              type="success"
              size="medium"
              :loading="arbitratingId === order.id"
              @click="handleArbitrate(order.id, 'buyer')"
            >
              裁定支持买家（退款）
            </n-button>
            <n-button
              type="primary"
              size="medium"
              :loading="arbitratingId === order.id"
              @click="handleArbitrate(order.id, 'seller')"
            >
              裁定支持卖家
            </n-button>
          </n-space>
        </n-space>
      </n-card>

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
import { orderAPI } from '@/api/orders'
import { formatPrice } from '@/utils/format'
import type { OrderDetail } from '@/api/orders'

const message = useMessage()
const dialog = useDialog()

const loading = ref(false)
const disputes = ref<OrderDetail[]>([])
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const totalPages = computed(() => Math.ceil(total.value / pageSize.value))
const arbitratingId = ref<number | null>(null)

async function loadDisputes() {
  loading.value = true
  try {
    const result = await orderAPI.listDisputes({
      page: currentPage.value,
      size: pageSize.value,
    })
    disputes.value = result.items
    total.value = result.total
  } catch {
    disputes.value = []
    message.error('加载纠纷列表失败')
  } finally {
    loading.value = false
  }
}

function changePage(page: number) {
  currentPage.value = page
  loadDisputes()
}

function handleArbitrate(orderId: number, decision: string) {
  const label = decision === 'buyer' ? '支持买家（退款给买家）' : '支持卖家（资金释放给卖家）'
  dialog.warning({
    title: '确认仲裁',
    content: `确定裁定${label}吗？仲裁后将不可更改，资金将根据裁定结果自动处理。`,
    positiveText: `确认${label}`,
    negativeText: '取消',
    onPositiveClick: async () => {
      arbitratingId.value = orderId
      try {
        await orderAPI.arbitrate(orderId, decision)
        message.success(`仲裁完成：${label}`)
        loadDisputes()
      } catch (err: any) {
        const msg = err?.response?.data?.message || '仲裁失败'
        message.error(msg)
      } finally {
        arbitratingId.value = null
      }
    },
  })
}

onMounted(() => {
  loadDisputes()
})
</script>
