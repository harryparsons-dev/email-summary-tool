<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { logout } from './services/authService'

const router = useRouter()
const isSigningOut = ref(false)
const signOutFailed = ref(false)

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
  <UApp>
    <div class="page-shell">
      <UHeader :toggle="false" class="site-header">
        <template #left>
          <a class="brand" href="/" aria-label="Briefly home">
            <span class="brand-mark" aria-hidden="true">B</span>
            <span>Briefly</span>
          </a>
        </template>

        <template #right>
          <nav class="auth-actions" aria-label="Account navigation">
            <UButton
              to="/profile"
              color="neutral"
              :variant="$route.path === '/profile' ? 'soft' : 'ghost'"
              size="md"
            >
              Profile
            </UButton>
            <UButton
              v-if="$route.path === '/profile'"
              color="neutral"
              variant="outline"
              size="md"
              :loading="isSigningOut"
              :disabled="isSigningOut"
              @click="signOut"
            >
              {{ signOutFailed ? 'Try sign out again' : 'Sign out' }}
            </UButton>
            <UButton
              v-if="$route.path !== '/profile'"
              to="/login"
              color="neutral"
              :variant="$route.path === '/login' ? 'soft' : 'ghost'"
              size="md"
            >
              Log in
            </UButton>
            <UButton
              v-if="$route.path !== '/profile'"
              to="/signup"
              color="primary"
              variant="solid"
              size="md"
              class="signup-button"
            >
              Sign up
            </UButton>
          </nav>
        </template>
      </UHeader>

      <main class="auth-main">
        <RouterView />
      </main>
    </div>
  </UApp>
</template>
