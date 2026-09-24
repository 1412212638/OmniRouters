export type MemberTierRule = {
  enabled: boolean
  group: string
  display_name: string
  description?: string
  min_topup_quota: number
  min_used_quota: number
}

export type MemberTierProgress = {
  rule: MemberTierRule
  qualified: boolean
  current: boolean
  next: boolean
  topup_remaining: number
  used_remaining: number
  group_ratio: number
  topup_group_ratio: number
  models: string[]
  model_count: number
}

export type MemberTierEvaluation = {
  enabled: boolean
  current_group: string
  target_group: string
  total_topup_quota: number
  used_quota: number
  current_tier?: MemberTierRule
  target_tier?: MemberTierRule
  next_tier?: MemberTierRule
  progress: MemberTierProgress[]
  upgraded: boolean
  skipped_by_active_subscription: boolean
}

export type MemberTierSelfResponse = {
  success: boolean
  message?: string
  data?: MemberTierEvaluation
}
