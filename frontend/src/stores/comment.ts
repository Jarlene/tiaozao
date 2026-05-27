import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Comment, CommentListResult } from '../types/comment'
import * as api from '../api/comment'

export const useCommentStore = defineStore('comment', () => {
  const comments = ref<Comment[]>([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(10)
  const loading = ref(false)
  const submitting = ref(false)

  const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

  async function loadComments(productId: string, p?: number) {
    if (p !== undefined) page.value = p
    loading.value = true
    try {
      const result: CommentListResult = await api.fetchComments(
        productId,
        page.value,
        pageSize.value,
      )
      comments.value = result.comments
      total.value = result.total
    } finally {
      loading.value = false
    }
  }

  async function addComment(productId: string, content: string) {
    submitting.value = true
    try {
      await api.createComment(productId, { content })
      await loadComments(productId, 1)
    } finally {
      submitting.value = false
    }
  }

  async function addReply(
    productId: string,
    commentId: number,
    content: string,
  ) {
    submitting.value = true
    try {
      await api.replyToComment(productId, commentId, { content })
      await loadComments(productId, page.value)
    } finally {
      submitting.value = false
    }
  }

  return {
    comments,
    total,
    page,
    pageSize,
    totalPages,
    loading,
    submitting,
    loadComments,
    addComment,
    addReply,
  }
})
