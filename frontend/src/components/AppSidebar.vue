<script setup lang="ts">
import type { DropdownMenuItem, NavigationMenuItem } from '@nuxt/ui'
import { useToast } from '@nuxt/ui/composables'
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { currentUser as user, logout } from '../services/authService'
import { navbarItems } from './navbarItems'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const open = ref(false)

// Nested pages such as /projects/:id keep their section highlighted.
const items = computed<NavigationMenuItem[]>(() =>
  navbarItems.map((item) => {
    const path = String(item.to)
    return {
      ...item,
      active: route.path === path || route.path.startsWith(`${path}/`),
      onSelect: () => (open.value = false),
    }
  }),
)

const userMenu = computed<DropdownMenuItem[][]>(() => [
  [{ type: 'label', label: user.value?.email ?? 'Signed in', avatar: { icon: 'i-lucide-user' } }],
  [{ label: 'Account', icon: 'i-lucide-circle-user-round', to: '/account' }],
  [{ label: 'Sign out', icon: 'i-lucide-log-out', color: 'error', onSelect: signOut }],
])

async function signOut() {
  try {
    await logout()
    await router.push('/login')
  } catch {
    toast.add({ title: 'Sign out failed', description: 'Try again.', color: 'error', icon: 'i-lucide-circle-alert' })
  }
}
</script>

<template>
  <UDashboardSidebar
    v-model:open="open"
    collapsible
    resizable
    :default-size="16"
    :min-size="12"
    :max-size="22"
    :ui="{ footer: 'border-t border-default' }"
  >
    <template #header="{ collapsed }">
      <!-- When collapsed there is only room for one button, so the logo makes way for the expand button. -->
      <RouterLink v-if="!collapsed" to="/projects" class="flex min-w-0 items-center gap-2.5" aria-label="Briefly home">
        <UAvatar icon="i-lucide-mail" size="sm" :ui="{ root: 'bg-primary', icon: 'text-inverted' }" />
        <span class="truncate font-semibold text-highlighted">Briefly</span>
      </RouterLink>
      <UDashboardSidebarCollapse :class="collapsed ? 'mx-auto' : 'ms-auto'" />
    </template>

    <template #default="{ collapsed }">
      <UNavigationMenu :items="items" orientation="vertical" :collapsed="collapsed" tooltip highlight />
    </template>

    <template #footer="{ collapsed }">
      <UDropdownMenu
        :items="userMenu"
        :content="{ align: 'start', side: collapsed ? 'right' : 'top' }"
        :ui="{ content: collapsed ? 'w-56' : 'w-(--reka-dropdown-menu-trigger-width)' }"
      >
        <UButton
          color="neutral"
          variant="ghost"
          :square="collapsed"
          :avatar="{ text: user?.email.charAt(0).toUpperCase() ?? '?', size: 'xs' }"
          :label="collapsed ? undefined : user?.email ?? 'Account'"
          :trailing-icon="collapsed ? undefined : 'i-lucide-chevrons-up-down'"
          class="min-w-0 flex-1 data-[state=open]:bg-elevated"
          :ui="{ label: 'truncate', trailingIcon: 'text-dimmed ms-auto' }"
        />
      </UDropdownMenu>
      <UColorModeButton v-if="!collapsed" color="neutral" variant="ghost" class="shrink-0" />
    </template>
  </UDashboardSidebar>
</template>
