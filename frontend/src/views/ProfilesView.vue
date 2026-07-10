<template>
  <div class="space-y-6 animate-fade-in">
    <div class="flex items-center justify-between">
      <div class="flex gap-2">
        <button class="btn-secondary text-xs" :disabled="syncing" @click="syncProfiles">
          <svg v-if="syncing" class="animate-spin w-3.5 h-3.5" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
          <svg v-else class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
          {{ syncing ? 'Menyinkronkan...' : 'Sync dari Router' }}
        </button>
        <button class="btn-primary text-xs" @click="showForm = !showForm">{{ showForm ? 'Batal' : '+ Tambah' }}</button>
      </div>
    </div>

    <form v-if="showForm" class="card p-6 space-y-4 max-w-2xl animate-slide-up" @submit.prevent="createProfile">
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div><label class="label">Nama</label><input v-model="newProfile.name" required class="input" placeholder="1-hour" /></div>
        <div><label class="label">Rate Limit</label><input v-model="newProfile.rate_limit" class="input" placeholder="1M/1M" /></div>
        <div><label class="label">Session Timeout</label><input v-model="newProfile.session_timeout" class="input" placeholder="1d" /></div>
        <div><label class="label">Idle Timeout</label><input v-model="newProfile.idle_timeout" class="input" placeholder="10m" /></div>
        <div><label class="label">Shared Users</label><input v-model.number="newProfile.shared_users" type="number" min="1" class="input" /></div>
      </div>
      <div v-if="formError" class="text-sm text-danger-400 bg-danger-500/10 px-3 py-2 rounded-lg">{{ formError }}</div>
      <button type="submit" class="btn-primary text-sm">Simpan</button>
    </form>

    <div class="card overflow-hidden">
      <div v-if="profiles.length === 0" class="px-5 py-12 text-center">
        <div class="w-12 h-12 rounded-full bg-ink-800 mx-auto flex items-center justify-center mb-3">
          <svg class="w-6 h-6 text-ink-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
        </div>
        <p class="text-sm text-ink-500">Belum ada profile.</p>
        <button class="text-sm text-brand-400 hover:text-brand-300 mt-2" :disabled="syncing" @click="syncProfiles">
          <svg v-if="syncing" class="animate-spin w-3.5 h-3.5 inline -mt-0.5 mr-1" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
          <svg v-else class="w-3.5 h-3.5 inline -mt-0.5 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
          {{ syncing ? 'Menyinkronkan...' : 'Sync dari router' }}
        </button>
      </div>
      <table v-else class="table-base">
        <thead><tr><th>Nama</th><th>Rate Limit</th><th>Session</th><th>Shared</th><th class="text-right">Aksi</th></tr></thead>
        <tbody>
          <tr v-for="p in profiles" :key="p.id">
            <td class="font-medium text-ink-100">{{ p.name }}</td>
            <td class="font-mono text-ink-400">{{ p.rate_limit || '-' }}</td>
            <td class="text-ink-400">{{ p.session_timeout || '-' }}</td>
            <td class="text-ink-400 font-mono">{{ p.shared_users }}</td>
            <td class="text-right"><button class="btn-ghost px-2 py-1 text-xs text-danger-400" @click="delProfile(p.id, p.name)">Hapus</button></td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useApi } from '@/composables/useApi'
import { useSwal } from '@/composables/useSwal'

const route = useRoute()
const id = computed(() => route.params.id as string)
const { apiGet, apiPost, apiDelete } = useApi()
const swal = useSwal()
const profiles = ref<any[]>([]); const showForm = ref(false); const formError = ref(''); const syncing = ref(false)
const newProfile = reactive({ name: '', rate_limit: '', session_timeout: '', idle_timeout: '', shared_users: 1 })

async function load() { try { const res = await apiGet<{ data: any[] }>(`/servers/${id.value}/profiles`); profiles.value = res.data || [] } catch (e: any) { swal.error(e.message) } }
async function syncProfiles() {
  syncing.value = true
  try { const res = await apiPost<{ count: number }>(`/servers/${id.value}/profiles/sync`); swal.successDialog('Sync Selesai', `${res.count} profile tersinkron.`); load() }
  catch (e: any) { swal.errorDialog('Gagal Sync', e.message) } finally { syncing.value = false }
}
async function createProfile() {
  formError.value = ''
  try { await apiPost(`/servers/${id.value}/profiles`, newProfile); showForm.value = false; Object.assign(newProfile, { name: '', rate_limit: '', session_timeout: '', idle_timeout: '', shared_users: 1 }); swal.success('Profile dibuat'); load() }
  catch (e: any) { formError.value = e.message }
}
async function delProfile(pid: string, name: string) {
  const ok = await swal.confirm(`Hapus profile "${name}"?`)
  if (!ok) return
  try { await apiDelete(`/servers/${id.value}/profiles/${pid}`); swal.success('Profile dihapus'); load() }
  catch (e: any) { swal.error(e.message) }
}
onMounted(load)
</script>
