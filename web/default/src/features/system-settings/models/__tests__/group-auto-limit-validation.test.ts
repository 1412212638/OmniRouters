import assert from 'node:assert/strict'
import { describe, test } from 'node:test'

import { positiveIntegerSchema } from '../../utils/numeric-field'

const t = (key: string) => key
const schema = positiveIntegerSchema(t('Enter a positive integer'))

describe('per-token Auto group limit validation', () => {
  test('accepts any positive integer without a product upper bound', () => {
    assert.equal(schema.safeParse(1000).success, true)
  })

  test('rejects zero, negative, and fractional limits', () => {
    for (const maxTokenAutoGroups of [0, -1, 1.5]) {
      const result = schema.safeParse(maxTokenAutoGroups)
      assert.equal(result.success, false)
      if (result.success) continue
      assert.equal(result.error.issues[0]?.message, 'Enter a positive integer')
    }
  })
})
