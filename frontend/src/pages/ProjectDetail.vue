<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import CopyButton from '../components/CopyButton.vue'
import EmailReader from '../components/EmailReader.vue'
import { formatRelative } from '../lib/format'
import type { Email } from '../models/email'
import type { Project, ProjectStatus } from '../models/project'
import { listProjectEmails } from '../services/projectEmailService'
import { deleteProject, getProject, updateProject } from '../services/projectService'

// New mail is checked for on this interval while the page is open.
const refreshIntervalMs = 10_000

const route = useRoute()
const router = useRouter()
const projectId = computed(() => {
  const id = route.params.id
  return String(Array.isArray(id) ? id[0] : id)
})

const project = ref<Project | null>(null)
const emails = ref<Email[]>([])
const selectedId = ref<string | null>(null)
const isLoading = ref(true)
const isRefreshing = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const refreshFailed = ref(false)
let refreshTimer: ReturnType<typeof setInterval> | undefined

const isEditDialogOpen = ref(false)
const isUpdating = ref(false)
const updateErrorMessage = ref('')
const projectName = ref('')
const projectDescription = ref('')
const projectStatus = ref<ProjectStatus>('pending')

const isDeleteDialogOpen = ref(false)
const isDeleting = ref(false)
const deleteErrorMessage = ref('')

const statusDetails: Record<ProjectStatus, { label: string; color: 'neutral' | 'warning' | 'info' | 'success' }> = {
  pending: { label: 'Pending', color: 'warning' },
  in_progress: { label: 'In progress', color: 'info' },
  completed: { label: 'Completed', color: 'success' },
  archived: { label: 'Archived', color: 'neutral' },
}

// Rows written before the current status set (e.g. 'active') must still render.
function getStatusDetails(status: string) {
  return statusDetails[status as ProjectStatus] ?? { label: status, color: 'neutral' as const }
}

const selectedEmail = computed(() => emails.value.find((email) => email.id === selectedId.value) ?? null)

function snippet(email: Email): string {
  return email.text_body.replace(/\s+/g, ' ').trim().slice(0, 140)
}

function sender(email: Email): string {
  return email.from_name || email.from_address || email.envelope_from || 'Unknown sender'
}

async function load() {
  isLoading.value = true
  errorMessage.value = ''
  successMessage.value = ''
  project.value = null
  emails.value = []
  selectedId.value = null

  try {
    const [loadedProject, loadedEmails] = await Promise.all([
      getProject(projectId.value),
      listProjectEmails(projectId.value),
    ])
    project.value = loadedProject
    emails.value = loadedEmails
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Unable to load this project.'
  } finally {
    isLoading.value = false
  }
}

async function refreshEmails() {
  if (!project.value || isRefreshing.value) return
  isRefreshing.value = true

  try {
    emails.value = await listProjectEmails(project.value.id)
    refreshFailed.value = false
  } catch {
    refreshFailed.value = true
  } finally {
    isRefreshing.value = false
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
    clearInterval(refreshTimer)
    await router.replace('/projects')
  } catch (error) {
    deleteErrorMessage.value = error instanceof Error ? error.message : 'Unable to delete the project.'
    isDeleting.value = false
  }
}

function startAutoRefresh() {
  clearInterval(refreshTimer)
  refreshTimer = setInterval(() => {
    if (document.visibilityState === 'visible') refreshEmails()
  }, refreshIntervalMs)
}

watch(
  projectId,
  async () => {
    await load()
    startAutoRefresh()
  },
  { immediate: true },
)

onBeforeUnmount(() => clearInterval(refreshTimer))
</script>

