<template>
  <div class="space-y-6 animate-fade-in">
    <div>
      <h1 class="page-title">Activity Log</h1>
      <p class="page-subtitle">Riwayat perubahan & aktivitas sistem</p>
    </div>

    <!-- Filters -->
    <div class="card p-4 space-y-3">
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-3">
        <div>
          <label class="label">Aksi</label>
          <select v-model="filter.action" class="input w-full">
            <option value="">Semua aksi</option>
            <option v-for="a in actions" :key="a.value" :value="a.value">{{ a.label }}</option>
          </select>
        </div>
        <div>
          <label class="label">Cari</label>
          <input v-model="filter.search" class="input w-full" placeholder="target / aksi" @keyup.enter="load" />
        </div>
        <div>
          <label class="label">Server</label>
          <select v-model="filter.server_id" class="input w-full">
            <option value="">Semua server</option>
            <option v-for="s in servers" :key="s.id" :value="s.id">{{ s.name }}</option>
          </select>
        </div>
        <div>
          <label class="label">Dari</label>
          <input v-model="filter.date_from" type="date" class="input w-full" />
        </div>
        <div>
          <label class="label">Sampai</label>
          <input v-model="filter.date_to" type="date" class="input w-full" />
        </div>
      </div>
      <div class="flex items-center gap-3">
        <button class="btn-primary text-sm" @click="load"><svg class="w-4 h-4 inline -mt-0.5 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>Cari</button>
        <button class="btn-ghost text-sm" @click="resetFilter">Reset</button>
        <span class="text-xs text-ink-500 ml-auto">{{ total }} entri</span>
      </div>
    </div>

    <!-- Table -->
    <div class="card overflow-hidden">
      <div v-if="logs.length === 0" class="px-5 py-12 text-center">
        <div class="w-12 h-12 rounded-full bg-ink-800 mx-auto flex items-center justify-center mb-3">
          <svg class="w-6 h-6 text-ink-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" /></svg>
        </div>
        <p class="text-sm text-ink-500">Belum ada aktivitas.</p>
      </div>
      <table v-else class="table-base">
        <thead>
          <tr>
            <th>Waktu</th>
            <th>User</th>
            <th>Aksi</th>
            <th>Target</th>
            <th>Server</th>
            <th>Sumber</th>
            <th>Detail</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="log in logs" :key="log.id" class="cursor-pointer hover:bg-ink-800/50" @click="showDetail(log)">
            <td class="font-mono text-xs text-ink-500 whitespace-nowrap">{{ fmt(log.created_at) }}</td>
            <td class="text-sm text-ink-300 max-w-[120px] truncate" :title="log.user_email">{{ log.user_email || '-' }}</td>
            <td><span :class="badgeClass(log.action)">{{ actionLabel(log.action) }}</span></td>
            <td class="text-sm text-ink-100 font-mono max-w-[150px] truncate" :title="log.target">{{ log.target || '-' }}</td>
            <td class="text-sm text-ink-400 max-w-[120px] truncate" :title="log.server_name">{{ log.server_name || '-' }}</td>
            <td><span :class="sourceBadge(log.source)">{{ log.source }}</span></td>
            <td class="text-xs text-ink-500 max-w-[160px] truncate" :title="log.detail">{{ log.detail || '-' }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    <div v-if="total > limit" class="flex items-center justify-between">
      <span class="text-xs text-ink-500">{{ offset + 1 }}-{{ Math.min(offset + limit, total) }} dari {{ total }}</span>
      <div class="flex gap-2">
        <button class="btn-ghost text-sm px-3 py-1" :disabled="offset === 0" @click="prevPage">← Sebelumnya</button>
        <button class="btn-ghost text-sm px-3 py-1" :disabled="offset + limit >= total" @click="nextPage">Selanjutnya →</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useApi } from '@/composables/useApi'
import { useSwal } from '@/composables/useSwal'
import { useSettings } from '@/composables/useSettings'

document.title = `Activity Log - ${useSettings().appName()}`

const { apiGet } = useApi()
const swal = useSwal()

const logs = ref<any[]>([])
const servers = ref<any[]>([])
const total = ref(0)
const limit = ref(50)
const offset = ref(0)

const filter = reactive({
  action: '',
  search: '',
  server_id: '',
  date_from: '',
  date_to: '',
})

