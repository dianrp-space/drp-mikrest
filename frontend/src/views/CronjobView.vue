<template>
  <div class="space-y-6 animate-fade-in">
    <div>
      <h1 class="page-title">Cronjob</h1>
      <p class="page-subtitle">Atur jadwal tugas otomatis</p>
    </div>

    <div class="space-y-6 max-w-2xl">
      <div class="card p-6 space-y-4">
        <h2 class="font-semibold text-ink-100">Cek Expired (Scheduler)</h2>
        <p class="text-sm text-ink-500">Interval pengecekan & auto-disable voucher expired. Diisi dalam menit.</p>
        <div class="flex items-center gap-4">
          <input v-model.number="cronInterval" type="number" min="1" class="input w-24" />
          <span class="text-sm text-ink-500">menit</span>
          <button class="btn-primary text-sm" @click="saveCron" :disabled="cronSaving">
            {{ cronSaving ? 'Menyimpan...' : 'Simpan' }}
          </button>
          <span v-if="cronSaved" class="text-sm text-accent-400">Tersimpan</span>
        </div>
      </div>

      <div class="card p-6 space-y-4">
        <h2 class="font-semibold text-ink-100">Auto Delete Log</h2>
        <p class="text-sm text-ink-500">Hapus otomatis activity log yang lebih lama dari jumlah hari ini (cron tiap 03:30).</p>
        <div class="flex items-center gap-4">
          <input v-model.number="logRetentionDays" type="number" min="1" class="input w-24" />
          <span class="text-sm text-ink-500">hari</span>
          <button class="btn-primary text-sm" @click="saveLogRetention" :disabled="logSaving">
            {{ logSaving ? 'Menyimpan...' : 'Simpan' }}
          </button>
          <span v-if="logSaved" class="text-sm text-accent-400">Tersimpan</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useApi } from '@/composables/useApi'
import { useSwal } from '@/composables/useSwal'
import { useSettings } from '@/composables/useSettings'

document.title = `Cronjob - ${useSettings().appName()}`

const { apiGet, apiPut, apiPost } = useApi()
const swal = useSwal()

const cronInterval = ref(1)
const cronSaving = ref(false)
const cronSaved = ref(false)

const logRetentionDays = ref(30)
const logSaving = ref(false)
const logSaved = ref(false)

async function loadData() {
  try {
    const info = await apiGet<{ cron_interval: number }>('/settings/scheduler')
    cronInterval.value = info.cron_interval || 1
  } catch {}
  const { fetchSettings } = useSettings()
  const s = await fetchSettings()
  logRetentionDays.value = parseInt(s?.auto_delete_log_days) || 30
}

async function saveCron() {
  cronSaving.value = true
  cronSaved.value = false
  try {
    await apiPut('/settings', { cron_interval: String(cronInterval.value) })
    await apiPost('/settings/scheduler/reload')
    cronSaved.value = true
    setTimeout(() => { cronSaved.value = false }, 3000)
  } catch (e: any) {
    swal.error(e.message)
  } finally {
    cronSaving.value = false
  }
}

async function saveLogRetention() {
  logSaving.value = true
  logSaved.value = false
  try {
    await apiPut('/settings', { auto_delete_log_days: String(logRetentionDays.value) })
    logSaved.value = true
    setTimeout(() => { logSaved.value = false }, 3000)
  } catch (e: any) {
    swal.error(e.message)
  } finally {
    logSaving.value = false
  }
}

onMounted(loadData)
</script>
