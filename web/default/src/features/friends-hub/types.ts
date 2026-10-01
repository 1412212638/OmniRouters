export interface FriendsSummary {
  aff_code?: string
  invite_code?: string
  invite_link?: string
  referral_rate?: number
  cashback_rate?: number
  invite_count?: number
  first_topup_count?: number
  first_top_up_invite_count?: number
  registration_pending_quota?: number
  registration_total_quota?: number
  cashback_pending_quota?: number
  cashback_total_quota?: number
  wallet_quota?: number
  [key: string]: unknown
}

export interface FriendRecord {
  id?: number | string
  user_id?: number | string
  username?: string
  display_name?: string
  email?: string
  created_at?: number | string
  register_time?: number | string
  first_topup_status?: string
  first_top_up_status?: string
  first_topup_quota?: number
  first_top_up_base_quota?: number
  registration_reward_quota?: number
  cashback_quota?: number
  cashback_rate_bps?: number
  cashback_status?: string
  reward_status?: string
  [key: string]: unknown
}

export interface RewardRecord {
  id?: number | string
  reward_type?: string
  invitee_id?: number | string
  invitee_name?: string
  source_trade_no?: string
  base_quota?: number
  rate_bps?: number
  reward_rate_bps?: number
  rate?: number
  reward_quota?: number
  status?: string
  created_at?: number | string
  transferred_at?: number | string
  [key: string]: unknown
}

export interface FriendsListResponse {
  items?: FriendRecord[]
  total?: number
  list?: FriendRecord[]
}

export interface RewardsListResponse {
  items?: RewardRecord[]
  total?: number
  list?: RewardRecord[]
}
