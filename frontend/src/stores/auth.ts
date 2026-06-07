import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authAPI } from '@/api/auth'
import type { LoginReq, RegisterReq, UserProfile, TokenPair } from '@/api/auth'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('access_token'))
  const refreshToken = ref<string | null>(localStorage.getItem('refresh_token'))
  const user = ref<UserProfile | null>(null)

  const isAuthenticated = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 2)

  async function login(req: LoginReq) {
    const res = await authAPI.login(req)
    setTokens(res.tokens)
    user.value = res.profile
    return res
  }

  async function register(req: RegisterReq) {
    const res = await authAPI.register(req)
    setTokens(res.tokens)
    user.value = res.profile
    return res
  }

  async function refresh() {
    if (!refreshToken.value) throw new Error('No refresh token')
    const res = await authAPI.refresh(refreshToken.value)
    setTokens(res.tokens)
    return res
  }

  async function getProfile() {
    const profile = await authAPI.getProfile()
    user.value = profile
    return profile
  }

  async function updateProfile(data: { nickname?: string }) {
    const profile = await authAPI.updateProfile(data)
    user.value = profile
    return profile
  }

  function setTokens(t: TokenPair) {
    token.value = t.access_token
    refreshToken.value = t.refresh_token
    localStorage.setItem('access_token', t.access_token)
    localStorage.setItem('refresh_token', t.refresh_token)
  }

  function logout() {
    token.value = null
    refreshToken.value = null
    user.value = null
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
  }

  return {
    token,
    refreshToken,
    user,
    isAuthenticated,
    isAdmin,
    login,
    register,
    refresh,
    getProfile,
    updateProfile,
    logout,
  }
})
