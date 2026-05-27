<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import type { Product, Category } from '@/api'
import { productApi, categoryApi } from '@/api'

const router = useRouter()

// 特色商品列表
const featuredProducts = ref<Product[]>([])
// 分类列表
const categories = ref<Category[]>([])
// 加载状态
const loading = ref(false)
// 错误信息
const error = ref('')

/** 模拟分类数据（当 API 不可用时使用） */
const defaultCategories: Category[] = [
  { id: '1', name: '电子产品', icon: 'Monitor' },
  { id: '2', name: '家具家居', icon: 'HomeFilled' },
  { id: '3', name: '服装鞋帽', icon: 'Clothes' },
  { id: '4', name: '图书教材', icon: 'Reading' },
  { id: '5', name: '运动户外', icon: 'Basketball' },
  { id: '6', name: '乐器设备', icon: 'Headset' },
]

onMounted(async () => {
  await loadData()
})

/** 加载首页数据 */
async function loadData() {
  loading.value = true
  error.value = ''
  try {
    const [productsRes, categoriesData] = await Promise.all([
      productApi.getFeatured(8),
      categoryApi.getList().catch(() => defaultCategories),
    ])
    featuredProducts.value = productsRes.items
    categories.value = categoriesData
  } catch (err: any) {
    error.value = err?.response?.data?.message || '加载数据失败，请稍后重试'
    categories.value = defaultCategories
  } finally {
    loading.value = false
  }
}

/** 跳转到分类商品列表 */
function goToCategory(category: Category) {
  router.push({ path: '/products', query: { category: category.name } })
}

/** 跳转到商品详情 */
function goToProduct(id: string) {
  router.push(`/products/${id}`)
}

/** 获取商品成色标签类型 */
function getConditionTag(condition: string) {
  const map: Record<string, string> = {
    new: '新品',
    like_new: '几乎全新',
    good: '良好',
    fair: '一般',
    poor: '较差',
  }
  return map[condition] || condition
}

/** 格式化价格 */
function formatPrice(price: number) {
  return `¥${price.toFixed(2)}`
}
</script>

<template>
  <div class="home-page">
    <!-- Banner 轮播 -->
    <el-carousel height="360px" class="banner-carousel" :interval="5000" indicator-position="outside">
      <el-carousel-item v-for="i in 3" :key="i">
        <div class="banner-slide" :class="`banner-${i}`">
          <div class="banner-content">
            <h2 v-if="i === 1">发现好物，就在跳蚤市场</h2>
            <h2 v-else-if="i === 2">环保交易，让闲置焕发新生</h2>
            <h2 v-else>安全交易，诚信平台</h2>
            <p>优质二手商品，等你来淘</p>
            <el-button type="primary" size="large" @click="router.push('/products')">
              立即逛一逛
            </el-button>
          </div>
        </div>
      </el-carousel-item>
    </el-carousel>

    <!-- 分类导航 -->
    <section class="section categories-section">
      <h3 class="section-title">商品分类</h3>
      <div class="categories-grid">
        <div
          v-for="cat in categories"
          :key="cat.id"
          class="category-card"
          @click="goToCategory(cat)"
        >
          <div class="category-icon">
            <el-icon :size="36">
              <component :is="cat.icon || 'Goods'" />
            </el-icon>
          </div>
          <span class="category-name">{{ cat.name }}</span>
        </div>
      </div>
    </section>

    <!-- 特色商品 -->
    <section class="section featured-section">
      <div class="section-header">
        <h3 class="section-title">推荐商品</h3>
        <el-button text type="primary" @click="router.push('/products')">
          查看全部 <el-icon><ArrowRight /></el-icon>
        </el-button>
      </div>

      <!-- 加载状态 -->
      <div v-if="loading" class="loading-state">
        <el-skeleton :rows="3" animated />
      </div>

      <!-- 错误状态 -->
      <div v-else-if="error" class="error-state">
        <el-result icon="error" title="加载失败" :sub-title="error">
          <template #extra>
            <el-button type="primary" @click="loadData">重新加载</el-button>
          </template>
        </el-result>
      </div>

      <!-- 商品网格 -->
      <div v-else class="product-grid">
        <el-card
          v-for="product in featuredProducts"
          :key="product.id"
          class="product-card"
          shadow="hover"
          @click="goToProduct(product.id)"
        >
          <div class="product-image-wrapper">
            <img
              :src="product.images[0] || '/placeholder.svg'"
              :alt="product.title"
              class="product-image"
            />
            <el-tag
              class="condition-tag"
              :type="product.condition === 'new' ? 'success' : 'warning'"
              size="small"
            >
              {{ getConditionTag(product.condition) }}
            </el-tag>
          </div>
          <div class="product-info">
            <h4 class="product-title">{{ product.title }}</h4>
            <div class="product-price">
              <span class="current-price">{{ formatPrice(product.price) }}</span>
              <span v-if="product.originalPrice" class="original-price">
                {{ formatPrice(product.originalPrice) }}
              </span>
            </div>
            <div class="product-meta">
              <span class="seller-name">{{ product.seller.username }}</span>
              <span class="publish-time">{{ new Date(product.createdAt).toLocaleDateString() }}</span>
            </div>
          </div>
        </el-card>
      </div>

      <!-- 空状态 -->
      <div v-if="!loading && !error && featuredProducts.length === 0" class="empty-state">
        <el-empty description="暂无推荐商品" />
      </div>
    </section>
  </div>
