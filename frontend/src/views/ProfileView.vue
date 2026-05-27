<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { api } from '@/api'

const router = useRouter()
const authStore = useAuthStore()

const nickname = ref('')
const editing = ref(false)
const saving = ref(false)
const error = ref('')
const success = ref('')

onMounted(() => {
  if (authStore.user) {
    nickname.value = authStore.user.nickname || ''
  }
})

async function saveProfile() {
  error.value = ''
  success.value = ''
  saving.value = true

  try {
    await api.updateProfile({ nickname: nickname.value })
    await authStore.refreshProfile()
    editing.value = false
    success.value = '个人资料更新成功'
    setTimeout(() => (success.value = ''), 3000)
  } catch (e: any) {
    error.value = e.message || '更新失败'
  } finally {
    saving.value = false
  }
}

function handleLogout() {
  authStore.logout()
  router.push('/login')
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  })
}
</script>

<template>
  <div class="profile-page">
    <div class="profile-card">
      <div class="card-header">
        <h2>个人中心</h2>
        <button class="logout-btn" @click="handleLogout">退出登录</button>
      </div>

      <div class="avatar-section">
        <div class="avatar">
          {{ authStore.user?.nickname?.charAt(0) || authStore.user?.phone?.slice(-1) || '?' }}
        </div>
      </div>

      <div class="info-list">
        <div class="info-item">
          <span class="label">手机号</span>
          <span class="value">{{ authStore.user?.phone || '-' }}</span>
        </div>

        <div class="info-item">
          <span class="label">昵称</span>
          <div class="value-row" v-if="!editing">
            <span>{{ authStore.user?.nickname || '未设置' }}</span>
            <button class="edit-btn" @click="editing = true">编辑</button>
          </div>
          <div class="value-row" v-else>
            <input
              v-model="nickname"
              maxlength="50"
              placeholder="请输入昵称"
              class="nickname-input"
            />
            <div class="edit-actions">
              <button class="save-btn" :disabled="saving" @click="saveProfile">
                {{ saving ? '保存中...' : '保存' }}
              </button>
              <button class="cancel-btn" @click="editing = false">取消</button>
            </div>
          </div>
        </div>

        <div class="info-item">
          <span class="label">注册时间</span>
          <span class="value">{{ formatDate(authStore.user?.created_at) || '-' }}</span>
        </div>
      </div>

      <p v-if="error" class="error-text">{{ error }}</p>
      <p v-if="success" class="success-text">{{ success }}</p>
    </div>
  </div>
</template>

<style scoped>
.profile-page {
  display: flex;
  justify-content: center;
  padding-top: 40px;
}

.profile-card {
  background: #fff;
  border-radius: 12px;
  padding: 32px;
  width: 100%;
  max-width: 500px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.card-header h2 {
  font-size: 20px;
  font-weight: 700;
  color: #1a1a1a;
}

.logout-btn {
  padding: 8px 16px;
  background: transparent;
  color: #e53935;
  border: 1px solid #e53935;
  border-radius: 6px;
  font-size: 13px;
  transition: all 0.2s;
}

.logout-btn:hover {
  background: #fff0f0;
}

.avatar-section {
  display: flex;
  justify-content: center;
  margin-bottom: 32px;
}

.avatar {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: linear-gradient(135deg, #ff6a00, #ff8c00);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  font-weight: 700;
}

.info-list {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 0;
  border-bottom: 1px solid #f0f0f0;
}

.info-item:last-child {
  border-bottom: none;
}

.label {
  font-size: 14px;
  color: #999;
  min-width: 80px;
}

.value {
  font-size: 14px;
  color: #333;
}

.value-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.nickname-input {
  padding: 8px 12px;
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  font-size: 14px;
  width: 180px;
}

.nickname-input:focus {
  border-color: #ff6a00;
}

.edit-actions {
  display: flex;
  gap: 8px;
}

.edit-btn {
  padding: 4px 12px;
  background: #f5f5f5;
  color: #666;
  border-radius: 4px;
  font-size: 12px;
}

.edit-btn:hover {
  background: #ebebeb;
}

.save-btn {
  padding: 6px 14px;
  background: #ff6a00;
  color: #fff;
  border-radius: 6px;
  font-size: 13px;
}

.save-btn:hover:not(:disabled) {
  background: #e85d00;
}

.save-btn:disabled {
  background: #ffb380;
  cursor: not-allowed;
}

.cancel-btn {
  padding: 6px 14px;
  background: #f5f5f5;
  color: #666;
  border-radius: 6px;
  font-size: 13px;
}

.cancel-btn:hover {
  background: #ebebeb;
}

.error-text {
  color: #e53935;
  font-size: 13px;
  text-align: center;
  margin-top: 16px;
}

.success-text {
  color: #43a047;
  font-size: 13px;
  text-align: center;
  margin-top: 16px;
}
</style>