const actions = [
  { value: 'voucher.generate', label: 'Voucher Generate' },
  { value: 'voucher.disable', label: 'Voucher Disable' },
  { value: 'voucher.enable', label: 'Voucher Enable' },
  { value: 'voucher.delete', label: 'Voucher Delete' },
  { value: 'voucher.expired', label: 'Voucher Expired' },
  { value: 'voucher.sync_remove', label: 'Voucher Sync Remove' },
  { value: 'server.create', label: 'Server Create' },
  { value: 'server.update', label: 'Server Update' },
  { value: 'server.delete', label: 'Server Delete' },
  { value: 'profile.create', label: 'Profile Create' },
  { value: 'profile.delete', label: 'Profile Delete' },
  { value: 'hotspot.kick', label: 'Hotspot Kick' },
  { value: 'member.create', label: 'Member Create' },
  { value: 'member.disable', label: 'Member Disable' },
  { value: 'member.enable', label: 'Member Enable' },
  { value: 'member.delete', label: 'Member Delete' },
  { value: 'voucher.cleanup_deleted', label: 'Cleanup Hapus' },
  { value: 'auth.login', label: 'Login' },
  { value: 'auth.register', label: 'Register' },
  { value: 'auth.change_email', label: 'Ubah Email' },
  { value: 'auth.change_password', label: 'Ubah Password' },
  { value: 'token.create', label: 'API Token Create' },
  { value: 'token.delete', label: 'API Token Delete' },
  { value: 'settings.update', label: 'Settings Update' },
]

function actionLabel(action: string): string {
  const a = actions.find(a => a.value === action)
  return a ? a.label : action
}

function badgeClass(action: string): string {
  const creates = ['voucher.generate', 'server.create', 'profile.create', 'token.create', 'auth.register', 'member.create']
  const deletes = ['voucher.delete', 'server.delete', 'profile.delete', 'token.delete', 'member.delete']
  if (action.startsWith('auth.login')) return 'badge-online'
  if (creates.includes(action)) return 'badge-online'
  if (deletes.includes(action)) return 'badge-failed'
  if (action.endsWith('.disable') || action.endsWith('.expired') || action === 'voucher.cleanup_deleted') return 'badge-unknown'
  if (action.endsWith('.enable')) return 'badge-online'
  return 'badge-unknown'
}

function sourceBadge(s: string): string {
  if (s === 'web') return 'badge-online'
  if (s === 'api') return 'badge-unknown'
  return 'badge-unknown'
}

function buildQuery(): string {
  const p = new URLSearchParams()
  p.set('limit', String(limit.value))
  p.set('offset', String(offset.value))
  if (filter.action) p.set('action', filter.action)
  if (filter.search) p.set('search', filter.search)
  if (filter.server_id) p.set('server_id', filter.server_id)
  if (filter.date_from) p.set('date_from', new Date(filter.date_from).toISOString())
  if (filter.date_to) p.set('date_to', new Date(filter.date_to + 'T23:59:59').toISOString())
  return p.toString()
}

async function load() {
  try {
    const res = await apiGet<{ data: any[]; total: number }>(`/audit-logs?${buildQuery()}`)
    logs.value = res.data || []
    total.value = res.total || 0
  } catch (e: any) { swal.error(e.message) }
}

async function loadServers() {
  try {
    const res = await apiGet<{ data: any[] }>('/servers')
    servers.value = res.data || []
  } catch {}
}

function resetFilter() {
  filter.action = ''
  filter.search = ''
  filter.server_id = ''
  filter.date_from = ''
  filter.date_to = ''
  offset.value = 0
  load()
}

function prevPage() {
  offset.value = Math.max(0, offset.value - limit.value)
  load()
}

function nextPage() {
  offset.value += limit.value
  load()
}

function fmt(t: string) {
  return new Date(t).toLocaleString('id-ID', { dateStyle: 'short', timeStyle: 'medium' })
}

function showDetail(log: any) {
  const html = `
    <div class="text-left" style="font-size:13px">
      <div class="grid grid-cols-[100px_1fr] gap-x-3 gap-y-2">
        <span style="color:#8DA0B8">Waktu</span>
        <span style="color:#D0DAE5;font-family:monospace">${fmt(log.created_at)}</span>
        <span style="color:#8DA0B8">Action</span>
        <span style="color:#D0DAE5">${actionLabel(log.action)}</span>
        <span style="color:#8DA0B8">Target</span>
        <span style="color:#D0DAE5;font-family:monospace;word-break:break-all">${escHtml(log.target || '-')}</span>
        <span style="color:#8DA0B8">User</span>
        <span style="color:#D0DAE5">${escHtml(log.user_email || '-')}</span>
        <span style="color:#8DA0B8">Server</span>
        <span style="color:#D0DAE5">${escHtml(log.server_name || '-')}</span>
        <span style="color:#8DA0B8">Source</span>
        <span style="color:#D0DAE5">${escHtml(log.source)}</span>
        <span style="color:#8DA0B8">Detail</span>
        <span style="color:#D0DAE5;font-family:monospace;font-size:12px;word-break:break-all;white-space:pre-wrap">${escHtml(log.detail || '-')}</span>
      </div>
    </div>
  `
  swal.Swal.fire({
    title: 'Detail Audit Log',
    html,
    confirmButtonText: 'Tutup',
    width: 600,
    background: '#26344A',
    color: '#D0DAE5',
    confirmButtonColor: '#2563EB',
    customClass: {
      popup: 'rounded-xl border border-ink-700 shadow-card',
      confirmButton: 'rounded-lg px-4 py-2 text-sm font-medium',
    },
  })
}

function escHtml(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;')
}

onMounted(() => { load(); loadServers() })
</script>
