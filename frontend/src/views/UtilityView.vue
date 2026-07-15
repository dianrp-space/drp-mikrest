<template>
  <div class="space-y-6 animate-fade-in">
    <div>
      <h1 class="page-title">Utility</h1>
      <p class="page-subtitle">Wake-on-LAN & Terminal RouterOS</p>
    </div>

    <!-- Tab Switcher -->
    <div class="flex gap-1 border-b border-ink-800">
      <button
        v-for="tab in tabs" :key="tab.key"
        class="px-4 py-2.5 text-sm font-medium transition-all duration-200 border-b-2 -mb-px"
        :class="activeTab === tab.key
          ? 'border-brand-500 text-brand-400'
          : 'border-transparent text-ink-400 hover:text-ink-200 hover:border-ink-600'"
        @click="activeTab = tab.key"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- Tab: WOL -->
    <div v-if="activeTab === 'wol'" class="space-y-6">
      <div class="card p-6 space-y-5 max-w-2xl">
        <h2 class="font-semibold text-ink-100">Wake-on-LAN</h2>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label class="label">Server</label>
            <select v-model="form.server_id" class="input" @change="onServerChange">
              <option value="" disabled>Pilih server</option>
              <option v-for="s in servers" :key="s.id" :value="s.id">{{ s.name }} ({{ s.host }})</option>
            </select>
          </div>
          <div>
            <label class="label">Interface</label>
            <select v-model="form.interface_name" class="input" :disabled="!form.server_id || loadingIfaces">
              <option value="" disabled>Pilih interface</option>
              <option v-for="iface in interfaces" :key="iface.name" :value="iface.name">{{ iface.name }}</option>
            </select>
          </div>
          <div>
            <label class="label">MAC Address</label>
            <input v-model="form.mac_address" class="input" placeholder="00:11:22:33:44:55" />
          </div>
          <div>
            <label class="label">Nama</label>
            <input v-model="form.name" class="input" placeholder="Label" />
          </div>
        </div>

        <div class="flex gap-2">
          <button class="btn-primary text-sm" @click="saveTarget" :disabled="saving">
            {{ saving ? 'Menyimpan...' : 'Simpan' }}
          </button>
        </div>
      </div>

      <!-- Daftar WOL Target -->
      <div v-if="targets.length > 0" class="max-w-2xl">
        <h3 class="font-semibold text-ink-100 mb-3">Target Tersimpan</h3>
        <div class="space-y-2">
          <div
            v-for="t in targets" :key="t.id"
            class="card p-4 flex items-center justify-between gap-4"
          >
            <div class="min-w-0 flex-1">
              <div class="font-medium text-ink-100">{{ t.name }}</div>
              <div class="text-sm text-ink-400 font-mono">{{ t.mac_address }} @ {{ t.interface_name }}</div>
              <div class="text-xs text-ink-500">{{ serverName(t.server_id) }}</div>
            </div>
            <div class="flex gap-2 shrink-0">
              <button class="btn-primary px-3 py-1.5 text-xs" @click="sendWOL(t.id)" :disabled="sending === t.id">
                {{ sending === t.id ? 'Mengirim...' : 'Kirim WOL' }}
              </button>
              <button class="btn-ghost px-3 py-1.5 text-xs text-danger-400 hover:text-danger-300" @click="deleteTarget(t.id, t.name)">
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
              </button>
            </div>
          </div>
        </div>
      </div>

      <div v-else class="max-w-2xl py-8 text-center">
        <p class="text-sm text-ink-500">Belum ada target WOL tersimpan.</p>
      </div>
    </div>

    <!-- Tab: Terminal -->
    <div v-if="activeTab === 'terminal'" class="space-y-4">
      <div class="flex flex-wrap items-end gap-3">
        <div class="w-72">
          <label class="label">Server</label>
          <select v-model="termServerId" class="input">
            <option value="" disabled>Pilih server</option>
            <option v-for="s in servers" :key="s.id" :value="s.id">{{ s.name }} ({{ s.host }})</option>
          </select>
        </div>
        <button class="btn-ghost text-xs" @click="clearTerminal">Clear</button>
      </div>

      <div class="card overflow-hidden border-ink-700">
        <!-- Terminal output -->
        <div
          ref="termOutputEl"
          class="bg-[#0d1117] p-4 font-mono text-sm leading-relaxed h-[420px] overflow-y-auto whitespace-pre-wrap"
        >
          <div v-if="termLines.length === 0" class="text-ink-600 select-none">
            &gt; Pilih server dan ketik command RouterOS di bawah, lalu tekan Enter.<br/>
            &gt; Contoh: /ip/hotspot/user/print , /system/identity/print , /interface/print<br/>
          </div>
          <div v-for="(line, i) in termLines" :key="i">
            <span v-if="line.type === 'input'" class="text-accent-400">$ {{ line.text }}</span>
            <span v-else-if="line.type === 'output'" class="text-ink-300">{{ line.text }}</span>
            <span v-else-if="line.type === 'error'" class="text-danger-400">{{ line.text }}</span>
            <span v-else-if="line.type === 'info'" class="text-ink-500 italic">{{ line.text }}</span>
          </div>
          <div class="flex items-center gap-1 text-accent-400">
            <span>$</span>
            <input
              ref="termInputEl"
              v-model="termInput"
              type="text"
              class="flex-1 bg-transparent border-none outline-none text-accent-200 font-mono text-sm placeholder-ink-700"
              placeholder="ketik command..."
              :disabled="!termServerId || termBusy"
              @keydown.enter="sendCommand"
              @keydown.up.prevent="historyBack"
              @keydown.down.prevent="historyForward"
            />
          </div>
        </div>
      </div>

      <div class="text-xs text-ink-500">
        Terhubung ke RouterOS via API port {{ selectedServerPort || '8728' }}.
        Gunakan command RouterOS standar (<span class="font-mono text-ink-400">/ip/... print</span>, <span class="font-mono text-ink-400">/system/... print</span>, dll).
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue'
import { useApi } from '@/composables/useApi'
import { useSwal } from '@/composables/useSwal'
import { useSettings } from '@/composables/useSettings'

