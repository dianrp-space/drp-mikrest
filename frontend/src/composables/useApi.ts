import { useAuthStore } from '@/stores/auth'

const API_BASE = import.meta.env.VITE_API_BASE || '/api/web'

export function useApi() {
  async function request<T>(method: string, path: string, body?: any): Promise<T> {
    const auth = useAuthStore()
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    }
    if (auth.token) {
      headers['Authorization'] = `Bearer ${auth.token}`
    }
    const res = await fetch(`${API_BASE}${path}`, {
      method,
      headers,
      body: body !== undefined ? JSON.stringify(body) : undefined,
    })
    if (res.status === 401) {
      auth.logout()
      window.location.href = '/login'
      throw new Error('Unauthorized')
    }
    const text = await res.text()
    let data: any
    try {
      data = text ? JSON.parse(text) : null
    } catch {
      data = text
    }
    if (!res.ok) {
      const msg = (data && (data.message || data.error)) || `HTTP ${res.status}`
      throw new Error(msg)
    }
    return data as T
  }

  const apiGet = <T>(path: string) => request<T>('GET', path)
  const apiPost = <T>(path: string, body?: any) => request<T>('POST', path, body)
  const apiPut = <T>(path: string, body?: any) => request<T>('PUT', path, body)
  const apiDelete = <T>(path: string) => request<T>('DELETE', path)

  async function apiUpload<T>(path: string, formData: FormData): Promise<T> {
    const auth = useAuthStore()
    const headers: Record<string, string> = {}
    if (auth.token) {
      headers['Authorization'] = `Bearer ${auth.token}`
    }
    const res = await fetch(`${API_BASE}${path}`, {
      method: 'POST',
      headers,
      body: formData,
    })
    if (res.status === 401) {
      auth.logout()
      window.location.href = '/login'
      throw new Error('Unauthorized')
    }
    const text = await res.text()
    let data: any
    try {
      data = text ? JSON.parse(text) : null
    } catch {
      data = text
    }
    if (!res.ok) {
      const msg = (data && (data.message || data.error)) || `HTTP ${res.status}`
      throw new Error(msg)
    }
    return data as T
  }

  return { apiGet, apiPost, apiPut, apiDelete, apiUpload }
}
