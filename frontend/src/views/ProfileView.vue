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
  <section class="profile-layout" aria-labelledby="profile-title">
    <div class="profile-heading">
      <p class="eyebrow">Your account</p>
      <h1 id="profile-title">Profile</h1>
      <p>Review the account details connected to your Briefly workspace.</p>
    </div>

    <div class="profile-card">
      <div v-if="isLoading" class="profile-status" role="status">
        <span class="profile-spinner" aria-hidden="true"></span>
        <p>Loading your profile…</p>
      </div>

      <div v-else-if="errorMessage" class="profile-status profile-error" role="alert">
        <span class="profile-status-icon" aria-hidden="true">!</span>
        <h2>We couldn’t load your profile</h2>
        <p>{{ errorMessage }}</p>
        <button class="retry-button" type="button" @click="loadProfile">Try again</button>
      </div>

      <template v-else-if="user">
        <div class="profile-card-heading">
          <span class="profile-avatar" aria-hidden="true">{{ user.email.charAt(0).toUpperCase() }}</span>
          <div>
            <h2>Account details</h2>
            <p>Your signed-in Briefly identity.</p>
          </div>
        </div>

        <dl class="profile-details">
          <div class="profile-detail-row">
            <dt>Email address</dt>
            <dd>{{ user.email }}</dd>
          </div>
          <div class="profile-detail-row">
            <dt>User ID</dt>
            <dd class="profile-id">{{ user.id }}</dd>
          </div>
        </dl>
      </template>
    </div>
  </section>
</template>
