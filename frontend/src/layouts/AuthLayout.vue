<template>
  <div class="min-h-screen bg-ink-950 flex items-center justify-center p-4">
    <div class="absolute inset-0 overflow-hidden pointer-events-none">
      <div class="absolute top-1/4 left-1/4 w-96 h-96 bg-brand-600/10 rounded-full blur-3xl"></div>
      <div class="absolute bottom-1/4 right-1/4 w-96 h-96 bg-accent-600/10 rounded-full blur-3xl"></div>
    </div>
    <div class="relative w-full max-w-md animate-fade-in">
      <slot />
    </div>
  </div>
</template>
