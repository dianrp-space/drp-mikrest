<template>
  <div class="space-y-6 animate-fade-in">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="page-title">Server Router</h1>
        <p class="page-subtitle">Daftar Mikrotik yang dikelola</p>
      </div>
      <button class="btn-primary" @click="showForm = !showForm">
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" /></svg>
        {{ showForm ? 'Batal' : 'Tambah Server' }}
      </button>
    </div>

    <ServerForm v-if="showForm" :server="null" @saved="onSaved" @cancel="showForm = false" />
    <ServerForm v-else-if="editing" :server="editing" @saved="onEdited" @cancel="editing = null" />

    <div v-if="loading" class="py-12 text-center">
      <div class="inline-flex items-center gap-2 text-ink-500 text-sm">
        <svg class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
        Memuat...
      </div>
    </div>
    <div v-else-if="servers.length === 0" class="py-12 text-center">
      <div class="w-12 h-12 rounded-full bg-ink-800 mx-auto flex items-center justify-center mb-3">
        <svg class="w-6 h-6 text-ink-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M5 12H3l9-9 9 9h-2M5 12v7a1 1 0 001 1h3v-4h6v4h3a1 1 0 001-1v-7" /></svg>
      </div>
      <p class="text-sm text-ink-500">Belum ada server.</p>
      <button class="text-sm text-brand-400 hover:text-brand-300 mt-2" @click="showForm = true">+ Tambah server</button>
    </div>
    <div v-else class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
      <div
        v-for="s in servers"
        :key="s.id"
        class="card card-hover p-5 block group cursor-pointer"
        @click="router.push(`/servers/${s.id}`)"
      >
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0 flex-1">
            <h3 class="font-semibold text-ink-100 group-hover:text-brand-400 truncate">{{ s.name }}</h3>
            <p class="text-sm font-mono text-ink-400 mt-0.5">{{ s.host }}:{{ s.api_port }}</p>
            <p class="text-xs text-ink-500 mt-0.5">{{ s.username }}</p>
          </div>
          <span :class="statusBadge(s.status)">
            <span class="w-1.5 h-1.5 rounded-full mr-1" :class="statusDot(s.status)"></span>
            {{ s.status }}
          </span>
        </div>
        <div class="flex gap-1 mt-4 pt-3 border-t border-ink-700">
          <button class="btn-ghost px-2 py-1 text-xs" @click.stop="testConnection(s.id)" title="Test koneksi">
            <svg class="w-4 h-4 text-accent-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>
          </button>
          <button class="btn-ghost px-2 py-1 text-xs" @click.stop="editServer(s)" title="Edit">
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg>
          </button>
          <button class="btn-ghost px-2 py-1 text-xs" @click.stop="removeServer(s.id, s.name)" title="Hapus">
            <svg class="w-4 h-4 text-danger-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useApi } from '@/composables/useApi'
import { useSwal } from '@/composables/useSwal'
import { useSettings } from '@/composables/useSettings'
import ServerForm from '@/components/ServerForm.vue'

document.title = `Server - ${useSettings().appName()}`

interface Server { id: string; name: string; host: string; api_port: number; username: string; status: string }

const router = useRouter()
const { apiGet, apiPost, apiDelete } = useApi()
const swal = useSwal()
const servers = ref<Server[]>([])
const loading = ref(true)
const showForm = ref(false)
const editing = ref<any>(null)

function statusBadge(s: string) { return s === 'online' ? 'badge-online' : s === 'offline' ? 'badge-offline' : 'badge-unknown' }
function statusDot(s: string) { return s === 'online' ? 'bg-accent-500' : s === 'offline' ? 'bg-danger-500' : 'bg-ink-600' }

async function load() {
  loading.value = true
  try { const res = await apiGet<{ data: Server[] }>('/servers'); servers.value = res.data || [] }
  catch (e: any) { swal.error(e.message) }
  finally { loading.value = false }
}
function onSaved() { showForm.value = false; swal.success('Server ditambahkan'); load() }
function onEdited() { editing.value = null; swal.success('Server diubah'); load() }
function editServer(s: any) { editing.value = s; window.scrollTo({ top: 0, behavior: 'smooth' }) }

async function testConnection(id: string) {
  try {
    const res = await apiPost<{ identity: string }>(`/servers/${id}/test`)
    swal.successDialog('Koneksi Berhasil!', `Identity: ${res.identity}`)
    load()
  } catch (e: any) { swal.errorDialog('Koneksi Gagal', e.message) }
}

async function removeServer(id: string, name: string) {
  const ok = await swal.confirm(`Hapus server "${name}"?`, 'Semua voucher & profile terkait akan terhapus.')
  if (!ok) return
  try { await apiDelete(`/servers/${id}`); swal.success('Server dihapus'); load() }
  catch (e: any) { swal.errorDialog('Gagal Hapus', e.message) }
}

onMounted(load)
</script>
