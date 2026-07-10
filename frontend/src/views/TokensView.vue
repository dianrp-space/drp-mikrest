<template>
  <div class="space-y-6 animate-fade-in">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="page-title">API Token</h1>
        <p class="page-subtitle">Token untuk integrasi sistem lain</p>
      </div>
      <button class="btn-primary text-sm" @click="showForm = !showForm">{{ showForm ? 'Batal' : '+ Buat Token' }}</button>
    </div>

    <form v-if="showForm" class="card p-6 space-y-4 max-w-xl animate-slide-up" @submit.prevent="createToken">
      <div><label class="label">Label</label><input v-model="newToken.label" required class="input" placeholder="cron-generator" /></div>
      <div class="grid grid-cols-2 gap-4">
        <div><label class="label">Rate Limit (req/menit)</label><input v-model.number="newToken.rate_limit" type="number" min="1" class="input" /></div>
        <div><label class="label">Scopes (comma)</label><input v-model="scopesText" class="input" placeholder="vouchers:rw,servers:ro" /></div>
      </div>
      <div v-if="formError" class="text-sm text-danger-400 bg-danger-500/10 px-3 py-2 rounded-lg">{{ formError }}</div>
      <button type="submit" class="btn-primary text-sm">Buat</button>
    </form>

    <!-- Tabel token -->
    <div class="card overflow-hidden">
      <div v-if="tokens.length === 0" class="px-5 py-12 text-center">
        <div class="w-12 h-12 rounded-full bg-ink-800 mx-auto flex items-center justify-center mb-3">
          <svg class="w-6 h-6 text-ink-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M15 7a4 4 0 11-8 0 4 4 0 018 0zM12 11v10m-3-3l3 3 3-3" /></svg>
        </div>
        <p class="text-sm text-ink-500">Belum ada token.</p>
      </div>
      <table v-else class="table-base">
        <thead>
          <tr>
            <th>Label</th>
            <th>Prefix</th>
            <th>Token</th>
            <th>Scopes</th>
            <th>Terakhir dipakai</th>
            <th>Status</th>
            <th class="text-right">Aksi</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="t in tokens" :key="t.id">
            <td class="font-medium text-ink-100">{{ t.label }}</td>
            <td class="font-mono text-xs text-ink-400">
              <div class="flex items-center gap-1.5">
                <code>{{ t.token_prefix }}</code>
                <button class="btn-ghost px-1 py-0.5 text-[10px]" title="Salin prefix" @click="copyText(t.token_prefix, `Prefix disalin`)">
                  <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" /></svg>
                </button>
              </div>
            </td>
            <td>
              <div v-if="visibleTokens.has(t.id) && storedPlain(t.id)" class="flex items-center gap-1.5">
                <code class="text-[11px] font-mono text-accent-400 bg-ink-900 px-1.5 py-0.5 rounded border border-ink-700 select-all truncate max-w-[200px]">{{ storedPlain(t.id) }}</code>
                <button class="btn-ghost px-1.5 py-0.5 text-[10px]" title="Salin full token" @click="copyText(storedPlain(t.id), 'Token disalin')">
                  <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" /></svg>
                </button>
                <button class="btn-ghost px-1.5 py-0.5 text-[10px]" title="Sembunyikan" @click="visibleTokens.delete(t.id)">
                  <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" /></svg>
                </button>
              </div>
              <div v-else-if="storedPlain(t.id)" class="flex items-center gap-1.5">
                <span class="text-xs text-ink-500 font-mono italic">tersimpan di browser</span>
                <button class="btn-ghost px-1.5 py-0.5 text-[10px]" title="Tampilkan & salin" @click="visibleTokens.add(t.id)">
                  <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path stroke-linecap="round" stroke-linejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg>
                </button>
              </div>
              <div v-else class="flex items-center gap-1.5">
                <span class="text-xs text-ink-600 italic">tidak ada di browser</span>
                <button class="btn-ghost px-1.5 py-0.5 text-[10px]" title="Buat ulang (revoke + create baru) untuk melihat token" disabled>
                  <svg class="w-3 h-3 opacity-40" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
                </button>
              </div>
            </td>
            <td class="text-xs text-ink-400">{{ (t.scopes || []).join(', ') }}</td>
            <td class="text-xs text-ink-500 font-mono">{{ t.last_used_at ? fmt(t.last_used_at) : '-' }}</td>
            <td>
              <span v-if="t.revoked_at" class="badge-failed">revoked</span>
              <span v-else class="badge-online"><span class="w-1.5 h-1.5 rounded-full bg-accent-500"></span>active</span>
            </td>
            <td class="text-right">
              <button v-if="!t.revoked_at" class="btn-ghost px-2 py-1 text-xs text-danger-400" @click="delToken(t.id, t.label)" title="Hapus token permanen">
                <svg class="w-3.5 h-3.5 inline -mt-0.5 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                Hapus
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Info -->
    <div class="card p-4 text-xs text-ink-400 space-y-2">
      <div class="flex items-start gap-2">
        <svg class="w-4 h-4 text-brand-400 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
        <span>Token plain hanya disimpan di <b>browser ini</b> (localStorage). Untuk produksi, simpan token di secret manager Anda setelah create. Kalau clear browser, token tidak bisa di-retrieve &mdash; hapus dan buat ulang.</span>
      </div>
    </div>

    <!-- Contoh pakai API -->
    <div class="card p-4 space-y-4">
      <div class="flex items-center gap-2">
        <svg class="w-4 h-4 text-brand-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" /></svg>
        <h3 class="font-semibold text-ink-100 text-sm">Contoh pakai API</h3>
      </div>
      <p class="text-xs text-ink-500">Base URL: <code class="font-mono text-ink-300 bg-ink-900 px-1.5 py-0.5 rounded">{{ baseUrl }}/api/v1</code> &middot; Auth: <code class="font-mono text-ink-300 bg-ink-900 px-1.5 py-0.5 rounded">Authorization: Bearer &lt;token&gt;</code></p>

      <div>
        <div class="text-xs font-medium text-ink-300 mb-1.5">Login — ambil bearer token</div>
        <pre class="text-xs bg-ink-950 text-ink-300 p-3 rounded-lg overflow-auto font-mono border border-ink-800 cursor-pointer hover:border-accent-500/50 transition-colors" @click="copyText($event.currentTarget.textContent.trim(), 'Contoh login tersalin')" v-text="examples.login"></pre>
      </div>

      <div>
        <div class="text-xs font-medium text-ink-300 mb-1.5">List server</div>
        <pre class="text-xs bg-ink-950 text-ink-300 p-3 rounded-lg overflow-auto font-mono border border-ink-800 cursor-pointer hover:border-accent-500/50 transition-colors" @click="copyText($event.currentTarget.textContent.trim(), 'Contoh list server tersalin')" v-text="examples.servers"></pre>
      </div>

      <div>
        <div class="text-xs font-medium text-ink-300 mb-1.5">Generate voucher (username = password otomatis)</div>
        <pre class="text-xs bg-ink-950 text-ink-300 p-3 rounded-lg overflow-auto font-mono border border-ink-800 cursor-pointer hover:border-accent-500/50 transition-colors" @click="copyText($event.currentTarget.textContent.trim(), 'Contoh generate voucher tersalin')" v-text="examples.voucher"></pre>
      </div>

      <div>
        <div class="text-xs font-medium text-ink-300 mb-1.5">Buat member (username & password berbeda)</div>
        <pre class="text-xs bg-ink-950 text-ink-300 p-3 rounded-lg overflow-auto font-mono border border-ink-800 cursor-pointer hover:border-accent-500/50 transition-colors" @click="copyText($event.currentTarget.textContent.trim(), 'Contoh buat member tersalin')" v-text="examples.member"></pre>
      </div>

      <div>
        <div class="text-xs font-medium text-ink-300 mb-1.5">Disable / enable / hapus voucher atau member</div>
        <pre class="text-xs bg-ink-950 text-ink-300 p-3 rounded-lg overflow-auto font-mono border border-ink-800 cursor-pointer hover:border-accent-500/50 transition-colors" @click="copyText($event.currentTarget.textContent.trim(), 'Contoh disable/enable/hapus tersalin')" v-text="examples.disableEnableDelete"></pre>
      </div>

      <div>
        <div class="text-xs font-medium text-ink-300 mb-1.5">Sinkronisasi voucher dari router</div>
        <pre class="text-xs bg-ink-950 text-ink-300 p-3 rounded-lg overflow-auto font-mono border border-ink-800 cursor-pointer hover:border-accent-500/50 transition-colors" @click="copyText($event.currentTarget.textContent.trim(), 'Contoh sync tersalin')" v-text="examples.sync"></pre>
      </div>

      <div class="text-xs text-ink-500 bg-ink-900/50 border border-ink-800 rounded-lg p-3 space-y-1">
        <div class="flex items-start gap-2">
          <svg class="w-4 h-4 text-warn-400 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01M5 19h14a2 2 0 001.84-2.75L13.74 4a2 2 0 00-3.48 0l-7.1 12.25A2 2 0 005 19z" /></svg>
          <span>Scope yang tersedia: <code class="font-mono text-ink-300">vouchers:rw</code> (generate/disable/enable/hapus voucher), <code class="font-mono text-ink-300">servers:ro</code> (list server & profile). Token tanpa scope = ditolak.</span>
        </div>
        <div class="flex items-start gap-2">
          <svg class="w-4 h-4 text-brand-400 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
          <span>Response error menggunakan format <code class="font-mono text-ink-300">{"error":"kode","message":"penjelasan"}</code>. Status: 200/201 sukses, 400 bad request, 401 invalid token, 403 scope kurang, 429 rate limit.</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useApi } from '@/composables/useApi'
