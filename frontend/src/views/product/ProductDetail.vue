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
            {{ formatPrice(product.price) }}
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

      <!-- 评分统计 -->
      <n-card title="商品评价" style="margin-top: 24px">
        <n-spin :show="statsLoading">
          <!-- 评分概览 -->
          <n-space align="center" size="large" style="margin-bottom: 20px" v-if="ratingStats">
            <div style="text-align: center; min-width: 100px">
              <n-text style="font-size: 48px; font-weight: bold; color: #f59e0b">
                {{ ratingStats.average.toFixed(1) }}
              </n-text>
              <n-text depth="3" style="display: block; font-size: 13px">
                {{ ratingStats.total }} 条评价
              </n-text>
            </div>

            <!-- 评分分布 -->
            <div style="flex: 1; min-width: 200px">
              <div v-for="star in 5" :key="star" style="display: flex; align-items: center; margin-bottom: 4px; gap: 8px">
                <n-text style="font-size: 13px; white-space: nowrap; width: 30px">{{ 6 - star }}星</n-text>
                <n-progress
                  :percentage="distributionPercent(6 - star)"
                  :height="10"
                  :rail-color="'#f0f0f0'"
                  color="#f59e0b"
                />
                <n-text depth="3" style="font-size: 12px; width: 30px; text-align: right">
                  {{ distributionCount(6 - star) }}
                </n-text>
              </div>
            </div>
          </n-space>

          <n-empty v-else-if="!statsLoading" description="暂无评价" />

          <!-- 发表评论（仅非卖家且已登录用户） -->
          <n-card
            v-if="authStore.isAuthenticated && !isOwner"
            size="small"
            style="margin-bottom: 16px"
            :title="hasReviewed ? '您已评价过该商品' : '发表评价'"
          >
            <template v-if="!hasReviewed">
              <n-space vertical>
                <div>
                  <n-text depth="3">评分：</n-text>
                  <StarRating v-model="newReview.rating" />
                </div>
                <n-input
                  v-model:value="newReview.content"
                  type="textarea"
                  placeholder="写下您的评价..."
                  :maxlength="1000"
                  show-count
                  :disabled="submitting"
                />
                <n-button
                  type="primary"
                  :loading="submitting"
                  :disabled="!canSubmit"
                  @click="handleSubmitReview"
                >
                  提交评价
                </n-button>
              </n-space>
            </template>
          </n-card>

          <!-- 评论列表 -->
          <n-list v-if="reviews.length > 0">
            <n-list-item v-for="review in reviews" :key="review.id">
              <template #prefix>
                <n-avatar
                  :src="review.user?.avatar_url"
                  fallback-src="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 32 32'><circle fill='%2318a058' r='16' cx='16' cy='16'/><text x='50%25' y='55%25' fill='white' text-anchor='middle' dy='.1em' font-size='14'>U</text></svg>"
                  circle
                  size="small"
                />
              </template>
              <n-space vertical :size="4">
                <n-space align="center" size="small">
                  <n-text strong style="font-size: 13px">{{ review.user?.nickname || '匿名用户' }}</n-text>
                  <n-rate :value="review.rating" :count="5" size="small" readonly />
                </n-space>
                <n-text>{{ review.content }}</n-text>
                <n-space align="center" size="small">
                  <n-text depth="3" style="font-size: 12px">{{ review.created_at }}</n-text>
                </n-space>

                <!-- 卖家回复 -->
                <n-card
                  v-if="review.reply_content"
                  size="small"
                  style="margin-top: 8px; background: #fafafa; margin-left: 0"
                >
                  <n-space vertical :size="4">
                    <n-text strong style="font-size: 12px; color: #18a058">卖家回复</n-text>
                    <n-text style="font-size: 13px">{{ review.reply_content }}</n-text>
                    <n-text v-if="review.replied_at" depth="3" style="font-size: 12px">
                      {{ review.replied_at }}
                    </n-text>
                  </n-space>
                </n-card>

                <!-- 卖家回复入口 -->
                <template v-if="isOwner && !review.reply_content && replyingTo !== review.id">
                  <n-button size="tiny" quaternary type="primary" @click="startReply(review.id)">
                    回复
                  </n-button>
                </template>
                <template v-if="isOwner && replyingTo === review.id">
                  <n-space vertical style="margin-top: 8px">
                    <n-input
                      v-model:value="replyContent"
                      type="textarea"
                      placeholder="输入回复内容..."
                      :maxlength="1000"
                      show-count
                      :disabled="submittingReply"
                    />
                    <n-space>
                      <n-button
                        size="small"
                        type="primary"
                        :loading="submittingReply"
                        :disabled="!replyContent.trim()"
                        @click="handleReply(review.id)"
                      >
                        提交回复
                      </n-button>
                      <n-button size="small" quaternary @click="cancelReply">
                        取消
                      </n-button>
                    </n-space>
                  </n-space>
                </template>
              </n-space>
            </n-list-item>
          </n-list>

          <!-- 分页 -->
          <n-space v-if="reviewTotalPages > 1" justify="center" style="margin-top: 16px">
            <n-pagination
              :page="reviewPage"
              :page-count="reviewTotalPages"
              :disabled="statsLoading"
              @update:page="handlePageChange"
            />
          </n-space>
        </n-spin>
      </n-card>

      <template #footer>
        <n-space justify="center">
          <n-button @click="$router.push('/products')">返回列表</n-button>
          <n-button
            v-if="authStore.isAuthenticated && !isOwner && product"
            type="primary"
            @click="contactSeller"
          >
            联系卖家
          </n-button>
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
import { reviewAPI } from '@/api/reviews'
import { formatPrice } from '@/utils/format'
import { useAuthStore } from '@/stores/auth'
import { useChatStore } from '@/stores/chat'
import StarRating from '@/components/StarRating.vue'
import type { ProductDetail } from '@/api/products'
import type { ReviewItem, RatingStats } from '@/api/reviews'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const authStore = useAuthStore()
const chatStore = useChatStore()

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

