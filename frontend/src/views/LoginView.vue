<template>
  <div class="card p-8 animate-slide-up">
    <div class="text-center mb-8">
      <div v-if="settings.logoUrl()" class="w-16 h-16 rounded-2xl flex items-center justify-center mx-auto mb-4 overflow-hidden">
        <img :src="settings.logoUrl()" class="w-full h-full object-contain" />
      </div>
      <div v-else class="inline-flex items-center justify-center w-14 h-14 rounded-2xl bg-brand-600 shadow-glow mb-4">
        <svg class="w-8 h-8 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
        </svg>
      </div>
      <h1 class="text-2xl font-bold text-ink-100">{{ settings.appName() }}</h1>
      <p class="text-sm text-ink-500 mt-1">Mikrotik RouterOS Manager</p>
    </div>

    <form class="space-y-4" @submit.prevent="onSubmit">
      <div>
        <label class="label" for="email">Email</label>
        <input id="email" v-model="email" type="email" required autocomplete="email"
          class="input" placeholder="admin@example.com" />
      </div>
      <div>
        <label class="label" for="password">Password</label>
        <PasswordField id="password" v-model="password" required autocomplete="current-password" placeholder="••••••••" />
      </div>
      <div v-if="error" class="text-sm text-danger-400 bg-danger-500/10 border border-danger-500/20 px-3 py-2 rounded-lg flex items-start gap-2">
        <svg class="w-4 h-4 mt-0.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span>{{ error }}</span>
      </div>
      <button type="submit" :disabled="loading" class="btn-primary w-full py-2.5">
        <svg v-if="loading" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
        <span>{{ loading ? 'Memproses...' : 'Masuk' }}</span>
      </button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSettings, initSettings } from '@/composables/useSettings'
import PasswordField from '@/components/PasswordField.vue'

const settings = useSettings()
settings.updateFavicon()
document.title = `Login - ${settings.appName()}`

const authStore = useAuthStore()
const router = useRouter()
const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

onMounted(async () => {
  authStore.init()
  if (authStore.isAuthenticated) router.push('/')
  await settings.fetchSettings()
})

async function onSubmit() {
  error.value = ''
  loading.value = true
  try {
    await authStore.login(email.value, password.value)
    initSettings()
    router.push('/')
  } catch (e: any) {
    error.value = e?.message || 'Login gagal'
  } finally {
    loading.value = false
  }
}
</script>
