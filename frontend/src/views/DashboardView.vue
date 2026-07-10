<template>
  <div class="space-y-6 animate-fade-in">
    <div>
      <h1 class="page-title">Dashboard</h1>
      <p class="page-subtitle">Ringkasan server & voucher</p>
    </div>

    <!-- KPI cards -->
    <div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
      <RouterLink to="/servers" class="card card-hover p-5 block">
        <div class="flex items-center justify-between">
          <div class="text-sm text-ink-500">Total Server</div>
          <div class="w-9 h-9 rounded-lg bg-brand-600/15 flex items-center justify-center">
            <svg class="w-5 h-5 text-brand-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M5 12H3l9-9 9 9h-2M5 12v7a1 1 0 001 1h3v-4h6v4h3a1 1 0 001-1v-7" />
            </svg>
          </div>
        </div>
        <div class="text-3xl font-bold text-ink-100 mt-3 font-mono">{{ servers.length }}</div>
        <div class="mt-2 text-xs flex gap-3">
          <span class="flex items-center gap-1 text-accent-400">
            <span class="w-1.5 h-1.5 rounded-full bg-accent-500 animate-pulse-glow"></span>{{ onlineCount }} online
          </span>
          <span class="text-ink-500">{{ offlineCount }} offline</span>
        </div>
      </RouterLink>

      <RouterLink to="/servers" class="card card-hover p-5 block">
        <div class="flex items-center justify-between">
          <div class="text-sm text-ink-500">Voucher Aktif</div>
          <div class="w-9 h-9 rounded-lg bg-accent-600/15 flex items-center justify-center">
            <svg class="w-5 h-5 text-accent-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M7 7h.01M7 3h5a1.99 1.99 0 011.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.99 1.99 0 013 12V7a4 4 0 014-4z" />
            </svg>
          </div>
        </div>
        <div class="text-3xl font-bold text-ink-100 mt-3 font-mono">{{ activeVouchers }}</div>
        <div class="mt-2 text-xs text-ink-500">voucher siap pakai</div>
      </RouterLink>

      <RouterLink to="/servers" class="card card-hover p-5 block">
        <div class="flex items-center justify-between">
          <div class="text-sm text-ink-500">Member Aktif</div>
          <div class="w-9 h-9 rounded-lg bg-accent-600/15 flex items-center justify-center">
            <svg class="w-5 h-5 text-accent-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" /></svg>
          </div>
        </div>
        <div class="text-3xl font-bold text-ink-100 mt-3 font-mono">{{ activeMembers }}</div>
        <div class="mt-2 text-xs text-ink-500">member aktif</div>
      </RouterLink>

      <RouterLink to="/tokens" class="card card-hover p-5 block">
        <div class="flex items-center justify-between">
          <div class="text-sm text-ink-500">API Token</div>
          <div class="w-9 h-9 rounded-lg bg-warn-500/15 flex items-center justify-center">
            <svg class="w-5 h-5 text-warn-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M15 7a4 4 0 11-8 0 4 4 0 018 0zM12 11v10m-3-3l3 3 3-3" />
            </svg>
          </div>
        </div>
        <div class="text-3xl font-bold text-ink-100 mt-3 font-mono">{{ tokenCount }}</div>
        <div class="mt-2 text-xs text-ink-500">token aktif</div>
      </RouterLink>
    </div>

    <!-- Server list -->
    <div class="card">
      <div class="px-5 py-4 border-b border-ink-800 flex items-center justify-between">
        <h2 class="font-semibold text-ink-100">Server Router</h2>
        <RouterLink to="/servers" class="text-sm text-brand-400 hover:text-brand-300 flex items-center gap-1">
          Lihat semua
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
          </svg>
        </RouterLink>
      </div>
      <div v-if="servers.length === 0" class="px-5 py-12 text-center">
        <div class="w-12 h-12 rounded-full bg-ink-800 mx-auto flex items-center justify-center mb-3">
          <svg class="w-6 h-6 text-ink-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01M5 12H3l9-9 9 9h-2M5 12v7a1 1 0 001 1h3v-4h6v4h3a1 1 0 001-1v-7" />
          </svg>
        </div>
        <p class="text-sm text-ink-500">Belum ada server</p>
        <RouterLink to="/servers" class="text-sm text-brand-400 hover:text-brand-300 mt-2 inline-block">+ Tambah server</RouterLink>
      </div>
      <ul v-else class="divide-y divide-ink-800">
        <li v-for="s in servers.slice(0, 5)" :key="s.id" class="px-5 py-3 flex items-center justify-between hover:bg-ink-800/40 transition-colors">
          <div class="min-w-0">
            <div class="font-medium text-ink-100 truncate">{{ s.name }}</div>
            <div class="text-xs text-ink-500 font-mono mt-0.5">{{ s.host }}:{{ s.api_port }} &middot; {{ s.username }}</div>
          </div>
          <span :class="statusBadge(s.status)">
            <span class="w-1.5 h-1.5 rounded-full" :class="statusDot(s.status)"></span>
            {{ s.status }}
          </span>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useApi } from '@/composables/useApi'
import { useSettings } from '@/composables/useSettings'

document.title = `Dashboard - ${useSettings().appName()}`

interface Server { id: string; name: string; host: string; api_port: number; username: string; status: string }

const { apiGet } = useApi()
const servers = ref<Server[]>([])
const activeVouchers = ref(0)
const activeMembers = ref(0)
const tokenCount = ref(0)

const onlineCount = computed(() => servers.value.filter((s) => s.status === 'online').length)
const offlineCount = computed(() => servers.value.filter((s) => s.status === 'offline').length)

function statusBadge(s: string) {
  return s === 'online' ? 'badge-online' : s === 'offline' ? 'badge-offline' : 'badge-unknown'
}
function statusDot(s: string) {
  return s === 'online' ? 'bg-accent-500' : s === 'offline' ? 'bg-danger-500' : 'bg-ink-600'
}

onMounted(async () => {
  try { const res = await apiGet<{ data: Server[] }>('/servers'); servers.value = res.data || [] } catch {}
  try { const v = await apiGet<{ total: number }>('/vouchers?status=active&limit=1&type=voucher'); activeVouchers.value = v.total || 0 } catch {}
  try { const m = await apiGet<{ total: number }>('/vouchers?status=active&limit=1&type=member'); activeMembers.value = m.total || 0 } catch {}
  try { const t = await apiGet<{ data: any[] }>('/tokens'); tokenCount.value = t.data?.length || 0 } catch {}
})
</script>
