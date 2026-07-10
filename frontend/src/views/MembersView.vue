<template>
  <div class="space-y-6 animate-fade-in">
    <div class="flex items-center justify-between flex-wrap gap-3">
      <div class="flex gap-2">
        <button class="btn-secondary text-xs" :disabled="syncing" @click="syncMembers">
          <svg v-if="syncing" class="animate-spin w-3.5 h-3.5" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
          <svg v-else class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
          {{ syncing ? 'Menyinkronkan...' : 'Sync dari Router' }}
        </button>
        <button class="btn-primary text-xs" @click="showForm = !showForm">
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" /></svg>
          {{ showForm ? 'Batal' : 'Buat Member' }}
        </button>
      </div>
    </div>

    <MemberForm v-if="showForm" :server-id="id" :profiles="profiles" @saved="onSaved" @cancel="showForm = false" />

    <div class="card overflow-hidden">
      <div class="px-5 py-3 border-b border-ink-800 flex flex-wrap gap-2 items-center">
        <input v-model="filter.q" placeholder="Cari username/comment..." class="input w-64" @keyup.enter="load" />
        <select v-model="filter.status" class="input w-auto" @change="load">
          <option value="">Semua</option>
          <option value="active">Active</option>
          <option value="used">Used</option>
          <option value="disabled">Disabled</option>
          <option value="expired">Expired</option>
          <option value="failed">Failed</option>
        </select>
        <button class="btn-secondary text-xs" @click="load">Filter</button>
        <span class="ml-auto text-xs text-ink-500 font-mono">{{ total }} member</span>
      </div>
      <div v-if="loading" class="px-5 py-12 text-center text-sm text-ink-500">Memuat...</div>
      <div v-else-if="members.length === 0" class="px-5 py-12 text-center">
        <div class="w-12 h-12 rounded-full bg-ink-800 mx-auto flex items-center justify-center mb-3">
          <svg class="w-6 h-6 text-ink-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" /></svg>
        </div>
        <p class="text-sm text-ink-500">Belum ada member. Sync dari router untuk mengimpor hotspot user dengan username != password.</p>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="table-base">
          <thead><tr><th>Username</th><th>Password</th><th>Limit</th><th>Status</th><th>Expired</th><th>Dibuat</th><th class="text-right">Aksi</th></tr></thead>
          <tbody>
            <tr v-for="v in members" :key="v.id" :class="v.status === 'disabled' ? 'opacity-50' : ''">
              <td class="font-mono font-medium text-ink-100">{{ v.username }}</td>
              <td class="font-mono text-ink-400">
                <span class="flex items-center gap-1.5">
                  <span>{{ visiblePasswords.has(v.id) ? v.password : '••••••••' }}</span>
                  <button class="btn-ghost px-1 py-0.5 text-ink-500 hover:text-ink-200" @click="togglePassword(v.id)" :title="visiblePasswords.has(v.id) ? 'Sembunyikan' : 'Lihat password'">
                    <svg v-if="visiblePasswords.has(v.id)" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" /></svg>
                    <svg v-else class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path stroke-linecap="round" stroke-linejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg>
                  </button>
                </span>
              </td>
              <td class="text-xs text-ink-500 font-mono">
                <span v-if="v.limit_uptime" :class="expiryClass(v)">{{ v.limit_uptime }}</span>
                <span v-else class="text-ink-600">-</span>
              </td>
              <td><span :class="`badge-${displayStatus(v)}`">{{ displayStatus(v) }}</span></td>
              <td class="text-xs font-mono" :class="expiryClass(v)">
                <span v-if="v.expires_at">{{ fmt(v.expires_at) }}</span>
                <span v-else-if="v.limit_uptime" class="text-ink-500 italic">setelah login</span>
                <span v-else class="text-ink-600">-</span>
              </td>
              <td class="text-xs text-ink-500 font-mono">{{ fmt(v.created_at) }}</td>
              <td class="text-right">
                <div class="inline-flex gap-1">
                  <template v-if="isExpired(v) || v.status === 'expired'">
                    <span class="text-xs text-ink-600 italic px-2 py-1">-</span>
                  </template>
                  <template v-else>
                    <button v-if="v.status !== 'disabled'" class="btn-ghost px-2 py-1 text-xs" @click="setMember(v.id, 'disable')">Disable</button>
                    <button v-else class="btn-ghost px-2 py-1 text-xs" @click="setMember(v.id, 'enable')">Enable</button>
                  </template>
                  <button class="btn-ghost px-2 py-1 text-xs text-danger-400" @click="delMember(v.id, v.username)">Hapus</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useApi } from '@/composables/useApi'
