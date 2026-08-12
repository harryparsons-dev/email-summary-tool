<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

type ApiState = 'loading' | 'connected' | 'error'
type AuthView = 'login' | 'register' | 'forgot' | 'reset'

interface User {
  id: string
  email: string
  created_at: string
}

interface AuthResponse {
  user: User
  access_token: string
  token_type: string
  expires_in: number
}

const apiState = ref<ApiState>('loading')
const apiMessage = ref('Checking the API connection…')
const authDialog = ref<HTMLDialogElement>()
const authView = ref<AuthView>('login')
const authEmail = ref('')
const authPassword = ref('')
const authError = ref('')
const authMessage = ref('')
const authBusy = ref(false)
const resetToken = ref('')
const accessToken = ref('')
const currentUser = ref<User | null>(null)

const authTitle = computed(() => ({
  login: 'Welcome back',
  register: 'Create your account',
  forgot: 'Reset your password',
  reset: 'Choose a new password',
}[authView.value]))

const authDescription = computed(() => ({
  login: 'Sign in to open your calm, focused inbox.',
  register: 'Start turning busy threads into a useful daily brief.',
  forgot: 'Enter your email and we’ll send you a secure reset link.',
  reset: 'Use at least 12 characters for your new password.',
}[authView.value]))

const statusLabel = computed(() => {
  if (apiState.value === 'connected') return 'Systems online'
  if (apiState.value === 'error') return 'API offline'
  return 'Checking systems'
})

const statusColor = computed(() => {
  if (apiState.value === 'connected') return 'success'
  if (apiState.value === 'error') return 'error'
  return 'warning'
})

const messages = [
  {
    initials: 'AM',
    sender: 'Amelia Martin',
    subject: 'Project Aurora — next steps',
    summary: 'Design is approved. Three action items are ready for your team before Friday.',
    time: '2m',
    tag: 'Action needed',
    tone: 'violet',
  },
  {
    initials: 'JL',
    sender: 'Jonas Lee',
    subject: 'Q3 planning notes',
    summary: 'The team aligned on two priorities and moved the launch review to 16 August.',
    time: '18m',
    tag: 'Planning',
    tone: 'amber',
  },
  {
    initials: 'NK',
    sender: 'Nora Kim',
    subject: 'A quick update from support',
    summary: 'Response time improved by 21%. No urgent escalations need your attention today.',
    time: '1h',
    tag: 'Good news',
    tone: 'green',
  },
]

const features = [
  {
    number: '01',
    title: 'The signal, not the noise',
    description: 'Long threads become crisp summaries with decisions, deadlines and owners pulled forward.',
  },
  {
    number: '02',
    title: 'Priorities at a glance',
    description: 'See what needs a reply, what can wait and what is simply useful context for later.',
  },
  {
    number: '03',
    title: 'A calmer daily rhythm',
    description: 'Start with one thoughtful brief instead of opening the day inside a crowded inbox.',
  },
]

async function checkApi() {
  apiState.value = 'loading'
  apiMessage.value = 'Checking the API connection…'

  try {
    const response = await fetch('/api/')

    if (!response.ok) {
      throw new Error(`API returned ${response.status}`)
    }

    apiMessage.value = await response.text()
    apiState.value = 'connected'
  } catch (error) {
    apiState.value = 'error'
    apiMessage.value = error instanceof Error ? error.message : 'Unable to reach the API'
  }
}

function scrollToPreview() {
  document.querySelector('#inbox-preview')?.scrollIntoView({ behavior: 'smooth', block: 'center' })
}

function scrollToFeatures() {
  document.querySelector('#how-it-works')?.scrollIntoView({ behavior: 'smooth' })
}

function showAuth(view: AuthView) {
  authView.value = view
  authError.value = ''
  authMessage.value = ''
  authPassword.value = ''
  authDialog.value?.showModal()
}

function switchAuth(view: AuthView) {
  authView.value = view
  authError.value = ''
  authMessage.value = ''
  authPassword.value = ''
}

async function authRequest(path: string, body?: object): Promise<Response> {
  return fetch(`/api/auth${path}`, {
    method: body ? 'POST' : 'GET',
    headers: body ? { 'Content-Type': 'application/json' } : undefined,
    body: body ? JSON.stringify(body) : undefined,
  })
}

async function responseError(response: Response): Promise<string> {
  try {
    const body = await response.json()
    return body.error?.message ?? 'Please try again.'
  } catch {
    return 'Please try again.'
  }
}

function acceptSession(session: AuthResponse) {
  currentUser.value = session.user
  accessToken.value = session.access_token
}

