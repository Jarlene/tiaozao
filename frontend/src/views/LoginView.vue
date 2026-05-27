<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { api } from '@/api'

const router = useRouter()
const authStore = useAuthStore()

const phone = ref('')
const code = ref('')
const step = ref<'phone' | 'code'>('phone')
const loading = ref(false)
const countdown = ref(0)
const error = ref('')
const successMsg = ref('')

let countdownTimer: ReturnType<typeof setInterval> | null = null

function validatePhone(): boolean {
  if (!/^1\d{10}$/.test(phone.value)) {
    error.value = '请输入正确的手机号'
    return false
  }
  return true
}

async function sendCode() {
  error.value = ''
  successMsg.value = ''
  if (!validatePhone()) return

  loading.value = true
  try {
    await api.sendCode(phone.value)
    step.value = 'code'
    startCountdown()
    successMsg.value = '验证码已发送'
  } catch (e: any) {
    error.value = e.message || '发送验证码失败'
  } finally {
    loading.value = false
  }
}

function startCountdown() {
  countdown.value = 60
  countdownTimer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      if (countdownTimer) clearInterval(countdownTimer)
    }
  }, 1000)
}

async function handleLogin() {
  error.value = ''
  if (!code.value || code.value.length !== 6) {
    error.value = '请输入6位验证码'
    return
  }

  loading.value = true
  try {
    await authStore.login(phone.value, code.value)
    router.push('/profile')
  } catch (e: any) {
    error.value = e.message || '登录失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <h1 class="title">欢迎来到跳蚤市场</h1>
      <p class="subtitle">登录后即可浏览和发布商品</p>

      <div class="form">
        <div class="input-group">
          <label>手机号</label>
          <input
            v-model="phone"
            type="tel"
            maxlength="11"
            placeholder="请输入手机号"
            :disabled="step === 'code'"
          />
        </div>

        <div class="input-group" v-if="step === 'code'">
          <label>验证码</label>
          <div class="code-row">
            <input
              v-model="code"
              type="text"
              maxlength="6"
              placeholder="请输入6位验证码"
            />
            <button
              class="send-code-btn"
              :disabled="countdown > 0"
              @click="sendCode"
            >
              {{ countdown > 0 ? `${countdown}s` : '重新发送' }}
            </button>
          </div>
        </div>

        <p v-if="error" class="error-text">{{ error }}</p>
        <p v-if="successMsg" class="success-text">{{ successMsg }}</p>

        <button
          v-if="step === 'phone'"
          class="submit-btn"
          :disabled="loading"
          @click="sendCode"
        >
          {{ loading ? '发送中...' : '获取验证码' }}
        </button>

        <button
          v-if="step === 'code'"
          class="submit-btn"
          :disabled="loading"
          @click="handleLogin"
        >
          {{ loading ? '登录中...' : '登录' }}
        </button>

        <button
          v-if="step === 'code'"
          class="back-btn"
          @click="step = 'phone'"
        >
          返回修改手机号
        </button>
      </div>

      <div class="hint">
        <p>测试模式验证码: 123456</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: calc(100vh - 112px);
}

.login-card {
  background: #fff;
  border-radius: 12px;
  padding: 40px;
  width: 100%;
  max-width: 400px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.title {
  font-size: 24px;
  font-weight: 700;
  text-align: center;
  margin-bottom: 8px;
  color: #1a1a1a;
}

.subtitle {
  font-size: 14px;
  color: #999;
  text-align: center;
  margin-bottom: 32px;
}

.form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.input-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.input-group label {
  font-size: 14px;
  color: #666;
  font-weight: 500;
}

.input-group input {
  padding: 12px 16px;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  font-size: 16px;
  transition: border-color 0.2s;
}

.input-group input:focus {
  border-color: #ff6a00;
}

.input-group input:disabled {
  background: #f5f5f5;
  color: #999;
}

.code-row {
  display: flex;
  gap: 12px;
}

.code-row input {
  flex: 1;
}

.send-code-btn {
  padding: 12px 16px;
  background: #fff;
  border: 1px solid #ff6a00;
  color: #ff6a00;
  border-radius: 8px;
  font-size: 14px;
  white-space: nowrap;
  transition: all 0.2s;
}

.send-code-btn:hover:not(:disabled) {
  background: #fff5eb;
}

.send-code-btn:disabled {
  border-color: #d9d9d9;
  color: #bbb;
  cursor: not-allowed;
}

.submit-btn {
  width: 100%;
  padding: 12px;
  background: #ff6a00;
  color: #fff;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 600;
  transition: background 0.2s;
}

.submit-btn:hover:not(:disabled) {
  background: #e85d00;
}

.submit-btn:disabled {
  background: #ffb380;
  cursor: not-allowed;
}

.back-btn {
  width: 100%;
  padding: 10px;
  background: transparent;
  color: #999;
  font-size: 14px;
}

.back-btn:hover {
  color: #666;
}

.error-text {
  color: #e53935;
  font-size: 13px;
  text-align: center;
}

.success-text {
  color: #43a047;
  font-size: 13px;
  text-align: center;
}

.hint {
  margin-top: 24px;
  padding: 12px;
  background: #fff8e1;
  border-radius: 8px;
  text-align: center;
}

.hint p {
  font-size: 12px;
  color: #f9a825;
}
</style>
