<script setup lang="ts">
import type { FormError, FormSubmitEvent } from '@nuxt/ui'
import { reactive } from 'vue'
import { statusOptions } from '../lib/projectStatus'
import type { CreateProjectInput } from '../models/project'

const props = defineProps<{
  id: string
  initial?: CreateProjectInput
  disabled?: boolean
  errorMessage?: string
  errorTitle: string
}>()

const emit = defineEmits<{ submit: [value: CreateProjectInput] }>()

const state = reactive<CreateProjectInput>({
  name: props.initial?.name ?? '',
  description: props.initial?.description ?? '',
  status: props.initial?.status ?? 'pending',
})

function validate(value: Partial<CreateProjectInput>): FormError[] {
  const errors: FormError[] = []
  if (!value.name?.trim()) errors.push({ name: 'name', message: 'Give the project a name.' })
  if (!value.description?.trim()) errors.push({ name: 'description', message: 'Add a short description.' })
  return errors
}

function submit(event: FormSubmitEvent<CreateProjectInput>) {
  emit('submit', {
    name: event.data.name.trim(),
    description: event.data.description.trim(),
    status: event.data.status,
  })
}
</script>

<template>
  <UForm :id="id" :state="state" :validate="validate" :disabled="disabled" class="grid gap-5" @submit="submit">
    <div class="grid gap-5 sm:grid-cols-[minmax(0,1fr)_11rem]">
      <UFormField label="Project name" name="name" required>
        <UInput
          v-model="state.name"
          placeholder="e.g. Weekly customer digest"
          icon="i-lucide-folder-kanban"
          maxlength="255"
          size="lg"
          class="w-full"
          autofocus
        />
      </UFormField>

      <UFormField label="Status" name="status" required>
        <USelect v-model="state.status" :items="statusOptions" size="lg" class="w-full" />
      </UFormField>
    </div>

    <UFormField label="Description" name="description" required>
      <UTextarea
        v-model="state.description"
        placeholder="Describe what this project will summarize."
        :rows="4"
        autoresize
        size="lg"
        class="w-full"
      />
    </UFormField>

    <Transition name="swap">
      <UAlert
        v-if="errorMessage"
        color="error"
        variant="subtle"
        icon="i-lucide-circle-alert"
        :title="errorTitle"
        :description="errorMessage"
        role="alert"
      />
    </Transition>
  </UForm>
</template>
