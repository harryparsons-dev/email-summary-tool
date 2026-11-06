<script setup lang="ts">
import { computed, ref } from 'vue'
import { formatBytes, formatDateTime } from '../lib/format'
import type { Email } from '../models/email'

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
<style>body{margin:0;font:14px/1.55 system-ui,sans-serif;color:#1c1917;background:#fff;overflow-wrap:anywhere}img{max-width:100%;height:auto}img[src^="http" i]{display:none}</style>
</head><body>${props.email.html_body}</body></html>`,
)

const hasRemoteImages = computed(() => /<img[^>]+src\s*=\s*["']?https?:/i.test(props.email.html_body))

const sender = computed(() => {
  const { from_name: name, from_address: address, envelope_from: envelope } = props.email
  if (name && address) return `${name} <${address}>`
  return address || envelope || 'Unknown sender'
})
</script>

<template>
  <article class="grid gap-5">
    <header class="grid gap-3">
      <h2 class="text-xl leading-snug font-bold tracking-[-0.02em] text-stone-950 [overflow-wrap:anywhere] dark:text-white">
        {{ email.subject || '(no subject)' }}
      </h2>
      <dl class="grid grid-cols-[3.5rem_minmax(0,1fr)] gap-x-3 gap-y-1 text-sm">
        <dt class="text-stone-500">From</dt>
        <dd class="[overflow-wrap:anywhere] text-stone-800 dark:text-stone-200">{{ sender }}</dd>
        <dt class="text-stone-500">To</dt>
        <dd class="[overflow-wrap:anywhere] text-stone-800 dark:text-stone-200">{{ email.to_addresses || email.envelope_to }}</dd>
        <dt class="text-stone-500">Date</dt>
        <dd class="text-stone-800 dark:text-stone-200">{{ formatDateTime(email.sent_at ?? email.received_at) }}</dd>
      </dl>
    </header>

    <div v-if="email.attachments.length" class="flex flex-wrap gap-2">
      <span
        v-for="(attachment, index) in email.attachments"
        :key="index"
        class="inline-flex max-w-full items-center gap-2 rounded-lg bg-stone-100 px-2.5 py-1.5 text-xs text-stone-700 dark:bg-white/5 dark:text-stone-300"
      >
        <UIcon name="i-lucide-paperclip" class="size-3.5 shrink-0" aria-hidden="true" />
        <span class="truncate font-medium">{{ attachment.filename || attachment.content_type || 'Attachment' }}</span>
        <span class="shrink-0 text-stone-500">{{ formatBytes(attachment.size) }}</span>
      </span>
    </div>

    <USeparator />

    <div v-if="hasHtml" class="flex flex-wrap items-center gap-1" role="group" aria-label="Email format">
      <UButton
        size="xs"
        color="neutral"
        :variant="view === 'html' ? 'soft' : 'ghost'"
        :aria-pressed="view === 'html'"
        @click="view = 'html'"
      >
        Formatted
      </UButton>
      <UButton
        size="xs"
        color="neutral"
        :variant="view === 'text' ? 'soft' : 'ghost'"
        :aria-pressed="view === 'text'"
        @click="view = 'text'"
      >
        Plain text
      </UButton>
      <span v-if="hasRemoteImages && view === 'html'" class="ml-auto flex items-center gap-1.5 text-xs text-stone-500">
        <UIcon name="i-lucide-image-off" class="size-3.5" aria-hidden="true" />
        Remote images blocked
      </span>
    </div>

    <iframe
      v-if="view === 'html'"
      :srcdoc="htmlDocument"
      sandbox="allow-popups allow-popups-to-escape-sandbox"
      referrerpolicy="no-referrer"
      title="Email content"
      class="h-[32rem] w-full rounded-lg bg-white ring-1 ring-stone-200 dark:ring-white/10"
    />
    <pre
      v-else
      class="font-sans text-sm leading-6 whitespace-pre-wrap text-stone-800 [overflow-wrap:anywhere] dark:text-stone-200"
    >{{ email.text_body || 'This email has no plain-text content.' }}</pre>
  </article>
</template>
