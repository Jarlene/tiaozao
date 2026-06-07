<template>
  <n-card title="我的消息" style="max-width: 800px; margin: 0 auto">
    <n-spin :show="loading">
      <n-list v-if="chatStore.conversations.length > 0">
        <n-list-item
          v-for="conv in chatStore.conversations"
          :key="conv.id"
          clickable
          @click="openConversation(conv)"
        >
          <template #prefix>
            <n-badge :value="getUnreadCount(conv)" :max="99">
              <n-avatar circle size="medium">
                {{ getOtherUserName(conv).charAt(0) }}
              </n-avatar>
            </n-badge>
          </template>
          <n-space vertical :size="4" style="flex: 1">
            <n-space align="center" justify="space-between">
              <n-text strong>{{ getOtherUserName(conv) }}</n-text>
              <n-text depth="3" style="font-size: 12px">{{ formatTime(conv.updated_at) }}</n-text>
            </n-space>
            <n-text depth="3" style="font-size: 13px" ellipsis>{{ conv.last_msg || '暂无消息' }}</n-text>
          </n-space>
        </n-list-item>
      </n-list>
      <n-empty v-else description="暂无会话" />
    </n-spin>
  </n-card>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useChatStore } from '@/stores/chat'
import { useAuthStore } from '@/stores/auth'
import type { Conversation } from '@/api/chat'

const router = useRouter()
const chatStore = useChatStore()
const authStore = useAuthStore()

const loading = chatStore.loading

function getUnreadCount(conv: Conversation): number {
  if (!authStore.user) return 0
  return conv.buyer_id === authStore.user.id ? conv.unread_buyer : conv.unread_seller
}

function getOtherUserName(conv: Conversation): string {
  return conv.seller_id === authStore.user?.id ? '买家' : '卖家'
}

function formatTime(timeStr: string): string {
  if (!timeStr) return ''
  const date = new Date(timeStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffDays = Math.floor(diffMs / 86400000)

  if (diffDays === 0) {
    return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  }
  if (diffDays === 1) return '昨天'
  if (diffDays < 7) return `${diffDays}天前`
  return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
}

function openConversation(conv: Conversation) {
  router.push(`/chat/${conv.id}`)
}

onMounted(() => {
  chatStore.fetchConversations()
})
</script>