<template>
  <section class="w-full self-start lg:pt-[4vh]" aria-labelledby="project-title">
    <div class="mb-3 flex items-center gap-2 text-xs font-medium text-stone-500">
      <RouterLink to="/projects" class="text-stone-500 no-underline hover:text-stone-800 dark:hover:text-stone-300">
        Projects
      </RouterLink>
      <UIcon name="i-lucide-chevron-right" class="size-3.5" aria-hidden="true" />
      <span class="truncate">{{ project?.name ?? 'Project' }}</span>
    </div>

    <div v-if="isLoading" class="grid gap-4" role="status" aria-live="polite">
      <USkeleton class="h-9 w-64 max-w-full" />
      <USkeleton class="h-14 w-full max-w-xl" />
      <USkeleton class="h-80 w-full rounded-xl" />
      <span class="sr-only">Loading project…</span>
    </div>

    <UAlert
      v-else-if="errorMessage || !project"
      color="error"
      variant="soft"
      icon="i-lucide-cloud-off"
      title="We couldn’t load this project"
      :description="errorMessage"
      role="alert"
    >
      <template #actions>
        <UButton color="error" variant="soft" size="sm" icon="i-lucide-refresh-cw" @click="load">Try again</UButton>
        <UButton color="neutral" variant="ghost" size="sm" to="/projects">Back to projects</UButton>
      </template>
    </UAlert>

    <template v-else>
      <div class="mb-7 flex flex-wrap items-end justify-between gap-4">
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-3">
            <h1 id="project-title" class="text-[2rem] leading-tight font-bold tracking-[-0.035em] text-stone-950 dark:text-white">
              {{ project.name }}
            </h1>
            <UBadge :color="getStatusDetails(project.status).color" variant="subtle" size="md">
              {{ getStatusDetails(project.status).label }}
            </UBadge>
          </div>
          <p v-if="project.description" class="mt-2 max-w-2xl text-[0.95rem] leading-6 text-stone-600 dark:text-stone-400">
            {{ project.description }}
          </p>
        </div>

        <div class="flex items-center gap-3">
          <UButton color="neutral" variant="outline" icon="i-lucide-pencil" @click="openEditDialog">
            Edit
          </UButton>
          <UButton color="error" variant="soft" icon="i-lucide-trash-2" @click="openDeleteDialog">
            Delete
          </UButton>
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

      <div class="mb-6 flex max-w-xl items-center gap-2 rounded-xl bg-orange-50 py-2 pr-2 pl-4 ring-1 ring-orange-200 dark:bg-orange-500/10 dark:ring-orange-400/20">
        <UIcon name="i-lucide-mail-plus" class="size-4.5 shrink-0 text-orange-700 dark:text-orange-300" aria-hidden="true" />
        <div class="min-w-0 flex-1">
          <p class="text-[0.68rem] font-semibold tracking-[0.06em] text-orange-800 uppercase dark:text-orange-300">
            Send or forward mail to
          </p>
          <p class="truncate font-mono text-sm text-stone-900 dark:text-stone-100">{{ project.email_address }}</p>
        </div>
        <CopyButton :value="project.email_address" label="Copy address" />
      </div>

      <div class="grid gap-4 lg:grid-cols-[minmax(0,22rem)_minmax(0,1fr)]">
        <UCard
          variant="outline"
          class="self-start rounded-xl bg-white ring-stone-200 dark:bg-[#161617] dark:ring-white/10"
          :ui="{ header: 'px-4 py-3 sm:px-4', body: 'p-0 sm:p-0' }"
        >
          <template #header>
            <div class="flex items-center justify-between gap-2">
              <h2 class="text-sm font-semibold text-stone-900 dark:text-stone-100">
                Inbox
                <span class="ml-1 font-normal text-stone-500">{{ emails.length }}</span>
              </h2>
              <UTooltip :text="refreshFailed ? 'Refresh failed — try again' : 'Checks every 10 seconds'">
                <UButton
                  color="neutral"
                  variant="ghost"
                  size="xs"
                  icon="i-lucide-refresh-cw"
                  :loading="isRefreshing"
                  :class="refreshFailed ? 'text-red-600 dark:text-red-400' : ''"
                  aria-label="Refresh inbox"
                  @click="refreshEmails"
                />
              </UTooltip>
            </div>
          </template>

          <UEmpty
            v-if="emails.length === 0"
            icon="i-lucide-inbox"
            title="Waiting for mail"
            description="Emails sent to this project’s address will show up here automatically."
            variant="naked"
            size="sm"
            class="px-4 py-10"
          />

          <ul v-else class="max-h-[34rem] divide-y divide-stone-200 overflow-y-auto dark:divide-white/10">
            <li v-for="email in emails" :key="email.id">
              <button
                type="button"
                class="block w-full px-4 py-3 text-left transition hover:bg-stone-50 focus-visible:bg-stone-50 focus-visible:outline-none dark:hover:bg-white/5 dark:focus-visible:bg-white/5"
                :class="email.id === selectedId ? 'bg-orange-50 hover:bg-orange-50 dark:bg-orange-500/10 dark:hover:bg-orange-500/10' : ''"
                :aria-current="email.id === selectedId ? 'true' : undefined"
                @click="selectedId = email.id"
              >
                <div class="flex items-baseline justify-between gap-3">
                  <span class="truncate text-sm font-semibold text-stone-900 dark:text-stone-100">{{ sender(email) }}</span>
                  <span class="shrink-0 text-xs text-stone-500">{{ formatRelative(email.received_at) }}</span>
                </div>
                <p class="mt-0.5 truncate text-sm text-stone-800 dark:text-stone-200">
                  {{ email.subject || '(no subject)' }}
                </p>
                <p class="mt-0.5 flex items-center gap-1.5 text-xs text-stone-500">
                  <UIcon
                    v-if="email.attachments.length"
                    name="i-lucide-paperclip"
                    class="size-3 shrink-0"
                    aria-label="Has attachments"
                  />
                  <span class="truncate">{{ snippet(email) }}</span>
                </p>
              </button>
            </li>
          </ul>
        </UCard>

        <UCard
          variant="outline"
          class="min-h-80 rounded-xl bg-white ring-stone-200 dark:bg-[#161617] dark:ring-white/10"
          :ui="{ body: 'p-5 sm:p-6' }"
        >
          <EmailReader v-if="selectedEmail" :key="selectedEmail.id" :email="selectedEmail" />
          <div v-else class="grid min-h-64 place-items-center content-center text-center">
            <UIcon name="i-lucide-mail-open" class="mb-3 size-6 text-stone-400" aria-hidden="true" />
            <p class="text-sm text-stone-500">
              {{ emails.length ? 'Select an email to read it.' : 'Nothing to read yet.' }}
            </p>
          </div>
        </UCard>
      </div>
    </template>

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
