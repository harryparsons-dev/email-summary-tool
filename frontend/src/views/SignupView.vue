<script setup lang="ts">
import { ref } from 'vue'
import { authenticate } from '../lib/auth'

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
    await authenticate('/signup', credentials)
    await authenticate('/login', credentials)
    isComplete.value = true
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Unable to create your account.'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <section class="auth-layout" aria-labelledby="signup-title">
    <div class="auth-intro">
      <p class="eyebrow">A calmer inbox starts here</p>
      <h1 id="signup-title">Make email feel manageable.</h1>
      <p>
        Create your account and turn long conversations into short, useful summaries.
      </p>
      <div class="intro-note" aria-hidden="true">
        <span class="intro-note-icon">✦</span>
        <span>Get the context. Keep your focus.</span>
      </div>
    </div>

    <div class="auth-card">
      <template v-if="isComplete">
        <div class="success-state" role="status">
          <span class="success-icon" aria-hidden="true">✓</span>
          <h2>Your account is ready</h2>
          <p>You’re signed up and logged in to Briefly.</p>
        </div>
      </template>

      <template v-else>
        <div class="card-heading">
          <h2>Create your account</h2>
          <p>It only takes a minute to get started.</p>
        </div>

        <form class="auth-form" @submit.prevent="submit">
          <div class="form-field">
            <label for="signup-email">Email address</label>
            <input
              id="signup-email"
              v-model="email"
              name="email"
              type="email"
              autocomplete="email"
              placeholder="you@example.com"
              required
            />
          </div>

          <div class="form-field">
            <label for="signup-password">Password</label>
            <div class="password-field">
              <input
                id="signup-password"
                v-model="password"
                name="password"
                :type="showPassword ? 'text' : 'password'"
                autocomplete="new-password"
                placeholder="At least 8 characters"
                minlength="8"
                required
              />
              <button
                class="password-toggle"
                type="button"
                :aria-label="showPassword ? 'Hide passwords' : 'Show passwords'"
                :aria-pressed="showPassword"
                @click="showPassword = !showPassword"
              >
                {{ showPassword ? 'Hide' : 'Show' }}
              </button>
            </div>
            <span class="field-hint">Use 8 or more characters.</span>
          </div>

          <div class="form-field">
            <label for="confirm-password">Confirm password</label>
            <input
              id="confirm-password"
              v-model="confirmPassword"
              name="confirm-password"
              :type="showPassword ? 'text' : 'password'"
              autocomplete="new-password"
              placeholder="Enter your password again"
              minlength="8"
              required
            />
          </div>

          <p v-if="errorMessage" class="form-message form-error" role="alert">
            {{ errorMessage }}
          </p>

          <button class="submit-button" type="submit" :disabled="isSubmitting">
            <span v-if="isSubmitting" class="spinner" aria-hidden="true"></span>
            {{ isSubmitting ? 'Creating account…' : 'Create account' }}
          </button>
        </form>

        <p class="terms-copy">
          By creating an account, you agree to the Terms of Service and Privacy Policy.
        </p>
        <p class="switch-auth">
          Already have an account? <RouterLink to="/login">Log in</RouterLink>
        </p>
      </template>
    </div>
  </section>
</template>
