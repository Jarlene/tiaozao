<template>
  <n-space vertical>
    <n-h2>商品列表</n-h2>

    <!-- 搜索栏 -->
    <n-card :bordered="false" size="small">
      <n-space>
        <n-input
          v-model:value="searchForm.keyword"
          placeholder="搜索商品..."
          clearable
          style="width: 300px"
          @keyup.enter="handleSearch"
        />
        <n-select
          v-if="categories.length > 0"
          v-model:value="searchForm.category_id"
          :options="categoryOptions"
          placeholder="选择分类"
          clearable
          style="width: 200px"
        />
        <n-input-number
          v-model:value="searchForm.price_min"
          placeholder="最低价"
          :min="0"
          clearable
          style="width: 120px"
        />
        <n-input-number
          v-model:value="searchForm.price_max"
          placeholder="最高价"
          :min="0"
          clearable
          style="width: 120px"
        />
        <n-button type="primary" @click="handleSearch">搜索</n-button>
        <n-button @click="handleReset">重置</n-button>
      </n-space>
    </n-card>

    <!-- 商品网格 -->
    <n-spin :show="loading">
      <n-row :gutter="[16, 16]" v-if="products.length > 0">
        <n-col :span="6" v-for="product in products" :key="product.id">
          <n-card
            hoverable
            @click="$router.push(`/products/${product.id}`)"
            :title="product.title"
            size="small"
          >
            <template #cover>
              <n-image
                v-if="product.images && product.images.length > 0"
                :src="product.images[0].url"
                :alt="product.title"
                style="height: 200px; object-fit: cover; width: 100%"
                fallback-src="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 200 200'><rect fill='%23e0e0e0' width='200' height='200'/><text x='50%25' y='50%25' fill='%23999' text-anchor='middle' dy='.3em'>暂无图片</text></svg>"
              />
              <div v-else style="height: 200px; background: #f0f0f0; display: flex; align-items: center; justify-content: center; color: #999;">
                暂无图片
              </div>
            </template>
            <n-space vertical :size="4">
              <n-text strong depth="primary" style="font-size: 18px; color: #f5222d">
                ¥{{ (product.price / 100).toFixed(2) }}
              </n-text>
              <n-text depth="3" style="font-size: 12px">
                {{ product.created_at }}
              </n-text>
            </n-space>
          </n-card>
        </n-col>
      </n-row>

      <n-empty v-else-if="!loading" description="暂无商品" style="padding: 60px 0" />

      <!-- 分页 -->
      <n-space justify="center" style="margin-top: 24px" v-if="totalPages > 1">
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
import { onMounted, reactive, ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { productAPI, categoryAPI } from '@/api/products'
import type { ProductListItem, CategoryTreeItem } from '@/api/products'
import { flattenCategories } from '@/utils/categories'

const route = useRoute()

const loading = ref(false)
const products = ref<ProductListItem[]>([])
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

const categories = ref<CategoryTreeItem[]>([])
const categoryOptions = ref<Array<{ label: string; value: number }>>([])

const searchForm = reactive({
  keyword: '',
  category_id: null as number | null,
  price_min: null as number | null,
  price_max: null as number | null,
})

async function fetchCategories() {
  try {
    const data = await categoryAPI.getTree()
    categories.value = data
    categoryOptions.value = flattenCategories(data)
  } catch {
    // 分类加载失败不影响商品列表
  }
}

async function loadProducts() {
  loading.value = true
  try {
    const hasSearchFilters = searchForm.keyword || searchForm.category_id || searchForm.price_min || searchForm.price_max
    if (hasSearchFilters) {
      const result = await productAPI.search({
        keyword: searchForm.keyword || undefined,
        category_id: searchForm.category_id || undefined,
        price_min: searchForm.price_min || undefined,
        price_max: searchForm.price_max || undefined,
        page: currentPage.value,
        size: pageSize.value,
      })
      products.value = result.items
      total.value = result.total
    } else {
      const result = await productAPI.list(currentPage.value, pageSize.value)
      products.value = result.items
      total.value = result.total
    }
  } catch {
    products.value = []
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  currentPage.value = 1
  loadProducts()
}

function handleReset() {
  searchForm.keyword = ''
  searchForm.category_id = null
  searchForm.price_min = null
  searchForm.price_max = null
  currentPage.value = 1
  loadProducts()
}

function changePage(page: number) {
  currentPage.value = page
  loadProducts()
}

onMounted(() => {
  // 读取 URL 查询参数中的分类 ID（从分类导航跳转过来）
  if (route.query.category_id) {
    searchForm.category_id = Number(route.query.category_id)
  }
  fetchCategories()
  loadProducts()
})
</script>
