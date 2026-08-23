<script setup lang="ts">
import { onMounted, ref } from 'vue'
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
  <section class="mx-auto w-full max-w-4xl self-start lg:pt-[8vh]" aria-labelledby="profile-title">
    <div class="mb-8 flex flex-wrap items-start justify-between gap-4">
      <div>
        <div class="mb-3 flex items-center gap-2 text-xs font-medium text-stone-500 dark:text-stone-500">
          <span>Workspace</span>
          <UIcon name="i-lucide-chevron-right" class="size-3.5" aria-hidden="true" />
          <span>Account</span>
        </div>
        <h1 id="profile-title" class="text-[2rem] leading-tight font-bold tracking-[-0.035em] text-stone-950 dark:text-white">
          Account
        </h1>
        <p class="mt-2 text-[0.95rem] leading-6 text-stone-600 dark:text-stone-400">
          Your identity and workspace access details.
        </p>
      </div>
      <UBadge color="success" variant="subtle" size="md" class="mt-7">
        <span class="mr-1.5 size-1.5 rounded-full bg-emerald-500" aria-hidden="true" />
        Active
      </UBadge>
    </div>

    <UCard
      variant="outline"
      class="rounded-xl bg-white ring-stone-200 dark:bg-[#161617] dark:ring-white/10"
      :ui="{ body: 'p-5 sm:p-7' }"
    >
      <div v-if="isLoading" class="grid min-h-52 content-center gap-5" role="status" aria-live="polite">
        <div class="flex items-center gap-4">
          <USkeleton class="size-12 rounded-xl" />
          <div class="grid flex-1 gap-2">
            <USkeleton class="h-4 w-36" />
            <USkeleton class="h-3.5 w-52 max-w-full" />
          </div>
        </div>
        <USeparator />
        <USkeleton class="h-12 w-full" />
        <USkeleton class="h-12 w-full" />
        <span class="sr-only">Loading your profile…</span>
      </div>

      <UAlert
        v-else-if="errorMessage"
        color="error"
        variant="soft"
        icon="i-lucide-cloud-off"
        title="We couldn’t load your account"
        :description="errorMessage"
        role="alert"
      >
        <template #actions>
          <UButton color="error" variant="soft" size="sm" icon="i-lucide-refresh-cw" @click="loadProfile">
            Try again
          </UButton>
        </template>
      </UAlert>

      <template v-else-if="user">
        <div class="flex items-center gap-4">
          <UAvatar
            :text="user.email.charAt(0).toUpperCase()"
            size="2xl"
            class="shrink-0 bg-orange-100 font-bold text-orange-800 ring-1 ring-orange-200 dark:bg-orange-500/15 dark:text-orange-300 dark:ring-orange-400/20"
          />
          <div class="min-w-0">
            <h2 class="truncate text-base font-semibold text-stone-950 dark:text-white">{{ user.email }}</h2>
            <p class="mt-1 text-sm text-stone-500">Workspace member</p>
          </div>
        </div>

        <USeparator class="my-7" />

        <dl class="grid">
          <div class="grid grid-cols-[9rem_minmax(0,1fr)] items-center gap-5 border-b border-stone-200 py-4 first:pt-0 max-sm:grid-cols-1 max-sm:gap-1.5 dark:border-white/10">
            <dt class="flex items-center gap-2 text-sm font-medium text-stone-500 dark:text-stone-500">
              <UIcon name="i-lucide-at-sign" class="size-4" aria-hidden="true" />
              Email address
            </dt>
            <dd class="min-w-0 [overflow-wrap:anywhere] text-sm font-medium text-stone-800 dark:text-stone-200">{{ user.email }}</dd>
          </div>
          <div class="grid grid-cols-[9rem_minmax(0,1fr)] items-center gap-5 py-4 pb-0 max-sm:grid-cols-1 max-sm:gap-1.5">
            <dt class="flex items-center gap-2 text-sm font-medium text-stone-500 dark:text-stone-500">
              <UIcon name="i-lucide-fingerprint" class="size-4" aria-hidden="true" />
              User ID
            </dt>
            <dd class="min-w-0 [overflow-wrap:anywhere] font-mono text-xs text-stone-700 dark:text-stone-300">{{ user.id }}</dd>
          </div>
        </dl>
      </template>
    </UCard>

    <UAlert
      color="neutral"
      variant="subtle"
      icon="i-lucide-info"
      title="Account changes"
      description="Email and password management will appear here when those controls are available."
      class="mt-5"
    />
  </section>
</template>
