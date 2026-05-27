<script setup lang="ts">
import { useUserStore } from '@/stores/user'
import { useRouter } from 'vue-router'
import { ref, watch } from 'vue'

const userStore = useUserStore()
const router = useRouter()

// 导航栏当前选中的菜单项
const currentRoute = ref(router.currentRoute.value.path)

watch(
  () => router.currentRoute.value.path,
  (path) => {
    currentRoute.value = path
  }
)

// 搜索关键字
const searchKeyword = ref('')

/** 执行搜索跳转到商品列表页 */
function handleSearch() {
  if (searchKeyword.value.trim()) {
    router.push({
      path: '/products',
      query: { search: searchKeyword.value.trim() },
    })
  }
}

/** 退出登录 */
function handleLogout() {
  userStore.logout()
  router.push('/')
}

/** 导航菜单点击处理 */
function handleSelect(key: string) {
  if (key === 'logout') {
    handleLogout()
  } else {
    router.push(key)
  }
}
</script>

<template>
  <div class="app-container">
    <!-- 顶部导航栏 -->
    <el-menu
      :default-active="currentRoute"
      mode="horizontal"
      class="app-header"
      @select="handleSelect"
    >
      <div class="header-content">
        <!-- Logo -->
        <div class="logo-section">
          <router-link to="/" class="logo-link">
            <el-icon :size="28"><Goods /></el-icon>
            <span class="logo-text">跳蚤市场</span>
          </router-link>
        </div>

        <!-- 搜索栏 -->
        <div class="search-section">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索商品..."
            clearable
            class="search-input"
            @keyup.enter="handleSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
            <template #append>
              <el-button @click="handleSearch">
                <el-icon><Search /></el-icon>
              </el-button>
            </template>
          </el-input>
        </div>

        <!-- 导航菜单 -->
        <div class="nav-section">
          <el-menu-item index="/">
            <el-icon><House /></el-icon>
            <span>首页</span>
          </el-menu-item>
          <el-menu-item index="/products">
            <el-icon><Goods /></el-icon>
            <span>商品</span>
          </el-menu-item>
          <el-menu-item v-if="userStore.isLoggedIn" index="/cart">
            <el-icon><ShoppingCart /></el-icon>
            <span>购物车</span>
          </el-menu-item>
          <el-menu-item v-if="userStore.isLoggedIn" index="/orders">
            <el-icon><List /></el-icon>
            <span>订单</span>
          </el-menu-item>

          <!-- 用户菜单 -->
          <el-sub-menu v-if="userStore.isLoggedIn" index="user">
            <template #title>
              <el-icon><User /></el-icon>
              <span>{{ userStore.profile?.username || '用户' }}</span>
            </template>
            <el-menu-item index="/profile">
              <el-icon><Setting /></el-icon>
              <span>个人中心</span>
            </el-menu-item>
            <el-menu-item index="/profile/products">
              <el-icon><SoldOut /></el-icon>
              <span>我的商品</span>
            </el-menu-item>
            <el-menu-item index="logout">
              <el-icon><SwitchButton /></el-icon>
              <span>退出登录</span>
            </el-menu-item>
          </el-sub-menu>

          <template v-else>
            <el-menu-item index="/login">
              <el-icon><User /></el-icon>
              <span>登录</span>
            </el-menu-item>
            <el-menu-item index="/register">
              <el-icon><UserPlus /></el-icon>
              <span>注册</span>
            </el-menu-item>
          </template>
        </div>
      </div>
    </el-menu>

    <!-- 主内容区域 -->
    <main class="app-main">
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>

    <!-- 页脚 -->
    <footer class="app-footer">
      <p>&copy; 2026 跳蚤市场 - 二手交易平台</p>
    </footer>
  </div>
</template>

<style scoped>
.app-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: #f5f7fa;
}

.app-header {
  position: sticky;
  top: 0;
  z-index: 1000;
  display: flex;
  justify-content: center;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.header-content {
  width: 100%;
  max-width: 1200px;
  display: flex;
  align-items: center;
  padding: 0 20px;
  gap: 20px;
}

.logo-section {
  flex-shrink: 0;
}

.logo-link {
  display: flex;
  align-items: center;
  gap: 8px;
  text-decoration: none;
  color: #409eff;
}

.logo-text {
  font-size: 20px;
  font-weight: bold;
  color: #409eff;
}

.search-section {
  flex: 1;
  max-width: 400px;
}

.search-input {
  --el-input-height: 36px;
}

.nav-section {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.app-main {
  flex: 1;
  max-width: 1200px;
  width: 100%;
  margin: 0 auto;
  padding: 20px;
}

.app-footer {
  text-align: center;
  padding: 20px;
  color: #999;
  font-size: 14px;
  border-top: 1px solid #ebeef5;
  background: #fff;
}

/* 路由切换过渡动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
