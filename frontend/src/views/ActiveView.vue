<template>
  <div class="space-y-6 animate-fade-in">
    <div class="flex items-center gap-3 flex-wrap">
      <button class="btn-secondary text-xs" @click="load">
        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
        Refresh
      </button>
      <span v-if="!loading && users.length > 0" class="ml-auto badge-online">
        <span class="w-1.5 h-1.5 rounded-full bg-accent-500 animate-pulse-glow"></span>
        {{ users.length }} user online
      </span>
    </div>

    <div class="card overflow-hidden">
      <!-- Loading -->
      <div v-if="loading" class="px-5 py-16 text-center">
        <div class="inline-flex items-center gap-2 text-ink-500 text-sm">
          <svg class="animate-spin w-5 h-5" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
          Memuat data dari router...
        </div>
      </div>

      <!-- Empty state: info jelas -->
      <div v-else-if="users.length === 0" class="px-5 py-16 text-center">
        <div class="w-16 h-16 rounded-full bg-ink-800 mx-auto flex items-center justify-center mb-4">
          <svg class="w-8 h-8 text-ink-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
        </div>
        <h3 class="text-ink-200 font-medium mb-1">Tidak ada user aktif</h3>
        <p class="text-sm text-ink-500 max-w-sm mx-auto">
          Saat ini tidak ada client yang terhubung ke hotspot.
          User akan muncul di sini setelah berhasil login menggunakan voucher/member/trial.
        </p>
        <button class="btn-secondary text-xs mt-4" @click="load">
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
          Cek ulang
        </button>
      </div>

      <!-- Table -->
      <table v-else class="table-base">
        <thead><tr><th>User</th><th>Address</th><th>MAC</th><th>Uptime</th><th>Bytes In/Out</th><th class="text-right">Aksi</th></tr></thead>
        <tbody>
          <tr v-for="u in users" :key="u.id">
            <td class="font-medium text-ink-100">
              <span class="flex items-center gap-2">
                <span class="w-1.5 h-1.5 rounded-full bg-accent-500 animate-pulse-glow"></span>
                {{ u.user || '(voucher)' }}
              </span>
            </td>
            <td class="font-mono text-ink-400">{{ u.address }}</td>
            <td class="font-mono text-xs text-ink-400">{{ u['mac-address'] }}</td>
            <td class="font-mono text-ink-400">{{ u.uptime }}</td>
            <td class="font-mono text-xs text-ink-500">{{ u['bytes-in'] || '0' }} / {{ u['bytes-out'] || '0' }}</td>
            <td class="text-right">
              <button class="btn-ghost px-2 py-1 text-xs text-danger-400" @click="kick(u.id, u.user)">Kick</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useApi } from '@/composables/useApi'
import { useSwal } from '@/composables/useSwal'

const route = useRoute()
const id = computed(() => route.params.id as string)
const { apiGet, apiPost } = useApi()
const swal = useSwal()
const users = ref<any[]>([]); const loading = ref(true)

async function load() {
  loading.value = true
  try { const res = await apiGet<{ data: any[] }>(`/servers/${id.value}/active`); users.value = res.data || [] }
  catch (e: any) { if (e.message !== 'Unauthorized') swal.errorDialog('Gagal Memuat Active Users', e.message) }
  finally { loading.value = false }
}
async function kick(rosId: string, userName: string) {
  const ok = await swal.confirm(
    `Putuskan user "${userName || '(voucher)'}"?`,
    'User akan dipaksa logout dari hotspot.',
    { confirmButtonText: 'Ya, kick', cancelButtonText: 'Batal' }
  )
  if (!ok) return
  try { await apiPost(`/servers/${id.value}/active/${rosId}/kick`); swal.success('User dikick'); load() }
  catch (e: any) { swal.error(e.message) }
}
onMounted(load)
</script>
