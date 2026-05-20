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
