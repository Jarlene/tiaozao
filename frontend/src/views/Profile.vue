<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'

const router = useRouter()
const userStore = useUserStore()

// 个人资料表单
const profileForm = reactive({
  username: '',
  email: '',
  phone: '',
  bio: '',
})

const loading = ref(false)
const profileFormRef = ref()

/** 表单校验规则 */
const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 2, max: 20, message: '用户名长度为 2-20 个字符', trigger: 'blur' },
  ],
  email: [
    { required: true, message: '请输入邮箱地址', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' },
  ],
  phone: [
    {
      pattern: /^1[3-9]\d{9}$/,
      message: '请输入正确的手机号码',
      trigger: 'blur',
    },
  ],
}

onMounted(() => {
  // 加载用户信息到表单
  if (userStore.profile) {
    profileForm.username = userStore.profile.username || ''
    profileForm.email = userStore.profile.email || ''
    profileForm.phone = userStore.profile.phone || ''
    profileForm.bio = userStore.profile.bio || ''
  }
})

/** 保存个人资料 */
async function handleSave() {
  if (!profileFormRef.value) return

  const valid = await profileFormRef.value.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await userStore.updateProfile({
      username: profileForm.username,
      email: profileForm.email,
      phone: profileForm.phone,
      bio: profileForm.bio,
    })
    ElMessage.success('个人资料已更新')
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.message || '更新失败')
  } finally {
    loading.value = false
  }
}

/** 退出登录 */
function handleLogout() {
  userStore.logout()
  ElMessage.success('已退出登录')
  router.push('/')
}
</script>

<template>
  <div class="profile-page">
    <h2 class="page-title">个人中心</h2>

    <div class="profile-content">
      <!-- 左侧：用户信息卡片 -->
      <div class="profile-sidebar">
        <el-card class="user-card">
          <div class="user-avatar-section">
            <el-avatar :size="80" class="user-avatar">
              {{ userStore.profile?.username?.[0] || 'U' }}
            </el-avatar>
            <h3 class="user-name">{{ userStore.profile?.username || '用户' }}</h3>
            <p class="user-email">{{ userStore.profile?.email }}</p>
          </div>
          <el-divider />
          <div class="user-stats">
            <div class="stat-item">
              <span class="stat-value">-</span>
              <span class="stat-label">在售商品</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">-</span>
              <span class="stat-label">已完成订单</span>
            </div>
          </div>
          <el-divider />
          <el-button
            type="danger"
            class="logout-btn"
            @click="handleLogout"
          >
            <el-icon><SwitchButton /></el-icon>
            退出登录
          </el-button>
        </el-card>

        <!-- 快捷导航 -->
        <el-card class="nav-card">
          <template #header>
            <span>快捷操作</span>
          </template>
          <div class="quick-nav">
            <el-button text @click="router.push('/profile/products')">
              <el-icon><SoldOut /></el-icon>
              我的商品
            </el-button>
            <el-button text @click="router.push('/orders')">
              <el-icon><List /></el-icon>
              我的订单
            </el-button>
            <el-button text @click="router.push('/cart')">
              <el-icon><ShoppingCart /></el-icon>
              购物车
            </el-button>
          </div>
        </el-card>
      </div>

      <!-- 右侧：编辑表单 -->
      <div class="profile-main">
        <el-card class="form-card">
          <template #header>
            <span>编辑个人资料</span>
          </template>

          <el-form
            ref="profileFormRef"
            :model="profileForm"
            :rules="rules"
            label-width="100px"
            class="profile-form"
          >
            <el-form-item label="用户名" prop="username">
              <el-input v-model="profileForm.username" />
            </el-form-item>

            <el-form-item label="邮箱" prop="email">
              <el-input v-model="profileForm.email" />
            </el-form-item>

            <el-form-item label="手机号" prop="phone">
              <el-input v-model="profileForm.phone" placeholder="选填" />
            </el-form-item>

            <el-form-item label="个人简介" prop="bio">
              <el-input
                v-model="profileForm.bio"
                type="textarea"
                :rows="4"
                placeholder="介绍一下自己吧..."
                maxlength="200"
                show-word-limit
              />
            </el-form-item>

            <el-form-item>
              <el-button
                type="primary"
                :loading="loading"
                @click="handleSave"
              >
                保存修改
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </div>
    </div>
  </div>
</template>

<style scoped>
.profile-page {
  padding-bottom: 40px;
}

.page-title {
  font-size: 22px;
  font-weight: 600;
  margin: 0 0 20px 0;
  color: #333;
}

.profile-content {
  display: flex;
  gap: 24px;
}

.profile-sidebar {
  width: 280px;
  flex-shrink: 0;
}

.user-card {
  text-align: center;
  margin-bottom: 16px;
}

.user-avatar-section {
  padding: 20px 0;
}

.user-avatar {
  margin-bottom: 12px;
}

.user-name {
  font-size: 18px;
  font-weight: 600;
  margin: 0 0 4px 0;
  color: #333;
}

.user-email {
  font-size: 13px;
  color: #999;
  margin: 0;
}

.user-stats {
  display: flex;
  justify-content: space-around;
  padding: 8px 0;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.stat-value {
  font-size: 20px;
  font-weight: bold;
  color: #333;
}

.stat-label {
  font-size: 12px;
  color: #999;
}

.logout-btn {
  width: 100%;
}

.nav-card {
  margin-bottom: 16px;
}

.quick-nav {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 8px;
}

.profile-main {
  flex: 1;
  min-width: 0;
}

.form-card {
  border-radius: 8px;
}

.profile-form {
  max-width: 500px;
}
</style>
