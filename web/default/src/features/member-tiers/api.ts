import { api } from '@/lib/api'

import type { MemberTierSelfResponse } from './types'

export async function getMemberTierSelf() {
  const res = await api.get<MemberTierSelfResponse>('/api/member_tiers/self')
  return res.data
}
