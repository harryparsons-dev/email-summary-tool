import { apiRequest } from '../lib/api'
import type { Email } from '../models/email'

export function listProjectEmails(projectId: string): Promise<Email[]> {
  return apiRequest<Email[]>(`/projects/${encodeURIComponent(projectId)}/emails`)
}
