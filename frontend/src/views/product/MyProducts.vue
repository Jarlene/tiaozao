<template>
  <n-space vertical>
    <n-h2>我的发布</n-h2>

    <!-- 数量统计 -->
    <n-card :bordered="true" size="small">
      <n-space justify="space-around">
        <n-statistic label="全部" :value="countsTotal" />
        <n-statistic label="在售" :value="counts.active">
          <template #suffix>
            <n-tag size="tiny" type="success">在售</n-tag>
          </template>
        </n-statistic>
        <n-statistic label="已售" :value="counts.sold">
          <template #suffix>
            <n-tag size="tiny" type="warning">已售</n-tag>
          </template>
        </n-statistic>
        <n-statistic label="下架" :value="counts.inactive">
          <template #suffix>
            <n-tag size="tiny">下架</n-tag>
          </template>
        </n-statistic>
      </n-space>
    </n-card>

    <!-- 状态筛选 -->
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

    <!-- 商品列表 -->
    <n-spin :show="loading">
      <n-empty v-if="products.length === 0 && !loading" description="暂无商品" style="padding: 60px 0" />

      <n-list v-else>
        <n-list-item v-for="product in products" :key="product.id">
          <template #prefix>
            <n-image
              v-if="product.images && product.images.length > 0"
              :src="product.images[0].url"
              style="width: 80px; height: 80px; object-fit: cover; border-radius: 4px"
              fallback-src="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 80 80'><rect fill='%23e0e0e0' width='80' height='80'/></svg>"
            />
            <div v-else style="width: 80px; height: 80px; background: #f0f0f0; border-radius: 4px; display: flex; align-items: center; justify-content: center; color: #999; font-size: 12px;">
              无图
            </div>
          </template>

          <n-space vertical :size="4">
            <n-space align="center">
              <n-text strong>{{ product.title }}</n-text>
              <n-tag :type="getStatusType(product.status)" size="tiny">
                {{ getStatusLabel(product.status) }}
              </n-tag>
            </n-space>
            <n-text style="color: #f5222d; font-weight: bold">
              ¥{{ (product.price / 100).toFixed(2) }}
            </n-text>
            <n-text depth="3" style="font-size: 12px">{{ product.created_at }}</n-text>
          </n-space>

          <template #suffix>
            <n-space vertical>
              <n-button size="tiny" @click="$router.push(`/products/${product.id}`)">查看</n-button>
              <n-button
                v-if="product.status === 1"
                size="tiny"
                @click="$router.push(`/products/${product.id}/edit`)"
              >
                编辑
              </n-button>
              <n-button
                v-if="product.status === 1"
                size="tiny"
                type="warning"
                @click="handleDeactivate(product.id)"
              >
                下架
              </n-button>
              <n-button
                v-if="product.status === 0"
                size="tiny"
                type="success"
                @click="handleActivate(product.id)"
              >
                重新上架
              </n-button>
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
import { useMessage } from 'naive-ui'
import { productAPI } from '@/api/products'
import type { ProductListItem } from '@/api/products'

const message = useMessage()
const loading = ref(false)
const products = ref<ProductListItem[]>([])
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const currentStatus = ref<number | undefined>(undefined)
const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

const counts = ref({ active: 0, sold: 0, inactive: 0 })
const countsTotal = computed(() => counts.value.active + counts.value.sold + counts.value.inactive)

const statusTabs = [
  { label: '全部', value: undefined },
  { label: '在售', value: 1 },
  { label: '已售', value: 2 },
  { label: '下架', value: 0 },
]

function getStatusType(status: number): 'success' | 'warning' | 'default' {
  switch (status) {
    case 1: return 'success'
    case 2: return 'warning'
    default: return 'default'
  }
}

function getStatusLabel(status: number): string {
  switch (status) {
    case 1: return '在售'
    case 2: return '已售'
    case 0: return '下架'
    default: return '未知'
  }
}

async function loadProducts() {
  loading.value = true
  try {
    const result = await productAPI.listMine({
      status: currentStatus.value,
      page: currentPage.value,
      size: pageSize.value,
    })
    products.value = result.items
    total.value = result.total
  } catch {
    products.value = []
  } finally {
    loading.value = false
  }
}

async function loadCounts() {
  try {
    const data = await productAPI.getCounts()
    counts.value = data as { active: number; sold: number; inactive: number }
  } catch {
    // 静默失败
  }
}

function changeStatus(status: number | undefined) {
  currentStatus.value = status
  currentPage.value = 1
  loadProducts()
}

function changePage(page: number) {
  currentPage.value = page
  loadProducts()
}

async function handleDeactivate(id: number) {
  try {
    await productAPI.updateStatus(id, 0)
    message.success('已下架')
    loadProducts()
    loadCounts()
  } catch {
    message.error('操作失败')
  }
}

async function handleActivate(id: number) {
  try {
    await productAPI.updateStatus(id, 1)
    message.success('已重新上架')
    loadProducts()
    loadCounts()
  } catch {
    message.error('操作失败')
  }
}

onMounted(() => {
  loadProducts()
  loadCounts()
})
</script>
