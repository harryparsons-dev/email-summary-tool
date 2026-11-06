<script setup lang="ts">
import type { AuthFormField, FormSubmitEvent } from '@nuxt/ui'
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { login } from '../services/authService'

type LoginState = { email: string; password: string }

const router = useRouter()
const isSubmitting = ref(false)
const errorMessage = ref('')

const fields: AuthFormField[] = [
  {
    name: 'email',
    type: 'email',
    label: 'Email address',
    placeholder: 'you@example.com',
    autocomplete: 'email',
    icon: 'i-lucide-at-sign',
    size: 'lg',
    required: true,
  },
  {
    name: 'password',
    type: 'password',
    label: 'Password',
    placeholder: 'Enter your password',
    autocomplete: 'current-password',
    icon: 'i-lucide-lock-keyhole',
    size: 'lg',
    required: true,
  },
]

async function submit(event: FormSubmitEvent<LoginState>) {
  errorMessage.value = ''
  isSubmitting.value = true

  try {
    await login({ email: event.data.email.trim(), password: event.data.password })
    await router.push('/projects')
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Unable to log in.'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <div class="w-full max-w-sm">
    <UPageCard variant="subtle" class="shadow-lg shadow-black/5">
      <UAuthForm
        icon="i-lucide-mail"
        title="Welcome back"
        description="Log in to your Briefly workspace."
        :fields="fields"
        :loading="isSubmitting"
        :submit="{ label: 'Log in', size: 'lg', block: true }"
        @submit="submit"
      >
        <template #validation>
          <Transition name="swap">
            <UAlert
              v-if="errorMessage"
              color="error"
              variant="subtle"
              icon="i-lucide-circle-alert"
              title="Couldn’t log in"
              :description="errorMessage"
              role="alert"
            />
          </Transition>
        </template>

        <template #footer>
          Don’t have an account?
          <ULink to="/signup" class="font-medium text-primary">Create one</ULink>
        </template>
      </UAuthForm>
    </UPageCard>

    <p class="mt-5 flex items-center justify-center gap-1.5 text-xs text-muted">
      <UIcon name="i-lucide-shield-check" class="size-3.5" aria-hidden="true" />
      Your session is stored in a secure, HTTP-only cookie.
    </p>
  </div>
</template>