document.title = `Utility - ${useSettings().appName()}`

interface Server { id: string; name: string; host: string; api_port?: number }
interface WOLTarget { id: string; server_id: string; interface_name: string; mac_address: string; name: string }
interface Interface { name: string; [key: string]: string }

const { apiGet, apiPost, apiDelete } = useApi()
const swal = useSwal()

// --- Tabs ---
const tabs = [
  { key: 'wol', label: 'Wake-on-LAN' },
  { key: 'terminal', label: 'Terminal' },
]
const activeTab = ref('wol')

// --- WOL ---
const servers = ref<Server[]>([])
const interfaces = ref<Interface[]>([])
const targets = ref<WOLTarget[]>([])
const loadingIfaces = ref(false)
const saving = ref(false)
const sending = ref<string | null>(null)

const form = ref({ server_id: '', interface_name: '', mac_address: '', name: '' })

async function loadServers() {
  try {
    const res = await apiGet<{ data: Server[] }>('/servers')
    servers.value = res.data || []
  } catch (e: any) {
    swal.error(e.message)
  }
}

async function loadTargets() {
  try {
    const res = await apiGet<{ data: WOLTarget[] }>('/wol')
    targets.value = res.data || []
  } catch (e: any) {
    swal.error(e.message)
  }
}

async function onServerChange() {
  form.value.interface_name = ''
  interfaces.value = []
  if (!form.value.server_id) return
  loadingIfaces.value = true
  try {
    const res = await apiGet<{ data: Interface[] }>(`/servers/${form.value.server_id}/interfaces`)
    interfaces.value = (res.data || []).filter(i => i.type === 'ether' || i.type === 'wlan' || !i.type)
  } catch (e: any) {
    swal.error(e.message)
  } finally {
    loadingIfaces.value = false
  }
}

