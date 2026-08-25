<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import type { PaginatedProjectResponse, ProjectStatus } from '../models/project'
import { createProject, listProjects } from '../services/projectService'

const DEFAULT_PAGE_SIZE = 10

const result = ref<PaginatedProjectResponse | null>(null)
const isLoading = ref(true)
const errorMessage = ref('')
const requestedPage = ref(1)
const isCreateFormOpen = ref(false)
const isCreating = ref(false)
const createErrorMessage = ref('')
const successMessage = ref('')
const projectName = ref('')
const projectDescription = ref('')
const projectStatus = ref<ProjectStatus>('pending')

const totalPages = computed(() => {
  if (!result.value) {
    return 1
  }

  return Math.max(1, Math.ceil(result.value.total / result.value.page_size))
})

const rangeStart = computed(() => {
  if (!result.value || result.value.total === 0) {
    return 0
  }

  return (result.value.page - 1) * result.value.page_size + 1
})

const rangeEnd = computed(() => {
  if (!result.value) {
    return 0
  }

  return Math.min(result.value.page * result.value.page_size, result.value.total)
})

const statusDetails: Record<ProjectStatus, { label: string; color: 'neutral' | 'warning' | 'info' | 'success' }> = {
  pending: { label: 'Pending', color: 'warning' },
  in_progress: { label: 'In progress', color: 'info' },
  completed: { label: 'Completed', color: 'success' },
  archived: { label: 'Archived', color: 'neutral' },
}

async function loadProjects(page = requestedPage.value) {
  isLoading.value = true
  errorMessage.value = ''
  requestedPage.value = page

  try {
    const response = await listProjects({ page, pageSize: DEFAULT_PAGE_SIZE })
    result.value = response
    requestedPage.value = response.page
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Unable to load your projects.'
  } finally {
    isLoading.value = false
  }
}

function toggleCreateForm() {
  isCreateFormOpen.value = !isCreateFormOpen.value
  createErrorMessage.value = ''
  successMessage.value = ''
}

function resetCreateForm() {
  projectName.value = ''
  projectDescription.value = ''
  projectStatus.value = 'pending'
  createErrorMessage.value = ''
}

function cancelCreate() {
  resetCreateForm()
  isCreateFormOpen.value = false
}

async function submitProject() {
  createErrorMessage.value = ''
  successMessage.value = ''

  const name = projectName.value.trim()
  const description = projectDescription.value.trim()

  if (!name || !description) {
    createErrorMessage.value = 'Enter a name and description for the project.'
    return
  }

  isCreating.value = true

  try {
    const project = await createProject({
      name,
      description,
      status: projectStatus.value,
    })

    if (result.value) {
      result.value = {
        ...result.value,
        projects: result.value.page === 1
          ? [project, ...result.value.projects].slice(0, result.value.page_size)
          : result.value.projects,
        total: result.value.total + 1,
      }
    } else {
      result.value = {
        projects: [project],
        page: 1,
        page_size: DEFAULT_PAGE_SIZE,
        total: 1,
      }
      requestedPage.value = 1
    }

    resetCreateForm()
    isCreateFormOpen.value = false
    successMessage.value = `“${project.name}” was created.`
  } catch (error) {
    createErrorMessage.value = error instanceof Error ? error.message : 'Unable to create the project.'
  } finally {
    isCreating.value = false
  }
}

onMounted(() => loadProjects())
</script>

