<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { Product } from '@/api'
import { productApi } from '@/api'
import { useUserStore } from '@/stores/user'
import { useCartStore } from '@/stores/cart'
import { ElMessage } from 'element-plus'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const cartStore = useCartStore()

const product = ref<Product | null>(null)
const loading = ref(false)
const error = ref('')
const quantity = ref(1)
const currentImageIndex = ref(0)

// 是否收藏
const isFavorited = ref(false)

/** 商品 ID */
const productId = computed(() => route.params.id as string)

/** 是否为商品主人 */
const isOwner = computed(() =>
  userStore.isLoggedIn && product.value && userStore.profile?.id === product.value.seller.id
)

onMounted(async () => {
  await loadProduct()
})

/** 加载商品详情 */
async function loadProduct() {
  loading.value = true
  error.value = ''
  try {
    product.value = await productApi.getById(productId.value)
  } catch (err: any) {
    error.value = err?.response?.data?.message || '加载商品详情失败'
  } finally {
    loading.value = false
  }
}

/** 添加到购物车 */
async function addToCart() {
  if (!userStore.isLoggedIn) {
    ElMessage.warning('请先登录')
    router.push({ path: '/login', query: { redirect: route.fullPath } })
    return
  }
  try {
    await cartStore.addItem(productId.value, quantity.value)
    ElMessage.success('已添加到购物车')
  } catch {
    ElMessage.error('添加失败，请稍后重试')
  }
}

/** 切换收藏状态 */
function toggleFavorite() {
  if (!userStore.isLoggedIn) {
    ElMessage.warning('请先登录')
    router.push({ path: '/login', query: { redirect: route.fullPath } })
    return
  }
  isFavorited.value = !isFavorited.value
  ElMessage.success(isFavorited.value ? '已收藏' : '已取消收藏')
}

/** 立即购买 */
function buyNow() {
  if (!userStore.isLoggedIn) {
    ElMessage.warning('请先登录')
    router.push({ path: '/login', query: { redirect: route.fullPath } })
    return
  }
  // 先加入购物车再跳转到结算
  addToCart().then(() => {
    router.push('/cart')
  })
}

/** 格式化价格 */
function formatPrice(price: number) {
  return `¥${price.toFixed(2)}`
}

/** 获取成色文本 */
function getConditionLabel(condition: string) {
  const map: Record<string, string> = {
    new: '全新',
    like_new: '几乎全新',
    good: '良好',
    fair: '一般',
    poor: '较差',
  }
  return map[condition] || condition
}

/** 格式化时间 */
function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}
</script>

