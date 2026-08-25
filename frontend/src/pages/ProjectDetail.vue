<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { Project, ProjectStatus } from '../models/project'
import { deleteProject, getProject, updateProject } from '../services/projectService'

const route = useRoute()
const router = useRouter()

const project = ref<Project | null>(null)
const isLoading = ref(true)
const errorMessage = ref('')
const successMessage = ref('')

const isEditDialogOpen = ref(false)
const isUpdating = ref(false)
const updateErrorMessage = ref('')
const projectName = ref('')
const projectDescription = ref('')
const projectStatus = ref<ProjectStatus>('pending')

const isDeleteDialogOpen = ref(false)
const isDeleting = ref(false)
const deleteErrorMessage = ref('')

const projectId = computed(() => {
  const id = route.params.id
  return Array.isArray(id) ? id[0] : id
})

const statusDetails: Record<ProjectStatus, { label: string; color: 'neutral' | 'warning' | 'info' | 'success' }> = {
  pending: { label: 'Pending', color: 'warning' },
  in_progress: { label: 'In progress', color: 'info' },
  completed: { label: 'Completed', color: 'success' },
  archived: { label: 'Archived', color: 'neutral' },
}

async function loadProject(id = projectId.value) {
  isLoading.value = true
  errorMessage.value = ''
  successMessage.value = ''

  if (!id) {
    project.value = null
    errorMessage.value = 'This project link is invalid.'
    isLoading.value = false
    return
  }

  try {
    project.value = await getProject(id)
  } catch (error) {
    project.value = null
    errorMessage.value = error instanceof Error ? error.message : 'Unable to load the project.'
  } finally {
    isLoading.value = false
  }
}

function openEditDialog() {
  if (!project.value) {
    return
  }

  projectName.value = project.value.name
  projectDescription.value = project.value.description
  projectStatus.value = project.value.status
  updateErrorMessage.value = ''
  successMessage.value = ''
  isEditDialogOpen.value = true
}

async function submitUpdate() {
  if (!project.value) {
    return
  }

  const name = projectName.value.trim()
  const description = projectDescription.value.trim()

  updateErrorMessage.value = ''

  if (!name || !description) {
    updateErrorMessage.value = 'Enter a name and description for the project.'
    return
  }

  isUpdating.value = true

  try {
    const updatedProject = await updateProject(project.value.id, {
      name,
      description,
      status: projectStatus.value,
    })

    project.value = updatedProject
    isEditDialogOpen.value = false
    successMessage.value = `“${updatedProject.name}” was updated.`
  } catch (error) {
    updateErrorMessage.value = error instanceof Error ? error.message : 'Unable to update the project.'
  } finally {
    isUpdating.value = false
  }
}

function openDeleteDialog() {
  deleteErrorMessage.value = ''
  successMessage.value = ''
  isDeleteDialogOpen.value = true
}

async function confirmDelete() {
  if (!project.value) {
    return
  }

  isDeleting.value = true
  deleteErrorMessage.value = ''

  try {
    await deleteProject(project.value.id)
    await router.replace('/projects')
  } catch (error) {
    deleteErrorMessage.value = error instanceof Error ? error.message : 'Unable to delete the project.'
    isDeleting.value = false
  }
}

watch(projectId, (id) => loadProject(id), { immediate: true })
</script>