import { useSwal } from '@/composables/useSwal'
import MemberForm from '@/components/MemberForm.vue'

const route = useRoute()
const id = computed(() => route.params.id as string)
const { apiGet, apiPost, apiDelete } = useApi()
const swal = useSwal()
const showForm = ref(false)
const loading = ref(true); const syncing = ref(false)
const members = ref<any[]>([]); const total = ref(0); const profiles = ref<any[]>([])
const visiblePasswords = ref(new Set<string>())
const filter = reactive({ q: '', status: '' })

function togglePassword(id: string) {
  const s = visiblePasswords.value
  if (s.has(id)) s.delete(id); else s.add(id)
}
function fmt(t: string) { return new Date(t).toLocaleString('id-ID', { dateStyle: 'short', timeStyle: 'short' }) }
function isExpired(v: any): boolean {
  return !!v.expires_at && new Date(v.expires_at).getTime() < Date.now()
}
function displayStatus(v: any): string {
  if (isExpired(v)) return 'expired'
  return v.status || 'unknown'
}
function expiryClass(v: any): string {
  if (!v.expires_at) return ''
  const exp = new Date(v.expires_at).getTime()
  const now = Date.now()
  if (exp < now) return 'text-danger-400'
  if (exp - now < 60 * 60 * 1000) return 'text-warn-400'
  return 'text-ink-500'
}

async function load() {
  loading.value = true
  try {
    const q = new URLSearchParams({ server_id: id.value, limit: '100', type: 'member' })
    if (filter.q) q.set('q', filter.q); if (filter.status) q.set('status', filter.status)
    const res = await apiGet<{ data: any[]; total: number }>(`/vouchers?${q.toString()}`)
    members.value = res.data || []; total.value = res.total || 0
  } catch (e: any) { swal.error(e.message) } finally { loading.value = false }
}
async function loadProfiles() { try { const res = await apiGet<{ data: any[] }>(`/servers/${id.value}/profiles`); profiles.value = res.data || [] } catch {} }
function onSaved() { showForm.value = false; load() }
async function syncMembers() {
  syncing.value = true
  try {
    const res = await apiPost<{ updated: number; imported: number; removed: number }>(`/servers/${id.value}/vouchers/sync`)
    if (res.removed > 0) {
      const ok = await swal.confirm(
        'Sinkronisasi Selesai',
        `<div class="text-left space-y-1 text-sm">
           <div><b class="text-brand-400">${res.updated}</b> member diupdate</div>
           <div><b class="text-accent-400">${res.imported}</b> member diimpor dari router</div>
           <div><b class="text-danger-400">${res.removed}</b> member dihapus dari DB (tidak ada di router)</div>
           <div class="text-xs text-ink-500 mt-2">DB sekarang mirror router.</div>
         </div>`,
        { icon: 'warning', confirmButtonText: 'OK, Mengerti' }
      )
      void ok
    } else {
      swal.successDialog(
        'Sync Selesai',
        `<div class="text-left text-sm space-y-1">
           <div><b class="text-brand-400">${res.updated}</b> diupdate</div>
           <div><b class="text-accent-400">${res.imported}</b> diimpor</div>
         </div>`
      )
    }
    load()
  } catch (e: any) { swal.errorDialog('Gagal Sync', e.message) } finally { syncing.value = false }
}
async function setMember(vid: string, action: 'disable' | 'enable') {
  try { await apiPost(`/vouchers/${vid}/${action}`); swal.success(`Member ${action === 'disable' ? 'disabled' : 'enabled'}`); load() }
  catch (e: any) { swal.error(e.message) }
}
async function delMember(vid: string, name: string) {
  const ok = await swal.confirm(`Hapus member ${name}?`, 'Member akan dihapus dari DB dan RouterOS.')
  if (!ok) return
  try { await apiDelete(`/vouchers/${vid}`); swal.success('Member dihapus'); load() }
  catch (e: any) { swal.error(e.message) }
}
onMounted(() => { load(); loadProfiles() })
</script>
