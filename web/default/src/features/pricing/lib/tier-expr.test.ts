import assert from 'node:assert/strict'
import { test } from 'node:test'

import { BILLING_CACHE_VAR_MAP } from './billing-expr'
import { compileBillingExpression } from './billing-expression/parser'
import {
  createDefaultVisualConfig,
  evalExprLocally,
  exprUsesExtraVars,
  generateExprFromVisualConfig,
  normalizeVisualConfig,
  normalizeVisualTier,
  tryParseVisualConfig,
} from './tier-expr'

test('absent categories stay absent, explicit zero categories survive editing and reload', () => {
  const original = 'tier("base", p * 5 + c * 30 + cr * 1.25 + img * 8)'
  const parsed = tryParseVisualConfig(original)
  assert.ok(parsed)
  assert.equal(
    generateExprFromVisualConfig(normalizeVisualConfig(parsed)),
    original
  )
  assert.equal(parsed.tiers[0].image_output_unit_cost, undefined)
  for (const variable of BILLING_CACHE_VAR_MAP) {
    const config = createDefaultVisualConfig()
    config.tiers[0] = normalizeVisualTier({
      input_unit_cost: 5,
      output_unit_cost: 30,
      label: 'base',
      [variable.field]: 0,
    })
    const expression = generateExprFromVisualConfig(config)
    const compiled = compileBillingExpression(expression)
    assert.equal(compiled.status, 'ready')
    if (compiled.status === 'ready') {
      assert.ok(new Set<string>(compiled.variables).has(variable.exprVar))
    }
    const restored = tryParseVisualConfig(expression)
    assert.ok(restored)
    assert.equal(restored.tiers[0][variable.field], 0)
    assert.equal(generateExprFromVisualConfig(restored), expression)
    delete restored.tiers[0][variable.field]
    assert.equal(
      generateExprFromVisualConfig(restored),
      'tier("base", p * 5 + c * 30)'
    )
  }
})

test('the linked cost estimator recognizes image-cache pricing', () => {
  const expression = 'tier("base", p * 5 + c * 30 + img_cr * 2)'
  assert.ok(exprUsesExtraVars(expression))
  const result = evalExprLocally(expression, 0, 0, {
    cacheReadTokens: 0,
    imageCacheTokens: 200,
    cacheCreateTokens: 0,
    cacheCreate1hTokens: 0,
    cacheCreate12hTokens: 0,
    imageTokens: 0,
    imageOutputTokens: 0,
    audioInputTokens: 0,
    audioOutputTokens: 0,
  })
  assert.equal(result.error, null)
  assert.equal(result.cost, 400)
})

test('existing tier conditions, cache TTL prices, fixed pricing and hidden fields round-trip', () => {
  for (const expression of [
    'tier("base", p * 5 + c * 30 + cr * 1.25 + img_cr * 2 + cc12h * 0 + img * 8)',
    'len <= 200000 ? tier("short", p * 3 + c * 15 + cr * 0.3 + cc * 3.75 + cc1h * 0) : tier("long", p * 6 + c * 30 + cr * 0.6 + cc * 7.5 + cc1h * 12)',
    'tier("request", fixed(0.2))',
  ]) {
    const parsed = tryParseVisualConfig(expression)
    assert.ok(parsed, expression)
    assert.equal(generateExprFromVisualConfig(parsed), expression)
  }
  const complex = 'tier("custom", p * 5 + max(c - 1, 0) * 2)'
  assert.equal(tryParseVisualConfig(complex), null)
})
