<script setup lang="ts">
import { computed, ref } from 'vue'
import { formatBytes, formatDateTime, formatRelative } from '../lib/format'
import type { Email } from '../models/email'
import SummaryPanel from './SummaryPanel.vue'

const props = defineProps<{
  email: Email
}>()

const hasHtml = computed(() => props.email.html_body.trim() !== '')
const view = ref<'html' | 'text'>(hasHtml.value ? 'html' : 'text')

// The HTML body comes from whoever sent the email, so it is shown in a
// sandboxed iframe (no scripts, no access to this page) with a policy that
// blocks remote content such as tracking pixels.
const htmlDocument = computed(
  () => `<!doctype html><html><head><meta charset="utf-8">
<meta http-equiv="Content-Security-Policy" content="default-src 'none'; img-src data:; style-src 'unsafe-inline'; font-src data:">
<base target="_blank">
<style>body{margin:0;padding:20px;font:14px/1.55 system-ui,sans-serif;color:#1c1917;background:#fff;overflow-wrap:anywhere}img{max-width:100%;height:auto}img[src^="http" i]{display:none}</style>
</head><body>${props.email.html_body}</body></html>`,
)

const hasRemoteImages = computed(() => /<img[^>]+src\s*=\s*["']?https?:/i.test(props.email.html_body))

const viewItems = [
  { label: 'Formatted', value: 'html', icon: 'i-lucide-layout-template' },
  { label: 'Plain text', value: 'text', icon: 'i-lucide-text' },
]

const sender = computed(() => {
  const { from_name: name, from_address: address, envelope_from: envelope } = props.email
  if (name && address) return `${name} <${address}>`
  return address || envelope || 'Unknown sender'
})

const senderInitial = computed(() => (props.email.from_name || props.email.from_address || '?').charAt(0).toUpperCase())
</script>

<template>
  <article class="grid gap-5">
    <header class="grid gap-4">
      <h2 class="text-xl leading-snug font-semibold text-highlighted [overflow-wrap:anywhere]">
        {{ email.subject || '(no subject)' }}
      </h2>

      <div class="flex flex-wrap items-start justify-between gap-3">
        <UUser
          :name="sender"
          :description="`To ${email.to_addresses || email.envelope_to}`"
          :avatar="{ text: senderInitial, ui: { root: 'bg-primary/10', fallback: 'text-primary font-semibold' } }"
          size="md"
          class="min-w-0"
          :ui="{ name: '[overflow-wrap:anywhere]', description: '[overflow-wrap:anywhere]' }"
        />
        <UTooltip :text="formatDateTime(email.sent_at ?? email.received_at)">
          <span class="shrink-0 text-xs text-muted">{{ formatRelative(email.sent_at ?? email.received_at) }}</span>
        </UTooltip>
      </div>

      <div v-if="email.attachments.length" class="flex flex-wrap gap-2">
        <UBadge
          v-for="(attachment, index) in email.attachments"
          :key="index"
          color="neutral"
          variant="subtle"
          size="lg"
          icon="i-lucide-paperclip"
          class="max-w-full"
          :ui="{ label: 'truncate' }"
        >
          {{ attachment.filename || attachment.content_type || 'Attachment' }}
          <span class="text-dimmed">· {{ formatBytes(attachment.size) }}</span>
        </UBadge>
      </div>
    </header>

    <SummaryPanel
      :project-id="email.project_id"
      :email-id="email.id"
      title="AI summary"
      empty-text="Get a short AI summary of this email."
    />

    <div v-if="hasHtml" class="flex flex-wrap items-center justify-between gap-2">
      <UTabs
        v-model="view"
        :items="viewItems"
        :content="false"
        size="xs"
        color="neutral"
        aria-label="Email format"
      />
      <Transition name="swap">
        <span v-if="hasRemoteImages && view === 'html'" class="flex items-center gap-1.5 text-xs text-muted">
          <UIcon name="i-lucide-image-off" class="size-3.5" aria-hidden="true" />
          Remote images blocked
        </span>
      </Transition>
    </div>

    <USeparator v-else />

    <Transition name="swap" mode="out-in">
      <iframe
        v-if="view === 'html'"
        :srcdoc="htmlDocument"
        sandbox="allow-popups allow-popups-to-escape-sandbox"
        referrerpolicy="no-referrer"
        title="Email content"
        class="h-[32rem] w-full rounded-lg bg-white ring ring-default"
      />
      <pre
        v-else
        class="font-sans text-sm leading-6 whitespace-pre-wrap text-default [overflow-wrap:anywhere]"
      >{{ email.text_body || 'This email has no plain-text content.' }}</pre>
    </Transition>
  </article>
</template>
