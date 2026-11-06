<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui'
import { useToast } from '@nuxt/ui/composables'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import CopyButton from '../components/CopyButton.vue'
import EmailReader from '../components/EmailReader.vue'
import ProjectForm from '../components/ProjectForm.vue'
import { formatRelative } from '../lib/format'
import { getStatusDetails } from '../lib/projectStatus'
import type { Email } from '../models/email'
import type { CreateProjectInput, Project } from '../models/project'
import { listProjectEmails } from '../services/projectEmailService'
import { deleteProject, getProject, updateProject } from '../services/projectService'

// New mail is checked for on this interval while the page is open.
const refreshIntervalMs = 10_000

const route = useRoute()
const router = useRouter()
const toast = useToast()
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
const refreshFailed = ref(false)
let refreshTimer: ReturnType<typeof setInterval> | undefined

const isEditOpen = ref(false)
const isUpdating = ref(false)
const updateErrorMessage = ref('')
const editFormKey = ref(0)

const isDeleteOpen = ref(false)
const isDeleting = ref(false)
const deleteErrorMessage = ref('')

// On small screens the reader opens in a slide-over instead of a second panel.
const desktopQuery = window.matchMedia('(min-width: 1024px)')
const isDesktop = ref(desktopQuery.matches)
const syncViewport = (event: MediaQueryListEvent) => (isDesktop.value = event.matches)
onMounted(() => desktopQuery.addEventListener('change', syncViewport))

const selectedEmail = computed(() => emails.value.find((email) => email.id === selectedId.value) ?? null)

const isReaderOpen = computed({
  get: () => !isDesktop.value && selectedEmail.value !== null,
  set: (open) => {
    if (!open) selectedId.value = null
  },
})

const projectMenu = computed<DropdownMenuItem[][]>(() => [
  [{ label: 'Edit project', icon: 'i-lucide-pencil', onSelect: openEdit }],
  [{ label: 'Delete project', icon: 'i-lucide-trash-2', color: 'error', onSelect: openDelete }],
])

function snippet(email: Email): string {
  return email.text_body.replace(/\s+/g, ' ').trim().slice(0, 140)
}

function sender(email: Email): string {
  return email.from_name || email.from_address || email.envelope_from || 'Unknown sender'
}

