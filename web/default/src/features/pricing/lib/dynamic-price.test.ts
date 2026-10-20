import assert from 'node:assert/strict'
import { describe, test } from 'node:test'

import { parseTiersFromExpr } from './billing-expr'
import {
  getDynamicPriceEntries,
  getDynamicPricingSummary,
} from './dynamic-price'
import { getDisplayedModelName } from './model-helpers'

describe('dynamic fixed request prices', () => {
  test('displays fixed prices as USD per request without token scaling', () => {
    const [tier] = parseTiersFromExpr('tier("request", fixed(0.005))')

    const [entry] = getDynamicPriceEntries(tier, {
      tokenUnit: 'M',
    })

    assert.equal(entry?.displayUnit, 'request')
    assert.equal(entry?.value, 0.005)
    assert.equal(entry?.formatted, '$0.005')

    const model = {
      billing_mode: 'tiered_expr',
      billing_expr: 'tier("request", fixed(0.005))',
    }
    const summary = getDynamicPricingSummary(model, { tokenUnit: 'M' })

    assert.equal(summary?.requestPriceRange, '$0.005')
  })
})

describe('displayed model names', () => {
  test('uses the developer instead of the channel supplier as the prefix', () => {
    assert.equal(
      getDisplayedModelName({
        model_name: 'minimax-m3.1-flash',
        developer_name: 'MiniMax',
        vendor_name: 'TencentCloud',
      } as Parameters<typeof getDisplayedModelName>[0]),
      'minimax/minimax-m3.1-flash'
    )
  })

  test('keeps supplier prefix behavior when no developer is configured', () => {
    assert.equal(
      getDisplayedModelName({
        model_name: 'gpt-4.1',
        vendor_name: 'OpenAI',
      } as Parameters<typeof getDisplayedModelName>[0]),
      'openai/gpt-4.1'
    )
  })
})
