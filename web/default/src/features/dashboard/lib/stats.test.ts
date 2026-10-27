import assert from 'node:assert/strict'
import { describe, test } from 'node:test'

import type { QuotaDataItem } from '../types'
import { calculateDashboardStats } from './stats'

describe('dashboard statistics', () => {
  test('aggregates cache tokens without adding them to total tokens', () => {
    const data: QuotaDataItem[] = [
      {
        created_at: 3600,
        token_used: 100,
        cache_read_tokens: 80,
        cache_creation_tokens: 10,
      },
      {
        created_at: 7200,
        token_used: 50,
        cache_read_tokens: 20,
        cache_creation_tokens: 5,
      },
    ]

    assert.deepEqual(calculateDashboardStats(data), {
      totalQuota: 0,
      totalCount: 0,
      totalTokens: 150,
      totalCacheReadTokens: 100,
      totalCacheCreationTokens: 15,
    })
  })

  test('ignores invalid and negative cache totals', () => {
    const data: QuotaDataItem[] = [
      {
        created_at: 3600,
        cache_read_tokens: -2,
        cache_creation_tokens: Number.NaN,
      },
    ]

    const stats = calculateDashboardStats(data)
    assert.equal(stats.totalCacheReadTokens, 0)
    assert.equal(stats.totalCacheCreationTokens, 0)
  })
})