import { useSwal } from '@/composables/useSwal'
import { useSettings } from '@/composables/useSettings'

const settings = useSettings()
document.title = `API Token - ${settings.appName()}`
const baseUrl = computed(() => settings.appUrl())

const examples = computed(() => ({
  login: `curl -X POST ${baseUrl.value}/api/web/auth/login \\\n  -H "Content-Type: application/json" \\\n  -d '{"email":"admin@example.com","password":"12345678"}'`,
  servers: `curl ${baseUrl.value}/api/v1/servers \\\n  -H "Authorization: Bearer drp_xxxxxxxxxxxx"`,
  voucher: `curl -X POST ${baseUrl.value}/api/v1/vouchers \\\n  -H "Authorization: Bearer drp_xxxxxxxxxxxx" \\\n  -H "Content-Type: application/json" \\\n  -d '{\n    "server_id":"<uuid-server>",\n    "profile_id":"<uuid-profile>",\n    "count":10,\n    "username_mode":"random",\n    "pattern":"####-####",\n    "limit_uptime":"1d",\n    "comment":"1jam-Rp5000"\n  }'`,
  member: `curl -X POST ${baseUrl.value}/api/v1/vouchers/member \\\n  -H "Authorization: Bearer drp_xxxxxxxxxxxx" \\\n  -H "Content-Type: application/json" \\\n  -d '{\n    "server_id":"<uuid-server>",\n    "username":"nama-member",\n    "password":"pass-member",\n    "profile_id":"<uuid-profile>",\n    "limit_uptime":"30d",\n    "comment":"member-bulanan"\n  }'`,
  disableEnableDelete: `# Disable\ncurl -X POST ${baseUrl.value}/api/v1/vouchers/<voucher-uuid>/disable \\\n  -H "Authorization: Bearer drp_xxxxxxxxxxxx"\n\n# Enable\ncurl -X POST ${baseUrl.value}/api/v1/vouchers/<voucher-uuid>/enable \\\n  -H "Authorization: Bearer drp_xxxxxxxxxxxx"\n\n# Hapus permanen\ncurl -X DELETE ${baseUrl.value}/api/v1/vouchers/<voucher-uuid> \\\n  -H "Authorization: Bearer drp_xxxxxxxxxxxx"`,
  sync: `curl -X POST ${baseUrl.value}/api/v1/servers/<server-uuid>/vouchers/sync \\\n  -H "Authorization: Bearer drp_xxxxxxxxxxxx"\n# Response: {"synced":true,"updated":3,"imported":0,"removed":0}`,
}))

