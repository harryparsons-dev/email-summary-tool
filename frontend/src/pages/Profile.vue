<script setup lang="ts">
import { onMounted, ref } from 'vue'
import AppBreadcrumb from '../components/AppBreadcrumb.vue'
import CopyButton from '../components/CopyButton.vue'
import type { User } from '../models/user'
import { getProfile } from '../services/profileService'

const user = ref<User | null>(null)
const isLoading = ref(true)
const errorMessage = ref('')

async function loadProfile() {
  isLoading.value = true
  errorMessage.value = ''

  try {
    user.value = await getProfile()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Unable to load your profile.'
  } finally {
    isLoading.value = false
  }
}

onMounted(loadProfile)
</script>

<template>
  <UDashboardPanel id="account">
    <template #header>
      <UDashboardNavbar>
        <template #title>
          <AppBreadcrumb :items="[{ label: 'Account' }]" />
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
      <div class="mx-auto grid w-full max-w-3xl gap-6">
        <Transition name="swap" mode="out-in">
          <UPageCard v-if="isLoading" key="loading" variant="subtle" role="status" aria-live="polite">
            <div class="flex items-center gap-4">
              <USkeleton class="size-12 rounded-full" />
              <div class="grid flex-1 gap-2">
                <USkeleton class="h-4 w-40" />
                <USkeleton class="h-3.5 w-24" />
              </div>
            </div>
            <USkeleton class="h-10 w-full" />
            <USkeleton class="h-10 w-full" />
            <span class="sr-only">Loading your profile…</span>
          </UPageCard>

          <UAlert
            v-else-if="errorMessage"
            key="error"
            color="error"
            variant="subtle"
            icon="i-lucide-cloud-off"
            title="We couldn’t load your account"
            :description="errorMessage"
            :actions="[{ label: 'Try again', icon: 'i-lucide-refresh-cw', color: 'error', variant: 'outline', onClick: loadProfile }]"
            role="alert"
          />

          <UPageCard
            v-else-if="user"
            key="profile"
            title="Profile"
            description="Your identity and workspace access details."
            variant="subtle"
            :ui="{ header: 'w-full', wrapper: 'w-full' }"
          >
            <template #header>
              <div class="flex items-center justify-between gap-4">
                <UUser
                  :name="user.email"
                  description="Workspace member"
                  :avatar="{ text: user.email.charAt(0).toUpperCase(), ui: { root: 'bg-primary/10', fallback: 'text-primary font-semibold' } }"
                  size="xl"
                  class="min-w-0"
                  :ui="{ name: 'truncate' }"
                />
                <UBadge color="success" variant="subtle" icon="i-lucide-badge-check" label="Active" />
              </div>
              <USeparator class="mt-6" />
            </template>

            <div class="divide-y divide-default">
              <div class="flex flex-wrap items-center justify-between gap-x-6 gap-y-1 py-3">
                <span class="flex items-center gap-2 text-sm text-muted">
                  <UIcon name="i-lucide-at-sign" class="size-4" aria-hidden="true" />
                  Email address
                </span>
                <span class="min-w-0 text-sm font-medium [overflow-wrap:anywhere] text-highlighted">{{ user.email }}</span>
              </div>
              <div class="flex flex-wrap items-center justify-between gap-x-6 gap-y-1 py-3">
                <span class="flex items-center gap-2 text-sm text-muted">
                  <UIcon name="i-lucide-fingerprint" class="size-4" aria-hidden="true" />
                  User ID
                </span>
                <span class="flex min-w-0 items-center gap-1">
                  <span class="font-mono text-xs [overflow-wrap:anywhere] text-toned">{{ user.id }}</span>
                  <CopyButton :value="user.id" label="Copy user ID" />
                </span>
              </div>
            </div>
          </UPageCard>
        </Transition>

        <UPageCard
          title="Appearance"
          description="Choose how Briefly looks on this device."
          variant="subtle"
          orientation="horizontal"
          :ui="{ container: 'items-center' }"
        >
          <UColorModeSelect class="w-40 justify-self-end max-lg:justify-self-start" />
        </UPageCard>

        <UAlert
          color="neutral"
          variant="subtle"
          icon="i-lucide-info"
          title="Account changes"
          description="Email and password management will appear here when those controls are available."
        />
      </div>
    </template>
  </UDashboardPanel>
</template>
