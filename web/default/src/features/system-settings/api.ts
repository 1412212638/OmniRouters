import { api } from '@/lib/api'

import type {
  FetchUpstreamRatiosRequest,
  LogCleanupTask,
  MarketingDigest,
  MarketingDigestPromotion,
  MarketingDigestRecipient,
  SearchMarketingEmailUsersParams,
  SearchMarketingEmailUsersResponse,
  SendMarketingEmailRequest,
  SendMarketingEmailResponse,
  SystemOptionsResponse,
  SystemTaskListResponse,
  SystemTaskResponse,
  UpdateOptionRequest,
  UpdateOptionResponse,
  UpstreamChannelsResponse,
  UpstreamRatiosResponse,
} from './types'

export type ModelPricingSnapshotResponse = {
  success: boolean
  message: string
  data: {
    entries: unknown[]
    options: Record<string, string>
    empty_version: string
  }
}

export async function getModelPricingSnapshot(models?: string[]) {
  const res = await api.get<ModelPricingSnapshotResponse>(
    '/api/option/model_pricing',
    { params: models?.length ? { model: models } : undefined }
  )
  return res.data
}

export async function previewModelPricingConversion(
  modelName: string,
  pricing: Record<string, unknown>
) {
  const res = await api.post('/api/option/model_pricing/convert', {
    model_name: modelName,
    pricing,
  })
  return res.data
}

export async function getSystemOptions() {
  const res = await api.get<SystemOptionsResponse>('/api/option/')
  return res.data
}

export async function updateSystemOption(request: UpdateOptionRequest) {
  const res = await api.put<UpdateOptionResponse>('/api/option/', request)
  return res.data
}

export async function getEmailSettingsOptions() {
  const res = await api.get<SystemOptionsResponse>(
    '/api/email_settings/options'
  )
  return res.data
}

export async function updateEmailSettingOption(request: UpdateOptionRequest) {
  const res = await api.put<UpdateOptionResponse>(
    '/api/email_settings/option',
    request
  )
  return res.data
}

export async function startLogCleanupTask(targetTimestamp: number) {
  const res = await api.post<SystemTaskResponse<LogCleanupTask>>(
    '/api/system-task/log-cleanup',
    null,
    {
      params: { target_timestamp: targetTimestamp },
    }
  )
  return res.data
}

export async function getCurrentLogCleanupTask() {
  const res = await api.get<SystemTaskResponse<LogCleanupTask | null>>(
    '/api/system-task/current',
    {
      params: { type: 'log_cleanup' },
    }
  )
  return res.data
}

export async function getSystemTask(taskId: string) {
  const res = await api.get<SystemTaskResponse<LogCleanupTask>>(
    `/api/system-task/${taskId}`
  )
  return res.data
}

export async function searchMarketingEmailUsers(
  params: SearchMarketingEmailUsersParams
) {
  const res = await api.get<SearchMarketingEmailUsersResponse>(
    '/api/user/search',
    { params }
  )
  return res.data
}

export async function sendMarketingEmail(request: SendMarketingEmailRequest) {
  const res = await api.post<SendMarketingEmailResponse>(
    '/api/option/marketing_email/send',
    request
  )
  return res.data
}

export async function sendEmailSettingsMarketingEmail(
  request: SendMarketingEmailRequest
) {
  const res = await api.post<SendMarketingEmailResponse>(
    '/api/email_settings/marketing_email/send',
    request
  )
  return res.data
}

export async function listMarketingDigestPromotions() {
  const res = await api.get<{
    success: boolean
    data?: MarketingDigestPromotion[]
  }>('/api/email_settings/newsletter/promotions')
  return res.data
}

export async function saveMarketingDigestPromotion(
  promotion: Omit<MarketingDigestPromotion, 'id'> & { id?: number }
) {
  const res = await api.request<{
    success: boolean
    data?: MarketingDigestPromotion
  }>({
    method: promotion.id ? 'PUT' : 'POST',
    url: '/api/email_settings/newsletter/promotions',
    data: promotion,
  })
  return res.data
}

export async function deleteMarketingDigestPromotion(id: number) {
  const res = await api.delete(
    `/api/email_settings/newsletter/promotions/${id}`
  )
  return res.data
}

export async function listMarketingDigests() {
  const res = await api.get<{ success: boolean; data?: MarketingDigest[] }>(
    '/api/email_settings/newsletter/digests'
  )
  return res.data
}

export async function generateMarketingDigest(periodDays: number) {
  const res = await api.post<{
    success: boolean
    message?: string
    data?: { digest: MarketingDigest; created: boolean }
  }>('/api/email_settings/newsletter/digests/generate', {
    period_days: periodDays,
  })
  return res.data
}

export async function updateMarketingDigest(
  id: number,
  request: {
    subject: string
    content: string
    audience_type: MarketingDigest['audience_type']
    audience_group: string
    user_ids: number[]
  }
) {
  const res = await api.put<{ success: boolean; data?: MarketingDigest }>(
    `/api/email_settings/newsletter/digests/${id}`,
    request
  )
  return res.data
}

export async function sendMarketingDigest(id: number, retry = false) {
  const action = retry ? 'retry' : 'send'
  const res = await api.post(
    `/api/email_settings/newsletter/digests/${id}/${action}`
  )
  return res.data
}

export async function listMarketingDigestRecipients(
  id: number,
  status?: MarketingDigestRecipient['status']
) {
  const res = await api.get<{
    success: boolean
    data?: { items: MarketingDigestRecipient[]; total: number }
  }>(`/api/email_settings/newsletter/digests/${id}/recipients`, {
    params: { status, limit: 100, offset: 0 },
  })
  return res.data
}

export async function listSystemTasks(limit = 20) {
  const res = await api.get<SystemTaskListResponse>('/api/system-task/list', {
    params: { limit },
  })
  return res.data
}

export async function resetModelRatios() {
  const res = await api.post<UpdateOptionResponse>(
    '/api/option/rest_model_ratio'
  )
  return res.data
}

export async function getUpstreamChannels() {
  const res = await api.get<UpstreamChannelsResponse>(
    '/api/ratio_sync/channels'
  )
  return res.data
}

export async function fetchUpstreamRatios(request: FetchUpstreamRatiosRequest) {
  const res = await api.post<UpstreamRatiosResponse>(
    '/api/ratio_sync/fetch',
    request
  )
  return res.data
}

export type MemberTierRecalculateResponse = {
  success: boolean
  message?: string
  data?: {
    scanned: number
    upgraded: number
    skipped_by_active_subscription: number
  }
}

export async function recalculateMemberTiers() {
  const res = await api.post<MemberTierRecalculateResponse>(
    '/api/member_tiers/admin/recalculate'
  )
  return res.data
}
