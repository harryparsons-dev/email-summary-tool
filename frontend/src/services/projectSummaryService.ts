import { apiRequest } from '../lib/api'
import type { Summary } from '../models/summary'

function emailSummaryPath(projectId: string, emailId: string): string {
  return `/projects/${encodeURIComponent(projectId)}/emails/${encodeURIComponent(emailId)}/summarise`
}

// Starts summarising an email. The result starts as pending; poll
// getEmailSummary until it is not.
export function createEmailSummary(projectId: string, emailId: string): Promise<Summary> {
  return apiRequest<Summary>(emailSummaryPath(projectId, emailId), { method: 'POST' })
}

export function getEmailSummary(projectId: string, emailId: string, summaryId: string): Promise<Summary> {
  return apiRequest<Summary>(`${emailSummaryPath(projectId, emailId)}/${encodeURIComponent(summaryId)}`)
}

// Lists the email's summaries, newest first.
export function listEmailSummaries(projectId: string, emailId: string): Promise<Summary[]> {
  return apiRequest<Summary[]>(emailSummaryPath(projectId, emailId))
}