<template>
  <div class="product-detail-page">
    <!-- 加载状态 -->
    <div v-if="loading" class="loading-state">
      <el-skeleton :rows="6" animated />
    </div>

    <!-- 错误状态 -->
    <div v-else-if="error" class="error-state">
      <el-result icon="error" title="加载失败" :sub-title="error">
        <template #extra>
          <el-button type="primary" @click="loadProduct">重新加载</el-button>
          <el-button @click="router.push('/products')">返回列表</el-button>
        </template>
      </el-result>
    </div>

    <!-- 商品详情 -->
    <template v-else-if="product">
      <div class="breadcrumb">
        <el-breadcrumb>
          <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
          <el-breadcrumb-item :to="{ path: '/products' }">商品列表</el-breadcrumb-item>
          <el-breadcrumb-item>{{ product.title }}</el-breadcrumb-item>
        </el-breadcrumb>
      </div>

      <div class="detail-content">
        <!-- 左侧：图片展示 -->
        <div class="image-section">
          <div class="main-image">
            <el-image
              :src="product.images[currentImageIndex] || '/placeholder.svg'"
              fit="contain"
              class="product-main-img"
            />
          </div>
          <div v-if="product.images.length > 1" class="image-thumbnails">
            <div
              v-for="(img, index) in product.images"
              :key="index"
              class="thumbnail"
              :class="{ active: currentImageIndex === index }"
              @click="currentImageIndex = index"
            >
              <el-image :src="img" fit="cover" class="thumbnail-img" />
            </div>
          </div>
        </div>

        <!-- 右侧：商品信息 -->
        <div class="info-section">
          <div class="product-header">
            <h1 class="product-title">{{ product.title }}</h1>
            <el-tag v-if="product.status === 'sold'" type="danger" size="large">已售出</el-tag>
          </div>

          <div class="price-section">
            <span class="current-price">{{ formatPrice(product.price) }}</span>
            <span v-if="product.originalPrice" class="original-price">
              原价 {{ formatPrice(product.originalPrice) }}
            </span>
          </div>

          <div class="meta-section">
            <div class="meta-item">
              <span class="meta-label">商品成色</span>
              <el-tag :type="product.condition === 'new' ? 'success' : 'warning'">
                {{ getConditionLabel(product.condition) }}
              </el-tag>
            </div>
            <div class="meta-item">
              <span class="meta-label">商品分类</span>
              <span>{{ product.category }}</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">发布时间</span>
              <span>{{ formatDate(product.createdAt) }}</span>
            </div>
          </div>

          <div class="description-section">
            <h3 class="section-label">商品描述</h3>
            <p class="description-text">{{ product.description || '暂无描述' }}</p>
          </div>

          <!-- 卖家信息 -->
          <div class="seller-section">
            <h3 class="section-label">卖家信息</h3>
            <div class="seller-info">
              <el-avatar :size="40">{{ product.seller.username[0] }}</el-avatar>
              <div class="seller-detail">
                <span class="seller-name">{{ product.seller.username }}</span>
              </div>
            </div>
          </div>

          <!-- 操作按钮 -->
          <div v-if="!isOwner && product.status !== 'sold'" class="action-section">
            <div class="quantity-selector">
              <span class="quantity-label">数量：</span>
              <el-input-number
                v-model="quantity"
                :min="1"
                :max="99"
                size="small"
              />
            </div>
            <div class="action-buttons">
              <el-button type="danger" size="large" @click="buyNow">
                <el-icon><Lightning /></el-icon>
                立即购买
              </el-button>
              <el-button type="primary" size="large" @click="addToCart">
                <el-icon><ShoppingCart /></el-icon>
                加入购物车
              </el-button>
              <el-button
                size="large"
                :type="isFavorited ? 'danger' : 'default'"
                @click="toggleFavorite"
              >
                <el-icon><Star /></el-icon>
                {{ isFavorited ? '已收藏' : '收藏' }}
              </el-button>
            </div>
          </div>

          <!-- 自己商品的提示 -->
          <div v-if="isOwner" class="owner-notice">
            <el-alert title="这是您发布的商品" type="info" :closable="false" show-icon />
          </div>
        </div>
      </div>

      <!-- 评论/评价区域 -->
      <div class="reviews-section">
        <h3 class="section-label">商品评价</h3>
        <el-empty description="暂无评价" />
      </div>
    </template>
  </div>
</template>

<style scoped>
.product-detail-page {
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

.detail-content {
  display: flex;
  gap: 32px;
  background: #fff;
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 24px;
}

/* 图片区域 */
.image-section {
  flex: 0 0 480px;
}

.main-image {
  width: 100%;
  height: 400px;
  border-radius: 8px;
  overflow: hidden;
  background: #f5f5f5;
  margin-bottom: 12px;
}

.product-main-img {
  width: 100%;
  height: 100%;
}

.image-thumbnails {
  display: flex;
  gap: 8px;
}

.thumbnail {
  width: 72px;
  height: 72px;
  border-radius: 6px;
  overflow: hidden;
  cursor: pointer;
  border: 2px solid transparent;
  transition: border-color 0.2s;
}

.thumbnail.active {
  border-color: #409eff;
}

.thumbnail-img {
  width: 100%;
  height: 100%;
}

/* 信息区域 */
.info-section {
  flex: 1;
  min-width: 0;
}

.product-header {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 16px;
}

.product-title {
  font-size: 22px;
  font-weight: 600;
  margin: 0;
  color: #333;
  flex: 1;
}

.price-section {
  margin-bottom: 24px;
  padding: 16px;
  background: #fdf6ec;
  border-radius: 8px;
}

.current-price {
  font-size: 28px;
  font-weight: bold;
  color: #f56c6c;
  margin-right: 12px;
}

.original-price {
  font-size: 15px;
  color: #999;
  text-decoration: line-through;
}

.meta-section {
  margin-bottom: 24px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 0;
  border-bottom: 1px solid #f5f5f5;
}

.meta-label {
  font-size: 14px;
  color: #999;
  min-width: 80px;
}

.description-section {
  margin-bottom: 24px;
}

.section-label {
  font-size: 16px;
  font-weight: 600;
  margin: 0 0 12px 0;
  color: #333;
}

.description-text {
  font-size: 14px;
  color: #666;
  line-height: 1.6;
}

.seller-section {
  margin-bottom: 24px;
}

.seller-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.seller-name {
  font-size: 15px;
  font-weight: 500;
  color: #333;
}

.action-section {
  border-top: 1px solid #ebeef5;
  padding-top: 20px;
}

.quantity-selector {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
}

.quantity-label {
  font-size: 14px;
  color: #666;
}

.action-buttons {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.owner-notice {
  margin-top: 16px;
}

/* 评价区域 */
.reviews-section {
  background: #fff;
  border-radius: 12px;
  padding: 24px;
}
</style>
