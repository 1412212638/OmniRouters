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
