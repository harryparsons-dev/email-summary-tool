<script setup lang="ts">
import { useToast } from '@nuxt/ui/composables'
import { onMounted, ref } from 'vue'
import AppBreadcrumb from '../components/AppBreadcrumb.vue'
import CopyButton from '../components/CopyButton.vue'
import ProjectForm from '../components/ProjectForm.vue'
import { getStatusDetails } from '../lib/projectStatus'
import type { CreateProjectInput, PaginatedProjectResponse } from '../models/project'
import { createProject, listProjects } from '../services/projectService'

const DEFAULT_PAGE_SIZE = 12

const toast = useToast()
const result = ref<PaginatedProjectResponse | null>(null)
const isLoading = ref(true)
const errorMessage = ref('')
const requestedPage = ref(1)
const isCreateOpen = ref(false)
const isCreating = ref(false)
const createErrorMessage = ref('')
// Remounts the form so each opening starts empty.
const formKey = ref(0)

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

function openCreate() {
  createErrorMessage.value = ''
  formKey.value++
  isCreateOpen.value = true
}

async function submitProject(input: CreateProjectInput) {
  createErrorMessage.value = ''
  isCreating.value = true

  try {
    const project = await createProject(input)

    if (result.value) {
      result.value = {
        ...result.value,
        projects: result.value.page === 1
          ? [project, ...result.value.projects].slice(0, result.value.page_size)
          : result.value.projects,
        total: result.value.total + 1,
      }
    } else {
      result.value = { projects: [project], page: 1, page_size: DEFAULT_PAGE_SIZE, total: 1 }
      requestedPage.value = 1
    }

    isCreateOpen.value = false
    toast.add({ title: 'Project created', description: `“${project.name}” is ready for mail.`, color: 'success', icon: 'i-lucide-circle-check' })
  } catch (error) {
    createErrorMessage.value = error instanceof Error ? error.message : 'Unable to create the project.'
  } finally {
    isCreating.value = false
  }
}

onMounted(() => loadProjects())
</script>

<template>
  <UDashboardPanel id="projects">
    <template #header>
      <UDashboardNavbar>
        <template #title>
          <AppBreadcrumb :items="[{ label: 'Projects' }]" />
        </template>
        <template #trailing>
          <Transition name="swap">
            <UBadge v-if="result && !isLoading" :label="String(result.total)" variant="subtle" color="neutral" />
          </Transition>
        </template>
        <template #right>
          <UButton icon="i-lucide-plus" label="New project" @click="openCreate" />
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
      <div class="mx-auto grid w-full max-w-6xl gap-6">
        <div v-if="isLoading && !result" class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3" role="status" aria-live="polite">
          <USkeleton v-for="n in 6" :key="n" class="h-44 rounded-lg" />
          <span class="sr-only">Loading projects…</span>
        </div>

        <UAlert
          v-else-if="errorMessage"
          color="error"
          variant="subtle"
          icon="i-lucide-cloud-off"
          title="We couldn’t load your projects"
          :description="errorMessage"
          :actions="[{ label: 'Try again', icon: 'i-lucide-refresh-cw', color: 'error', variant: 'outline', onClick: () => loadProjects() }]"
          role="alert"
        />

        <UEmpty
          v-else-if="!result?.projects.length"
          icon="i-lucide-folder-open"
          title="No projects yet"
          description="Create your first project to get an inbox address for summaries."
          :actions="[{ label: 'Create project', icon: 'i-lucide-plus', onClick: openCreate }]"
          variant="naked"
          size="lg"
          class="py-16"
        />

        <TransitionGroup
          v-else
          tag="div"
          name="list"
          appear
          class="grid gap-4 transition-opacity duration-150 sm:grid-cols-2 xl:grid-cols-3"
          :class="{ 'pointer-events-none opacity-60': isLoading }"
        >
          <UPageCard
            v-for="(project, index) in result.projects"
            :key="project.id"
            :to="`/projects/${project.id}`"
            :aria-label="`View ${project.name}`"
            :title="project.name"
            :description="project.description || 'No description provided.'"
            variant="outline"
            :style="{ '--i': index }"
            class="group"
            :ui="{ container: 'h-full', wrapper: 'h-full', description: 'line-clamp-2', footer: 'mt-auto w-full pt-4' }"
          >
            <template #header>
              <div class="flex items-center justify-between gap-3">
                <UAvatar icon="i-lucide-folder-kanban" size="md" :ui="{ root: 'bg-primary/10 rounded-md', icon: 'text-primary' }" />
                <UBadge
                  :color="getStatusDetails(project.status).color"
                  :icon="getStatusDetails(project.status).icon"
                  :label="getStatusDetails(project.status).label"
                  variant="subtle"
                />
              </div>
            </template>

            <template #footer>
              <div class="relative z-[1] flex items-center gap-1 rounded-md bg-elevated/60 py-1 ps-2.5 pe-1">
                <UIcon name="i-lucide-at-sign" class="size-3.5 shrink-0 text-dimmed" aria-hidden="true" />
                <span class="min-w-0 flex-1 truncate font-mono text-xs text-toned">{{ project.email_address }}</span>
                <CopyButton :value="project.email_address" label="Copy address" />
              </div>
            </template>
          </UPageCard>
        </TransitionGroup>

        <div
          v-if="result && result.total > result.page_size && !errorMessage"
          class="flex items-center justify-center border-t border-default pt-4"
        >
          <UPagination
            :page="result.page"
            :total="result.total"
            :items-per-page="result.page_size"
            :disabled="isLoading"
            @update:page="(page: number) => loadProjects(page)"
          />
        </div>
      </div>

      <UModal
        v-model:open="isCreateOpen"
        title="New project"
        description="Each project gets its own address for incoming mail."
        :dismissible="!isCreating"
        :close="!isCreating"
        :ui="{ content: 'sm:max-w-xl', footer: 'justify-end' }"
      >
        <template #body>
          <ProjectForm
            id="create-project-form"
            :key="formKey"
            :disabled="isCreating"
            :error-message="createErrorMessage"
            error-title="Couldn’t create the project"
            @submit="submitProject"
          />
        </template>
        <template #footer>
          <UButton color="neutral" variant="ghost" label="Cancel" :disabled="isCreating" @click="isCreateOpen = false" />
          <UButton
            type="submit"
            form="create-project-form"
            icon="i-lucide-plus"
            :label="isCreating ? 'Creating…' : 'Create project'"
            :loading="isCreating"
          />
        </template>
      </UModal>
    </template>
  </UDashboardPanel>
</template>
