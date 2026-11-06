<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import AppSidebar from './components/AppSidebar.vue'

const route = useRoute()
const isAuthLayout = computed(() => route.meta.layout === 'auth')
</script>

<template>
  <UApp :toaster="{ position: 'bottom-right', duration: 3000 }">
    <div v-if="isAuthLayout" class="relative grid min-h-svh place-items-center overflow-hidden px-4 py-12">
      <div
        class="pointer-events-none absolute inset-0 bg-[radial-gradient(60rem_30rem_at_50%_-10%,color-mix(in_oklab,var(--ui-primary)_16%,transparent),transparent)]"
        aria-hidden="true"
      />
      <RouterView v-slot="{ Component }">
        <Transition name="page" mode="out-in">
          <component :is="Component" :key="route.path" class="relative" />
        </Transition>
      </RouterView>
    </div>

    <UDashboardGroup v-else unit="rem" storage-key="briefly-dashboard">
      <AppSidebar />

      <RouterView v-slot="{ Component }">
        <Transition name="page" mode="out-in">
          <div :key="route.path" class="flex min-w-0 flex-1">
            <component :is="Component" />
          </div>
        </Transition>
      </RouterView>
    </UDashboardGroup>
  </UApp>
</template>