const { apiGet, apiPost, apiDelete } = useApi()
const swal = useSwal()
const tokens = ref<any[]>([]); const showForm = ref(false); const formError = ref('')
const scopesText = ref('vouchers:rw,servers:ro')
const newToken = reactive({ label: '', rate_limit: 60 })

// Set berisi token id yang sedang ditampilkan (toggle show/hide per baris).
const visibleTokens = ref(new Set<string>())

// localStorage key: drp_token_plain_<id> -> plain token
function tokenStoreKey(id: string) { return `drp_token_plain_${id}` }

function storedPlain(id: string): string {
  if (typeof localStorage === 'undefined') return ''
  return localStorage.getItem(tokenStoreKey(id)) || ''
}

function savePlain(id: string, plain: string) {
  if (typeof localStorage === 'undefined') return
  localStorage.setItem(tokenStoreKey(id), plain)
}

function removePlain(id: string) {
  if (typeof localStorage === 'undefined') return
  localStorage.removeItem(tokenStoreKey(id))
}

async function load() {
  try {
    const res = await apiGet<{ data: any[] }>('/tokens')
    tokens.value = res.data || []
    // bersihkan cache token yang sudah di-revoke/dihapus dari server
    const serverIds = new Set(res.data?.map((t: any) => t.id) || [])
    if (typeof localStorage !== 'undefined') {
      for (let i = localStorage.length - 1; i >= 0; i--) {
        const k = localStorage.key(i)
        if (k && k.startsWith('drp_token_plain_')) {
          const id = k.replace('drp_token_plain_', '')
          if (!serverIds.has(id)) localStorage.removeItem(k)
        }
      }
    }
  } catch (e: any) { swal.error(e.message) }
}

