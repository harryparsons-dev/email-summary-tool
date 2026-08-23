<script setup lang="ts">
import { ref } from 'vue'
import { login } from '../services/authService'

const email = ref('')
const password = ref('')
const showPassword = ref(false)
const isSubmitting = ref(false)
const errorMessage = ref('')
const isComplete = ref(false)

async function submit() {
  errorMessage.value = ''
  isSubmitting.value = true

  try {
    await login({
      email: email.value.trim(),
      password: password.value,
    })
    isComplete.value = true
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Unable to log in.'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <section class="auth-layout" aria-labelledby="login-title">
    <div class="auth-intro">
      <p class="eyebrow">Welcome back</p>
      <h1 id="login-title">Your inbox, distilled.</h1>
      <p>
        Log in to catch up on the messages that matter, without working through every thread.
      </p>
      <div class="intro-note" aria-hidden="true">
        <span class="intro-note-icon">✓</span>
        <span>Clear summaries. Less inbox noise.</span>
      </div>
    </div>

    <div class="auth-card">
      <template v-if="isComplete">
        <div class="success-state" role="status">
          <span class="success-icon" aria-hidden="true">✓</span>
          <h2>You’re logged in</h2>
          <p>Your Briefly session is ready.</p>
        </div>
      </template>

      <template v-else>
        <div class="card-heading">
          <h2>Log in to Briefly</h2>
          <p>Enter your account details below.</p>
        </div>

        <form class="auth-form" @submit.prevent="submit">
          <div class="form-field">
            <label for="login-email">Email address</label>
            <input
              id="login-email"
              v-model="email"
              name="email"
              type="email"
              autocomplete="email"
              placeholder="you@example.com"
              required
            />
          </div>

          <div class="form-field">
            <label for="login-password">Password</label>
            <div class="password-field">
              <input
                id="login-password"
                v-model="password"
                name="password"
                :type="showPassword ? 'text' : 'password'"
                autocomplete="current-password"
                placeholder="Enter your password"
                required
              />
              <button
                class="password-toggle"
                type="button"
                :aria-label="showPassword ? 'Hide password' : 'Show password'"
                :aria-pressed="showPassword"
                @click="showPassword = !showPassword"
              >
                {{ showPassword ? 'Hide' : 'Show' }}
              </button>
            </div>
          </div>

          <p v-if="errorMessage" class="form-message form-error" role="alert">
            {{ errorMessage }}
          </p>

          <button class="submit-button" type="submit" :disabled="isSubmitting">
            <span v-if="isSubmitting" class="spinner" aria-hidden="true"></span>
            {{ isSubmitting ? 'Logging in…' : 'Log in' }}
          </button>
        </form>

        <p class="switch-auth">
          New to Briefly? <RouterLink to="/signup">Create an account</RouterLink>
        </p>
      </template>
    </div>
  </section>
</template>
