/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
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
    TICKET_STATUSES.find((item) => item.value === status) ??
    TICKET_STATUSES[0]
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
    TICKET_CATEGORIES.find((item) => item.value === category)?.label ??
    category
  )
}
