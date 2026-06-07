import client from './client'
import type { ApiResponse } from './client'

export interface LoginReq {
  email: string
  password: string
}

export interface RegisterReq {
  email: string
  password: string
}

export interface TokenPair {
  access_token: string
  refresh_token: string
  expires_in: number
}

export interface UserProfile {
  id: number
  email: string
  nickname: string
  avatar_url: string
  role: number
  created_at: string
}

interface AuthResponse {
  tokens: TokenPair
  profile: UserProfile
}

interface RefreshResponse {
  tokens: TokenPair
}

export const authAPI = {
  login(data: LoginReq) {
    return client.post<ApiResponse<AuthResponse>>('/auth/login', data).then((r) => r.data.data)
  },

  register(data: RegisterReq) {
    return client.post<ApiResponse<AuthResponse>>('/auth/register', data).then((r) => r.data.data)
  },

  refresh(refreshToken: string) {
    return client
      .post<ApiResponse<RefreshResponse>>('/auth/refresh', { refresh_token: refreshToken })
      .then((r) => r.data.data)
  },

  getProfile() {
    return client.get<ApiResponse<UserProfile>>('/auth/profile').then((r) => r.data.data)
  },

  updateProfile(data: { nickname?: string }) {
    return client.put<ApiResponse<UserProfile>>('/auth/profile', data).then((r) => r.data.data)
  },
}
