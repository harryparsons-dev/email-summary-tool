import { apiRequest } from '../lib/api'
import type { CreateProjectRequest, Project } from '../models/project'

export function listProjects(): Promise<Project[]> {
  return apiRequest<Project[]>('/projects')
}

export function getProject(projectId: string): Promise<Project> {
  return apiRequest<Project>(`/projects/${encodeURIComponent(projectId)}`)
}

export function createProject(request: CreateProjectRequest): Promise<Project> {
  return apiRequest<Project, CreateProjectRequest>('/projects', {
    method: 'POST',
    body: request,
  })
}
