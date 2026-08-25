export type ProjectStatus = 'pending' | 'in_progress' | 'completed' | 'archived'

export interface Project {
  id: string
  name: string
  description: string
  status: ProjectStatus
}

export interface CreateProjectInput {
  name: string
  description: string
  status: ProjectStatus
}

export interface UpdateProjectInput {
  name: string
  description: string
  status: ProjectStatus
}

export interface PaginatedProjectResponse {
  projects: Project[]
  page: number
  page_size: number
  total: number
}
