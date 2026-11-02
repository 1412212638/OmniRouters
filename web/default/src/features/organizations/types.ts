export type OrganizationRole = 'owner' | 'admin' | 'member'

export interface Organization {
  id: number
  name: string
  owner_user_id: number
  group: string
  quota?: number
  used_quota?: number
  request_count?: number
  created_at: number
  updated_at: number
}

export interface OrganizationMembership {
  organization_id: number
  user_id: number
  role: OrganizationRole
  status: string
  monthly_quota_limit: number
}

export interface OrganizationWorkspace {
  organization: Organization
  membership: OrganizationMembership
  monthly_used_quota: number
  period: string
  member_count?: number
}

export interface OrganizationMember extends OrganizationMembership {
  username: string
  display_name: string
  email: string
  monthly_used_quota: number
}

export interface OrganizationInvitation {
  id: number
  organization_id: number
  organization_name?: string
  email: string
  role: OrganizationRole
  status: string
  expires_at: number
  created_at: number
}

export interface OrganizationToken {
  id: number
  user_id: number
  organization_id: number
  name: string
  key: string
  status: number
  used_quota: number
  created_time: number
  accessed_time: number
  expired_time: number
  creator_name?: string
}

export interface OrganizationWalletTransaction {
  id: number
  user_id: number
  amount: number
  balance_after: number
  type: string
  description: string
  created_at: number
  user_name?: string
}

export interface OrganizationUsage {
  used_quota?: number
  request_count?: number
  monthly_used_quota?: number
  monthly_quota_limit?: number
  members?: Array<{ used_quota: number }>
  total_tokens?: number
}

export interface OrganizationApiResponse<T = unknown> {
  success: boolean
  message?: string
  data?: T
}

export interface OrganizationLogPage {
  items: Array<{
    id: number
    created_at: number
    username: string
    token_name: string
    model_name: string
    quota: number
    prompt_tokens: number
    completion_tokens: number
  }>
  total: number
}
