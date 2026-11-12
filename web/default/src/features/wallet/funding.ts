import { api } from '@/lib/api'

export const fundingKinds = [
  'paid',
  'gift',
  'adjustment',
  'unclassified',
] as const
export type FundingKind = (typeof fundingKinds)[number]

export interface FundingSummary {
  paid: string
  gift: string
  adjustment: string
  unclassified: string
}

export interface FundingEntry {
  id: number
  amount: string
  kind: FundingKind
  source: string
  reference: string
  operator_id: number
  note: string
  created_at: number
}

export interface FundingLedger {
  items: FundingEntry[]
  total: number
  summary: FundingSummary
}

export async function getFundingSummary(): Promise<FundingSummary> {
  const { data } = await api.get('/api/user/self/wallet/funding')
  if (!data.success) throw new Error(data.message || 'Failed to load')
  return data.data
}

export async function getFundingLedger(
  userId: number,
  page: number,
  kind: string
): Promise<FundingLedger> {
  const { data } = await api.get(`/api/user/${userId}/wallet/funding`, {
    params: { p: page, page_size: 20, kind },
  })
  if (!data.success) throw new Error(data.message || 'Failed to load')
  return data.data
}
