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
