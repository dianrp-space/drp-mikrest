<template>
  <div class="card p-6 space-y-5 animate-slide-up">
    <div class="flex items-center gap-3">
      <div class="w-10 h-10 rounded-lg bg-accent-600/15 flex items-center justify-center">
        <svg class="w-5 h-5 text-accent-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" /></svg>
      </div>
      <div>
        <h2 class="font-semibold text-ink-100">Tambah Member</h2>
        <p class="text-xs text-ink-500">Member dibuat di DB & di-push ke RouterOS. Username & password harus berbeda.</p>
      </div>
    </div>
    <form class="grid grid-cols-1 sm:grid-cols-2 gap-4" @submit.prevent="onSubmit">
      <div>
        <label class="label">Username</label>
        <input v-model="form.username" required class="input" placeholder="nama-member" />
      </div>
      <div>
        <label class="label">Password</label>
        <input v-model="form.password" required class="input" placeholder="password berbeda" />
      </div>

      <div>
        <label class="label">Profile</label>
        <select v-model="form.profile_id" class="input">
          <option :value="null">- tanpa profile -</option>
          <option v-for="p in profiles" :key="p.id" :value="p.id">{{ p.name }} ({{ p.rate_limit || 'no-limit' }})</option>
        </select>
      </div>
      <div><label class="label">Comment (opsional)</label><input v-model="form.comment" class="input" placeholder="member-name" /></div>

      <!-- Limit Uptime -->
      <div class="sm:col-span-2 card p-3 border-ink-700 bg-ink-900/50 space-y-2">
        <div class="flex items-center justify-between">
          <label class="text-sm font-medium text-ink-200 flex items-center gap-2">
            <svg class="w-4 h-4 text-brand-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
            Masa Aktif
          </label>
          <label class="flex items-center gap-1.5 text-xs text-ink-400 cursor-pointer">
            <input v-model="limitEnabled" type="checkbox" class="rounded border-ink-600 bg-ink-800 text-brand-500 focus:ring-brand-500" />
            Aktifkan limit
          </label>
        </div>
        <div v-if="limitEnabled" class="grid grid-cols-1 sm:grid-cols-2 gap-2">
          <div>
            <label class="text-xs text-ink-500 mb-1 block">Preset</label>
            <div class="flex flex-wrap gap-1.5">
              <button
                v-for="p in presets"
                :key="p.value"
                type="button"
                class="px-2.5 py-1 text-xs rounded-md border transition-colors"
                :class="form.limit_uptime === p.value
                  ? 'bg-brand-600 text-white border-brand-600'
                  : 'bg-ink-800 text-ink-300 border-ink-700 hover:bg-ink-700 hover:border-ink-600'"
                @click="form.limit_uptime = p.value"
              >{{ p.label }}</button>
            </div>
          </div>
          <div>
            <label class="text-xs text-ink-500 mb-1 block">Custom (format RouterOS)</label>
            <input v-model="form.limit_uptime" class="input text-sm font-mono" placeholder="1d12h" />
            <p class="text-xs text-ink-600 mt-1">w=minggu, d=hari, h=jam, m=menit. Cth: 30m, 1h, 1d, 1d12h</p>
          </div>
        </div>
      </div>

      <!-- Limit Bytes -->
      <div>
        <label class="label">Limit Kuota (opsional)</label>
        <input v-model="form.limit_bytes" class="input" placeholder="500M / 1G (kosongkan = unlimited)" />
      </div>

      <div v-if="error" class="sm:col-span-2 text-sm text-danger-400 bg-danger-500/10 px-3 py-2 rounded-lg">{{ error }}</div>

      <div class="sm:col-span-2 flex gap-3">
        <button type="submit" :disabled="loading" class="btn-primary">
          <svg v-if="loading" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
          {{ loading ? 'Menyimpan...' : 'Simpan' }}
        </button>
        <button type="button" class="btn-secondary" @click="$emit('cancel')">Batal</button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { useApi } from '@/composables/useApi'
import { useSwal } from '@/composables/useSwal'

const props = defineProps<{ serverId: string; profiles: any[] }>()
const emit = defineEmits<{ saved: []; cancel: [] }>()

const { apiPost } = useApi()
const swal = useSwal()

const presets = [
  { label: '30 menit', value: '30m' },
  { label: '1 jam',   value: '1h' },
  { label: '3 jam',   value: '3h' },
  { label: '12 jam',  value: '12h' },
  { label: '1 hari',  value: '1d' },
  { label: '3 hari',  value: '3d' },
  { label: '7 hari',  value: '7d' },
  { label: '30 hari', value: '30d' },
]

const form = reactive({
  server_id: props.serverId,
  username: '',
  password: '',
  profile_id: null as string | null,
  comment: '',
  limit_uptime: '1d',
  limit_bytes: '',
})
const limitEnabled = ref(true)
const loading = ref(false)
const error = ref('')

watch(limitEnabled, (v) => { if (!v) form.limit_uptime = ''; else if (!form.limit_uptime) form.limit_uptime = '1d' })

async function onSubmit() {
  loading.value = true; error.value = ''
  try {
    const payload: any = { ...form }
    if (!payload.profile_id) delete payload.profile_id
    if (!limitEnabled.value || !form.limit_uptime) delete payload.limit_uptime
    if (!form.limit_bytes) delete payload.limit_bytes
    await apiPost('/vouchers/member', payload)
    swal.success(`Member "${form.username}" berhasil dibuat`)
    emit('saved')
  } catch (e: any) { error.value = e.message } finally { loading.value = false }
}
</script>
