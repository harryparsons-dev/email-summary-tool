<script setup lang="ts">
import type { AuthFormField, FormError, FormSubmitEvent } from '@nuxt/ui'
import { ref } from 'vue'
import { login, signup } from '../services/authService'

type SignupState = { email?: string; password?: string; confirmPassword?: string }

const isSubmitting = ref(false)
const errorMessage = ref('')
const isComplete = ref(false)

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
    description: 'Use 8 or more characters.',
    placeholder: 'At least 8 characters',
    autocomplete: 'new-password',
    icon: 'i-lucide-lock-keyhole',
    size: 'lg',
    required: true,
  },
  {
    name: 'confirmPassword',
    type: 'password',
    label: 'Confirm password',
    placeholder: 'Enter your password again',
    autocomplete: 'new-password',
    icon: 'i-lucide-check-circle-2',
    size: 'lg',
    required: true,
  },
]

function validate(state: Partial<SignupState>): FormError[] {
  const errors: FormError[] = []
  if (!state.email?.trim()) errors.push({ name: 'email', message: 'Enter your email address.' })
  if ((state.password ?? '').length < 8) errors.push({ name: 'password', message: 'Password must be at least 8 characters.' })
  if (state.confirmPassword !== state.password) errors.push({ name: 'confirmPassword', message: 'Passwords do not match.' })
  return errors
}

async function submit(event: FormSubmitEvent<SignupState>) {
  errorMessage.value = ''
  isSubmitting.value = true

  try {
    const credentials = { email: (event.data.email ?? '').trim(), password: event.data.password ?? '' }
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
  <div class="w-full max-w-sm">
    <UPageCard variant="subtle" class="shadow-lg shadow-black/5">
      <Transition name="swap" mode="out-in">
        <UEmpty
          v-if="isComplete"
          icon="i-lucide-party-popper"
          title="Account created"
          description="You’re logged in and ready to go."
          variant="naked"
          :actions="[{ label: 'Open projects', to: '/projects', trailingIcon: 'i-lucide-arrow-right' }]"
          role="status"
        />

        <UAuthForm
          v-else
          icon="i-lucide-user-plus"
          title="Create your account"
          description="Set up your Briefly workspace in seconds."
          :fields="fields"
          :validate="validate"
          :loading="isSubmitting"
          :submit="{ label: 'Create account', size: 'lg', block: true }"
          @submit="submit"
        >
          <template #validation>
            <Transition name="swap">
              <UAlert
                v-if="errorMessage"
                color="error"
                variant="subtle"
                icon="i-lucide-circle-alert"
                title="Check your details"
                :description="errorMessage"
                role="alert"
              />
            </Transition>
          </template>

          <template #footer>
            Already registered?
            <ULink to="/login" class="font-medium text-primary">Log in</ULink>
          </template>
        </UAuthForm>
      </Transition>
    </UPageCard>

    <p class="mt-5 text-center text-xs text-muted">
      By continuing, you agree to the Terms of Service and Privacy Policy.
    </p>
  </div>
</template>
