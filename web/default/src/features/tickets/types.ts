export type TicketStatus = 'open' | 'pending' | 'answered' | 'closed'
export type TicketPriority = 'low' | 'normal' | 'high' | 'urgent'
export type TicketSenderRole = 'user' | 'admin' | 'system'

export interface TicketAttachment {
  name?: string
  url?: string
  size?: number
  type?: string
}

export interface Ticket {
  id: number
  user_id: number
  username?: string
  user_display_name?: string
  title: string
  category: string
  priority: TicketPriority
  status: TicketStatus
  closed_at: number
  closed_by_id: number
  closed_by_role?: TicketSenderRole
  close_reason?: string
  reopen_until: number
  assigned_admin_id: number
  assigned_admin_name?: string
  related_type?: string
  related_id?: number
  last_reply_at: number
  last_user_reply_at: number
  last_admin_reply_at: number
  user_unread_count: number
  admin_unread_count: number
  created_at: number
  updated_at: number
}

export interface TicketMessage {
  id: number
  ticket_id: number
  sender_id: number
  sender_role: TicketSenderRole
  sender_name?: string
  content: string
  internal: boolean
  attachments?: TicketAttachment[]
  created_at: number
}

export interface TicketListResponse {
  items: Ticket[]
  total: number
  page: number
  page_size: number
}

export interface TicketDetailResponse {
  ticket: Ticket
  messages: TicketMessage[]
}

export interface ApiResponse<T = unknown> {
  success: boolean
  message?: string
  data?: T
}

export interface ListTicketsParams {
  admin?: boolean
  p?: number
  page_size?: number
  status?: string
  category?: string
  priority?: string
  keyword?: string
  assigned_admin_id?: number
}

export interface CreateTicketPayload {
  title: string
  category: string
  priority: string
  content: string
  related_type?: string
  related_id?: number
  attachments?: TicketAttachment[]
}

export interface SendTicketMessagePayload {
  content: string
  internal?: boolean
  attachments?: TicketAttachment[]
}
