<script setup lang="ts">
import { useAuthStore } from '@/store/auth'
import { onMounted } from 'vue'

const authStore = useAuthStore()

onMounted(() => {
  authStore.restoreSession()
})
</script>

<template>
  <div id="app-container">
    <header class="app-header">
      <div class="header-content">
        <router-link to="/" class="logo">跳蚤市场</router-link>
        <nav class="nav-links">
          <router-link to="/profile" v-if="authStore.isLoggedIn" class="nav-link">
            个人中心
          </router-link>
          <router-link to="/login" v-if="!authStore.isLoggedIn" class="nav-link">
            登录
          </router-link>
        </nav>
      </div>
    </header>
    <main class="app-main">
      <router-view />
    </main>
  </div>
</template>

<style scoped>
.app-header {
  background: #fff;
  border-bottom: 1px solid #e8e8e8;
  padding: 0 24px;
  position: sticky;
  top: 0;
  z-index: 100;
}
.header-content {
  max-width: 1200px;
  margin: 0 auto;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.logo {
  font-size: 22px;
  font-weight: 700;
  color: #ff6a00;
  text-decoration: none;
}
.nav-links {
  display: flex;
  gap: 16px;
}
.nav-link {
  color: #333;
  text-decoration: none;
  font-size: 14px;
  padding: 8px 16px;
  border-radius: 6px;
  transition: background 0.2s;
}
.nav-link:hover {
  background: #f5f5f5;
}
.app-main {
  max-width: 1200px;
  margin: 0 auto;
  padding: 24px;
}
</style>
