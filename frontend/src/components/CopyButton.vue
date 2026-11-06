<script setup lang="ts">
import { useToast } from '@nuxt/ui/composables'
import { ref } from 'vue'

const props = withDefaults(
  defineProps<{
    value: string
    label?: string
  }>(),
  { label: 'Copy' },
)

const toast = useToast()
const copied = ref(false)
let resetTimer: ReturnType<typeof setTimeout> | undefined

async function copy() {
  try {
    await navigator.clipboard.writeText(props.value)
    copied.value = true
    clearTimeout(resetTimer)
    resetTimer = setTimeout(() => (copied.value = false), 1800)
  } catch {
    toast.add({ title: 'Couldn’t copy', description: 'Select the address and copy it manually.', color: 'error' })
  }
}
</script>

<template>
  <UTooltip :text="copied ? 'Copied' : label">
    <UButton
      type="button"
      color="neutral"
      variant="ghost"
      size="xs"
      :icon="copied ? 'i-lucide-check' : 'i-lucide-copy'"
      :aria-label="copied ? 'Copied' : label"
      @click.stop.prevent="copy"
    />
  </UTooltip>
</template>
