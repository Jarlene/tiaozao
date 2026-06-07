<template>
  <n-space vertical>
    <!-- Hero 区域 -->
    <n-card :bordered="false" style="text-align: center; padding: 40px 0">
      <n-h2 style="margin-bottom: 8px">欢迎来到跳蚤市场</n-h2>
      <n-p depth="3">发现好物，二手也有价值</n-p>
      <n-space justify="center" style="margin-top: 16px">
        <n-button type="primary" size="large" @click="$router.push('/products')">
          浏览全部商品
        </n-button>
        <n-button v-if="authStore.isAuthenticated" size="large" @click="$router.push('/products/create')">
          发布商品
        </n-button>
        <n-button v-else size="large" @click="$router.push('/register')">
          立即注册
        </n-button>
      </n-space>
    </n-card>

    <!-- 分类导航 -->
    <n-h2>商品分类</n-h2>
    <n-spin :show="loading">
      <n-grid :cols="4" :x-gap="16" :y-gap="16" v-if="categories.length > 0">
        <n-grid-item v-for="item in categories" :key="item.id">
          <n-card
            hoverable
            size="small"
            :title="item.name"
            @click="goToCategory(item)"
            style="cursor: pointer"
          >
            <n-space v-if="item.children && item.children.length > 0" :size="6" wrap>
              <n-tag
                v-for="child in item.children"
                :key="child.id"
                size="small"
                :bordered="false"
                style="cursor: pointer"
                @click.stop="goToCategory(child)"
              >
                {{ child.name }}
              </n-tag>
            </n-space>
            <n-text v-else depth="3" style="font-size: 13px">点击查看该分类下的商品</n-text>
          </n-card>
        </n-grid-item>
      </n-grid>
      <n-empty v-else-if="!loading" description="暂无分类" style="padding: 40px 0" />
    </n-spin>
  </n-space>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { categoryAPI } from '@/api/products'
import type { CategoryTreeItem } from '@/api/products'

const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)
const categories = ref<CategoryTreeItem[]>([])

function goToCategory(item: CategoryTreeItem) {
  router.push({ name: 'ProductList', query: { category_id: item.id } })
}

onMounted(async () => {
  loading.value = true
  try {
    categories.value = await categoryAPI.getTree()
  } catch {
    // 分类加载失败不影响首页
  } finally {
    loading.value = false
  }
})
</script>
