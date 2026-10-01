import { api } from '@/lib/api'

import type {
  FriendsListResponse,
  FriendsSummary,
  RewardsListResponse,
} from './types'

type ApiResponse<T = unknown> = { success?: boolean; message?: string; data?: T }

export async function getFriendsSummary(): Promise<ApiResponse<FriendsSummary>> {
  const response = await api.get('/api/user/friends/summary')
  return response.data
}

export async function getFriends(
  page: number,
  pageSize: number
): Promise<ApiResponse<FriendsListResponse>> {
  const response = await api.get('/api/user/friends', {
    params: { p: page, page_size: pageSize },
  })
  return response.data
}

export async function getFriendRewards(
  page: number,
  pageSize: number
): Promise<ApiResponse<RewardsListResponse>> {
  const response = await api.get('/api/user/friends/rewards', {
    params: { p: page, page_size: pageSize },
  })
  return response.data
}

export async function transferFriendRewards(amount?: number) {
  const response = await api.post('/api/user/friends/cashback/transfer',
    amount == null ? undefined : { quota: amount }
  )
  return response.data as ApiResponse
}
