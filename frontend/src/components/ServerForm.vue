<template>
  <div class="card p-6 space-y-5 animate-slide-up">
    <div class="flex items-center gap-3">
      <div class="w-10 h-10 rounded-lg bg-brand-600/15 flex items-center justify-center">
        <svg class="w-5 h-5 text-brand-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" /></svg>
      </div>
      <div>
        <h2 class="font-semibold text-ink-100">{{ server ? 'Edit Server' : 'Tambah Server' }}</h2>
        <p class="text-xs text-ink-500">Kredensial disimpan terenkripsi (AES-256-GCM)</p>
      </div>
    </div>
    <form class="space-y-4" @submit.prevent="onSubmit">
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div><label class="label">Nama</label><input v-model="form.name" required class="input" placeholder="Hotspot Kantor" /></div>
        <div><label class="label">Host / IP</label><input v-model="form.host" required class="input" placeholder="192.168.88.1" /></div>
        <div><label class="label">API Port</label><input v-model.number="form.api_port" type="number" required class="input" placeholder="8728" /></div>
        <div><label class="label">Username RouterOS</label><input v-model="form.username" required class="input" placeholder="admin" /></div>
        <div class="sm:col-span-2"><label class="label">Password RouterOS</label><input v-model="form.password" :required="!server" class="input" :placeholder="server ? 'Kosongkan jika tidak diubah' : '••••••••'" /></div>
      </div>
      <div v-if="error" class="text-sm text-danger-400 bg-danger-500/10 px-3 py-2 rounded-lg">{{ error }}</div>
      <div class="flex items-center gap-3">
        <button type="submit" :disabled="loading" class="btn-primary">
          <svg v-if="loading" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
          {{ loading ? 'Menyimpan...' : 'Simpan' }}
        </button>
        <button type="button" class="btn-secondary" @click="$emit('cancel')">Batal</button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useApi } from '@/composables/useApi'

const props = defineProps<{ server: any | null }>()
const emit = defineEmits<{ saved: []; cancel: [] }>()

const { apiPost, apiPut } = useApi()
const form = reactive({ name: props.server?.name || '', host: props.server?.host || '', api_port: props.server?.api_port || 8728, username: props.server?.username || '', password: '' })
const loading = ref(false); const error = ref('')

async function onSubmit() {
  loading.value = true; error.value = ''
  try { if (props.server) await apiPut(`/servers/${props.server.id}`, form); else await apiPost('/servers', form); emit('saved') }
  catch (e: any) { error.value = e.message } finally { loading.value = false }
}
</script>
