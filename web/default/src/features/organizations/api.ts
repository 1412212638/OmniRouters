import { api } from '@/lib/api'

import type {
  OrganizationApiResponse,
  OrganizationInvitation,
  OrganizationMember,
  OrganizationRole,
  OrganizationToken,
  OrganizationUsage,
  OrganizationWalletTransaction,
  OrganizationWorkspace,
  OrganizationLogPage,
} from './types'

async function getData<T>(url: string): Promise<T> {
  const response = await api.get<OrganizationApiResponse<T>>(url)
  if (!response.data.success || response.data.data === undefined) {
    throw new Error(response.data.message || 'Request failed')
  }
  return response.data.data
}

async function mutate<T>(
  method: 'post' | 'patch' | 'delete',
  url: string,
  data?: unknown
): Promise<T | undefined> {
  const response = await api.request<OrganizationApiResponse<T>>({
    method,
    url,
    data,
  })
  if (!response.data.success) {
    throw new Error(response.data.message || 'Request failed')
  }
  return response.data.data
}

export const getOrganizations = () =>
  getData<OrganizationWorkspace[]>('/api/organizations')
export const createOrganization = (name: string) =>
  mutate<OrganizationWorkspace>('post', '/api/organizations', { name })
export const getOrganization = (id: number) =>
  getData<OrganizationWorkspace>(`/api/organizations/${id}`)
export const updateOrganization = (id: number, name: string) =>
  mutate('patch', `/api/organizations/${id}`, { name })
export const getOrganizationMembers = (id: number) =>
  getData<OrganizationMember[]>(`/api/organizations/${id}/members`)
export const updateOrganizationMember = (
  id: number,
  userId: number,
  data: { role?: OrganizationRole; monthly_quota_limit?: number }
) => mutate('patch', `/api/organizations/${id}/members/${userId}`, data)
export const removeOrganizationMember = (id: number, userId: number) =>
  mutate('delete', `/api/organizations/${id}/members/${userId}`)
export const leaveOrganization = (id: number) =>
  mutate('post', `/api/organizations/${id}/leave`)
export const transferOrganizationOwnership = (id: number, userId: number) =>
  mutate('post', `/api/organizations/${id}/owner`, { user_id: userId })
export const getMyOrganizationInvitations = () =>
  getData<OrganizationInvitation[]>('/api/organizations/invitations')
export const acceptOrganizationInvitation = (id: number) =>
  mutate<OrganizationWorkspace>(
    'post',
    `/api/organizations/invitations/${id}/accept`
  )
export const getOrganizationInvitations = (id: number) =>
  getData<OrganizationInvitation[]>(`/api/organizations/${id}/invitations`)
export const inviteOrganizationMember = (
  id: number,
  email: string,
  role: 'admin' | 'member'
) => mutate('post', `/api/organizations/${id}/invitations`, { email, role })
export const revokeOrganizationInvitation = (
  id: number,
  invitationId: number
) => mutate('delete', `/api/organizations/${id}/invitations/${invitationId}`)
export const getOrganizationTokens = (id: number) =>
  getData<OrganizationToken[]>(`/api/organizations/${id}/tokens`)
export const createOrganizationToken = (id: number, name: string) =>
  mutate<OrganizationToken>('post', `/api/organizations/${id}/tokens`, { name })
export const updateOrganizationToken = (
  id: number,
  tokenId: number,
  status: number
) => mutate('patch', `/api/organizations/${id}/tokens/${tokenId}`, { status })
export const deleteOrganizationToken = (id: number, tokenId: number) =>
  mutate('delete', `/api/organizations/${id}/tokens/${tokenId}`)
export const revealOrganizationToken = (id: number, tokenId: number) =>
  mutate<{ key: string }>(
    'post',
    `/api/organizations/${id}/tokens/${tokenId}/key`
  )
export const transferToOrganization = (id: number, quota: number) =>
  mutate('post', `/api/organizations/${id}/wallet/transfer`, { quota })
export const getOrganizationWalletTransactions = (id: number) =>
  getData<OrganizationWalletTransaction[]>(
    `/api/organizations/${id}/wallet/transactions`
  )
export const getOrganizationUsage = (id: number) =>
  getData<OrganizationUsage>(`/api/organizations/${id}/usage`)
export const getOrganizationLogs = (id: number, page: number) =>
  getData<OrganizationLogPage>(
    `/api/organizations/${id}/logs?p=${page}&page_size=20`
  )
