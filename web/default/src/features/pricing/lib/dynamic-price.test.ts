import assert from 'node:assert/strict'
import { describe, test } from 'node:test'

import { parseTiersFromExpr } from './billing-expr'
import { getDynamicPriceEntries } from './dynamic-price'

describe('dynamic fixed request prices', () => {
  test('displays fixed prices as USD per request without token scaling', () => {
    const [tier] = parseTiersFromExpr('tier("request", fixed(0.005))')

    const [entry] = getDynamicPriceEntries(tier, {
      tokenUnit: 'M',
    })

    assert.equal(entry?.displayUnit, 'request')
    assert.equal(entry?.value, 0.005)
    assert.equal(entry?.formatted, '$0.005')
  })
})
