/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
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
