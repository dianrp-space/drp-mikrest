import { defineStore } from 'pinia'
import { useApi } from '@/composables/useApi'

interface User {
  id: string
  email: string
  role: string
}

interface LoginResponse {
  access_token: string
  expires_at: string
  token_type: string
  user: User
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: '' as string,
    user: null as User | null,
    expiresAt: '' as string,
  }),
  getters: {
    isAuthenticated: (state) => !!state.token,
    email: (state) => state.user?.email ?? '',
    role: (state) => state.user?.role ?? '',
  },
  actions: {
    init() {
      this.token = localStorage.getItem('drp_token') || ''
      this.expiresAt = localStorage.getItem('drp_expires') || ''
      const userRaw = localStorage.getItem('drp_user') || ''
      try {
        this.user = userRaw ? JSON.parse(userRaw) : null
      } catch {
        this.user = null
      }
    },
    async login(email: string, password: string) {
      const { apiPost } = useApi()
      const res = await apiPost<LoginResponse>('/auth/login', { email, password })
      this.token = res.access_token
      this.expiresAt = res.expires_at
      this.user = res.user
      localStorage.setItem('drp_token', res.access_token)
      localStorage.setItem('drp_user', JSON.stringify(res.user))
      localStorage.setItem('drp_expires', res.expires_at)
    },
    async register(email: string, password: string) {
      const { apiPost } = useApi()
      return apiPost<User>('/auth/register', { email, password })
    },
    logout() {
      this.token = ''
      this.user = null
      this.expiresAt = ''
      localStorage.removeItem('drp_token')
      localStorage.removeItem('drp_user')
      localStorage.removeItem('drp_expires')
    },
  },
})
