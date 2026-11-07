<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { formatDateTime, formatRelative } from '../lib/format'
import type { Summary } from '../models/summary'
import { createEmailSummary, getEmailSummary, listEmailSummaries } from '../services/projectSummaryService'

const props = withDefaults(
  defineProps<{
    projectId: string
    emailId: string
    title?: string
    emptyText?: string
  }>(),
  {
    title: 'AI summary',
    emptyText: 'Get a short AI summary of this email.',
  },
)

// Summaries are generated in the background, so a pending one is checked on
// this interval until it finishes. The server gives up after 3 minutes.
const pollIntervalMs = 2_000
const pollTimeoutMs = 4 * 60_000
const maxPollFailures = 3

const summary = ref<Summary | null>(null)
const isLoading = ref(true)
const isStarting = ref(false)
const errorMessage = ref('')
// True while a pending summary is being checked. Polling can stop before the
// summary finishes (timeout or repeated errors); the button then works again.
const isPolling = ref(false)

let pollTimer: ReturnType<typeof setTimeout> | undefined
let pollStartedAt = 0
let pollFailures = 0
// Bumped whenever the target changes so late responses for the old one are ignored.
let generation = 0

const isBusy = computed(() => isStarting.value || isPolling.value)
const buttonLabel = computed(() => {
  if (isBusy.value) return 'Summarising…'
  if (summary.value?.status === 'pending') return 'Check again'
  return summary.value ? 'Regenerate' : 'Summarise'
})

function stopPolling() {
  clearTimeout(pollTimer)
  pollTimer = undefined
  isPolling.value = false
}

function schedulePoll() {
  clearTimeout(pollTimer)
  isPolling.value = true
  pollTimer = setTimeout(poll, pollIntervalMs)
}

async function poll() {
  const current = summary.value
  if (!current || current.status !== 'pending') {
    stopPolling()
    return
  }

  if (Date.now() - pollStartedAt > pollTimeoutMs) {
    stopPolling()
    errorMessage.value = 'This is taking longer than expected. Try again in a moment.'
    return
  }
  // Don't poll in a background tab; check again when it's next due.
  if (document.visibilityState !== 'visible') {
    schedulePoll()
    return
  }

  const requestGeneration = generation
  try {
    const updated = await getEmailSummary(props.projectId, props.emailId, current.id)
    if (requestGeneration !== generation) return
    summary.value = updated
    pollFailures = 0
  } catch (error) {
    if (requestGeneration !== generation) return
    pollFailures++
    if (pollFailures >= maxPollFailures) {
      stopPolling()
      errorMessage.value = error instanceof Error ? error.message : 'Unable to check the summary.'
      return
    }
  }

  if (summary.value?.status === 'pending') schedulePoll()
  else stopPolling()
}

function startPolling() {
  pollStartedAt = Date.now()
  pollFailures = 0
  errorMessage.value = ''
  schedulePoll()
}

async function loadLatest() {
  stopPolling()
  const requestGeneration = ++generation
  summary.value = null
  errorMessage.value = ''
  isLoading.value = true

  try {
    const summaries = await listEmailSummaries(props.projectId, props.emailId)
    if (requestGeneration !== generation) return
    summary.value = summaries[0] ?? null
    if (summary.value?.status === 'pending') startPolling()
  } catch (error) {
    if (requestGeneration !== generation) return
    errorMessage.value = error instanceof Error ? error.message : 'Unable to load the summary.'
  } finally {
    if (requestGeneration === generation) isLoading.value = false
  }
}

async function generate() {
  if (isBusy.value) return
  stopPolling()
  const requestGeneration = generation
  errorMessage.value = ''
  isStarting.value = true

  try {
    const started = await createEmailSummary(props.projectId, props.emailId)
    if (requestGeneration !== generation) return
    summary.value = started
    if (started.status === 'pending') startPolling()
  } catch (error) {
    if (requestGeneration !== generation) return
    errorMessage.value = error instanceof Error ? error.message : 'Unable to start the summary.'
  } finally {
    if (requestGeneration === generation) isStarting.value = false
  }
}

watch(() => [props.projectId, props.emailId], loadLatest, { immediate: true })

onBeforeUnmount(() => {
  generation++
  stopPolling()
})
</script>

<template>
  <section class="grid gap-2 rounded-lg bg-primary/5 p-3 ring ring-primary/20" :aria-busy="isBusy">
    <div class="flex items-center justify-between gap-2">
      <div class="flex min-w-0 items-center gap-1.5">
        <UIcon name="i-lucide-sparkles" class="size-4 shrink-0 text-primary" aria-hidden="true" />
        <h3 class="truncate text-xs font-semibold tracking-wide text-primary uppercase">{{ title }}</h3>
        <UTooltip v-if="summary?.status === 'completed'" :text="formatDateTime(summary.updated_at)">
          <span class="shrink-0 text-xs text-dimmed">· {{ formatRelative(summary.updated_at) }}</span>
        </UTooltip>
      </div>
      <UButton
        size="xs"
        color="primary"
        :variant="summary ? 'ghost' : 'soft'"
        :icon="summary ? 'i-lucide-refresh-cw' : 'i-lucide-sparkles'"
        :label="buttonLabel"
        :loading="isBusy"
        :disabled="isLoading"
        @click="generate"
      />
    </div>

    <div aria-live="polite">
      <div v-if="isLoading" class="grid gap-1.5" role="status">
        <USkeleton class="h-3.5 w-full" />
        <USkeleton class="h-3.5 w-2/3" />
        <span class="sr-only">Loading summary…</span>
      </div>

      <div v-else-if="isBusy" class="grid gap-1.5" role="status">
        <USkeleton class="h-3.5 w-full" />
        <USkeleton class="h-3.5 w-5/6" />
        <USkeleton class="h-3.5 w-1/2" />
        <span class="sr-only">Summarising…</span>
      </div>

      <p v-else-if="summary?.status === 'failed'" class="flex items-start gap-1.5 text-sm text-error">
        <UIcon name="i-lucide-circle-alert" class="mt-0.5 size-4 shrink-0" aria-hidden="true" />
        {{ summary.error || 'The summary could not be generated. Please try again.' }}
      </p>

      <p
        v-else-if="summary?.status === 'completed'"
        class="text-sm leading-6 whitespace-pre-line text-default [overflow-wrap:anywhere]"
      >{{ summary.summary }}</p>

      <p v-else-if="!summary && !errorMessage" class="text-sm text-muted">{{ emptyText }}</p>

      <p v-if="errorMessage" class="flex items-start gap-1.5 text-sm text-error" role="alert">
        <UIcon name="i-lucide-circle-alert" class="mt-0.5 size-4 shrink-0" aria-hidden="true" />
        {{ errorMessage }}
      </p>
    </div>
  </section>
</template>
