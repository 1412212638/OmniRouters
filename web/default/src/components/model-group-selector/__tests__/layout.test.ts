import assert from 'node:assert/strict'
import { describe, test } from 'node:test'

import {
  modelGroupSelectorLayoutClasses,
  scrollSelectedOptionIntoView,
} from '../../model-group-selector-layout'

describe('model group selector layout', () => {
  test('keeps group options at a fixed height and aligned to the top', () => {
    const classes = modelGroupSelectorLayoutClasses.groupScroll.split(' ')
    assert.ok(classes.includes('auto-rows-[2rem]'))
    assert.ok(classes.includes('content-start'))
  })

  test('centers the selected group inside its scroll container', () => {
    const calls: ScrollToOptions[] = []
    const selectedOption = {
      offsetHeight: 32,
      offsetTop: 160,
      scrollIntoView() {},
    }
    const scrollContainer = {
      clientHeight: 200,
      scrollTop: 0,
      scrollTo(options: ScrollToOptions) {
        calls.push(options)
      },
    }

    scrollSelectedOptionIntoView(selectedOption, scrollContainer)
    assert.deepEqual(calls, [{ top: 76, behavior: 'auto' }])
  })

  test('falls back to scrollIntoView without a group container', () => {
    const calls: ScrollIntoViewOptions[] = []
    const selectedOption = {
      scrollIntoView(options?: ScrollIntoViewOptions) {
        calls.push(options ?? {})
      },
    }

    scrollSelectedOptionIntoView(selectedOption)
    assert.deepEqual(calls, [{ block: 'center', inline: 'nearest' }])
  })
})
