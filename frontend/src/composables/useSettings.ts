import { ref } from 'vue'

const cache = ref<Record<string, string> | null>(null)
const defaults = { app_name: 'drp-mikrest', logo_path: '', favicon_path: '', app_url: 'http://localhost:8080' }

function faviconUrl(): string {
  const p = cache.value?.favicon_path
  if (!p) return '/favicon.svg'
  return `/${p}`
}

function updateFavicon() {
  const link = document.querySelector<HTMLLinkElement>('link[rel="icon"]')
  if (link) link.href = faviconUrl()
}

export function useSettings() {
  const API_BASE = import.meta.env.VITE_API_BASE || '/api/web'

  async function fetchSettings() {
    try {
      const headers: Record<string, string> = {}
      const token = localStorage.getItem('drp_token')
      if (token) headers['Authorization'] = `Bearer ${token}`
      const res = await fetch(`${API_BASE}/settings`, { headers })
      if (!res.ok) throw new Error()
      cache.value = await res.json()
      updateFavicon()
    } catch {
      if (!cache.value) cache.value = { ...defaults }
      updateFavicon()
    }
    return cache.value
  }

  function appName(): string {
    return cache.value?.app_name || defaults.app_name
  }

  function logoUrl(): string {
    const p = cache.value?.logo_path
    if (!p) return ''
    return `/${p}`
  }

  function appUrl(): string {
    const u = cache.value?.app_url
    if (!u) return defaults.app_url
    return u.replace(/\/+$/, '')
  }

  return { fetchSettings, appName, logoUrl, appUrl, faviconUrl, updateFavicon }
}

export function initSettings() {
  const token = localStorage.getItem('drp_token')
  if (token && !cache.value) {
    const API_BASE = import.meta.env.VITE_API_BASE || '/api/web'
    fetch(`${API_BASE}/settings`, {
      headers: { Authorization: `Bearer ${token}` },
    })
      .then((r) => r.ok ? r.json() : null)
      .then((data) => { if (data) cache.value = data; updateFavicon() })
      .catch(() => {})
  }
  if (!cache.value) cache.value = { ...defaults }
  updateFavicon()
}
