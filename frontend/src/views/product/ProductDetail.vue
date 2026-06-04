<template>
  <n-spin :show="loading">
    <n-card v-if="product" style="max-width: 900px; margin: 0 auto">
      <template #header>
        <n-space align="center" justify="space-between">
          <n-h2 style="margin: 0">{{ product.title }}</n-h2>
          <n-space v-if="isOwner">
            <n-button size="small" @click="$router.push(`/products/${product.id}/edit`)">
              编辑
            </n-button>
            <n-popconfirm @positive-click="handleDelete">
              <template #trigger>
                <n-button size="small" type="warning">下架</n-button>
              </template>
              确认下架该商品？
            </n-popconfirm>
          </n-space>
        </n-space>
      </template>

      <!-- 图片轮播 -->
      <n-carousel v-if="product.images && product.images.length > 0" style="height: 400px" autoplay>
        <n-carousel-item v-for="img in product.images" :key="img.id">
          <n-image
            :src="img.url"
            style="width: 100%; height: 400px; object-fit: contain"
            preview
          />
        </n-carousel-item>
      </n-carousel>
      <n-empty v-else description="暂无图片" style="height: 200px" />

      <!-- 商品信息 -->
      <n-descriptions label-placement="left" bordered style="margin-top: 24px">
        <n-descriptions-item label="价格">
          <n-text strong style="font-size: 24px; color: #f5222d">
            ¥{{ (product.price / 100).toFixed(2) }}
          </n-text>
        </n-descriptions-item>
        <n-descriptions-item label="状态">
          <n-tag :type="statusType" size="small">{{ statusLabel }}</n-tag>
        </n-descriptions-item>
        <n-descriptions-item label="发布时间">
          {{ product.created_at }}
        </n-descriptions-item>
        <n-descriptions-item label="描述" :span="2">
          <n-text>{{ product.description || '暂无描述' }}</n-text>
        </n-descriptions-item>
      </n-descriptions>

      <!-- 卖家信息 -->
      <n-card title="卖家信息" size="small" style="margin-top: 24px" v-if="product.user">
        <n-space align="center">
          <n-avatar
            :src="product.user.avatar_url"
            fallback-src="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 32 32'><circle fill='%2318a058' r='16' cx='16' cy='16'/><text x='50%25' y='55%25' fill='white' text-anchor='middle' dy='.1em' font-size='14'>U</text></svg>"
            circle
          />
          <n-space vertical :size="2">
            <n-text strong>{{ product.user.nickname }}</n-text>
            <n-text depth="3" style="font-size: 12px">
              注册时间：{{ product.user.created_at }}
            </n-text>
          </n-space>
        </n-space>
      </n-card>

      <template #footer>
        <n-space justify="center">
          <n-button @click="$router.push('/products')">返回列表</n-button>
        </n-space>
      </template>
    </n-card>
  </n-spin>
</template>

<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { productAPI } from '@/api/products'
import { useAuthStore } from '@/stores/auth'
import type { ProductDetail } from '@/api/products'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const authStore = useAuthStore()

const loading = ref(true)
const product = ref<ProductDetail | null>(null)

const isOwner = computed(() => {
  if (!product.value || !authStore.user) return false
  return product.value.user_id === authStore.user.id
})

const statusType = computed(() => {
  if (!product.value) return 'default'
  switch (product.value.status) {
    case 1: return 'success'
    case 2: return 'warning'
    case 0: return 'default'
    default: return 'default'
  }
})

const statusLabel = computed(() => {
  if (!product.value) return ''
  switch (product.value.status) {
    case 1: return '在售'
    case 2: return '已售'
    case 0: return '已下架'
    default: return '未知'
  }
})

async function loadProduct() {
  loading.value = true
  try {
    const id = Number(route.params.id)
    const data = await productAPI.getById(id)
    product.value = data
  } catch {
    message.error('加载商品信息失败')
    router.push('/products')
  } finally {
    loading.value = false
  }
}

async function handleDelete() {
  try {
    await productAPI.delete(product.value!.id)
    message.success('已下架')
    router.push('/products/mine')
  } catch {
    message.error('操作失败')
  }
}

onMounted(loadProduct)
</script>
