import type { StatusVariant } from '@/components/status-badge'

import type { TicketPriority, TicketStatus } from './types'

export const TICKET_STATUSES: Array<{
  value: TicketStatus
  label: string
  variant: StatusVariant
}> = [
  { value: 'open', label: 'Open', variant: 'info' },
  { value: 'pending', label: 'Pending', variant: 'warning' },
  { value: 'answered', label: 'Answered', variant: 'success' },
  { value: 'closed', label: 'Closed', variant: 'neutral' },
]

export const TICKET_PRIORITIES: Array<{
  value: TicketPriority
  label: string
  variant: StatusVariant
}> = [
  { value: 'low', label: 'Low', variant: 'neutral' },
  { value: 'normal', label: 'Normal', variant: 'blue' },
  { value: 'high', label: 'High', variant: 'warning' },
  { value: 'urgent', label: 'Urgent', variant: 'danger' },
]

export const TICKET_CATEGORIES = [
  { value: 'general', label: 'General' },
  { value: 'billing', label: 'Billing' },
  { value: 'technical', label: 'Technical' },
  { value: 'account', label: 'Account' },
  { value: 'channel', label: 'Channel' },
  { value: 'model', label: 'Model' },
  { value: 'quota', label: 'Quota' },
  { value: 'other', label: 'Other' },
]

export function getTicketStatusMeta(status: string) {
  return (
    TICKET_STATUSES.find((item) => item.value === status) ?? TICKET_STATUSES[0]
  )
}

export function getTicketPriorityMeta(priority: string) {
  return (
    TICKET_PRIORITIES.find((item) => item.value === priority) ??
    TICKET_PRIORITIES[1]
  )
}

export function getTicketCategoryLabel(category: string) {
  return (
    TICKET_CATEGORIES.find((item) => item.value === category)?.label ?? category
  )
}