async function createToken() {
  formError.value = ''
  const scopes = scopesText.value.split(',').map((s) => s.trim()).filter(Boolean)
  try {
    const res = await apiPost<{ plain_token: string; token: any }>('/tokens', {
      label: newToken.label, rate_limit: newToken.rate_limit, scopes,
    })
    // simpan plain token di localStorage supaya bisa dilihat/disalin lagi nanti
    savePlain(res.token.id, res.plain_token)
    visibleTokens.value.add(res.token.id)
    showForm.value = false
    newToken.label = ''
    swal.successDialog(
      'Token dibuat',
      `<div class="text-left text-sm space-y-2">
         <div>Token sudah disimpan di browser ini dan bisa dilihat & disalin kapan saja dari tabel.</div>
         <div class="text-xs text-warn-400">Untuk produksi: salin token ke secret manager sekarang, karena tidak akan pernah bisa dilihat lagi dari server.</div>
       </div>`
    )
    load()
  } catch (e: any) { formError.value = e.message }
}

async function delToken(tid: string, label: string) {
  const ok = await swal.confirm(
    `Hapus token "${label}"?`,
    'Token akan dihapus permanen dari DB dan tidak bisa dipakai untuk request API. Plain token juga akan dihapus dari browser ini.',
    { confirmButtonText: 'Ya, hapus', cancelButtonText: 'Batal' }
  )
  if (!ok) return
  try {
    await apiDelete(`/tokens/${tid}`)
    removePlain(tid)
    visibleTokens.value.delete(tid)
    swal.success(`Token "${label}" dihapus`)
    load()
  } catch (e: any) { swal.error(e.message) }
}

async function copyText(text: string, msg: string) {
  if (!text) return
  try { await navigator.clipboard.writeText(text); swal.success(msg) }
  catch { swal.error('Gagal menyalin, salin manual dari teks') }
}

function fmt(t: string) { return new Date(t).toLocaleString('id-ID', { dateStyle: 'short', timeStyle: 'short' }) }
onMounted(load)
</script>