</template>

<style scoped>
.home-page {
  padding-bottom: 40px;
}

/* Banner 区域 */
.banner-carousel {
  border-radius: 12px;
  overflow: hidden;
  margin-bottom: 32px;
}

.banner-slide {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #fff;
}

.banner-1 { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
.banner-2 { background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%); }
.banner-3 { background: linear-gradient(135deg, #fa709a 0%, #fee140 100%); }

.banner-content {
  text-align: center;
}

.banner-content h2 {
  font-size: 32px;
  margin-bottom: 12px;
}

.banner-content p {
  font-size: 16px;
  margin-bottom: 24px;
  opacity: 0.9;
}

/* 通用 section 样式 */
.section {
  margin-bottom: 40px;
  background: #fff;
  border-radius: 12px;
  padding: 24px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.section-title {
  font-size: 20px;
  font-weight: 600;
  margin: 0 0 16px 0;
  padding-left: 12px;
  border-left: 3px solid #409eff;
}

.section-header .section-title {
  margin-bottom: 0;
}

/* 分类导航 */
.categories-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
  gap: 16px;
}

.category-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
  background: #f8f9fa;
}

.category-card:hover {
  background: #ecf5ff;
  transform: translateY(-2px);
}

.category-icon {
  width: 60px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fff;
  border-radius: 50%;
  margin-bottom: 8px;
  color: #409eff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.15);
}

.category-name {
  font-size: 14px;
  color: #333;
}

/* 商品网格 */
.product-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 20px;
}

.product-card {
  cursor: pointer;
  transition: transform 0.2s;
  border-radius: 8px;
  overflow: hidden;
}

.product-card:hover {
  transform: translateY(-4px);
}

.product-image-wrapper {
  position: relative;
  width: 100%;
  height: 200px;
  overflow: hidden;
}

.product-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s;
}

.product-card:hover .product-image {
  transform: scale(1.05);
}

.condition-tag {
  position: absolute;
  top: 8px;
  right: 8px;
}

.product-info {
  padding: 12px;
}

.product-title {
  font-size: 15px;
  font-weight: 500;
  margin: 0 0 8px 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #333;
}

.product-price {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-bottom: 8px;
}

.current-price {
  font-size: 18px;
  font-weight: bold;
  color: #f56c6c;
}

.original-price {
  font-size: 13px;
  color: #999;
  text-decoration: line-through;
}

.product-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #999;
}

/* 状态 */
.loading-state {
  padding: 40px;
}

.error-state {
  padding: 20px;
}

.empty-state {
  padding: 40px;
}
</style>
