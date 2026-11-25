import { isAxiosError } from 'axios'

import { api } from '@/lib/api'

import type {
  BillingExport,
  BillingExportFilters,
  BillingExportLevel,
  BillingExportResponse,
} from './types'

const billingExportRequestConfig = {
  skipBusinessError: true,
  skipErrorHandler: true,
}

export function billingExportErrorKey(error: unknown, fallback: string) {
  const message: unknown = isAxiosError(error)
    ? error.response?.data?.message
    : ((error as Error | null)?.message ?? error)
  switch (message) {
    case 'already_queued':
      return 'Wait for your current billing export to finish.'
    case 'manager_required':
      return 'Only organization owners and managers can export this bill.'
    case 'invalid_period':
      return 'Select a billing period of up to 12 months.'
    case 'invalid_language':
      return 'Select Simplified Chinese or English.'
    case 'limit_exceeded':
      return 'This export is too large. Select a shorter billing period.'
    default:
      return fallback
  }
}

export async function listBillingExports(
  organizationId?: number | null,
  userId?: number | null
): Promise<BillingExport[]> {
  const response = await api.get<BillingExportResponse>(
    '/api/billing-exports/',
    {
      ...billingExportRequestConfig,
      params:
        organizationId || userId
          ? {
              ...(organizationId ? { organization_id: organizationId } : {}),
              ...(userId ? { user_id: userId } : {}),
            }
          : undefined,
    }
  )
  if (!response.data.success) {
    throw new Error(response.data.message || 'Failed to load billing exports')
  }
  return response.data.data || []
}

export async function createBillingExport({
  startMonth,
  endMonth,
  language,
  organizationId,
  userId,
}: BillingExportFilters) {
  const response = await api.post(
    '/api/billing-exports/',
    {
      start_month: startMonth,
      end_month: endMonth,
      language,
      ...(organizationId ? { organization_id: organizationId } : {}),
      ...(userId ? { user_id: userId } : {}),
    },
    billingExportRequestConfig
  )
  if (!response.data.success) {
    throw new Error(response.data.message || 'Failed to create billing export')
  }
}

export async function downloadBillingExport(
  taskId: string,
  level: BillingExportLevel,
  organizationId?: number | null,
  userId?: number | null
) {
  const response = await api.get<Blob>(
    `/api/billing-exports/${encodeURIComponent(taskId)}/${level}`,
    {
      ...billingExportRequestConfig,
      responseType: 'blob',
      params:
        organizationId || userId
          ? {
              ...(organizationId ? { organization_id: organizationId } : {}),
              ...(userId ? { user_id: userId } : {}),
            }
          : undefined,
    }
  )
  return response.data
}
