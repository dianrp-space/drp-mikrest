<template>
  <div class="card p-6 space-y-5 animate-slide-up">
    <div class="flex items-center gap-3">
      <div class="w-10 h-10 rounded-lg bg-accent-600/15 flex items-center justify-center">
        <svg class="w-5 h-5 text-accent-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M7 7h.01M7 3h5a1.99 1.99 0 011.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.99 1.99 0 013 12V7a4 4 0 014-4z" /></svg>
      </div>
      <div>
        <h2 class="font-semibold text-ink-100">Generate Voucher</h2>
        <p class="text-xs text-ink-500">Voucher dibuat di DB & di-push ke RouterOS</p>
      </div>
    </div>
    <form class="grid grid-cols-1 sm:grid-cols-2 gap-4" @submit.prevent="onSubmit">
      <div>
        <label class="label">Profile</label>
        <select v-model="form.profile_id" class="input">
          <option :value="null">- tanpa profile -</option>
          <option v-for="p in profiles" :key="p.id" :value="p.id">{{ p.name }} ({{ p.rate_limit || 'no-limit' }})</option>
        </select>
      </div>
      <div>
        <label class="label">Jumlah</label>
        <input v-model.number="form.count" type="number" min="1" max="500" required class="input" />
      </div>

      <!-- Limit Uptime (masa aktif) -->
      <div class="sm:col-span-2 card p-3 border-ink-700 bg-ink-900/50 space-y-2">
        <div class="flex items-center justify-between">
          <label class="text-sm font-medium text-ink-200 flex items-center gap-2">
            <svg class="w-4 h-4 text-brand-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
            Masa Aktif Voucher
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
        <p v-if="limitEnabled" class="text-xs text-ink-500 flex items-center gap-1.5 mt-1">
          <svg class="w-3.5 h-3.5 text-ink-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
          Masa aktif dihitung sejak user login pertama kali ke hotspot
        </p>
      </div>

      <!-- Limit Bytes (opsional) -->
      <div>
        <label class="label">Limit Kuota (opsional)</label>
        <input v-model="form.limit_bytes" class="input" placeholder="500M / 1G (kosongkan = unlimited)" />
      </div>

      <div>
        <label class="label">Mode Username</label>
        <p class="text-xs text-ink-500 mb-2">Password = username (voucher). Username & password berbeda = member.</p>
        <select v-model="form.username_mode" class="input">
          <option value="random">Random (acak)</option>
          <option value="prefix">Prefix + nomor urut</option>
        </select>
      </div>
      <div v-if="form.username_mode === 'random'">
        <label class="label">Pola (opsional)</label>
        <input v-model="form.pattern" class="input" placeholder="####-####" />
        <p class="text-xs text-ink-600 mt-1"># diganti karakter acak</p>
      </div>
      <div v-else>
        <label class="label">Prefix</label>
        <input v-model="form.prefix" class="input" placeholder="HOTSPOT" />
      </div>
      <div class="sm:col-span-2"><label class="label">Comment (opsional)</label><input v-model="form.comment" class="input" placeholder="1hour-Rp5000" /></div>

      <div v-if="error" class="sm:col-span-2 text-sm text-danger-400 bg-danger-500/10 px-3 py-2 rounded-lg">{{ error }}</div>
      <div v-if="result" class="sm:col-span-2 text-sm bg-accent-500/10 border border-accent-500/20 px-3 py-2 rounded-lg">
        <div class="text-accent-400 font-medium">Dibuat {{ result.vouchers.length }} voucher, gagal {{ (result.failed || []).length }}.</div>
        <details class="mt-2">
          <summary class="cursor-pointer text-xs text-accent-400">Lihat daftar</summary>
          <pre class="mt-2 text-xs bg-ink-900 p-2 rounded font-mono text-ink-300 max-h-40 overflow-auto">{{ result.vouchers.map((v: any) => `${v.username}`).join('\n') }}</pre>
        </details>
      </div>
      <div class="sm:col-span-2 flex gap-3">
        <button type="submit" :disabled="loading" class="btn-primary">
          <svg v-if="loading" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
          {{ loading ? 'Generate...' : 'Generate' }}
        </button>
        <button type="button" class="btn-secondary" @click="$emit('cancel')">Tutup</button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { useApi } from '@/composables/useApi'
import { useSwal } from '@/composables/useSwal'

const props = defineProps<{ serverId: string; profiles: any[] }>()
const emit = defineEmits<{ generated: []; cancel: [] }>()

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
  profile_id: null as string | null,
  count: 10,
  pattern: '',
  prefix: '',
  username_mode: 'random',
  comment: '',
  limit_uptime: '1d',
  limit_bytes: '',
})
const limitEnabled = ref(true)
const loading = ref(false)
const error = ref('')
const result = ref<any>(null)

// toggle limit: kosongkan field kalau disable
watch(limitEnabled, (v) => { if (!v) form.limit_uptime = ''; else if (!form.limit_uptime) form.limit_uptime = '1d' })

async function onSubmit() {
  loading.value = true; error.value = ''; result.value = null
  try {
    const payload: any = { ...form }
    if (!form.pattern) delete payload.pattern
    if (!form.prefix) delete payload.prefix
    if (!form.profile_id) delete payload.profile_id
    if (!limitEnabled.value || !form.limit_uptime) delete payload.limit_uptime
    if (!form.limit_bytes) delete payload.limit_bytes
    result.value = await apiPost('/vouchers/generate', payload)
    const ok = (result.value.vouchers || []).length
    const fail = (result.value.failed || []).length
    if (fail === 0) swal.success(`${ok} voucher berhasil dibuat`)
    else if (ok === 0) swal.errorDialog('Generate Gagal', `${fail} voucher gagal dibuat.`)
    else swal.warning(`${ok} berhasil, ${fail} gagal`)
    emit('generated')
  } catch (e: any) { error.value = e.message } finally { loading.value = false }
}
</script>
