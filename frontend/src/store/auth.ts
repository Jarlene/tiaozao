import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '@/api'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(null)
  const user = ref<any | null>(null)

  const isLoggedIn = computed(() => !!token.value)

  function restoreSession() {
    const savedToken = localStorage.getItem('token')
    const savedUser = localStorage.getItem('user')
    if (savedToken && savedUser) {
      token.value = savedToken
      user.value = JSON.parse(savedUser)
    }
  }

  async function login(phone: string, code: string) {
    const res = await api.login(phone, code)
    token.value = res.token
    user.value = res.user
    localStorage.setItem('token', res.token)
    localStorage.setItem('user', JSON.stringify(res.user))
  }

  function logout() {
    token.value = null
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  async function refreshProfile() {
    const res = await api.getProfile()
    user.value = res.user
    localStorage.setItem('user', JSON.stringify(res.user))
  }

  return { token, user, isLoggedIn, restoreSession, login, logout, refreshProfile }
})