<template>
  <section class="mx-auto w-full max-w-5xl self-start lg:pt-[5vh]" aria-labelledby="projects-title">
    <div class="mb-8 flex flex-wrap items-end justify-between gap-4">
      <div>
        <div class="mb-3 flex items-center gap-2 text-xs font-medium text-stone-500 dark:text-stone-500">
          <span>Workspace</span>
          <UIcon name="i-lucide-chevron-right" class="size-3.5" aria-hidden="true" />
          <span>Projects</span>
        </div>
        <h1 id="projects-title" class="text-[2rem] leading-tight font-bold tracking-[-0.035em] text-stone-950 dark:text-white">
          Projects
        </h1>
        <p class="mt-2 text-[0.95rem] leading-6 text-stone-600 dark:text-stone-400">
          Review the email summary projects in your workspace.
        </p>
      </div>

      <div class="flex items-center gap-3">
        <UBadge v-if="result && !isLoading" color="neutral" variant="subtle" size="md">
          {{ result.total }} {{ result.total === 1 ? 'project' : 'projects' }}
        </UBadge>
        <UButton
          color="primary"
          :icon="isCreateFormOpen ? 'i-lucide-x' : 'i-lucide-plus'"
          :aria-expanded="isCreateFormOpen"
          aria-controls="create-project-form"
          @click="toggleCreateForm"
        >
          {{ isCreateFormOpen ? 'Close' : 'New project' }}
        </UButton>
      </div>
    </div>

    <Transition name="fade">
      <UCard
        v-if="isCreateFormOpen"
        id="create-project-form"
        variant="outline"
        class="mb-5 rounded-xl bg-white ring-stone-200 dark:bg-[#161617] dark:ring-white/10"
        :ui="{ body: 'p-5 sm:p-7' }"
      >
        <div class="mb-6">
          <h2 class="text-base font-semibold text-stone-950 dark:text-white">Create a project</h2>
          <p class="mt-1 text-sm leading-5 text-stone-500 dark:text-stone-400">
            Add a project to organize a new email summary workflow.
          </p>
        </div>

        <form class="grid gap-5" @submit.prevent="submitProject">
          <div class="grid gap-5 sm:grid-cols-[minmax(0,1fr)_12rem]">
            <UFormField label="Project name" name="project-name" required>
              <UInput
                id="project-name"
                v-model="projectName"
                name="project-name"
                placeholder="e.g. Weekly customer digest"
                icon="i-lucide-folder-kanban"
                maxlength="255"
                size="xl"
                class="w-full"
                :disabled="isCreating"
                required
              />
            </UFormField>

            <UFormField label="Status" name="project-status" required>
              <div class="relative">
                <select
                  id="project-status"
                  v-model="projectStatus"
                  name="project-status"
                  class="h-12 w-full appearance-none rounded-md border border-stone-300 bg-white px-3.5 pr-10 text-sm text-stone-900 shadow-sm outline-none transition focus:border-orange-500 focus:ring-2 focus:ring-orange-500/20 disabled:cursor-not-allowed disabled:opacity-60 dark:border-white/15 dark:bg-[#161617] dark:text-stone-100"
                  :disabled="isCreating"
                  required
                >
                  <option v-for="(details, status) in statusDetails" :key="status" :value="status">
                    {{ details.label }}
                  </option>
                </select>
                <UIcon
                  name="i-lucide-chevron-down"
                  class="pointer-events-none absolute top-1/2 right-3 size-4 -translate-y-1/2 text-stone-500"
                  aria-hidden="true"
                />
              </div>
            </UFormField>
          </div>

          <UFormField label="Description" name="project-description" required>
            <textarea
              id="project-description"
              v-model="projectDescription"
              name="project-description"
              rows="4"
              placeholder="Describe what this project will summarize."
              class="block w-full resize-y rounded-md border border-stone-300 bg-white px-3.5 py-3 text-sm leading-5 text-stone-900 shadow-sm outline-none transition placeholder:text-stone-400 focus:border-orange-500 focus:ring-2 focus:ring-orange-500/20 disabled:cursor-not-allowed disabled:opacity-60 dark:border-white/15 dark:bg-[#161617] dark:text-stone-100 dark:placeholder:text-stone-600"
              :disabled="isCreating"
              required
            />
          </UFormField>

          <UAlert
            v-if="createErrorMessage"
            color="error"
            variant="soft"
            icon="i-lucide-circle-alert"
            title="Couldn’t create the project"
            :description="createErrorMessage"
            role="alert"
          />

          <div class="flex flex-wrap justify-end gap-3">
            <UButton type="button" color="neutral" variant="ghost" :disabled="isCreating" @click="cancelCreate">
              Cancel
            </UButton>
            <UButton type="submit" color="primary" icon="i-lucide-plus" :loading="isCreating" :disabled="isCreating">
              {{ isCreating ? 'Creating…' : 'Create project' }}
            </UButton>
          </div>
        </form>
      </UCard>
    </Transition>

    <UAlert
      v-if="successMessage"
      class="mb-5"
      color="success"
      variant="soft"
      icon="i-lucide-circle-check"
      title="Project created"
      :description="successMessage"
      role="status"
    />

    <UCard
      variant="outline"
      class="overflow-hidden rounded-xl bg-white ring-stone-200 dark:bg-[#161617] dark:ring-white/10"
      :ui="{ body: 'p-0 sm:p-0' }"
    >
      <div v-if="isLoading" class="grid gap-0" role="status" aria-live="polite">
        <div v-for="row in 4" :key="row" class="flex items-center gap-4 border-b border-stone-200 p-5 last:border-b-0 dark:border-white/10">
          <USkeleton class="size-10 shrink-0 rounded-lg" />
          <div class="grid flex-1 gap-2">
            <USkeleton class="h-4 w-44 max-w-full" />
            <USkeleton class="h-3.5 w-80 max-w-full" />
          </div>
          <USkeleton class="h-6 w-20 rounded-full max-sm:hidden" />
        </div>
        <span class="sr-only">Loading projects…</span>
      </div>

      <div v-else-if="errorMessage" class="p-5 sm:p-7">
        <UAlert
          color="error"
          variant="soft"
          icon="i-lucide-cloud-off"
          title="We couldn’t load your projects"
          :description="errorMessage"
          role="alert"
        >
          <template #actions>
            <UButton color="error" variant="soft" size="sm" icon="i-lucide-refresh-cw" @click="loadProjects()">
              Try again
            </UButton>
          </template>
        </UAlert>
      </div>

      <div v-else-if="result?.projects.length" class="divide-y divide-stone-200 dark:divide-white/10">
        <RouterLink
          v-for="project in result.projects"
          :key="project.id"
          :to="`/projects/${project.id}`"
          class="group flex items-start gap-4 p-5 no-underline transition-colors hover:bg-stone-50 focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-orange-500 sm:p-6 dark:hover:bg-white/[0.025]"
          :aria-label="`View ${project.name}`"
        >
          <div class="grid size-10 shrink-0 place-items-center rounded-lg bg-orange-50 text-orange-700 ring-1 ring-orange-200 dark:bg-orange-500/10 dark:text-orange-300 dark:ring-orange-400/15">
            <UIcon name="i-lucide-folder-kanban" class="size-5" aria-hidden="true" />
          </div>

          <div class="min-w-0 flex-1">
            <div class="flex flex-wrap items-center justify-between gap-x-4 gap-y-2">
              <h2 class="truncate text-sm font-semibold text-stone-950 dark:text-white">
                {{ project.name }}
              </h2>
              <UBadge :color="statusDetails[project.status].color" variant="subtle" size="sm">
                {{ statusDetails[project.status].label }}
              </UBadge>
            </div>
            <p class="mt-1.5 line-clamp-2 text-sm leading-5 text-stone-600 dark:text-stone-400">
              {{ project.description || 'No description provided.' }}
            </p>
          </div>
          <UIcon
            name="i-lucide-chevron-right"
            class="mt-3 size-4 shrink-0 text-stone-400 transition-transform group-hover:translate-x-0.5 group-hover:text-orange-600 dark:text-stone-600 dark:group-hover:text-orange-400"
            aria-hidden="true"
          />
        </RouterLink>
      </div>

      <div v-else class="grid min-h-64 place-items-center p-8 text-center">
        <div>
          <div class="mx-auto grid size-12 place-items-center rounded-xl bg-stone-100 text-stone-500 dark:bg-white/5 dark:text-stone-400">
            <UIcon name="i-lucide-folder-open" class="size-6" aria-hidden="true" />
          </div>
          <h2 class="mt-4 text-sm font-semibold text-stone-950 dark:text-white">No projects yet</h2>
          <p class="mt-1.5 text-sm text-stone-500 dark:text-stone-400">
            Create your first project to start organizing email summaries.
          </p>
          <UButton class="mt-5" color="primary" icon="i-lucide-plus" @click="toggleCreateForm">
            Create project
          </UButton>
        </div>
      </div>
    </UCard>

    <div v-if="result && result.total > 0 && !errorMessage" class="mt-5 flex flex-wrap items-center justify-between gap-3">
      <p class="text-xs text-stone-500 dark:text-stone-500">
        Showing {{ rangeStart }}–{{ rangeEnd }} of {{ result.total }}
      </p>
      <div class="flex items-center gap-2" aria-label="Project pagination">
        <UButton
          color="neutral"
          variant="outline"
          size="sm"
          icon="i-lucide-chevron-left"
          aria-label="Previous page"
          :disabled="result.page <= 1 || isLoading"
          @click="loadProjects(result.page - 1)"
        />
        <span class="min-w-20 text-center text-xs font-medium text-stone-600 dark:text-stone-400">
          {{ result.page }} of {{ totalPages }}
        </span>
        <UButton
          color="neutral"
          variant="outline"
          size="sm"
          icon="i-lucide-chevron-right"
          aria-label="Next page"
          :disabled="result.page >= totalPages || isLoading"
          @click="loadProjects(result.page + 1)"
        />
      </div>
    </div>
  </section>
</template>
