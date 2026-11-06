export interface Project {
  id: string
  user_id: string
  title: string
  description: string
  email_address: string
  status: string
  created_at: string
  updated_at: string
}

export interface CreateProjectRequest {
  title: string
  description?: string
}
