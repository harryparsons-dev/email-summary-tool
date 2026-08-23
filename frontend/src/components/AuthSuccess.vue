<script setup lang="ts">
defineProps<{
  title: string
  description: string
}>()
</script>

<template>
  <div class="grid min-h-72 place-items-center content-center text-center" role="status">
    <div class="mb-5 grid size-12 place-items-center rounded-xl bg-orange-100 text-orange-700 ring-1 ring-orange-200 dark:bg-orange-500/15 dark:text-orange-300 dark:ring-orange-400/20">
      <UIcon name="i-lucide-check" class="size-6" aria-hidden="true" />
    </div>
    <h2 class="text-xl font-bold tracking-[-0.02em] text-stone-950 dark:text-white">{{ title }}</h2>
    <p class="mt-2 text-sm text-stone-600 dark:text-stone-400">{{ description }}</p>
    <UButton to="/profile" color="primary" size="lg" trailing-icon="i-lucide-arrow-right" class="mt-6">
      Open account
    </UButton>
  </div>
</template>
