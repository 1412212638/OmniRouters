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
  DeleteLogsResponse,
  FetchUpstreamRatiosRequest,
  SearchMarketingEmailUsersParams,
  SearchMarketingEmailUsersResponse,
  SendMarketingEmailRequest,
  SendMarketingEmailResponse,
  SystemOptionsResponse,
  UpdateOptionRequest,
  UpdateOptionResponse,
  UpstreamChannelsResponse,
  UpstreamRatiosResponse,
} from './types'

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

export async function deleteLogsBefore(targetTimestamp: number) {
  const res = await api.delete<DeleteLogsResponse>('/api/log/', {
    params: { target_timestamp: targetTimestamp },
  })
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
