const dateTimeFormat = new Intl.DateTimeFormat(undefined, {
  dateStyle: 'medium',
  timeStyle: 'short',
})

const relativeFormat = new Intl.RelativeTimeFormat(undefined, { numeric: 'auto' })

export function formatDateTime(value: string): string {
  return dateTimeFormat.format(new Date(value))
}

// Short, inbox-style time: "5 minutes ago", "yesterday", then a full date.
export function formatRelative(value: string): string {
  const seconds = Math.round((new Date(value).getTime() - Date.now()) / 1000)
  const absolute = Math.abs(seconds)

  if (absolute < 60) return 'just now'
  if (absolute < 3600) return relativeFormat.format(Math.round(seconds / 60), 'minute')
  if (absolute < 86400) return relativeFormat.format(Math.round(seconds / 3600), 'hour')
  if (absolute < 7 * 86400) return relativeFormat.format(Math.round(seconds / 86400), 'day')
  return dateTimeFormat.format(new Date(value))
}

export function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}
