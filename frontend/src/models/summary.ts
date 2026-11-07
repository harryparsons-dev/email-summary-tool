export type SummaryStatus = 'pending' | 'completed' | 'failed'

export interface Summary {
  id: string
  project_id: string
  // null when the summary covers the whole project.
  email_id: string | null
  status: SummaryStatus
  summary: string
  error: string
  created_at: string
  updated_at: string
}
