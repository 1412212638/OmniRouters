import assert from 'node:assert/strict'
import { describe, test } from 'node:test'

import type { Channel } from '../../types'
import { getChannelTableRowId, type TagRow } from '../channel-utils'

describe('channel table row identity', () => {
  test('uses stable, separate channel and tag namespaces', () => {
    const channel = { id: 202 } as Channel
    const tag = {
      id: '202' as unknown as number,
      tag: '202',
      children: [channel],
    } as TagRow
    assert.equal(getChannelTableRowId(channel), 'channel:202')
    assert.equal(getChannelTableRowId(tag), 'tag:202')
  })
})
