<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import type { Product, ProductQuery } from '@/api'
import { productApi } from '@/api'
import { ElMessage } from 'element-plus'

const router = useRouter()
const route = useRoute()

// 商品列表数据
const products = ref<Product[]>([])
const loading = ref(false)
const error = ref('')
const total = ref(0)

// 查询参数
const queryParams = reactive<ProductQuery>({
  page: 1,
  pageSize: 20,
  search: '',
  category: '',
  condition: '',
  minPrice: undefined,
  maxPrice: undefined,
  sortBy: 'createdAt',
  sortOrder: 'desc',
})

// 搜索关键字
const searchKeyword = ref('')

// 展开高级筛选
const showAdvancedFilters = ref(false)

/** 商品成色选项 */
const conditionOptions = [
  { value: '', label: '全部' },
  { value: 'new', label: '新品' },
  { value: 'like_new', label: '几乎全新' },
  { value: 'good', label: '良好' },
  { value: 'fair', label: '一般' },
  { value: 'poor', label: '较差' },
]

/** 排序选项 */
const sortOptions = [
  { value: 'createdAt', label: '最新发布' },
  { value: 'price', label: '价格' },
]

/** 监听路由参数变化 */
watch(
  () => route.query,
  (query) => {
    if (query.search) {
      queryParams.search = query.search as string
      searchKeyword.value = query.search as string
    }
    if (query.category) {
      queryParams.category = query.category as string
    }
    loadProducts()
  },
  { immediate: true }
)

/** 加载商品列表 */
async function loadProducts() {
  loading.value = true
  error.value = ''
  try {
    const res = await productApi.getList(queryParams)
    products.value = res.items
    total.value = res.total
  } catch (err: any) {
    error.value = err?.response?.data?.message || '加载商品列表失败'
    products.value = []
  } finally {
    loading.value = false
  }
}

/** 搜索 */
function handleSearch() {
  queryParams.search = searchKeyword.value
  queryParams.page = 1
  loadProducts()
}

/** 重置筛选条件 */
function resetFilters() {
  queryParams.category = ''
  queryParams.condition = ''
  queryParams.minPrice = undefined
  queryParams.maxPrice = undefined
  queryParams.sortBy = 'createdAt'
  queryParams.sortOrder = 'desc'
  queryParams.search = ''
  searchKeyword.value = ''
  queryParams.page = 1
  loadProducts()
}