async function load() {
  isLoading.value = true
  errorMessage.value = ''
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

function openEdit() {
  updateErrorMessage.value = ''
  editFormKey.value++
  isEditOpen.value = true
}

async function submitUpdate(input: CreateProjectInput) {
  if (!project.value) return

  updateErrorMessage.value = ''
  isUpdating.value = true

  try {
    project.value = await updateProject(project.value.id, input)
    isEditOpen.value = false
    toast.add({ title: 'Project updated', description: `“${project.value.name}” was saved.`, color: 'success', icon: 'i-lucide-circle-check' })
  } catch (error) {
    updateErrorMessage.value = error instanceof Error ? error.message : 'Unable to update the project.'
  } finally {
    isUpdating.value = false
  }
}

function openDelete() {
  deleteErrorMessage.value = ''
  isDeleteOpen.value = true
}

async function confirmDelete() {
  if (!project.value) return

  isDeleting.value = true
  deleteErrorMessage.value = ''

  try {
    const name = project.value.name
    await deleteProject(project.value.id)
    clearInterval(refreshTimer)
    isDeleteOpen.value = false
    toast.add({ title: 'Project deleted', description: `“${name}” was removed.`, icon: 'i-lucide-trash-2' })
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

onBeforeUnmount(() => {
  clearInterval(refreshTimer)
  desktopQuery.removeEventListener('change', syncViewport)
})
</script>

<template>
  <UDashboardPanel id="project-inbox" :default-size="26" :min-size="20" :max-size="36" resizable>
    <template #header>
      <UDashboardNavbar :ui="{ title: 'min-w-0' }">
        <template #leading>
          <UDashboardSidebarCollapse />
          <UButton to="/projects" icon="i-lucide-arrow-left" color="neutral" variant="ghost" aria-label="Back to projects" class="lg:hidden" />
        </template>

        <template #title>
          <USkeleton v-if="isLoading" class="h-5 w-32" />
          <span v-else class="truncate">{{ project?.name ?? 'Project' }}</span>
        </template>

        <template #right>
          <UTooltip :text="refreshFailed ? 'Refresh failed — try again' : 'Checks every 10 seconds'">
            <UButton
              color="neutral"
              variant="ghost"
              icon="i-lucide-refresh-cw"
              :loading="isRefreshing"
              :disabled="!project"
              :class="{ 'text-error': refreshFailed }"
              aria-label="Refresh inbox"
              @click="refreshEmails"
            />
          </UTooltip>
          <UDropdownMenu v-if="project" :items="projectMenu" :content="{ align: 'end' }">
            <UButton color="neutral" variant="ghost" icon="i-lucide-ellipsis-vertical" aria-label="Project actions" />
          </UDropdownMenu>
        </template>
      </UDashboardNavbar>

      <Transition name="swap">
        <div v-if="project" class="grid gap-3 border-b border-default px-4 py-3 sm:px-6">
          <div class="flex items-center gap-2">
            <UBadge
              :color="getStatusDetails(project.status).color"
              :icon="getStatusDetails(project.status).icon"
              :label="getStatusDetails(project.status).label"
              variant="subtle"
            />
            <span class="text-xs text-muted">{{ emails.length }} {{ emails.length === 1 ? 'email' : 'emails' }}</span>
          </div>
          <p v-if="project.description" class="line-clamp-2 text-sm text-muted">{{ project.description }}</p>
          <div class="flex items-center gap-2 rounded-md bg-primary/5 py-1.5 ps-3 pe-1 ring ring-primary/20">
            <UIcon name="i-lucide-mail-plus" class="size-4 shrink-0 text-primary" aria-hidden="true" />
            <div class="min-w-0 flex-1">
              <p class="text-[0.65rem] font-semibold tracking-wide text-primary uppercase">Send or forward mail to</p>
              <p class="truncate font-mono text-xs text-highlighted">{{ project.email_address }}</p>
            </div>
            <CopyButton :value="project.email_address" label="Copy address" />
          </div>
        </div>
      </Transition>
    </template>

    <template #body>
      <div v-if="isLoading" class="grid gap-3" role="status" aria-live="polite">
        <USkeleton v-for="n in 5" :key="n" class="h-16 w-full" />
        <span class="sr-only">Loading project…</span>
      </div>

      <UAlert
        v-else-if="errorMessage || !project"
        color="error"
        variant="subtle"
        icon="i-lucide-cloud-off"
        title="We couldn’t load this project"
        :description="errorMessage"
        :actions="[
          { label: 'Try again', icon: 'i-lucide-refresh-cw', color: 'error', variant: 'outline', onClick: load },
          { label: 'Back to projects', color: 'neutral', variant: 'ghost', to: '/projects' },
        ]"
        role="alert"
      />

      <UEmpty
        v-else-if="emails.length === 0"
        icon="i-lucide-inbox"
        title="Waiting for mail"
        description="Emails sent to this project’s address will show up here automatically."
        variant="naked"
        class="py-12"
      />

      <TransitionGroup v-else tag="ul" name="list" appear class="-mx-4 -my-4 divide-y divide-default sm:-mx-6 sm:-my-6">
        <li v-for="(email, index) in emails" :key="email.id" :style="{ '--i': index }">
          <button
            type="button"
            class="relative block w-full border-s-2 px-4 py-3 text-left transition-colors duration-150 focus-visible:bg-elevated/50 focus-visible:outline-none sm:px-6"
            :class="email.id === selectedId
              ? 'border-primary bg-primary/5'
              : 'border-transparent hover:border-accented hover:bg-elevated/50'"
            :aria-current="email.id === selectedId ? 'true' : undefined"
            @click="selectedId = email.id"
          >
            <div class="flex items-baseline justify-between gap-3">
              <span class="truncate text-sm font-semibold text-highlighted">{{ sender(email) }}</span>
              <span class="shrink-0 text-xs text-dimmed">{{ formatRelative(email.received_at) }}</span>
            </div>
            <p class="mt-0.5 truncate text-sm text-default">{{ email.subject || '(no subject)' }}</p>
            <p class="mt-0.5 flex items-center gap-1.5 text-xs text-muted">
              <UIcon v-if="email.attachments.length" name="i-lucide-paperclip" class="size-3 shrink-0" aria-label="Has attachments" />
              <span class="truncate">{{ snippet(email) }}</span>
            </p>
          </button>
        </li>
      </TransitionGroup>
    </template>
  </UDashboardPanel>

  <UDashboardPanel v-if="isDesktop" id="project-reader">
    <template #header>
      <UDashboardNavbar :title="selectedEmail ? 'Message' : ''" :toggle="false">
        <template v-if="selectedEmail" #right>
          <UTooltip text="Close">
            <UButton icon="i-lucide-x" color="neutral" variant="ghost" aria-label="Close message" @click="selectedId = null" />
          </UTooltip>
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
      <Transition name="swap" mode="out-in">
        <EmailReader v-if="selectedEmail" :key="selectedEmail.id" :email="selectedEmail" />
        <UEmpty
          v-else
          icon="i-lucide-mail-open"
          :title="emails.length ? 'Select an email' : 'Nothing to read yet'"
          :description="emails.length ? 'Pick a message from the inbox to read it here.' : 'New mail will appear in the inbox on the left.'"
          variant="naked"
          class="flex-1"
        />
      </Transition>
    </template>
  </UDashboardPanel>

  <USlideover v-if="!isDesktop" v-model:open="isReaderOpen" title="Message" :ui="{ content: 'max-w-full' }">
    <template #body>
      <EmailReader v-if="selectedEmail" :key="selectedEmail.id" :email="selectedEmail" />
    </template>
  </USlideover>

  <UModal
    v-model:open="isEditOpen"
    title="Edit project"
    description="Update the project name, description, or status."
    :dismissible="!isUpdating"
    :close="!isUpdating"
    :ui="{ content: 'sm:max-w-xl', footer: 'justify-end' }"
  >
    <template #body>
      <ProjectForm
        v-if="project"
        id="edit-project-form"
        :key="editFormKey"
        :initial="{ name: project.name, description: project.description, status: project.status }"
        :disabled="isUpdating"
        :error-message="updateErrorMessage"
        error-title="Couldn’t update the project"
        @submit="submitUpdate"
      />
    </template>
    <template #footer>
      <UButton color="neutral" variant="ghost" label="Cancel" :disabled="isUpdating" @click="isEditOpen = false" />
      <UButton
        type="submit"
        form="edit-project-form"
        icon="i-lucide-save"
        :label="isUpdating ? 'Saving…' : 'Save changes'"
        :loading="isUpdating"
      />
    </template>
  </UModal>

  <UModal
    v-model:open="isDeleteOpen"
    title="Delete project?"
    description="This action cannot be undone."
    :dismissible="!isDeleting"
    :close="!isDeleting"
    :ui="{ content: 'sm:max-w-md', footer: 'justify-end' }"
  >
    <template #body>
      <div class="grid gap-4">
        <UAlert
          color="error"
          variant="subtle"
          icon="i-lucide-triangle-alert"
          :description="`“${project?.name}” and its project data will be permanently deleted.`"
        />
        <Transition name="swap">
          <UAlert
            v-if="deleteErrorMessage"
            color="error"
            variant="soft"
            icon="i-lucide-circle-alert"
            title="Couldn’t delete the project"
            :description="deleteErrorMessage"
            role="alert"
          />
        </Transition>
      </div>
    </template>
    <template #footer>
      <UButton color="neutral" variant="ghost" label="Cancel" :disabled="isDeleting" @click="isDeleteOpen = false" />
      <UButton
        color="error"
        icon="i-lucide-trash-2"
        :label="isDeleting ? 'Deleting…' : 'Delete project'"
        :loading="isDeleting"
        @click="confirmDelete"
      />
    </template>
  </UModal>
</template>
