import { api } from '@/lib/api'

import type {
  ApiResponse,
  CreateTicketPayload,
  ListTicketsParams,
  SendTicketMessagePayload,
  TicketDetailResponse,
  TicketListResponse,
} from './types'

function getTicketBase(admin?: boolean) {
  return admin ? '/api/tickets/admin' : '/api/tickets'
}

export async function listTickets(
  params: ListTicketsParams
): Promise<ApiResponse<TicketListResponse>> {
  const { admin, ...query } = params
  const res = await api.get(getTicketBase(admin), { params: query })
  return res.data
}

export async function createTicket(
  payload: CreateTicketPayload
): Promise<ApiResponse<TicketDetailResponse>> {
  const res = await api.post('/api/tickets', payload)
  return res.data
}

export async function getTicket(
  id: number,
  admin?: boolean
): Promise<ApiResponse<TicketDetailResponse>> {
  const res = await api.get(`${getTicketBase(admin)}/${id}`)
  return res.data
}

export async function sendTicketMessage(
  id: number,
  payload: SendTicketMessagePayload,
  admin?: boolean
): Promise<ApiResponse<TicketDetailResponse>> {
  const res = await api.post(`${getTicketBase(admin)}/${id}/messages`, payload)
  return res.data
}

export async function updateTicketStatus(
  id: number,
  status: string
): Promise<ApiResponse<TicketDetailResponse>> {
  const res = await api.patch(`/api/tickets/admin/${id}/status`, { status })
  return res.data
}

export async function assignTicket(
  id: number,
  assignedAdminId: number
): Promise<ApiResponse<TicketDetailResponse>> {
  const res = await api.patch(`/api/tickets/admin/${id}/assign`, {
    assigned_admin_id: assignedAdminId,
  })
  return res.data
}

export async function closeTicket(
  id: number,
  admin?: boolean
): Promise<ApiResponse<TicketDetailResponse>> {
  const res = await api.post(`${getTicketBase(admin)}/${id}/close`)
  return res.data
}

export async function reopenTicket(
  id: number
): Promise<ApiResponse<TicketDetailResponse>> {
  const res = await api.post(`/api/tickets/${id}/reopen`)
  return res.data
}
