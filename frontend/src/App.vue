<script setup lang="ts">
import AppNavbar from './components/AppNavbar.vue'
</script>

<template>
  <UApp>
    <div
      class="app-shell min-h-screen bg-stone-50 font-sans text-stone-950 transition-colors duration-200 dark:bg-[#0c0c0d] dark:text-stone-100 [--ui-primary:var(--color-orange-600)] dark:[--ui-primary:var(--color-orange-500)]"
    >
      <AppNavbar />

      <main class="min-h-[calc(100vh-4rem)] lg:min-h-screen lg:pl-66">
        <div class="mx-auto flex min-h-[calc(100vh-4rem)] w-full max-w-6xl items-center px-5 py-10 sm:px-8 lg:min-h-screen lg:px-12 lg:py-14">
          <RouterView />
        </div>
      </main>
    </div>
  </UApp>
</template>
