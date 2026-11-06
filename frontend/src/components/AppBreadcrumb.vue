<script setup lang="ts">
import type { BreadcrumbItem } from '@nuxt/ui'

// The last item is the current page. Set `loading` to show a placeholder for it.
defineProps<{
  items: BreadcrumbItem[]
  loading?: boolean
}>()
</script>

<template>
  <UBreadcrumb :items="items" :ui="{ root: 'min-w-0', list: 'min-w-0', item: 'min-w-0 last:font-semibold', link: '[&[href]]:hover:underline underline-offset-4', linkLabel: 'truncate' }">
    <template #item-label="{ item, active }">
      <USkeleton v-if="active && loading" class="h-4 w-32" />
      <template v-else>{{ item.label }}</template>
    </template>
  </UBreadcrumb>
</template>
