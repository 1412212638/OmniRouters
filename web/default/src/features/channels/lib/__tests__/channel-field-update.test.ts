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

import {
  CHANNEL_FIELD_UPDATE_DELAY_MS,
  createChannelFieldUpdateScheduler,
} from '../channel-field-update'

function createFakeTimers() {
  const pending = new Map<number, () => void>()
  let nextId = 1
  return {
    timers: {
      setTimeout: (callback: () => void, delay: number) => {
        assert.equal(delay, CHANNEL_FIELD_UPDATE_DELAY_MS)
        const id = nextId++
        pending.set(id, callback)
        return id
      },
      clearTimeout: (id: number) => pending.delete(id),
    },
    fireAll() {
      const callbacks = [...pending.values()]
      pending.clear()
      callbacks.forEach((callback) => callback())
    },
    pending,
  }
}

describe('channel field update scheduler', () => {
  test('coalesces rapid values and preserves zero', () => {
    const fake = createFakeTimers()
    const updates: number[] = []
    const scheduler = createChannelFieldUpdateScheduler(
      (value) => updates.push(value),
      fake.timers
    )
    scheduler.schedule(3)
    scheduler.schedule(0)
    assert.equal(fake.pending.size, 1)
    fake.fireAll()
    assert.deepEqual(updates, [0])
  })

  test('flush commits once and cancels the timer', () => {
    const fake = createFakeTimers()
    const updates: number[] = []
    const scheduler = createChannelFieldUpdateScheduler(
      (value) => updates.push(value),
      fake.timers
    )
    scheduler.schedule(7)
    scheduler.flush()
    scheduler.flush()
    fake.fireAll()
    assert.deepEqual(updates, [7])
  })
})
