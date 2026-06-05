<template>
  <div class="auth-container">
    <n-card title="注册" style="width: 400px">
      <n-form ref="formRef" :rules="rules" :model="form" @submit.prevent="handleRegister">
        <n-form-item label="邮箱" path="email">
          <n-input v-model:value="form.email" placeholder="请输入邮箱" />
        </n-form-item>
        <n-form-item label="密码" path="password">
          <n-input
            v-model:value="form.password"
            type="password"
            placeholder="至少8位，含大小写字母和数字"
            show-password-on="click"
          />
        </n-form-item>
        <n-form-item label="确认密码" path="confirmPassword">
          <n-input
            v-model:value="form.confirmPassword"
            type="password"
            placeholder="请再次输入密码"
            show-password-on="click"
          />
        </n-form-item>
        <n-button type="primary" block :loading="loading" attr-type="submit">
          注册
        </n-button>
      </n-form>
      <n-p style="text-align: center; margin-top: 16px">
        已有账号？
        <n-button text type="primary" @click="$router.push('/login')">
          立即登录
        </n-button>
      </n-p>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import type { FormRules } from 'naive-ui'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const message = useMessage()
const authStore = useAuthStore()
const loading = ref(false)

interface RegisterForm {
  email: string
  password: string
  confirmPassword: string
}

const form = reactive<RegisterForm>({
  email: '',
  password: '',
  confirmPassword: '',
})

const rules: FormRules = {
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 8, message: '密码至少8位', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator: (_rule, value) => value === form.password,
      message: '两次密码不一致',
      trigger: 'blur',
    },
  ],
}

async function handleRegister() {
  loading.value = true
  try {
    await authStore.register({ email: form.email, password: form.password })
    message.success('注册成功')
    router.push('/')
  } catch (err: any) {
    const msg = err?.response?.data?.message || '注册失败'
    message.error(msg)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background-color: #f5f5f5;
}
</style>
