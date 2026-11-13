import assert from 'node:assert/strict'
import { after, test } from 'node:test'

import { Window } from 'happy-dom'

import en from '@/i18n/locales/en.json'
import fr from '@/i18n/locales/fr.json'
import ja from '@/i18n/locales/ja.json'
import ru from '@/i18n/locales/ru.json'
import vi from '@/i18n/locales/vi.json'
import zhTW from '@/i18n/locales/zh-TW.json'
import zh from '@/i18n/locales/zh.json'

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
const { TokenPriceFields } = await import('./token-price-fields')
const { TieredPricingEditor } = await import('./tiered-pricing-editor')
const { normalizeVisualTier, generateExprFromVisualConfig } =
  await import('@/features/pricing/lib/tier-expr')
const i18n = createInstance()
const resources = { en, zh, zhTW, fr, ja, ru, vi }
await i18n.init({ lng: 'zh', resources })
after(() => dom.close())

test('scenario switches preserve all prices and category toggles distinguish fallback from free', async () => {
  let changes = 0
  let tier = normalizeVisualTier({
    input_unit_cost: 5,
    output_unit_cost: 30,
    image_unit_cost: 8,
    image_cache_unit_cost: 2,
    cache_create_12h_unit_cost: 0,
  })
  function Harness() {
    const [value, setValue] = useState(tier)
    return (
      <TokenPriceFields
        tier={value}
        onChange={(next) => {
          changes++
          tier = next
          setValue(next)
        }}
      />
    )
  }
  const host = document.createElement('div')
  document.body.append(host)
  const root = createRoot(host)
  const clickText = async (text: string) => {
    const button = [...host.querySelectorAll('button')].find(
      (entry) => entry.textContent === text
    )
    assert.ok(button, text)
    await act(async () => button.click())
  }
  try {
    await act(async () =>
      root.render(
        <I18nextProvider i18n={i18n}>
          <Harness />
        </I18nextProvider>
      )
    )
    const original = generateExprFromVisualConfig({ tiers: [tier] })
    for (const scenario of [
      'Claude 缓存',
      'Gemini 多模态',
      'OpenAI 图片',
      '自定义／全部字段',
      '通用文本',
    ]) {
      await clickText(scenario)
      assert.equal(changes, 0)
      assert.equal(generateExprFromVisualConfig({ tiers: [tier] }), original)
      assert.ok(
        host.querySelector('input[aria-label="图片缓存输入 Tokens 单价"]')
      )
      assert.ok(
        host.querySelector(
          'input[aria-label="缓存写入 Tokens 单价（12 小时）"]'
        )
      )
    }
    await clickText('Gemini 多模态')
    const categorySwitch = host.querySelector<HTMLElement>(
      '[role="switch"][aria-label="单独计价：图像输出 Tokens 单价"]'
    )
    assert.ok(categorySwitch)
    assert.equal(categorySwitch.getAttribute('aria-checked'), 'false')
    await act(async () => categorySwitch.click())
    assert.equal(tier.image_output_unit_cost, 0)
    assert.match(generateExprFromVisualConfig({ tiers: [tier] }), /img_o \* 0/)
    assert.match(host.textContent ?? '', /此类别免费/)
    await act(async () => categorySwitch.click())
    assert.equal(tier.image_output_unit_cost, undefined)
    assert.equal(generateExprFromVisualConfig({ tiers: [tier] }), original)
    await clickText('通用文本')
    await clickText('展开其他 Tokens 类别')
    assert.match(host.textContent ?? '', /音频输出 Tokens 单价/)
    await clickText('收起其他 Tokens 类别')
    assert.doesNotMatch(host.textContent ?? '', /音频输出 Tokens 单价/)
    assert.equal(generateExprFromVisualConfig({ tiers: [tier] }), original)
    await act(async () => i18n.changeLanguage('en'))
    assert.match(host.textContent ?? '', /Configuration scenario/)
    assert.doesNotMatch(host.textContent ?? '', /pricingEditor\./)
  } finally {
    await act(async () => root.unmount())
    host.remove()
    await i18n.changeLanguage('zh')
  }
})

test('opening the full editor preserves an existing expression with an explicitly free category', async () => {
  const expression =
    'tier("base", p * 5 + c * 30 + cr * 1.25 + img * 8 + img_o * 0)'
  const emitted: string[] = []
  const host = document.createElement('div')
  document.body.append(host)
  const root = createRoot(host)
  try {
    await act(async () =>
      root.render(
        <I18nextProvider i18n={i18n}>
          <TieredPricingEditor
            modelName='custom-public-model'
            billingExpr={expression}
            requestRuleExpr=''
            onBillingExprChange={(value) => emitted.push(value)}
            onRequestRuleExprChange={() => {}}
          />
        </I18nextProvider>
      )
    )
    assert.deepEqual(emitted, [])
    assert.match(host.textContent ?? '', /此类别免费/)
    assert.ok(host.querySelector('input[aria-label="图像输出 Tokens 单价"]'))
  } finally {
    await act(async () => root.unmount())
    host.remove()
  }
})

test('every guided pricing key has a translation in all seven languages', () => {
  const keys = Object.keys(en.translation).filter((key) =>
    key.startsWith('pricingEditor.')
  )
  assert.ok(keys.length >= 40)
  for (const [language, resource] of Object.entries(resources)) {
    for (const key of keys) {
      const value = (resource.translation as Record<string, string>)[key]
      assert.ok(value, `${language}: ${key}`)
      if (language !== 'en' && key.endsWith('Help')) {
        assert.notEqual(
          value,
          (en.translation as Record<string, string>)[key],
          `${language}: ${key}`
        )
      }
    }
  }
})
