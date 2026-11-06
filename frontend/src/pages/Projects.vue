<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import CopyButton from '../components/CopyButton.vue'
import { formatRelative } from '../lib/format'
import type { Project } from '../models/project'
import { createProject, listProjects } from '../services/projectService'

const router = useRouter()

const projects = ref<Project[]>([])
const isLoading = ref(true)
const errorMessage = ref('')

const createOpen = ref(false)
const title = ref('')
const description = ref('')
const isCreating = ref(false)
const createError = ref('')

async function loadProjects() {
  isLoading.value = true
  errorMessage.value = ''

  try {
    projects.value = await listProjects()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Unable to load your projects.'
  } finally {
    isLoading.value = false
  }
}

function openCreate() {
  title.value = ''
  description.value = ''
  createError.value = ''
  createOpen.value = true
}

async function submitCreate() {
  createError.value = ''
  isCreating.value = true

  try {
    const project = await createProject({
      title: title.value.trim(),
      description: description.value.trim(),
    })
    createOpen.value = false
    await router.push(`/projects/${project.id}`)
  } catch (error) {
    createError.value = error instanceof Error ? error.message : 'Unable to create the project.'
  } finally {
    isCreating.value = false
  }
}

onMounted(loadProjects)
</script>

<template>
  <section class="mx-auto w-full max-w-5xl self-start lg:pt-[6vh]" aria-labelledby="projects-title">
    <div class="mb-8 flex flex-wrap items-end justify-between gap-4">
      <div>
        <div class="mb-3 flex items-center gap-2 text-xs font-medium text-stone-500">
          <span>Workspace</span>
          <UIcon name="i-lucide-chevron-right" class="size-3.5" aria-hidden="true" />
          <span>Projects</span>
        </div>
        <h1 id="projects-title" class="text-[2rem] leading-tight font-bold tracking-[-0.035em] text-stone-950 dark:text-white">
          Projects
        </h1>
        <p class="mt-2 text-[0.95rem] leading-6 text-stone-600 dark:text-stone-400">
          Each project has its own email address. Anything sent to it lands here.
        </p>
      </div>
      <UButton v-if="projects.length > 0" color="primary" size="lg" icon="i-lucide-plus" @click="openCreate">
        New project
      </UButton>
    </div>

    <div v-if="isLoading" class="grid gap-3 sm:grid-cols-2" role="status" aria-live="polite">
      <USkeleton v-for="n in 4" :key="n" class="h-36 rounded-xl" />
      <span class="sr-only">Loading your projects…</span>
    </div>

    <UAlert
      v-else-if="errorMessage"
      color="error"
      variant="soft"
      icon="i-lucide-cloud-off"
      title="We couldn’t load your projects"
      :description="errorMessage"
      role="alert"
    >
      <template #actions>
        <UButton color="error" variant="soft" size="sm" icon="i-lucide-refresh-cw" @click="loadProjects">
          Try again
        </UButton>
      </template>
    </UAlert>

    <UCard
      v-else-if="projects.length === 0"
      variant="outline"
      class="rounded-xl bg-white ring-stone-200 dark:bg-[#161617] dark:ring-white/10"
    >
      <UEmpty
        icon="i-lucide-inbox"
        title="No projects yet"
        description="Create a project to get an email address you can forward mail to."
        :actions="[{ label: 'Create your first project', icon: 'i-lucide-plus', color: 'primary', onClick: openCreate }]"
        variant="naked"
        class="py-10"
      />
    </UCard>

    <ul v-else class="grid gap-3 sm:grid-cols-2">
      <li v-for="project in projects" :key="project.id">
        <RouterLink
          :to="`/projects/${project.id}`"
          class="group flex h-full flex-col rounded-xl bg-white p-5 no-underline ring-1 ring-stone-200 transition hover:ring-orange-300 focus-visible:ring-2 focus-visible:ring-orange-500 focus-visible:outline-none dark:bg-[#161617] dark:ring-white/10 dark:hover:ring-orange-400/40"
        >
          <div class="flex items-start justify-between gap-3">
            <h2 class="min-w-0 truncate text-base font-semibold text-stone-950 group-hover:text-orange-800 dark:text-white dark:group-hover:text-orange-300">
              {{ project.title }}
            </h2>
            <UIcon
              name="i-lucide-arrow-up-right"
              class="size-4 shrink-0 text-stone-400 transition group-hover:text-orange-600 dark:group-hover:text-orange-400"
              aria-hidden="true"
            />
          </div>
          <p v-if="project.description" class="mt-1.5 line-clamp-2 text-sm text-stone-600 dark:text-stone-400">
            {{ project.description }}
          </p>

          <div class="mt-auto pt-5">
            <div class="flex items-center gap-1 rounded-lg bg-stone-100 py-1 pr-1 pl-3 dark:bg-white/5">
              <UIcon name="i-lucide-at-sign" class="size-3.5 shrink-0 text-stone-500" aria-hidden="true" />
              <span class="min-w-0 flex-1 truncate font-mono text-xs text-stone-700 dark:text-stone-300">
                {{ project.email_address }}
              </span>
              <CopyButton :value="project.email_address" label="Copy address" />
            </div>
            <p class="mt-3 text-xs text-stone-500">Created {{ formatRelative(project.created_at) }}</p>
          </div>
        </RouterLink>
      </li>
    </ul>

    <UModal
      v-model:open="createOpen"
      title="New project"
      description="We’ll generate an email address for it straight away."
    >
      <template #body>
        <form id="create-project-form" class="grid gap-5" @submit.prevent="submitCreate">
          <UFormField label="Title" name="title" required>
            <UInput
              v-model="title"
              name="title"
              placeholder="e.g. Weekly supplier updates"
              maxlength="255"
              size="lg"
              class="w-full"
              autofocus
              required
            />
          </UFormField>
          <UFormField label="Description" name="description" hint="Optional">
            <UTextarea
              v-model="description"
              name="description"
              placeholder="What kind of mail will this project receive?"
              :rows="3"
              autoresize
              class="w-full"
            />
          </UFormField>
          <UAlert
            v-if="createError"
            color="error"
            variant="soft"
            icon="i-lucide-circle-alert"
            :description="createError"
            role="alert"
          />
        </form>
      </template>
      <template #footer>
        <div class="flex w-full justify-end gap-2">
          <UButton color="neutral" variant="ghost" :disabled="isCreating" @click="createOpen = false">Cancel</UButton>
          <UButton
            type="submit"
            form="create-project-form"
            color="primary"
            :loading="isCreating"
            :disabled="isCreating || !title.trim()"
          >
            Create project
          </UButton>
        </div>
      </template>
    </UModal>
  </section>
</template>