async function saveTarget() {
  if (!form.value.server_id || !form.value.interface_name || !form.value.mac_address || !form.value.name) {
    swal.warning('Semua field wajib diisi')
    return
  }
  saving.value = true
  try {
    await apiPost('/wol', form.value)
    swal.success('Target WOL ditambahkan')
    form.value = { server_id: form.value.server_id, interface_name: '', mac_address: '', name: '' }
    await loadTargets()
  } catch (e: any) {
    swal.error(e.message)
  } finally {
    saving.value = false
  }
}

async function sendWOL(id: string) {
  sending.value = id
  try {
    await apiPost(`/wol/${id}/send`)
    swal.success('WOL berhasil dikirim')
  } catch (e: any) {
    swal.errorDialog('Gagal Kirim WOL', e.message)
  } finally {
    sending.value = null
  }
}

async function deleteTarget(id: string, name: string) {
  const ok = await swal.confirm(`Hapus target "${name}"?`, '')
  if (!ok) return
  try {
    await apiDelete(`/wol/${id}`)
    swal.success('Target dihapus')
    await loadTargets()
  } catch (e: any) {
    swal.errorDialog('Gagal Hapus', e.message)
  }
}

function serverName(id: string) {
  const s = servers.value.find(s => s.id === id)
  return s ? s.name : id
}

// --- Terminal ---
const termServerId = ref('')
const termInput = ref('')
const termBusy = ref(false)
const termLines = ref<{ type: string; text: string }[]>([])
const termHistory = ref<string[]>([])
const historyIdx = ref(-1)
const termOutputEl = ref<HTMLElement | null>(null)
const termInputEl = ref<HTMLInputElement | null>(null)

const selectedServerPort = computed(() => {
  const s = servers.value.find(s => s.id === termServerId.value)
  return s?.api_port
})

function scrollTerminal() {
  nextTick(() => {
    if (termOutputEl.value) {
      termOutputEl.value.scrollTop = termOutputEl.value.scrollHeight
    }
  })
}

function addLine(type: string, text: string) {
  termLines.value.push({ type, text })
  scrollTerminal()
}

function clearTerminal() {
  termLines.value = []
  termHistory.value = []
  historyIdx.value = -1
  termInput.value = ''
  nextTick(() => termInputEl.value?.focus())
}

function historyBack() {
  if (termHistory.value.length === 0) return
  if (historyIdx.value === -1) {
    historyIdx.value = termHistory.value.length - 1
  } else if (historyIdx.value > 0) {
    historyIdx.value--
  }
  termInput.value = termHistory.value[historyIdx.value]
}

function historyForward() {
  if (historyIdx.value === -1) return
  if (historyIdx.value < termHistory.value.length - 1) {
    historyIdx.value++
    termInput.value = termHistory.value[historyIdx.value]
  } else {
    historyIdx.value = -1
    termInput.value = ''
  }
}

async function sendCommand() {
  const cmd = termInput.value.trim()
  if (!cmd || !termServerId.value || termBusy.value) return

  termInput.value = ''
  historyIdx.value = -1
  termHistory.value.push(cmd)

  addLine('input', cmd)
  termBusy.value = true

  try {
    const res = await apiPost<{ output: string }>(`/servers/${termServerId.value}/terminal`, { command: cmd })
    if (res.output) {
      const lines = res.output.split('\n')
      for (const line of lines) {
        if (line.trim()) addLine('output', line)
      }
    }
  } catch (e: any) {
    addLine('error', `Error: ${e.message}`)
  } finally {
    termBusy.value = false
    nextTick(() => termInputEl.value?.focus())
  }
}

onMounted(() => {
  loadServers()
  loadTargets()
})
</script>

<style scoped>
/* Terminal input override */
.bg-\[\#0d1117\] input::placeholder {
  color: #333;
}
</style>
