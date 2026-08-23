<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { isAuthenticated, logout } from '../services/authService'
import { navbarItems } from './navbarItems'

const route = useRoute()
const router = useRouter()
const isSigningOut = ref(false)
const signOutFailed = ref(false)
const mobileNavigationOpen = ref(false)
const desktopMediaQuery = window.matchMedia('(min-width: 1024px)')
const isDesktop = ref(desktopMediaQuery.matches)

const savedTheme = window.localStorage.getItem('briefly-theme')
const isDark = ref(
  savedTheme === 'dark' ||
    (savedTheme === null && window.matchMedia('(prefers-color-scheme: dark)').matches),
)

const visibleNavbarItems = computed(() =>
  navbarItems.filter((item) => {
    if (item.requiresAuthentication && !isAuthenticated.value) {
      return false
    }

    if (item.guestOnly && isAuthenticated.value) {
      return false
    }

    return true
  }),
)

function syncViewport(event: MediaQueryListEvent) {
  isDesktop.value = event.matches

  if (event.matches) {
    mobileNavigationOpen.value = false
  }
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    mobileNavigationOpen.value = false
  }
}

onMounted(() => {
  desktopMediaQuery.addEventListener('change', syncViewport)
  window.addEventListener('keydown', handleKeydown)
})

onBeforeUnmount(() => {
  desktopMediaQuery.removeEventListener('change', syncViewport)
  window.removeEventListener('keydown', handleKeydown)
})

watch(
  isDark,
  (dark) => {
    document.documentElement.classList.toggle('dark', dark)
    document.documentElement.style.colorScheme = dark ? 'dark' : 'light'
    document.querySelector('meta[name="theme-color"]')?.setAttribute('content', dark ? '#0c0c0d' : '#fafaf9')
    window.localStorage.setItem('briefly-theme', dark ? 'dark' : 'light')
  },
  { immediate: true },
)

watch(
  () => route.fullPath,
  () => {
    mobileNavigationOpen.value = false
    signOutFailed.value = false
  },
)

async function signOut() {
  isSigningOut.value = true
  signOutFailed.value = false

  try {
    await logout()
    await router.push('/login')
  } catch {
    signOutFailed.value = true
  } finally {
    isSigningOut.value = false
  }
}
</script>

<template>
  <header
    class="sticky top-0 z-30 flex h-16 items-center justify-between border-b border-stone-200 bg-stone-50/95 px-4 backdrop-blur lg:hidden dark:border-white/10 dark:bg-[#0c0c0d]/95"
  >
    <RouterLink to="/" class="inline-flex items-center gap-2.5 no-underline" aria-label="Briefly home">
      <span class="grid size-8 place-items-center rounded-lg bg-orange-600 text-white dark:bg-orange-500">
        <UIcon name="i-lucide-mail" class="size-4.5" aria-hidden="true" />
      </span>
      <span class="text-sm font-bold tracking-[-0.01em] text-stone-950 dark:text-white">Briefly</span>
    </RouterLink>

    <UButton
      color="neutral"
      variant="ghost"
      size="md"
      :icon="mobileNavigationOpen ? 'i-lucide-x' : 'i-lucide-menu'"
      :aria-label="mobileNavigationOpen ? 'Close navigation' : 'Open navigation'"
      :aria-expanded="mobileNavigationOpen"
      @click="mobileNavigationOpen = !mobileNavigationOpen"
    />
  </header>

  <Transition name="fade">
    <button
      v-if="mobileNavigationOpen"
      class="fixed inset-x-0 bottom-0 top-16 z-30 bg-black/45 lg:hidden"
      type="button"
      aria-label="Close navigation"
      @click="mobileNavigationOpen = false"
    />
  </Transition>

  <aside
    class="fixed inset-y-0 left-0 z-40 flex w-66 flex-col border-r border-stone-200 bg-[#f4f3f1] px-3 py-4 transition-transform duration-200 max-lg:top-16 max-lg:w-72 max-lg:-translate-x-full lg:translate-x-0 dark:border-white/10 dark:bg-[#131314]"
    :class="{ 'max-lg:translate-x-0': mobileNavigationOpen }"
    :inert="!isDesktop && !mobileNavigationOpen"
    :aria-hidden="!isDesktop && !mobileNavigationOpen"
  >
    <RouterLink
      to="/"
      class="mb-8 hidden items-center gap-3 px-2.5 py-1 no-underline lg:flex"
      aria-label="Briefly home"
    >
      <span class="grid size-9 place-items-center rounded-[0.7rem] bg-orange-600 text-white shadow-sm dark:bg-orange-500">
        <UIcon name="i-lucide-mail" class="size-5" aria-hidden="true" />
      </span>
      <span>
        <span class="block text-sm font-bold tracking-[-0.01em] text-stone-950 dark:text-white">Briefly</span>
        <span class="block text-[0.68rem] font-medium text-stone-500 dark:text-stone-500">Email summaries</span>
      </span>
    </RouterLink>

    <nav class="flex-1" aria-label="Primary navigation">
      <p class="px-3 pb-2 text-[0.68rem] font-semibold tracking-[0.08em] text-stone-500 uppercase dark:text-stone-500">
        {{ isAuthenticated ? 'Workspace' : 'Account' }}
      </p>
      <div class="grid gap-1">
        <UButton
          v-for="item in visibleNavbarItems"
          :key="item.to"
          :to="item.to"
          :icon="item.icon"
          color="neutral"
          :variant="route.path === item.to ? 'soft' : 'ghost'"
          size="lg"
          class="justify-start px-3 font-medium"
          :class="route.path === item.to ? 'nav-active' : ''"
        >
          {{ item.label }}
        </UButton>
      </div>
    </nav>

    <div class="grid gap-2 border-t border-stone-200 pt-3 dark:border-white/10">
      <div class="flex min-h-11 items-center justify-between gap-3 rounded-lg px-3">
        <div class="flex min-w-0 items-center gap-3">
          <UIcon
            :name="isDark ? 'i-lucide-moon' : 'i-lucide-sun'"
            class="size-4.5 shrink-0 text-stone-500 dark:text-stone-400"
            aria-hidden="true"
          />
          <span class="truncate text-sm font-medium text-stone-700 dark:text-stone-300">
            {{ isDark ? 'Dark mode' : 'Light mode' }}
          </span>
        </div>
        <USwitch v-model="isDark" size="sm" aria-label="Toggle dark mode" />
      </div>

      <UAlert
        v-if="signOutFailed"
        color="error"
        variant="soft"
        description="Sign out failed. Try again."
        class="mb-1"
      />

      <UButton
        v-if="isAuthenticated"
        color="neutral"
        variant="ghost"
        icon="i-lucide-log-out"
        size="lg"
        class="justify-start px-3"
        :loading="isSigningOut"
        :disabled="isSigningOut"
        @click="signOut"
      >
        Sign out
      </UButton>

      <div class="flex items-center gap-2 px-3 pt-2 text-[0.68rem] text-stone-400 dark:text-stone-600">
        <span class="size-1.5 rounded-full bg-orange-500" aria-hidden="true" />
        <span>Briefly workspace</span>
      </div>
    </div>
  </aside>
</template>
