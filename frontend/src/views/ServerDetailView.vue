<template>
  <div class="space-y-6 animate-fade-in">
    <div class="flex items-center gap-3">
      <RouterLink to="/servers" class="text-sm text-ink-500 hover:text-ink-300 flex items-center gap-1">
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" /></svg>
        Server
      </RouterLink>
      <span class="text-ink-700">/</span>
      <h1 class="page-title">{{ server?.name || '...' }}</h1>
    </div>

    <div class="flex border-b border-ink-800">
      <button
        v-for="tab in tabs" :key="tab.key"
        @click="activeTab = tab.key"
        class="px-5 py-2.5 text-sm font-medium transition-colors relative"
        :class="activeTab === tab.key
          ? 'text-brand-400'
          : 'text-ink-500 hover:text-ink-300'"
      >
        {{ tab.label }}
        <span
          v-if="activeTab === tab.key"
          class="absolute bottom-0 left-0 right-0 h-0.5 bg-brand-400 rounded-full"
        ></span>
      </button>
    </div>

    <VouchersView v-if="activeTab === 'vouchers'" />
    <MembersView v-if="activeTab === 'members'" />
    <ProfilesView v-if="activeTab === 'profiles'" />
    <ActiveView v-if="activeTab === 'active'" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useApi } from '@/composables/useApi'
import { useSwal } from '@/composables/useSwal'
import VouchersView from './VouchersView.vue'
import MembersView from './MembersView.vue'
import ProfilesView from './ProfilesView.vue'
import ActiveView from './ActiveView.vue'

defineOptions({ name: 'ServerDetailView' })

const route = useRoute()
const id = computed(() => route.params.id as string)
const { apiGet } = useApi()
const swal = useSwal()
const server = ref<any>(null)
const activeTab = ref('vouchers')
const tabs = [
  { key: 'vouchers', label: 'Voucher' },
  { key: 'members', label: 'Member' },
  { key: 'profiles', label: 'Profile' },
  { key: 'active', label: 'Active' },
]

async function load() { try { server.value = await apiGet<any>(`/servers/${id.value}`) } catch (e: any) { if (e.message !== 'Unauthorized') swal.error(e.message) } }
onMounted(load)
</script>
