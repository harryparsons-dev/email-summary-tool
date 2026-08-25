import { apiRequest } from '../lib/api'
import type {
  CreateProjectInput,
  PaginatedProjectResponse,
  Project,
  UpdateProjectInput,
} from '../models/project'

export interface ListProjectsParams {
  page?: number
  pageSize?: number
}

export function listProjects({
  page = 1,
  pageSize = 10,
}: ListProjectsParams = {}): Promise<PaginatedProjectResponse> {
  const query = new URLSearchParams({
    page: page.toString(),
    page_size: pageSize.toString(),
  })

  return apiRequest<PaginatedProjectResponse>(`/projects?${query.toString()}`)
}

export function createProject(input: CreateProjectInput): Promise<Project> {
  return apiRequest<Project, CreateProjectInput>('/projects', {
    method: 'POST',
    body: input,
  })
}

export function getProject(projectId: string): Promise<Project> {
  return apiRequest<Project>(`/projects/${encodeURIComponent(projectId)}`)
}

export function updateProject(projectId: string, input: UpdateProjectInput): Promise<Project> {
  return apiRequest<Project, UpdateProjectInput>(`/projects/${encodeURIComponent(projectId)}`, {
    method: 'PUT',
    body: input,
  })
}

export function deleteProject(projectId: string): Promise<void> {
  return apiRequest<void>(`/projects/${encodeURIComponent(projectId)}`, {
    method: 'DELETE',
  })
}
