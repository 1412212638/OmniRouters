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

export async function getAuditLogs(page = 1, pageSize = 20) {
  const response = await api.get<AuditLogApiResponse>('/api/audit', {
    params: { p: page, page_size: pageSize },
  })
  return response.data.data ?? { items: [], total: 0, page: page, page_size: pageSize }
}
