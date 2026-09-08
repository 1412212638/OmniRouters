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

export async function getAuditLogs(page = 1, pageSize = 20) {
  const response = await api.get<AuditLogResponse>('/api/audit', {
    params: { p: page, page_size: pageSize },
  })
  return response.data
}
