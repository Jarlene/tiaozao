<template>
  <n-card style="max-width: 800px; margin: 0 auto; height: calc(100vh - 140px); display: flex; flex-direction: column">
    <template #header>
      <n-space align="center">
        <n-button quaternary @click="goBack" size="small">
          ← 返回
        </n-button>
        <n-text strong>聊天</n-text>
        <n-tag v-if="isOnline" size="tiny" type="success">在线</n-tag>
        <n-tag v-else size="tiny" type="default">离线</n-tag>
      </n-space>
    </template>

    <!-- 消息列表 -->
    <div ref="messageListRef" class="message-list" style="flex: 1; overflow-y: auto; padding: 12px 0">
      <n-space vertical :size="12">
        <div v-for="msg in chatStore.messages" :key="msg.id" style="display: flex; justify-content: center">
          <!-- 系统消息 -->
          <n-tag v-if="msg.msg_type === 3" size="tiny">{{ msg.content }}</n-tag>

          <!-- 普通消息 -->
          <div v-else :style="getMessageStyle(msg)" class="message-bubble">
            <div :style="{ maxWidth: '70%', display: 'flex', flexDirection: isMine(msg) ? 'row-reverse' : 'row', alignItems: 'flex-end', gap: '8px' }">
              <div
                :style="{
                  background: isMine(msg) ? '#18a058' : '#f0f0f0',
                  color: isMine(msg) ? '#fff' : '#333',
                  padding: '8px 14px',
                  borderRadius: '16px',
                  borderBottomRightRadius: isMine(msg) ? '4px' : '16px',
                  borderBottomLeftRadius: isMine(msg) ? '16px' : '4px',
                  maxWidth: '400px',
                  wordBreak: 'break-word',
                }"
              >
                <n-text>{{ msg.content }}</n-text>
              </div>
              <div v-if="isMine(msg)" style="display: flex; flex-direction: column; align-items: flex-end">
                <n-text depth="3" style="font-size: 11px">{{ formatMessageTime(msg.created_at) }}</n-text>
                <n-text v-if="msg.status === 1" depth="3" style="font-size: 11px">已发送</n-text>
                <n-text v-else-if="msg.status === 2" depth="3" style="font-size: 11px">已送达</n-text>
                <n-text v-else-if="msg.status === 3" depth="3" style="font-size: 11px">已读</n-text>
              </div>
            </div>
          </div>
        </div>

        <!-- 正在输入 -->
        <div v-if="isTyping" style="text-align: left; padding-left: 16px">
          <n-text depth="3" style="font-size: 13px">对方正在输入...</n-text>
        </div>
      </n-space>
    </div>

    <!-- 输入区域 -->
    <n-space style="border-top: 1px solid #eee; padding-top: 12px" align="center">
      <n-input
        v-model:value="inputText"
        type="textarea"
        :rows="2"
        placeholder="输入消息..."
        :disabled="!chatStore.connected"
        @keydown.enter.prevent="handleSend"
        @input="handleTyping"
        style="flex: 1"
      />
      <n-button
        type="primary"
        :disabled="!inputText.trim() || !chatStore.connected"
        @click="handleSend"
        style="align-self: flex-end"
      >
        发送
      </n-button>
    </n-space>
  </n-card>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref, computed, nextTick, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useChatStore } from '@/stores/chat'
import { useAuthStore } from '@/stores/auth'
import type { Message } from '@/api/chat'

const route = useRoute()
const router = useRouter()
const chatStore = useChatStore()
const authStore = useAuthStore()
const messageListRef = ref<HTMLElement | null>(null)
const inputText = ref('')
let typingTimer: ReturnType<typeof setTimeout> | null = null

const conversationId = computed(() => Number(route.params.id))

const isOnline = computed(() => {
  if (!chatStore.currentConversation) return false
  const otherId = getOtherUserId(chatStore.currentConversation!)
  return chatStore.onlineUsers.has(otherId)
})

const isTyping = computed(() => {
  if (!chatStore.currentConversation) return false
  const otherId = getOtherUserId(chatStore.currentConversation!)
  return chatStore.typingUsers.has(otherId)
})

function getOtherUserId(conv: any): number {
  if (!authStore.user) return 0
  return conv.buyer_id === authStore.user.id ? conv.seller_id : conv.buyer_id
}

function isMine(msg: Message): boolean {
  return authStore.user?.id === msg.sender_id
}

function getMessageStyle(msg: Message) {
  const mine = isMine(msg)
  return {
    display: 'flex',
    justifyContent: mine ? 'flex-end' as const : 'flex-start' as const,
    width: '100%',
    padding: '0 16px',
  }
}

function formatMessageTime(timeStr: string): string {
  if (!timeStr) return ''
  return new Date(timeStr).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

function scrollToBottom() {
  nextTick(() => {
    if (messageListRef.value) {
      messageListRef.value.scrollTop = messageListRef.value.scrollHeight
    }
  })
}

function handleSend() {
  const text = inputText.value.trim()
  if (!text) return
  chatStore.sendMessage(text)
  inputText.value = ''
  scrollToBottom()
}

function handleTyping() {
  chatStore.sendTyping()
  if (typingTimer) clearTimeout(typingTimer)
  typingTimer = setTimeout(() => {}, 2000)
}

function goBack() {
  chatStore.disconnectWebSocket()
  router.push('/messages')
}

watch(conversationId, (newId) => {
  if (newId) {
    loadConversation()
  }
})

async function loadConversation() {
  const convs = chatStore.conversations
  let conv = convs.find(c => c.id === conversationId.value)

  // 如果 store 中没有，尝试刷新
  if (!conv) {
    await chatStore.fetchConversations()
    conv = chatStore.conversations.find(c => c.id === conversationId.value)
  }

  if (conv) {
    chatStore.currentConversation = conv
    await chatStore.fetchMessages(conversationId.value)
    chatStore.connectWebSocket(conversationId.value)
    chatStore.sendReadReceipt(conversationId.value)
    scrollToBottom()
  }
}

onMounted(() => {
  loadConversation()
})

onUnmounted(() => {
  chatStore.disconnectWebSocket()
  if (typingTimer) clearTimeout(typingTimer)
})
</script>

<style scoped>
.message-list {
  scroll-behavior: smooth;
}

.message-bubble {
  width: 100%;
}
</style>
