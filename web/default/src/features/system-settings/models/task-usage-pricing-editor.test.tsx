import assert from 'node:assert/strict'
import { after, test } from 'node:test'

import { Window } from 'happy-dom'

import type { BillingUsageSchema } from '@/features/pricing/types'

const dom = new Window({ url: 'http://localhost' })
for (const key of [
  'window',
  'document',
  'navigator',
  'HTMLElement',
  'Element',
  'Event',
  'MouseEvent',
  'Node',
  'MutationObserver',
  'ResizeObserver',
  'getComputedStyle',
  'requestAnimationFrame',
  'cancelAnimationFrame',
] as const) {
  Object.defineProperty(globalThis, key, {
    configurable: true,
    value: key === 'window' ? dom : dom[key],
  })
}
Object.assign(globalThis, { IS_REACT_ACT_ENVIRONMENT: true })
const { act, useState } = await import('react')
const { createRoot } = await import('react-dom/client')
const { createInstance } = await import('i18next')
const { I18nextProvider } = await import('react-i18next')
const { TaskUsagePricingEditor } = await import('./task-usage-pricing-editor')
const {
  tryParseTaskMatrixConfig,
  evaluateTaskVisualConfig,
  taskMatrixToTiers,
} = await import('@/features/pricing/lib/task-expr')
const i18n = createInstance()
await i18n.init({ lng: 'en', resources: {} })

after(() => dom.close())

test('task prices retain typed decimals through expression updates, blur and reopening', async () => {
  const schema: BillingUsageSchema = {
    resolution: { enum: ['720p', '1080p'] },
    billable_tokens: { type: 'number', unit: 'token' },
    seconds: { type: 'number', unit: 'second' },
    images: { type: 'number', unit: 'count', freeQuantity: 1 },
  }
  let saved = ''
  function Harness() {
    const [expression, setExpression] = useState(saved)
    return (
      <TaskUsagePricingEditor
        schema={schema}
        expression={expression}
        onExpressionChange={(next) => {
          saved = next
          setExpression(next)
        }}
      />
    )
  }
  const host = document.createElement('div')
  document.body.append(host)
  const root = createRoot(host)
  const inputFor = (label: string) => {
    const input = host.querySelector<HTMLInputElement>(
      `input[aria-label="${label}"]`
    )
    assert.ok(input, label)
    return input
  }
  // Use the native setter to simulate browser edits without updating React's tracker.
  const setValue = Object.getOwnPropertyDescriptor(
    dom.HTMLInputElement.prototype,
    'value'
  )?.set
  assert.ok(setValue)
  const edit = async (label: string, drafts: string[]) => {
    const input = inputFor(label)
    await act(async () => input.focus())
    for (const draft of drafts) {
      await act(async () => {
        setValue.call(input, draft)
        input.dispatchEvent(new Event('input', { bubbles: true }))
      })
      assert.equal(
        inputFor(label),
        input,
        'expression updates must not remount'
      )
      assert.equal(input.value, draft, `keep intermediate text ${draft}`)
    }
    await act(async () => input.blur())
    return input
  }
  const render = async (key: string) => {
    await act(async () =>
      root.render(
        <I18nextProvider i18n={i18n}>
          <Harness key={key} />
        </I18nextProvider>
      )
    )
  }
  try {
    await render('initial')
    assert.equal(saved, '', 'opening must not change existing settings')
    await edit('Fixed price 720p', ['0', '0.', '0.1'])
    await edit('billable_tokens 720p', ['2', '2.', '2.3'])
    await edit('seconds 720p', ['0', '0.', '0.0', '0.04', '0.042'])
    await edit('images 720p', ['0.02']) // Pasted fractional image price.
    await edit('Free quantity images 720p', ['5'])
    await edit('billable_tokens 1080p', ['0', '0.', '0.1'])
    const parsed = tryParseTaskMatrixConfig(saved, schema)
    assert.ok(parsed)
    assert.equal(parsed.rows[0].constant, 0.1)
    assert.deepEqual(parsed.rows[0].unitPrices, {
      billable_tokens: 2.3,
      images: 0.02,
      seconds: 0.042,
    })
    assert.equal(parsed.rows[0].freeQuantities.images, 5)
    assert.equal(parsed.rows[1].constant, 0)
    assert.equal(parsed.rows[1].unitPrices.billable_tokens, 0.1)
    const result = evaluateTaskVisualConfig(
      { tiers: taskMatrixToTiers(parsed, schema) },
      {
        resolution: '720p',
        billable_tokens: 1_000_000,
        seconds: 10,
        images: 7,
      },
      schema
    )
    assert.ok(result)
    assert.ok(Math.abs(result.total - 2.86) < 1e-10)
    const expressionBeforeReload = saved
    await render('reopened')
    assert.equal(saved, expressionBeforeReload)
    for (const [label, value] of [
      ['Fixed price 720p', '0.1'],
      ['billable_tokens 720p', '2.3'],
      ['seconds 720p', '0.042'],
      ['images 720p', '0.02'],
      ['Free quantity images 720p', '5'],
      ['billable_tokens 1080p', '0.1'],
    ]) {
      assert.equal(inputFor(label).value, value)
    }
    // Zero prices stay valid, and existing finite/non-negative guards remain active.
    await edit('Fixed price 720p', [''])
    assert.equal(inputFor('Fixed price 720p').value, '0')
    await edit('seconds 720p', ['-2'])
    assert.equal(inputFor('seconds 720p').value, '0')
    await edit('images 720p', ['1e999'])
    assert.equal(inputFor('images 720p').value, '0')
    const reset = [...host.querySelectorAll('button')].find(
      (button) => button.textContent === 'Reset'
    )
    assert.ok(reset)
    await act(async () => reset.click())
    assert.equal(inputFor('billable_tokens 720p').value, '0')
    assert.equal(inputFor('Free quantity images 720p').value, '1')
  } finally {
    await act(async () => root.unmount())
    host.remove()
  }
})
