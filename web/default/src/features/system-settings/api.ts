import { api } from '@/lib/api'

import type {
  FetchUpstreamRatiosRequest,
  LogCleanupTask,
  MarketingEmailSendTask,
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

export async function syncWeChatOfficialMenu() {
  const res = await api.post<{ success: boolean; message?: string }>(
    '/api/oauth/wechat/official/menu/sync'
  )
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

export async function getMarketingEmailAudienceCount(params: {
  audience_type: 'all' | 'topup' | 'active' | 'group'
  audience_group?: string
}) {
  const res = await api.get<{
    success: boolean
    message?: string
    data?: { total: number }
  }>('/api/email_settings/marketing_email/audience', { params })
  return res.data
}

export async function getMarketingEmailSendTask(taskId: string) {
  const res = await api.get<SystemTaskResponse<MarketingEmailSendTask>>(
    `/api/email_settings/marketing_email/tasks/${encodeURIComponent(taskId)}`
  )
  return res.data
}

export type MarketingDeliverySettings = {
  mode: 'inherit' | 'smtp'
  server: string
  port: number
  account: string
  password?: string
  from: string
  display_name: string
  reply_to: string
  per_second: number
}

type MarketingDeliveryResponse = {
  success: boolean
  message?: string
  data?: {
    settings: MarketingDeliverySettings
    password_configured: boolean
    callback_url: string
  }
}

const marketingDeliveryPath = '/api/email_settings/marketing_email/delivery'

export async function getMarketingDeliverySettings() {
  return (await api.get<MarketingDeliveryResponse>(marketingDeliveryPath)).data
}

export async function saveMarketingDeliverySettings(settings: MarketingDeliverySettings) {
  return (await api.put<MarketingDeliveryResponse>(marketingDeliveryPath, settings)).data
}

export async function testMarketingDelivery(recipient: string) {
  return (await api.post<UpdateOptionResponse>(`${marketingDeliveryPath}/test`, { recipient })).data
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
