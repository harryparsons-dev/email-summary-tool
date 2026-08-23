<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { login } from '../services/authService'

const router = useRouter()
const email = ref('')
const password = ref('')
const showPassword = ref(false)
const isSubmitting = ref(false)
const errorMessage = ref('')

async function submit() {
  errorMessage.value = ''
  isSubmitting.value = true

  try {
    await login({
      email: email.value.trim(),
      password: password.value,
    })
    await router.push('/profile')
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Unable to log in.'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <section class="mx-auto w-full max-w-[29rem]" aria-labelledby="login-title">
    <div class="mb-7">
      <UBadge color="neutral" variant="soft" size="md" class="mb-4">Account access</UBadge>
      <h1 id="login-title" class="text-[2rem] leading-tight font-bold tracking-[-0.035em] text-stone-950 dark:text-white">
        Log in
      </h1>
      <p class="mt-2 text-[0.95rem] leading-6 text-stone-600 dark:text-stone-400">
        Use your Briefly account to open your workspace.
      </p>
    </div>

    <UCard
      variant="outline"
      class="rounded-xl bg-white ring-stone-200 dark:bg-[#161617] dark:ring-white/10"
      :ui="{ body: 'p-5 sm:p-7' }"
    >
      <form class="grid gap-5" @submit.prevent="submit">
        <UFormField label="Email address" name="email" required>
          <UInput
            id="login-email"
            v-model="email"
            name="email"
            type="email"
            autocomplete="email"
            placeholder="you@example.com"
            icon="i-lucide-at-sign"
            size="xl"
            class="w-full"
            required
          />
        </UFormField>

        <UFormField label="Password" name="password" required>
          <UInput
            id="login-password"
            v-model="password"
            name="password"
            :type="showPassword ? 'text' : 'password'"
            autocomplete="current-password"
            placeholder="Enter your password"
            icon="i-lucide-lock-keyhole"
            size="xl"
            class="w-full"
            required
          >
            <template #trailing>
              <UTooltip :text="showPassword ? 'Hide password' : 'Show password'">
                <UButton
                  type="button"
                  color="neutral"
                  variant="ghost"
                  size="xs"
                  :icon="showPassword ? 'i-lucide-eye-off' : 'i-lucide-eye'"
                  :aria-label="showPassword ? 'Hide password' : 'Show password'"
                  :aria-pressed="showPassword"
                  @click="showPassword = !showPassword"
                />
              </UTooltip>
            </template>
          </UInput>
        </UFormField>

        <UAlert
          v-if="errorMessage"
          color="error"
          variant="soft"
          icon="i-lucide-circle-alert"
          title="Couldn’t log in"
          :description="errorMessage"
          role="alert"
        />

        <UButton type="submit" color="primary" size="xl" block :loading="isSubmitting" :disabled="isSubmitting">
          {{ isSubmitting ? 'Logging in…' : 'Log in' }}
        </UButton>
      </form>

      <USeparator class="my-6" />

      <p class="text-center text-sm text-stone-600 dark:text-stone-400">
        Don’t have an account?
        <ULink to="/signup" class="font-semibold text-orange-700 hover:text-orange-800 dark:text-orange-400 dark:hover:text-orange-300">
          Create one
        </ULink>
      </p>
    </UCard>

    <p class="mt-5 flex items-center justify-center gap-2 text-xs text-stone-500 dark:text-stone-500">
      <UIcon name="i-lucide-shield-check" class="size-3.5" aria-hidden="true" />
      Your session is stored in a secure, HTTP-only cookie.
    </p>
  </section>
</template>