<template>
  <section class="mx-auto w-full max-w-5xl self-start lg:pt-[7vh]" aria-labelledby="project-title">
    <div class="mb-8">
      <div class="mb-5 flex items-center gap-2 text-xs font-medium text-stone-500 dark:text-stone-500">
        <span>Workspace</span>
        <UIcon name="i-lucide-chevron-right" class="size-3.5" aria-hidden="true" />
        <RouterLink to="/projects" class="rounded-sm text-stone-500 no-underline hover:text-orange-700 focus-visible:outline-2 focus-visible:outline-orange-500 dark:hover:text-orange-300">
          Projects
        </RouterLink>
        <UIcon name="i-lucide-chevron-right" class="size-3.5" aria-hidden="true" />
        <span class="max-w-48 truncate">{{ project?.name || 'Project details' }}</span>
      </div>

      <div class="flex flex-wrap items-start justify-between gap-5">
        <div class="min-w-0">
          <div class="mb-3 flex items-center gap-3">
            <RouterLink
              to="/projects"
              class="grid size-9 shrink-0 place-items-center rounded-lg text-stone-500 no-underline ring-1 ring-stone-200 transition hover:bg-white hover:text-stone-900 focus-visible:outline-2 focus-visible:outline-orange-500 dark:text-stone-400 dark:ring-white/10 dark:hover:bg-white/5 dark:hover:text-white"
              aria-label="Back to projects"
            >
              <UIcon name="i-lucide-arrow-left" class="size-4" aria-hidden="true" />
            </RouterLink>
            <h1 id="project-title" class="truncate text-[2rem] leading-tight font-bold tracking-[-0.035em] text-stone-950 dark:text-white">
              {{ project?.name || (isLoading ? 'Loading project…' : 'Project details') }}
            </h1>
          </div>
          <p class="text-[0.95rem] leading-6 text-stone-600 dark:text-stone-400">
            View and manage this email summary project.
          </p>
        </div>

        <div v-if="project && !isLoading" class="flex items-center gap-3">
          <UButton color="neutral" variant="outline" icon="i-lucide-pencil" @click="openEditDialog">
            Edit
          </UButton>
          <UButton color="error" variant="soft" icon="i-lucide-trash-2" @click="openDeleteDialog">
            Delete
          </UButton>
        </div>
      </div>
    </div>

    <UAlert
      v-if="successMessage"
      class="mb-5"
      color="success"
      variant="soft"
      icon="i-lucide-circle-check"
      title="Project updated"
      :description="successMessage"
      role="status"
    />

    <UCard
      variant="outline"
      class="rounded-xl bg-white ring-stone-200 dark:bg-[#161617] dark:ring-white/10"
      :ui="{ body: 'p-5 sm:p-7' }"
    >
      <div v-if="isLoading" class="grid min-h-60 content-center gap-6" role="status" aria-live="polite">
        <div class="flex items-center gap-4">
          <USkeleton class="size-12 shrink-0 rounded-xl" />
          <div class="grid flex-1 gap-2">
            <USkeleton class="h-5 w-56 max-w-full" />
            <USkeleton class="h-4 w-24" />
          </div>
        </div>
        <USeparator />
        <div class="grid gap-3">
          <USkeleton class="h-4 w-24" />
          <USkeleton class="h-4 w-full" />
          <USkeleton class="h-4 w-4/5" />
        </div>
        <span class="sr-only">Loading project details…</span>
      </div>

      <UAlert
        v-else-if="errorMessage"
        color="error"
        variant="soft"
        icon="i-lucide-folder-x"
        title="We couldn’t load this project"
        :description="errorMessage"
        role="alert"
      >
        <template #actions>
          <div class="flex flex-wrap gap-2">
            <UButton color="error" variant="soft" size="sm" icon="i-lucide-refresh-cw" @click="loadProject()">
              Try again
            </UButton>
            <UButton to="/projects" color="neutral" variant="ghost" size="sm">
              Back to projects
            </UButton>
          </div>
        </template>
      </UAlert>

      <template v-else-if="project">
        <div class="flex flex-wrap items-center justify-between gap-4">
          <div class="flex min-w-0 items-center gap-4">
            <div class="grid size-12 shrink-0 place-items-center rounded-xl bg-orange-50 text-orange-700 ring-1 ring-orange-200 dark:bg-orange-500/10 dark:text-orange-300 dark:ring-orange-400/15">
              <UIcon name="i-lucide-folder-kanban" class="size-6" aria-hidden="true" />
            </div>
            <div class="min-w-0">
              <p class="truncate text-base font-semibold text-stone-950 dark:text-white">{{ project.name }}</p>
              <p class="mt-1 text-xs text-stone-500">Project overview</p>
            </div>
          </div>
          <UBadge :color="statusDetails[project.status].color" variant="subtle" size="md">
            {{ statusDetails[project.status].label }}
          </UBadge>
        </div>

        <USeparator class="my-7" />

        <dl class="grid gap-7">
          <div>
            <dt class="flex items-center gap-2 text-xs font-semibold tracking-[0.06em] text-stone-500 uppercase dark:text-stone-500">
              <UIcon name="i-lucide-align-left" class="size-4" aria-hidden="true" />
              Description
            </dt>
            <dd class="mt-3 whitespace-pre-wrap text-sm leading-6 text-stone-700 dark:text-stone-300">
              {{ project.description }}
            </dd>
          </div>
          <div>
            <dt class="flex items-center gap-2 text-xs font-semibold tracking-[0.06em] text-stone-500 uppercase dark:text-stone-500">
              <UIcon name="i-lucide-fingerprint" class="size-4" aria-hidden="true" />
              Project ID
            </dt>
            <dd class="mt-3 break-all font-mono text-xs text-stone-600 dark:text-stone-400">{{ project.id }}</dd>
          </div>
        </dl>
      </template>
    </UCard>

    <UModal
      v-model:open="isEditDialogOpen"
      title="Edit project"
      description="Update the project name, description, or status."
      :dismissible="!isUpdating"
      :close="!isUpdating"
      :ui="{ content: 'sm:max-w-xl', footer: 'justify-end' }"
    >
      <template #body>
        <form id="edit-project-form" class="grid gap-5" @submit.prevent="submitUpdate">
          <UFormField label="Project name" name="edit-project-name" required>
            <UInput
              id="edit-project-name"
              v-model="projectName"
              name="edit-project-name"
              icon="i-lucide-folder-kanban"
              maxlength="255"
              size="xl"
              class="w-full"
              :disabled="isUpdating"
              required
            />
          </UFormField>

          <UFormField label="Status" name="edit-project-status" required>
            <div class="relative">
              <select
                id="edit-project-status"
                v-model="projectStatus"
                name="edit-project-status"
                class="h-12 w-full appearance-none rounded-md border border-stone-300 bg-white px-3.5 pr-10 text-sm text-stone-900 shadow-sm outline-none transition focus:border-orange-500 focus:ring-2 focus:ring-orange-500/20 disabled:cursor-not-allowed disabled:opacity-60 dark:border-white/15 dark:bg-[#161617] dark:text-stone-100"
                :disabled="isUpdating"
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

          <UFormField label="Description" name="edit-project-description" required>
            <textarea
              id="edit-project-description"
              v-model="projectDescription"
              name="edit-project-description"
              rows="5"
              class="block w-full resize-y rounded-md border border-stone-300 bg-white px-3.5 py-3 text-sm leading-5 text-stone-900 shadow-sm outline-none transition placeholder:text-stone-400 focus:border-orange-500 focus:ring-2 focus:ring-orange-500/20 disabled:cursor-not-allowed disabled:opacity-60 dark:border-white/15 dark:bg-[#161617] dark:text-stone-100 dark:placeholder:text-stone-600"
              :disabled="isUpdating"
              required
            />
          </UFormField>

          <UAlert
            v-if="updateErrorMessage"
            color="error"
            variant="soft"
            icon="i-lucide-circle-alert"
            title="Couldn’t update the project"
            :description="updateErrorMessage"
            role="alert"
          />
        </form>
      </template>

      <template #footer>
        <UButton color="neutral" variant="ghost" :disabled="isUpdating" @click="isEditDialogOpen = false">
          Cancel
        </UButton>
        <UButton
          type="submit"
          form="edit-project-form"
          color="primary"
          icon="i-lucide-save"
          :loading="isUpdating"
          :disabled="isUpdating"
        >
          {{ isUpdating ? 'Saving…' : 'Save changes' }}
        </UButton>
      </template>
    </UModal>

    <UModal
      v-model:open="isDeleteDialogOpen"
      title="Delete project?"
      description="This action cannot be undone."
      :dismissible="!isDeleting"
      :close="!isDeleting"
      :ui="{ content: 'sm:max-w-md', footer: 'justify-end' }"
    >
      <template #body>
        <div class="grid gap-4">
          <div class="flex items-start gap-3 rounded-lg bg-red-50 p-4 text-red-900 ring-1 ring-red-200 dark:bg-red-500/10 dark:text-red-200 dark:ring-red-400/15">
            <UIcon name="i-lucide-triangle-alert" class="mt-0.5 size-5 shrink-0" aria-hidden="true" />
            <p class="text-sm leading-5">
              <span class="font-semibold">“{{ project?.name }}”</span> and its project data will be permanently deleted.
            </p>
          </div>

          <UAlert
            v-if="deleteErrorMessage"
            color="error"
            variant="soft"
            icon="i-lucide-circle-alert"
            title="Couldn’t delete the project"
            :description="deleteErrorMessage"
            role="alert"
          />
        </div>
      </template>

      <template #footer>
        <UButton color="neutral" variant="ghost" :disabled="isDeleting" @click="isDeleteDialogOpen = false">
          Cancel
        </UButton>
        <UButton color="error" icon="i-lucide-trash-2" :loading="isDeleting" :disabled="isDeleting" @click="confirmDelete">
          {{ isDeleting ? 'Deleting…' : 'Delete project' }}
        </UButton>
      </template>
    </UModal>
  </section>
</template>