/** 分页变化 */
function handlePageChange(page: number) {
  queryParams.page = page
  loadProducts()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

/** 排序变化 */
function handleSortChange(sortBy: string) {
  queryParams.sortBy = sortBy
  queryParams.page = 1
  loadProducts()
}

/** 跳转到商品详情 */
function goToProduct(id: string) {
  router.push(`/products/${id}`)
}

/** 格式化价格 */
function formatPrice(price: number) {
  return `¥${price.toFixed(2)}`
}

/** 获取成色标签内容 */
function getConditionLabel(condition: string) {
  const map: Record<string, string> = {
    new: '新品',
    like_new: '几乎全新',
    good: '良好',
    fair: '一般',
    poor: '较差',
  }
  return map[condition] || condition
}

/** 获取成色标签类型 */
function getConditionType(condition: string) {
  const map: Record<string, string> = {
    new: 'success',
    like_new: 'info',
    good: 'warning',
    fair: 'danger',
    poor: 'danger',
  }
  return map[condition] || 'info'
}
</script>

<template>
  <div class="product-list-page">
    <h2 class="page-title">商品列表</h2>

    <!-- 搜索和筛选栏 -->
    <div class="filter-bar">
      <div class="search-row">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索商品名称、描述..."
          clearable
          class="search-input"
          @keyup.enter="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-button type="primary" @click="handleSearch">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
        <el-button @click="showAdvancedFilters = !showAdvancedFilters">
          <el-icon><Filter /></el-icon>
          高级筛选
        </el-button>
        <el-button text @click="resetFilters">重置</el-button>
      </div>

      <!-- 高级筛选 -->
      <el-collapse-transition>
        <div v-show="showAdvancedFilters" class="advanced-filters">
          <el-form :inline="true" label-width="80px">
            <el-form-item label="分类">
              <el-input v-model="queryParams.category" placeholder="输入分类名称" clearable />
            </el-form-item>

            <el-form-item label="成色">
              <el-select v-model="queryParams.condition" placeholder="选择成色" clearable style="width: 140px">
                <el-option
                  v-for="opt in conditionOptions"
                  :key="opt.value"
                  :label="opt.label"
                  :value="opt.value"
                />
              </el-select>
            </el-form-item>

            <el-form-item label="价格区间">
              <el-input-number
                v-model="queryParams.minPrice"
                :min="0"
                :precision="2"
                placeholder="最低价"
                style="width: 130px"
                controls-position="right"
              />
              <span class="price-separator">-</span>
              <el-input-number
                v-model="queryParams.maxPrice"
                :min="0"
                :precision="2"
                placeholder="最高价"
                style="width: 130px"
                controls-position="right"
              />
            </el-form-item>
          </el-form>
        </div>
      </el-collapse-transition>

      <!-- 排序和结果统计 -->
      <div class="sort-bar">
        <div class="sort-options">
          <span class="sort-label">排序：</span>
          <el-radio-group
            :model-value="queryParams.sortBy"
            @change="handleSortChange"
          >
            <el-radio-button
              v-for="opt in sortOptions"
              :key="opt.value"
              :value="opt.value"
            >
              {{ opt.label }}
            </el-radio-button>
          </el-radio-group>
        </div>
        <span class="total-count">共 {{ total }} 件商品</span>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-state">
      <el-skeleton :rows="5" animated />
    </div>

    <!-- 错误状态 -->
    <div v-else-if="error" class="error-state">
      <el-result icon="error" title="加载失败" :sub-title="error">
        <template #extra>
          <el-button type="primary" @click="loadProducts">重新加载</el-button>
        </template>
      </el-result>
    </div>

    <!-- 商品网格 -->
    <template v-else>
      <div v-if="products.length > 0" class="product-grid">
        <el-card
          v-for="product in products"
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
              :type="getConditionType(product.condition)"
              size="small"
            >
              {{ getConditionLabel(product.condition) }}
            </el-tag>
          </div>
          <div class="product-info">
            <h4 class="product-title">{{ product.title }}</h4>
            <p class="product-desc">{{ product.description }}</p>
            <div class="product-price">
              <span class="current-price">{{ formatPrice(product.price) }}</span>
              <span v-if="product.originalPrice" class="original-price">
                {{ formatPrice(product.originalPrice) }}
              </span>
            </div>
            <div class="product-meta">
              <span>
                <el-icon><User /></el-icon>
                {{ product.seller.username }}
              </span>
              <span>
                <el-icon><Clock /></el-icon>
                {{ new Date(product.createdAt).toLocaleDateString() }}
              </span>
            </div>
          </div>
        </el-card>
      </div>

      <!-- 空状态 -->
      <div v-else class="empty-state">
        <el-empty description="没有找到相关商品">
          <el-button type="primary" @click="resetFilters">清除筛选条件</el-button>
        </el-empty>
      </div>
    </template>

    <!-- 分页 -->
    <div v-if="total > queryParams.pageSize!" class="pagination-wrapper">
      <el-pagination
        v-model:current-page="queryParams.page"
        :page-size="queryParams.pageSize"
        :total="total"
        layout="prev, pager, next, jumper"
        background
        @current-change="handlePageChange"
      />
    </div>
  </div>
</template>

<style scoped>
.product-list-page {
  padding-bottom: 40px;
}

.page-title {
  font-size: 22px;
  font-weight: 600;
  margin: 0 0 20px 0;
  color: #333;
}

.filter-bar {
  background: #fff;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 20px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
}

.search-row {
  display: flex;
  gap: 12px;
  align-items: center;
}

.search-input {
  flex: 1;
  max-width: 400px;
}

.advanced-filters {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #ebeef5;
}

.price-separator {
  margin: 0 8px;
  color: #999;
}

.sort-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #ebeef5;
}

.sort-options {
  display: flex;
  align-items: center;
  gap: 8px;
}

.sort-label {
  font-size: 14px;
  color: #666;
}

.total-count {
  font-size: 13px;
  color: #999;
}

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
  background: #f5f5f5;
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
  margin: 0 0 6px 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #333;
}

.product-desc {
  font-size: 13px;
  color: #999;
  margin: 0 0 8px 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
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

.loading-state {
  background: #fff;
  border-radius: 8px;
  padding: 40px;
}

.error-state {
  padding: 40px;
}

.empty-state {
  padding: 60px;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 32px;
}
</style>
