const BASE_URL = '/api/v1'

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const token = localStorage.getItem('token')
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options?.headers as Record<string, string>),
  }

  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const response = await fetch(`${BASE_URL}${url}`, {
    ...options,
    headers,
  })

  const data = await response.json()

  if (!response.ok) {
    throw new Error(data.error || 'Request failed')
  }

  return data
}

export const api = {
  sendCode(phone: string): Promise<{ message: string }> {
    return request('/auth/send-code', {
      method: 'POST',
      body: JSON.stringify({ phone }),
    })
  },

  login(phone: string, code: string): Promise<{ token: string; user: any }> {
    return request('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ phone, code }),
    })
  },

  getProfile(): Promise<{ user: any }> {
    return request('/user/profile')
  },

  updateProfile(data: { nickname?: string; avatar?: string }): Promise<{ user: any }> {
    return request('/user/profile', {
      method: 'PUT',
      body: JSON.stringify(data),
    })
  },
}
