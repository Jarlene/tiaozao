<template>
  <div class="comment-item">
    <div class="comment-header">
      <div class="avatar-placeholder">{{ initial }}</div>
      <div class="comment-meta">
        <span class="nickname">{{ comment.user_nick || '匿名用户' }}</span>
        <span class="time">{{ formatTime(comment.created_at) }}</span>
      </div>
    </div>
    <div class="comment-body">{{ comment.content }}</div>
    <div class="comment-actions">
      <button class="btn-reply" @click="$emit('toggle-reply', comment.id)">
        {{ showReply ? '收起回复' : '回复' }}
      </button>
    </div>

    <!-- Reply form -->
    <ReplyForm
      v-if="showReply"
      :username="comment.user_nick || '匿名用户'"
      @submit="(content) => $emit('reply', comment.id, content)"
      @cancel="$emit('toggle-reply', comment.id)"
    />

    <!-- Nested replies -->
    <div v-if="comment.replies && comment.replies.length > 0" class="replies">
      <div v-for="reply in comment.replies" :key="reply.id" class="reply-item">
        <div class="reply-header">
          <div class="avatar-small-placeholder">{{ reply.user_nick?.[0] || '?' }}</div>
          <span class="reply-nickname">{{
            reply.user_nick || '匿名用户'
          }}</span>
          <span class="reply-time">{{ formatTime(reply.created_at) }}</span>
        </div>
        <div class="reply-body">{{ reply.content }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Comment } from '../types/comment'
import ReplyForm from './ReplyForm.vue'

const props = defineProps<{
  comment: Comment
  showReply: boolean
}>()

defineEmits<{
  'toggle-reply': [commentId: number]
  reply: [commentId: number, content: string]
}>()

const initial = computed(() => props.comment.user_nick?.[0] || '?')

function formatTime(iso: string): string {
  const d = new Date(iso)
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  const minutes = Math.floor(diff / 60000)
  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}小时前`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days}天前`
  return d.toLocaleDateString('zh-CN')
}
</script>

<style scoped>
.comment-item {
  padding: 16px 0;
  border-bottom: 1px solid #f0f0f0;
}
.comment-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}
.avatar-placeholder {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #1677ff;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  font-weight: 600;
  flex-shrink: 0;
}
.comment-meta {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.nickname {
  font-weight: 600;
  font-size: 14px;
  color: #333;
}
.time {
  font-size: 12px;
  color: #999;
}
.comment-body {
  font-size: 15px;
  line-height: 1.6;
  color: #333;
  margin-bottom: 8px;
}
.comment-actions {
  display: flex;
  gap: 16px;
}
.btn-reply {
  background: none;
  border: none;
  color: #1677ff;
  cursor: pointer;
  font-size: 13px;
  padding: 0;
}
.btn-reply:hover {
  color: #4096ff;
}
.replies {
  margin-top: 12px;
  padding: 12px 12px 12px 48px;
  background: #fafafa;
  border-radius: 8px;
}
.reply-item {
  padding: 8px 0;
}
.reply-item + .reply-item {
  border-top: 1px solid #f0f0f0;
}
.reply-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}
.avatar-small-placeholder {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #1677ff;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
  flex-shrink: 0;
}
.reply-nickname {
  font-weight: 600;
  font-size: 13px;
  color: #333;
}
.reply-time {
  font-size: 12px;
  color: #999;
}
.reply-body {
  font-size: 14px;
  line-height: 1.5;
  color: #333;
}
</style>
