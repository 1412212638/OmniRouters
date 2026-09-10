import { api } from '@/lib/api'

export type AuditLog = {
  id: number
  event_id: string
  user_id: number
  username: string
  created_at: number
  category: string
  action: string
  token_ref?: string
  ip?: string
  success: boolean
  outcome?: string
  request_id?: string
  other?: string
}

export type AuditLogResponse = {
  items: AuditLog[]
  total: number
  page: number
  page_size: number
}

type AuditLogApiResponse = {
  success: boolean
  message?: string
  data?: AuditLogResponse
}

export type AuditFilters = {
  view: string
  username?: string
  user_id?: string
  action?: string
  ip?: string
  request_id?: string
  outcome?: string
  start?: number
  end?: number
}

export async function getAuditLogs(
  page = 1,
  pageSize = 20,
  filters: AuditFilters = { view: 'all' }
) {
  const response = await api.get<AuditLogApiResponse>('/api/audit', {
    params: { p: page, page_size: pageSize, ...filters },
    skipBusinessError: true,
    skipErrorHandler: true,
  })
  if (!response.data.success || !response.data.data) {
    throw new Error('Audit request failed')
  }
  return response.data.data
}