// 评论相关状态
const statsLoading = ref(true)
const ratingStats = ref<RatingStats | null>(null)
const reviews = ref<ReviewItem[]>([])
const reviewPage = ref(1)
const reviewTotalPages = ref(0)
const reviewPageSize = 10

// 发表评论
const newReview = ref({ rating: 0, content: '' })
const submitting = ref(false)
const hasReviewed = ref(false)

// 卖家回复
const replyingTo = ref<number | null>(null)
const replyContent = ref('')
const submittingReply = ref(false)

const canSubmit = computed(() => {
  return newReview.value.rating > 0 && newReview.value.content.trim().length > 0
})

function distributionPercent(star: number): number {
  if (!ratingStats.value || ratingStats.value.total === 0) return 0
  const count = distributionCount(star)
  return Math.round((count / ratingStats.value.total) * 100)
}

function distributionCount(star: number): number {
  if (!ratingStats.value) return 0
  const key = `dist_${star}` as keyof RatingStats
  return ratingStats.value[key] as number
}

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

async function loadReviews() {
  const productId = Number(route.params.id)
  statsLoading.value = true
  try {
    const [stats, reviewResult] = await Promise.all([
      reviewAPI.getStats(productId),
      reviewAPI.list(productId, reviewPage.value, reviewPageSize),
    ])
    ratingStats.value = stats
    reviews.value = reviewResult.items
    reviewTotalPages.value = reviewResult.total_pages
  } catch {
    message.error('加载评价失败')
  } finally {
    statsLoading.value = false
  }
}

async function checkHasReviewed() {
  if (!authStore.isAuthenticated || isOwner.value) return
  const productId = Number(route.params.id)
  try {
    const result = await reviewAPI.checkCanReview(productId)
    hasReviewed.value = !result.can_review
  } catch (e) {
    console.warn('检查评论资格失败，将继续显示表单', e)
  }
}

async function handleSubmitReview() {
  if (!canSubmit.value) return
  submitting.value = true
  try {
    await reviewAPI.create(Number(route.params.id), {
      rating: newReview.value.rating,
      content: newReview.value.content.trim(),
    })
    message.success('评价发表成功')
    newReview.value = { rating: 0, content: '' }
    hasReviewed.value = true
    reviewPage.value = 1
    await loadReviews()
  } catch (err: any) {
    const msg = err?.response?.data?.message || '评价发表失败'
    message.error(msg)
  } finally {
    submitting.value = false
  }
}

function startReply(reviewId: number) {
  replyingTo.value = reviewId
  replyContent.value = ''
}

function cancelReply() {
  replyingTo.value = null
  replyContent.value = ''
}

async function handleReply(reviewId: number) {
  if (!replyContent.value.trim()) return
  submittingReply.value = true
  try {
    await reviewAPI.reply(reviewId, { content: replyContent.value.trim() })
    message.success('回复成功')
    replyingTo.value = null
    replyContent.value = ''
    await loadReviews()
  } catch (err: any) {
    const msg = err?.response?.data?.message || '回复失败'
    message.error(msg)
  } finally {
    submittingReply.value = false
  }
}

function handlePageChange(page: number) {
  reviewPage.value = page
  loadReviews()
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

async function contactSeller() {
  if (!product.value) return
  try {
    const conv = await chatStore.createConversation(
      product.value.id,
      product.value.user_id
    )
    router.push(`/chat/${conv.id}`)
  } catch {
    message.error('创建会话失败')
  }
}

onMounted(async () => {
  await loadProduct()
  await Promise.all([loadReviews(), checkHasReviewed()])
})
</script>
