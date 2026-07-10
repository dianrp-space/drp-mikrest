<template>
  <div class="space-y-6 animate-fade-in">
    <div>
      <h1 class="page-title">Settings</h1>
      <p class="page-subtitle">Kelola aplikasi & akun</p>
    </div>

    <div class="flex border-b border-ink-800">
      <button
        v-for="t in tabs" :key="t.key"
        @click="activeTab = t.key"
        class="px-5 py-2.5 text-sm font-medium transition-colors relative"
        :class="activeTab === t.key ? 'text-brand-400' : 'text-ink-500 hover:text-ink-300'"
      >
        {{ t.label }}
        <span v-if="activeTab === t.key" class="absolute bottom-0 left-0 right-0 h-0.5 bg-brand-400 rounded-full"></span>
      </button>
    </div>

    <!-- General Tab -->
    <div v-if="activeTab === 'general'" class="space-y-6 max-w-2xl">
      <div class="card p-6 space-y-4">
        <h2 class="font-semibold text-ink-100">Informasi Aplikasi</h2>
        <div>
          <label class="label">Nama Aplikasi</label>
          <input v-model="form.app_name" class="input w-full" placeholder="drp-mikrest" />
        </div>
        <div>
          <label class="label">Base URL Aplikasi</label>
          <input v-model="form.app_url" type="url" class="input w-full" placeholder="http://localhost:8080" />
          <p class="text-xs text-ink-500 mt-1">Digunakan untuk contoh di halaman API Token.</p>
        </div>
        <div class="flex gap-4 items-end">
          <button class="btn-primary text-sm" @click="saveSettings" :disabled="saving">
            {{ saving ? 'Menyimpan...' : 'Simpan' }}
          </button>
          <span v-if="saved" class="text-sm text-accent-400">Tersimpan</span>
        </div>
      </div>

      <div class="card p-6 space-y-4">
        <h2 class="font-semibold text-ink-100">Logo Aplikasi</h2>
        <p class="text-sm text-ink-500">Digunakan di sidebar & halaman login.</p>
        <div v-if="logoPreview" class="w-16 h-16 rounded-lg bg-ink-900 border border-ink-800 flex items-center justify-center overflow-hidden">
          <img :src="logoPreview" class="w-full h-full object-contain" />
        </div>
        <div class="flex gap-3 items-center">
          <label class="btn-secondary text-sm cursor-pointer">
            Pilih File
            <input type="file" accept="image/*" class="hidden" @change="onLogoChange" />
          </label>
          <button v-if="form.logo_path" class="btn-ghost text-xs text-danger-400" @click="removeLogo">Hapus</button>
        </div>
      </div>

      <div class="card p-6 space-y-4">
        <h2 class="font-semibold text-ink-100">Favicon</h2>
        <p class="text-sm text-ink-500">Icon tab browser.</p>
        <div v-if="faviconPreview" class="w-10 h-10 rounded-lg bg-ink-900 border border-ink-800 flex items-center justify-center overflow-hidden">
          <img :src="faviconPreview" class="w-full h-full object-contain" />
        </div>
        <div class="flex gap-3 items-center">
          <label class="btn-secondary text-sm cursor-pointer">
            Pilih File
            <input type="file" accept="image/*" class="hidden" @change="onFaviconChange" />
          </label>
          <button v-if="form.favicon_path" class="btn-ghost text-xs text-danger-400" @click="removeFavicon">Hapus</button>
        </div>
      </div>

    </div>

    <!-- Profile Tab -->
    <div v-if="activeTab === 'profile'" class="space-y-6 max-w-2xl">
      <div class="card p-6 space-y-4">
        <h2 class="font-semibold text-ink-100">Ubah Email</h2>
        <div>
          <label class="label">Email Saat Ini</label>
          <input :value="authStore.email" class="input w-full bg-ink-850 text-ink-500" disabled />
        </div>
        <div>
          <label class="label">Email Baru</label>
          <input v-model="emailForm.new_email" type="email" class="input w-full" placeholder="baru@example.com" />
        </div>
        <div>
          <label class="label">Password (konfirmasi)</label>
          <PasswordField v-model="emailForm.password" placeholder="Masukkan password" />
        </div>
        <button class="btn-primary text-sm" @click="changeEmail" :disabled="emailLoading">
          {{ emailLoading ? 'Menyimpan...' : 'Simpan Email' }}
        </button>
        <p v-if="emailError" class="text-sm text-danger-400">{{ emailError }}</p>
      </div>

      <div class="card p-6 space-y-4">
        <h2 class="font-semibold text-ink-100">Ubah Password</h2>
        <div>
          <label class="label">Password Lama</label>
          <PasswordField v-model="passForm.old_password" placeholder="Password saat ini" />
        </div>
        <div>
          <label class="label">Password Baru</label>
          <PasswordField v-model="passForm.new_password" placeholder="Minimal 8 karakter" />
        </div>
        <button class="btn-primary text-sm" @click="changePassword" :disabled="passLoading">
          {{ passLoading ? 'Menyimpan...' : 'Simpan Password' }}
        </button>
        <p v-if="passError" class="text-sm text-danger-400">{{ passError }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import PasswordField from '@/components/PasswordField.vue'
import { useApi } from '@/composables/useApi'
import { useSwal } from '@/composables/useSwal'
import { useSettings } from '@/composables/useSettings'
import { useAuthStore } from '@/stores/auth'

document.title = `Settings - ${useSettings().appName()}`

const { apiPut, apiUpload } = useApi()
const swal = useSwal()
const { fetchSettings, appName } = useSettings()
const authStore = useAuthStore()

const activeTab = ref('general')
const tabs = [
  { key: 'general', label: 'General' },
  { key: 'profile', label: 'Profile' },
]

// General form
const saving = ref(false)
const saved = ref(false)
const form = reactive({ app_name: '', app_url: '', logo_path: '', favicon_path: '' })
const logoPreview = ref('')
const faviconPreview = ref('')

// Email form
const emailForm = reactive({ new_email: '', password: '' })
const emailLoading = ref(false)
const emailError = ref('')

// Password form
const passForm = reactive({ old_password: '', new_password: '' })
const passLoading = ref(false)
const passError = ref('')

async function loadSettings() {
  const s = await fetchSettings()
  form.app_name = s?.app_name || ''
  form.app_url = s?.app_url || ''
  form.logo_path = s?.logo_path || ''
  form.favicon_path = s?.favicon_path || ''
  logoPreview.value = form.logo_path ? '/' + form.logo_path : ''
  faviconPreview.value = form.favicon_path ? '/' + form.favicon_path : ''
}

async function saveSettings() {
  saving.value = true
  saved.value = false
  try {
    await apiPut('/settings', { app_name: form.app_name, app_url: form.app_url })
    await fetchSettings()
    saved.value = true
    setTimeout(() => { saved.value = false }, 3000)
    document.title = `${appName()} - Settings`
  } catch (e: any) {
    swal.error(e.message)
  } finally {
    saving.value = false
  }
}

async function uploadFile(endpoint: string, file: File, previewRef: typeof logoPreview, pathRef: keyof typeof form) {
  const fd = new FormData()
  fd.append('file', file)
  try {
    const res = await apiUpload<{ ok: boolean; path: string }>(endpoint, fd)
    previewRef.value = '/' + res.path
    form[pathRef] = res.path as any
    await fetchSettings()
  } catch (e: any) {
    swal.error(e.message)
  }
}

function onLogoChange(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  uploadFile('/settings/upload-logo', file, logoPreview, 'logo_path')
}

function onFaviconChange(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  uploadFile('/settings/upload-favicon', file, faviconPreview, 'favicon_path')
}

function removeLogo() {
  form.logo_path = ''
  logoPreview.value = ''
  apiPut('/settings', { logo_path: '' }).then(() => fetchSettings()).catch(() => {})
}

function removeFavicon() {
  form.favicon_path = ''
  faviconPreview.value = ''
  apiPut('/settings', { favicon_path: '' }).then(() => fetchSettings()).catch(() => {})
}

async function changeEmail() {
  emailError.value = ''
  emailLoading.value = true
  try {
    const res = await apiPut<any>('/auth/email', emailForm)
    swal.success('Email berhasil diubah')
    if (res.access_token) {
      authStore.token = res.access_token
      authStore.expiresAt = res.expires_at
      localStorage.setItem('drp_token', res.access_token)
      localStorage.setItem('drp_expires', res.expires_at)
    }
    if (res.user) {
      authStore.user = res.user
      localStorage.setItem('drp_user', JSON.stringify(res.user))
    }
    emailForm.new_email = ''
    emailForm.password = ''
  } catch (e: any) {
    emailError.value = e.message
  } finally {
    emailLoading.value = false
  }
}

async function changePassword() {
  passError.value = ''
  passLoading.value = true
  try {
    await apiPut('/auth/password', passForm)
    swal.success('Password berhasil diubah')
    passForm.old_password = ''
    passForm.new_password = ''
  } catch (e: any) {
    passError.value = e.message
  } finally {
    passLoading.value = false
  }
}

onMounted(loadSettings)
</script>
