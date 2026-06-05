<template>
  <n-layout position="absolute">
    <n-layout-header bordered>
      <n-space style="padding: 0 24px; height: 64px" align="center" justify="space-between">
        <n-space align="center">
          <n-h2 style="margin: 0" @click="$router.push('/')" class="logo">
            跳蚤市场
          </n-h2>
          <n-space style="margin-left: 32px">
            <n-button quaternary @click="$router.push('/products')">
              商品市场
            </n-button>
            <template v-if="authStore.isAuthenticated">
              <n-button quaternary @click="$router.push('/products/create')">
                发布商品
              </n-button>
              <n-button quaternary @click="$router.push('/products/mine')">
                我的发布
              </n-button>
              <n-button quaternary @click="$router.push('/categories')">
                分类管理
              </n-button>
              <n-badge :value="chatStore.totalUnread" :max="99">
                <n-button quaternary @click="$router.push('/messages')">
                  消息
                </n-button>
              </n-badge>
            </template>
          </n-space>
        </n-space>
        <n-space>
          <template v-if="authStore.isAuthenticated">
            <n-button quaternary @click="$router.push('/profile')">
              <template #icon>
                <n-icon><person-outline /></n-icon>
              </template>
              {{ authStore.user?.nickname || '个人中心' }}
            </n-button>
            <n-button quaternary @click="handleLogout">
              <template #icon>
                <n-icon><log-out-outline /></n-icon>
              </template>
              退出
            </n-button>
          </template>
          <template v-else>
            <n-button @click="$router.push('/login')">登录</n-button>
            <n-button @click="$router.push('/register')" type="primary">注册</n-button>
          </template>
        </n-space>
      </n-space>
    </n-layout-header>

    <n-layout-content content-style="padding: 24px; min-height: calc(100vh - 64px - 60px);">
      <router-view />
    </n-layout-content>

    <n-layout-footer bordered style="height: 60px; display: flex; align-items: center; justify-content: center;">
      <n-text depth="3">© 2024 跳蚤市场. All rights reserved.</n-text>
    </n-layout-footer>
  </n-layout>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useChatStore } from '@/stores/chat'
import { LogOutOutline, PersonOutline } from '@vicons/ionicons5'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const chatStore = useChatStore()
const router = useRouter()

onMounted(() => {
  if (authStore.isAuthenticated) {
    chatStore.fetchConversations()
  }
})

function handleLogout() {
  authStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.logo {
  cursor: pointer;
  color: #18a058;
}
</style>
