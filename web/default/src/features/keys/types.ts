import { z } from 'zod'

// ============================================================================
// API Key Schema & Types
// ============================================================================

const apiKeySubscriptionSchema = z.object({
  id: z.number(),
  plan_id: z.number(),
  plan_title: z.string().default(''),
  billing_mode: z.string().default('quota'),
  status: z.string(),
  start_time: z.number(),
  end_time: z.number(),
  amount_total: z.number(),
  amount_used: z.number(),
  amount_remaining: z.number(),
  next_reset_time: z.number(),
})

export const apiKeySchema = z.object({
  id: z.number(),
  name: z.string(),
  key: z.string(),
  status: z.number(), // 1: enabled, 2: disabled, 3: expired, 4: exhausted
  remain_quota: z.number(),
  used_quota: z.number(),
  unlimited_quota: z.boolean(),
  expired_time: z.number(), // -1 for never expires
  created_time: z.number(),
  accessed_time: z.number(),
  group: z.string().nullish().default(''),
  auto_groups: z.array(z.string()).nullish().default(null),
  cross_group_retry: z
    .preprocess((v) => {
      if (v === 1) return true
      if (v === 0) return false
      return v
    }, z.boolean())
    .optional()
    .default(false),
  auto_route_strategy: z.string().optional().default(''),
  model_limits_enabled: z.boolean(),
  model_limits: z.string().nullish().default(''),
  allow_ips: z.string().nullish().default(''),
  allow_user_quota_query: z.boolean().optional().default(false),
  key_type: z.string().optional().default('general'),
  user_subscription_id: z.number().optional().default(0),
  subscription: apiKeySubscriptionSchema.nullish(),
  effective_status: z.string().optional().default(''),
})

export type ApiKey = z.infer<typeof apiKeySchema>

// ============================================================================
// API Request/Response Types
// ============================================================================

export interface ApiResponse<T = unknown> {
  success: boolean
  message?: string
  data?: T
}

export interface GetApiKeysParams {
  p?: number
  size?: number
}

export interface GetApiKeysResponse {
  success: boolean
  message?: string
  data?: {
    items: ApiKey[]
    total: number
    page: number
    page_size: number
  }
}

export interface SearchApiKeysParams {
  keyword?: string
  token?: string
  p?: number
  size?: number
}

export interface ApiKeyFormData {
  name: string
  remain_quota: number
  expired_time: number
  unlimited_quota: boolean
  model_limits_enabled: boolean
  model_limits: string
  allow_ips: string
  group: string
  auto_groups: string[]
  cross_group_retry: boolean
  auto_route_strategy: string
  allow_user_quota_query: boolean
}

export interface TokenAutoGroupsConfig {
  groups: string[]
  max_count: number
}

// ============================================================================
// Dialog Types
// ============================================================================

export type ApiKeysDialogType =
  | 'create'
  | 'update'
  | 'delete'
  | 'batch-delete'
  | 'cc-switch'
  | 'endpoints'
