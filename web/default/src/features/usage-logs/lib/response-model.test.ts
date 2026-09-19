import assert from 'node:assert/strict'
import { test } from 'node:test'

import { hasResponseModelMismatch } from './response-model'

test('response model comparison accepts aliases, versions and provider paths', () => {
  assert.equal(hasResponseModelMismatch(), false)
  for (const returned of [
    '',
    'public',
    'GPT-TEST',
    'gpt-test-2026',
    'provider/gpt-test',
  ]) {
    assert.equal(
      hasResponseModelMismatch({
        requested_model: 'public',
        upstream_model: 'gpt-test',
        returned_model: returned,
      }),
      false
    )
  }
  assert.equal(
    hasResponseModelMismatch({
      requested_model: 'public',
      upstream_model: 'gpt-test',
      returned_model: 'other-model',
    }),
    true
  )
  assert.equal(
    hasResponseModelMismatch({
      requested_model: '',
      upstream_model: '',
      returned_model: 'other-model',
    }),
    true
  )
})
