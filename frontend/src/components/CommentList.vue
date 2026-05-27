<template>
  <div class="comment-list">
    <h3 class="section-title">评论 ({{ store.total }})</h3>

    <CommentForm
      :submitting="store.submitting"
      @submit="handleAddComment"
    />

    <div v-if="store.loading" class="loading">加载中...</div>

    <div v-else-if="store.comments.length === 0" class="empty">
      暂无评论，快来发表第一条评论吧
    </div>

    <CommentItem
      v-for="comment in store.comments"
      :key="comment.id"
      :comment="comment"
      :show-reply="replyingTo === comment.id"
      @toggle-reply="toggleReply(comment.id)"
      @reply="handleReply"
    />

    <!-- Pagination -->
    <div v-if="store.totalPages > 1" class="pagination">
      <button
        :disabled="store.page <= 1"
        class="btn-page"
        @click="changePage(store.page - 1)"
      >
        上一页
      </button>
      <span class="page-info">{{ store.page }} / {{ store.totalPages }}</span>
      <button
        :disabled="store.page >= store.totalPages"
        class="btn-page"
        @click="changePage(store.page + 1)"
      >
        下一页
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useCommentStore } from '../stores/comment'
import CommentForm from './CommentForm.vue'
import CommentItem from './CommentItem.vue'

const props = defineProps<{
  productId: string
}>()

const store = useCommentStore()
const replyingTo = ref<number | null>(null)

onMounted(() => {
  if (props.productId) {
    store.loadComments(props.productId)
  }
})

function toggleReply(commentId: number) {
  replyingTo.value = replyingTo.value === commentId ? null : commentId
}

function handleAddComment(content: string) {
  store.addComment(props.productId, content)
}

function handleReply(commentId: number, content: string) {
  store.addReply(props.productId, commentId, content)
  replyingTo.value = null
}

function changePage(page: number) {
  store.loadComments(props.productId, page)
}
</script>

<style scoped>
.comment-list {
  max-width: 800px;
  margin: 0 auto;
  padding: 24px;
}
.section-title {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 20px;
  color: #333;
}
.loading,
.empty {
  text-align: center;
  padding: 40px;
  color: #999;
  font-size: 14px;
}
.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 16px;
  margin-top: 24px;
}
.btn-page {
  padding: 6px 16px;
  background: #fff;
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  color: #333;
}
.btn-page:hover:not(:disabled) {
  border-color: #1677ff;
  color: #1677ff;
}
.btn-page:disabled {
  color: #d9d9d9;
  cursor: not-allowed;
}
.page-info {
  font-size: 14px;
  color: #666;
}
</style>
