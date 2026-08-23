<script setup lang="ts">
import { ref } from 'vue'
import AuthSuccess from '../components/AuthSuccess.vue'
import { login, signup } from '../services/authService'

const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const showPassword = ref(false)
const isSubmitting = ref(false)
const errorMessage = ref('')
const isComplete = ref(false)

async function submit() {
  errorMessage.value = ''

  if (password.value.length < 8) {
    errorMessage.value = 'Password must be at least 8 characters.'
    return
  }

  if (password.value !== confirmPassword.value) {
    errorMessage.value = 'Passwords do not match.'
    return
  }

  isSubmitting.value = true

  try {
    const credentials = { email: email.value.trim(), password: password.value }
    await signup(credentials)
    await login(credentials)
    isComplete.value = true
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Unable to create your account.'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <section class="mx-auto w-full max-w-[29rem]" aria-labelledby="signup-title">
    <div class="mb-7">
      <UBadge color="neutral" variant="soft" size="md" class="mb-4">New account</UBadge>
      <h1 id="signup-title" class="text-[2rem] leading-tight font-bold tracking-[-0.035em] text-stone-950 dark:text-white">
        Create your account
      </h1>
      <p class="mt-2 text-[0.95rem] leading-6 text-stone-600 dark:text-stone-400">
        Set up your workspace credentials.
      </p>
    </div>

    <UCard
      variant="outline"
      class="rounded-xl bg-white ring-stone-200 dark:bg-[#161617] dark:ring-white/10"
      :ui="{ body: 'p-5 sm:p-7' }"
    >
      <AuthSuccess
        v-if="isComplete"
        title="Account created"
        description="You’re logged in and ready to continue."
      />

      <form v-else class="grid gap-5" @submit.prevent="submit">
        <UFormField label="Email address" name="email" required>
          <UInput
            id="signup-email"
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

        <UFormField label="Password" name="password" description="Use 8 or more characters." required>
          <UInput
            id="signup-password"
            v-model="password"
            name="password"
            :type="showPassword ? 'text' : 'password'"
            autocomplete="new-password"
            placeholder="At least 8 characters"
            icon="i-lucide-lock-keyhole"
            minlength="8"
            size="xl"
            class="w-full"
            required
          >
            <template #trailing>
              <UTooltip :text="showPassword ? 'Hide passwords' : 'Show passwords'">
                <UButton
                  type="button"
                  color="neutral"
                  variant="ghost"
                  size="xs"
                  :icon="showPassword ? 'i-lucide-eye-off' : 'i-lucide-eye'"
                  :aria-label="showPassword ? 'Hide passwords' : 'Show passwords'"
                  :aria-pressed="showPassword"
                  @click="showPassword = !showPassword"
                />
              </UTooltip>
            </template>
          </UInput>
        </UFormField>

        <UFormField label="Confirm password" name="confirm-password" required>
          <UInput
            id="confirm-password"
            v-model="confirmPassword"
            name="confirm-password"
            :type="showPassword ? 'text' : 'password'"
            autocomplete="new-password"
            placeholder="Enter your password again"
            icon="i-lucide-check-circle-2"
            minlength="8"
            size="xl"
            class="w-full"
            required
          />
        </UFormField>

        <UAlert
          v-if="errorMessage"
          color="error"
          variant="soft"
          icon="i-lucide-circle-alert"
          title="Check your details"
          :description="errorMessage"
          role="alert"
        />

        <UButton type="submit" color="primary" size="xl" block :loading="isSubmitting" :disabled="isSubmitting">
          {{ isSubmitting ? 'Creating account…' : 'Create account' }}
        </UButton>

        <p class="text-center text-xs leading-5 text-stone-500 dark:text-stone-500">
          By continuing, you agree to the Terms of Service and Privacy Policy.
        </p>

        <USeparator />

        <p class="text-center text-sm text-stone-600 dark:text-stone-400">
          Already registered?
          <ULink to="/login" class="font-semibold text-orange-700 hover:text-orange-800 dark:text-orange-400 dark:hover:text-orange-300">
            Log in
          </ULink>
        </p>
      </form>
    </UCard>
  </section>
</template>
