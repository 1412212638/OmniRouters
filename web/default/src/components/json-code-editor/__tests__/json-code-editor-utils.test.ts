import assert from 'node:assert/strict'
import { describe, test } from 'node:test'

import {
  applyJsonSmartEnter,
  createScrollLayerSynchronizer,
  formatJsonDraft,
  getCursorLocation,
  getJsonValidationState,
} from '../json-code-editor-utils'

describe('json code editor utils', () => {
  test('validates and formats drafts without discarding invalid input', () => {
    assert.equal(getJsonValidationState('{"model": }').isValid, false)
    assert.deepEqual(formatJsonDraft('{"model": }'), {
      didFormat: false,
      value: '{"model": }',
    })
    assert.equal(formatJsonDraft('{"model":1}').value, '{\n  "model": 1\n}')
  })

  test('reports cursor position and expands paired braces', () => {
    assert.deepEqual(getCursorLocation('{\n  "model": 1\n}', 5), {
      line: 2,
      column: 4,
    })
    assert.deepEqual(applyJsonSmartEnter('{}', 1, 1), {
      value: '{\n  \n}',
      selectionStart: 4,
      selectionEnd: 4,
    })
  })

  test('coalesces scroll layer updates', () => {
    const source = { scrollLeft: 12, scrollTop: 40 }
    const contentLayer = { style: { transform: '' } }
    const lineNumberLayer = { style: { transform: '' } }
    const frames: Array<() => void> = []
    const synchronizer = createScrollLayerSynchronizer(
      source,
      { contentLayer, lineNumberLayer },
      (callback) => {
        frames.push(callback)
        return frames.length
      }
    )
    synchronizer.sync()
    source.scrollLeft = 24
    source.scrollTop = 80
    synchronizer.sync()
    assert.equal(frames.length, 1)
    frames[0]()
    assert.equal(contentLayer.style.transform, 'translate3d(-24px, -80px, 0)')
  })
})
