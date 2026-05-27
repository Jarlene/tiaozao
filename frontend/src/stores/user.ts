import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { UserProfile, LoginCredentials, RegisterData } from '@/api'
import { authApi, userApi } from '@/api'

/**
 * 用户认证状态管理
 * - 使用 localStorage 持久化 token
 * - 提供登录、注册、退出登录操作
 * - 自动恢复登录状态
 */
export const useUserStore = defineStore('user', () => {
  // 从 localStorage 恢复 token
  const token = ref<string | null>(localStorage.getItem('tiaozao_token'))
  const profile = ref<UserProfile | null>(null)

  // 计算属性：是否已登录
  const isLoggedIn = computed(() => !!token.value)

  /** 初始化时尝试加载用户信息 */
  async function loadProfile() {
    if (!token.value) return
    try {
      profile.value = await userApi.getProfile()
    } catch {
      // token 失效，清除登录状态
      token.value = null
      localStorage.removeItem('tiaozao_token')
    }
  }

  /**
   * 用户登录
   * @param credentials 登录凭证（用户名/邮箱 + 密码）
   */
  async function login(credentials: LoginCredentials) {
    const res = await authApi.login(credentials)
    token.value = res.token
    localStorage.setItem('tiaozao_token', res.token)
    await loadProfile()
  }

  /**
   * 用户注册
   * @param data 注册信息
   */
  async function register(data: RegisterData) {
    const res = await authApi.register(data)
    token.value = res.token
    localStorage.setItem('tiaozao_token', res.token)
    await loadProfile()
  }

  /** 退出登录：清除 token 和个人信息 */
  function logout() {
    token.value = null
    profile.value = null
    localStorage.removeItem('tiaozao_token')
  }

  /** 更新个人资料 */
  async function updateProfile(data: Partial<UserProfile>) {
    profile.value = await userApi.updateProfile(data)
  }

  // 自动加载用户信息
  if (token.value) {
    loadProfile()
  }

  return {
    token,
    profile,
    isLoggedIn,
    login,
    register,
    logout,
    loadProfile,
    updateProfile,
  }
})