async function submitAuth() {
  authBusy.value = true
  authError.value = ''
  authMessage.value = ''
  try {
    if (authView.value === 'forgot') {
      const response = await authRequest('/password/forgot', { email: authEmail.value })
      if (!response.ok) throw new Error(await responseError(response))
      const body = await response.json()
      authMessage.value = body.message
      return
    }

    if (authView.value === 'reset') {
      const response = await authRequest('/password/reset', {
        token: resetToken.value,
        password: authPassword.value,
      })
      if (!response.ok) throw new Error(await responseError(response))
      window.history.replaceState({}, '', window.location.pathname)
      resetToken.value = ''
      switchAuth('login')
      authMessage.value = 'Password updated. Sign in with your new password.'
      return
    }

    const endpoint = authView.value === 'register' ? '/register' : '/login'
    const response = await authRequest(endpoint, { email: authEmail.value, password: authPassword.value })
    if (!response.ok) throw new Error(await responseError(response))
    acceptSession(await response.json())
    authDialog.value?.close()
  } catch (error) {
    authError.value = error instanceof Error ? error.message : 'Please try again.'
  } finally {
    authBusy.value = false
  }
}

async function restoreSession() {
  try {
    const response = await authRequest('/refresh', {})
    if (response.ok) acceptSession(await response.json())
  } catch {
    // No active session is a normal state for the public landing page.
  }
}

async function logout() {
  try {
    await authRequest('/logout', {})
  } finally {
    currentUser.value = null
    accessToken.value = ''
  }
}

function primaryAction() {
  if (currentUser.value) scrollToPreview()
  else showAuth('register')
}

onMounted(() => {
  checkApi()
  restoreSession()
  const token = new URLSearchParams(window.location.search).get('token')
  if (token) {
    resetToken.value = token
    showAuth('reset')
  }
})
</script>

