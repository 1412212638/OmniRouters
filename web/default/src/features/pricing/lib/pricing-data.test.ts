import assert from 'node:assert/strict'
import { describe, test } from 'node:test'

import type { PricingModel } from '../types'
import { getPricingQueryKey, resolveModelGroupRatios } from './pricing-data'

const model: PricingModel = {
  id: 1,
  model_name: 'example-model',
  quota_type: 0,
  model_ratio: 1,
  completion_ratio: 1,
  enable_groups: ['default'],
}

describe('pricing data resolution', () => {
  test('uses the backend-resolved user multiplier instead of multiplying it again', () => {
    const ratios = resolveModelGroupRatios(
      { ...model, model_group_ratio: { default: 0.6 } },
      { default: 0.8 },
      { default: { 'example-model': 0.6 } }
    )

    assert.equal(ratios.default, 0.6)
  })

  test('falls back to generic and model-specific ratios for older responses', () => {
    const ratios = resolveModelGroupRatios(
      model,
      { default: 0.8 },
      { default: { 'example-model': 0.6 } }
    )

    assert.equal(ratios.default, 0.48)
  })

  test('isolates pricing query caches by user and workspace', () => {
    assert.notDeepEqual(
      getPricingQueryKey(7, null),
      getPricingQueryKey(8, null)
    )
    assert.notDeepEqual(getPricingQueryKey(7, null), getPricingQueryKey(7, 42))
  })
})
