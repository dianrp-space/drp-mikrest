<template>
  <div class="space-y-6 animate-fade-in">
    <div class="flex items-center justify-between flex-wrap gap-3">
      <div class="flex gap-2">
        <button class="btn-secondary text-xs" :disabled="syncing" @click="syncVouchers">
          <svg v-if="syncing" class="animate-spin w-3.5 h-3.5" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
          <svg v-else class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
          {{ syncing ? 'Menyinkronkan...' : 'Sync dari Router' }}
        </button>
        <button class="btn-primary text-xs" @click="showGen = !showGen">
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" /></svg>
          {{ showGen ? 'Batal' : 'Buat Voucher' }}
        </button>
      </div>
    </div>

    <GenerateDialog v-if="showGen" :server-id="id" :profiles="profiles" @generated="onGenerated" @cancel="showGen = false" />

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
        <span class="ml-auto text-xs text-ink-500 font-mono">{{ total }} voucher</span>
      </div>
      <div v-if="loading" class="px-5 py-12 text-center text-sm text-ink-500">Memuat...</div>
      <div v-else-if="vouchers.length === 0" class="px-5 py-12 text-center">
        <div class="w-12 h-12 rounded-full bg-ink-800 mx-auto flex items-center justify-center mb-3">
          <svg class="w-6 h-6 text-ink-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M7 7h.01M7 3h5a1.99 1.99 0 011.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.99 1.99 0 013 12V7a4 4 0 014-4z" /></svg>
        </div>
        <p class="text-sm text-ink-500">Belum ada voucher.</p>
        <button class="text-sm text-brand-400 hover:text-brand-300 mt-2" @click="showGen = true">+ Generate Voucher</button>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="table-base">
          <thead><tr><th>Voucher</th><th>Limit</th><th>Status</th><th>Expired</th><th>First Login</th><th>Dibuat</th><th class="text-right">Aksi</th></tr></thead>
          <tbody>
            <tr v-for="v in vouchers" :key="v.id" :class="v.status === 'disabled' ? 'opacity-50' : ''">
              <td class="font-mono font-medium text-ink-100">{{ v.username }}</td>
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
              <td class="text-xs font-mono text-ink-400">
                <span v-if="v.comment && /^\d{4}-\d{2}-\d{2}/.test(v.comment)">{{ v.comment }}</span>
                <span v-else class="text-ink-600 italic">Never used</span>
              </td>
              <td class="text-xs text-ink-500 font-mono">{{ fmt(v.created_at) }}</td>
              <td class="text-right">
                <div class="inline-flex gap-1">
                  <template v-if="isExpired(v) || v.status === 'expired'">
                    <span class="text-xs text-ink-600 italic px-2 py-1">-</span>
                  </template>
                  <template v-else>
                    <button v-if="v.status !== 'disabled'" class="btn-ghost px-2 py-1 text-xs" @click="setVoucher(v.id, 'disable')">Disable</button>
                    <button v-else class="btn-ghost px-2 py-1 text-xs" @click="setVoucher(v.id, 'enable')">Enable</button>
                  </template>
                  <button class="btn-ghost px-2 py-1 text-xs text-danger-400" @click="delVoucher(v.id, v.username)">Hapus</button>
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
import GenerateDialog from '@/components/GenerateDialog.vue'

const route = useRoute()
const id = computed(() => route.params.id as string)
const { apiGet, apiPost, apiDelete } = useApi()
const swal = useSwal()
const showGen = ref(false); const loading = ref(true); const syncing = ref(false)
const vouchers = ref<any[]>([]); const total = ref(0); const profiles = ref<any[]>([])
const filter = reactive({ q: '', status: '' })

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
    const q = new URLSearchParams({ server_id: id.value, limit: '100', type: 'voucher' })
    if (filter.q) q.set('q', filter.q); if (filter.status) q.set('status', filter.status)
    const res = await apiGet<{ data: any[]; total: number }>(`/vouchers?${q.toString()}`)
    vouchers.value = res.data || []; total.value = res.total || 0
  } catch (e: any) { swal.error(e.message) } finally { loading.value = false }
}
async function loadProfiles() { try { const res = await apiGet<{ data: any[] }>(`/servers/${id.value}/profiles`); profiles.value = res.data || [] } catch {} }
async function syncVouchers() {
  syncing.value = true
  try {
    const res = await apiPost<{ updated: number; imported: number; removed: number }>(`/servers/${id.value}/vouchers/sync`)

    // kalau ada yang dihapus, tampilkan detail untuk transparansi
    if (res.removed > 0) {
      const ok = await swal.confirm(
        'Sinkronisasi Selesai',
        `<div class="text-left space-y-1 text-sm">
           <div><b class="text-brand-400">${res.updated}</b> voucher diupdate</div>
           <div><b class="text-accent-400">${res.imported}</b> voucher diimpor dari router</div>
           <div><b class="text-danger-400">${res.removed}</b> voucher dihapus dari DB (tidak ada di router)</div>
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
function onGenerated() { showGen.value = false; load() }
async function setVoucher(vid: string, action: 'disable' | 'enable') {
  try { await apiPost(`/vouchers/${vid}/${action}`); swal.success(`Voucher ${action === 'disable' ? 'disabled' : 'enabled'}`); load() }
  catch (e: any) { swal.error(e.message) }
}
async function delVoucher(vid: string, name: string) {
  const ok = await swal.confirm(`Hapus voucher ${name}?`, 'Voucher akan dihapus dari DB dan RouterOS.')
  if (!ok) return
  try { await apiDelete(`/vouchers/${vid}`); swal.success('Voucher dihapus'); load() }
  catch (e: any) { swal.error(e.message) }
}
onMounted(() => { load(); loadProfiles() })
</script>