<template>
  <UApp>
    <div class="site-shell">
      <div class="ambient ambient-one" aria-hidden="true"></div>
      <div class="ambient ambient-two" aria-hidden="true"></div>

      <UHeader :toggle="false" class="site-header">
        <template #left>
          <a class="brand" href="#" aria-label="Briefly home">
            <span class="brand-mark" aria-hidden="true">
              <span></span><span></span><span></span><span></span>
            </span>
            <span>Briefly</span>
          </a>
        </template>

        <nav class="desktop-nav" aria-label="Main navigation">
          <button type="button" @click="scrollToFeatures">How it works</button>
          <button type="button" @click="scrollToPreview">Preview</button>
          <a href="mailto:hello@briefly.app">Contact</a>
        </nav>

        <template #right>
          <UBadge :color="statusColor" variant="subtle" size="sm" class="api-badge">
            <span class="status-pip" :class="`status-pip--${apiState}`"></span>
            {{ statusLabel }}
          </UBadge>
          <UButton v-if="currentUser" color="neutral" variant="ghost" size="md" @click="logout">
            Sign out
          </UButton>
          <UButton v-else color="neutral" variant="outline" size="md" @click="showAuth('login')">
            Sign in
          </UButton>
        </template>
      </UHeader>

      <main>
        <UContainer class="hero-container">
          <section class="hero-grid" aria-labelledby="page-title">
            <div class="hero-copy">
              <UBadge color="neutral" variant="outline" size="md" class="hero-badge">
                <span class="spark" aria-hidden="true">✦</span>
                A clearer way through your inbox
              </UBadge>

              <h1 id="page-title">
                Make space for
                <em>what matters.</em>
              </h1>

              <p class="hero-intro">
                Briefly reads the long threads, finds the decisions and turns a busy inbox into a calm,
                useful daily brief.
              </p>

              <div class="hero-actions">
                <UButton color="primary" size="xl" class="primary-cta" @click="primaryAction">
                  {{ currentUser ? 'Open my daily brief' : 'Create my account' }}
                  <span aria-hidden="true">→</span>
                </UButton>
                <UButton
                  color="neutral"
                  variant="ghost"
                  size="xl"
                  class="secondary-cta"
                  @click="scrollToFeatures"
                >
                  See how it works
                </UButton>
              </div>

              <div class="social-proof">
                <UAvatarGroup size="md">
                  <UAvatar text="AM" color="neutral" class="proof-avatar proof-avatar--one" />
                  <UAvatar text="JL" color="neutral" class="proof-avatar proof-avatar--two" />
                  <UAvatar text="NK" color="neutral" class="proof-avatar proof-avatar--three" />
                </UAvatarGroup>
                <div>
                  <div class="rating" aria-label="Five out of five stars">★★★★★</div>
                  <span>Less inbox. More headspace.</span>
                </div>
              </div>
            </div>

            <div id="inbox-preview" class="preview-wrap">
              <div class="preview-orbit preview-orbit--top" aria-hidden="true">✦</div>
              <div class="preview-orbit preview-orbit--bottom" aria-hidden="true">✦</div>

              <UCard class="inbox-card" variant="outline">
                <template #header>
                  <div class="inbox-header">
                    <div>
                      <span class="overline">Good morning, Alex</span>
                      <h2>Your daily brief</h2>
                    </div>
                    <UAvatar text="AP" size="lg" color="primary" class="profile-avatar" />
                  </div>

                  <div class="brief-stats">
                    <div><strong>12</strong><span>emails read</span></div>
                    <USeparator orientation="vertical" />
                    <div><strong>3</strong><span>need attention</span></div>
                    <USeparator orientation="vertical" />
                    <div><strong>24m</strong><span>time saved</span></div>
                  </div>
                </template>

                <div class="message-list">
                  <article v-for="message in messages" :key="message.subject" class="message-row">
                    <UAvatar
                      :text="message.initials"
                      color="neutral"
                      size="lg"
                      :class="`message-avatar message-avatar--${message.tone}`"
                    />
                    <div class="message-content">
                      <div class="message-meta">
                        <strong>{{ message.sender }}</strong>
                        <span>{{ message.time }}</span>
                      </div>
                      <h3>{{ message.subject }}</h3>
                      <p>{{ message.summary }}</p>
                      <UBadge color="neutral" variant="subtle" size="sm">{{ message.tag }}</UBadge>
                    </div>
                  </article>
                </div>

                <template #footer>
                  <UButton block color="neutral" variant="soft" size="lg" @click="scrollToFeatures">
                    View full daily brief
                    <span aria-hidden="true">→</span>
                  </UButton>
                </template>
              </UCard>
            </div>
          </section>

          <section id="how-it-works" class="feature-section" aria-labelledby="feature-title">
            <div class="section-heading">
              <span class="overline">Thoughtful by design</span>
              <h2 id="feature-title">Everything important.<br />Nothing overwhelming.</h2>
            </div>

            <UPageGrid class="feature-grid">
              <UPageCard
                v-for="feature in features"
                :key="feature.number"
                variant="subtle"
                class="feature-card"
              >
                <template #header>
                  <span class="feature-number">{{ feature.number }}</span>
                </template>
                <template #title>{{ feature.title }}</template>
                <template #description>{{ feature.description }}</template>
              </UPageCard>
            </UPageGrid>

            <UAlert
              v-if="apiState === 'error'"
              color="error"
              variant="subtle"
              title="The demo API is not responding"
              :description="apiMessage"
              class="api-alert"
            >
              <template #actions>
                <UButton color="error" variant="soft" size="sm" @click="checkApi">Try again</UButton>
              </template>
            </UAlert>
          </section>
        </UContainer>
      </main>

      <footer class="site-footer">
        <UContainer class="footer-inner">
          <a class="brand" href="#" aria-label="Briefly home">
            <span class="brand-mark" aria-hidden="true">
              <span></span><span></span><span></span><span></span>
            </span>
            <span>Briefly</span>
          </a>
          <p>Your inbox, distilled.</p>
          <span>© 2026 Briefly</span>
        </UContainer>
      </footer>

      <dialog ref="authDialog" class="auth-dialog" @click.self="authDialog?.close()">
        <button class="auth-close" type="button" aria-label="Close" @click="authDialog?.close()">×</button>
        <div class="auth-brand">
          <span class="brand-mark" aria-hidden="true">
            <span></span><span></span><span></span><span></span>
          </span>
        </div>
        <span class="overline">Your Briefly account</span>
        <h2>{{ authTitle }}</h2>
        <p class="auth-description">{{ authDescription }}</p>

        <form class="auth-form" @submit.prevent="submitAuth">
          <label v-if="authView !== 'reset'">
            <span>Email address</span>
            <input v-model="authEmail" type="email" autocomplete="email" required />
          </label>
          <label v-if="authView !== 'forgot'">
            <span>{{ authView === 'reset' ? 'New password' : 'Password' }}</span>
            <input
              v-model="authPassword"
              type="password"
              :autocomplete="authView === 'login' ? 'current-password' : 'new-password'"
              :minlength="authView === 'login' ? undefined : 12"
              required
            />
            <small v-if="authView !== 'login'">At least 12 characters</small>
          </label>

          <div v-if="authError" class="auth-notice auth-notice--error" role="alert">{{ authError }}</div>
          <div v-if="authMessage" class="auth-notice auth-notice--success" role="status">{{ authMessage }}</div>

          <button class="auth-submit" type="submit" :disabled="authBusy">
            {{ authBusy ? 'Please wait…' : authView === 'login' ? 'Sign in' : authView === 'register' ? 'Create account' : authView === 'forgot' ? 'Send reset link' : 'Update password' }}
          </button>
        </form>

        <div class="auth-links">
          <template v-if="authView === 'login'">
            <button type="button" @click="switchAuth('forgot')">Forgot password?</button>
            <button type="button" @click="switchAuth('register')">Create an account</button>
          </template>
          <button v-else-if="authView !== 'reset'" type="button" @click="switchAuth('login')">Back to sign in</button>
          <button v-else type="button" @click="switchAuth('login')">Cancel and sign in</button>
        </div>
      </dialog>
    </div>
  </UApp>
</template>
