<template>
  <div class="min-h-screen bg-ink-950 flex">
    <!-- Sidebar -->
    <aside
      class="fixed lg:sticky top-0 left-0 z-40 h-screen w-64 bg-ink-900 border-r border-ink-800 flex flex-col transition-transform duration-300"
      :class="sidebarOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'"
    >
      <div class="h-16 flex items-center gap-2 px-5 border-b border-ink-800">
        <div v-if="settings.logoUrl()" class="w-8 h-8 rounded-lg flex items-center justify-center overflow-hidden">
          <img :src="settings.logoUrl()" class="w-full h-full object-contain" />
        </div>
        <div v-else class="w-8 h-8 rounded-lg bg-brand-600 flex items-center justify-center shadow-glow-sm">
          <svg class="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
          </svg>
        </div>
        <div>
          <div class="font-bold text-ink-100 leading-none">{{ settings.appName() }}</div>
          <div class="text-[10px] text-ink-500 leading-none mt-0.5">Mikrotik Manager</div>
        </div>
      </div>

      <nav class="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
        <RouterLink
          v-for="item in navItems"
          :key="item.to"
          :to="item.to"
          class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-all duration-200"
          :class="$route.path === item.to || (item.to !== '/' && $route.path.startsWith(item.to))
            ? 'bg-brand-600/15 text-brand-400 shadow-glow-sm'
            : 'text-ink-400 hover:text-ink-200 hover:bg-ink-800'"
        >
          <component :is="item.icon" class="w-5 h-5 shrink-0" />
          <span>{{ item.label }}</span>
        </RouterLink>
      </nav>

      <div class="px-3 py-4 border-t border-ink-800">
        <div class="flex items-center gap-3 px-3 py-2 rounded-lg bg-ink-850">
          <div class="w-8 h-8 rounded-full bg-brand-600/20 flex items-center justify-center text-brand-400 font-bold text-sm uppercase">
            {{ authStore.email.charAt(0) }}
          </div>
          <div class="min-w-0 flex-1">
            <div class="text-sm text-ink-200 truncate">{{ authStore.email }}</div>
            <div class="text-xs text-ink-500">{{ authStore.role }}</div>
          </div>
          <button class="btn-ghost p-1.5" title="Keluar" @click="logout">
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
            </svg>
          </button>
        </div>
      </div>
    </aside>

    <!-- Sidebar overlay (mobile) -->
    <div v-if="sidebarOpen" class="fixed inset-0 z-30 bg-black/60 lg:hidden" @click="sidebarOpen = false" />

    <!-- Main -->
    <div class="flex-1 flex flex-col min-w-0">
      <!-- Topbar (mobile) -->
      <header class="lg:hidden h-14 flex items-center gap-3 px-4 bg-ink-900 border-b border-ink-800 sticky top-0 z-20">
        <button class="btn-ghost p-2" @click="sidebarOpen = true">
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>
        <span class="font-bold text-ink-100">{{ settings.appName() }}</span>
      </header>

      <main class="flex-1 px-4 sm:px-6 lg:px-8 py-6 max-w-7xl w-full mx-auto">
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, h } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSettings } from '@/composables/useSettings'

const authStore = useAuthStore()
const router = useRouter()
const settings = useSettings()
const sidebarOpen = ref(false)

const icon = (path: string) => h('svg', {
  class: 'w-5 h-5', fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': 2,
  innerHTML: path,
})

const navItems = [
  { to: '/',        label: 'Dashboard', icon: icon('<path stroke-linecap="round" stroke-linejoin="round" d="M4 5a1 1 0 011-1h5a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM13 5a1 1 0 011-1h5a1 1 0 011 1v3a1 1 0 01-1 1h-5a1 1 0 01-1-1V5zM13 12a1 1 0 011-1h5a1 1 0 011 1v6a1 1 0 01-1 1h-5a1 1 0 01-1-1v-6zM4 16a1 1 0 011-1h5a1 1 0 011 1v3a1 1 0 01-1 1H5a1 1 0 01-1-1v-3z" />') },
  { to: '/servers', label: 'Server',    icon: icon('<path stroke-linecap="round" stroke-linejoin="round" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />') },
  { to: '/tokens',  label: 'API Token', icon: icon('<path stroke-linecap="round" stroke-linejoin="round" d="M15 7a4 4 0 11-8 0 4 4 0 018 0zM21 21l-4-4m0 0l-3-3m3 3l-3 3m3-3l3-3" />') },
  { to: '/cronjobs', label: 'Cronjob',  icon: icon('<path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6l4 2m6-2a10 10 0 11-20 0 10 10 0 0120 0z" />') },
  { to: '/audit-logs', label: 'Activity Log', icon: icon('<path stroke-linecap="round" stroke-linejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />') },
  { to: '/settings', label: 'Settings', icon: icon('<path stroke-linecap="round" stroke-linejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" /><path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />') },
]

function logout() {
  authStore.logout()
  router.push('/login')
}
</script>
