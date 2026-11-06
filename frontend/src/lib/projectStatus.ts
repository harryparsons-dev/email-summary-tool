import type { ProjectStatus } from '../models/project'

type StatusColor = 'neutral' | 'warning' | 'info' | 'success'

export const statusDetails: Record<ProjectStatus, { label: string; color: StatusColor; icon: string }> = {
  pending: { label: 'Pending', color: 'warning', icon: 'i-lucide-circle-dashed' },
  in_progress: { label: 'In progress', color: 'info', icon: 'i-lucide-loader-circle' },
  completed: { label: 'Completed', color: 'success', icon: 'i-lucide-circle-check' },
  archived: { label: 'Archived', color: 'neutral', icon: 'i-lucide-archive' },
}

export const statusOptions = (Object.keys(statusDetails) as ProjectStatus[]).map((value) => ({
  value,
  label: statusDetails[value].label,
  icon: statusDetails[value].icon,
}))

// Rows written before the current status set (e.g. 'active') must still render.
export function getStatusDetails(status: string) {
  return statusDetails[status as ProjectStatus] ?? { label: status, color: 'neutral' as const, icon: 'i-lucide-circle' }
}
