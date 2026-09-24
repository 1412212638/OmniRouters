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
