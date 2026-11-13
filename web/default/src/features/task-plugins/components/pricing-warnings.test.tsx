import assert from 'node:assert/strict'
import { after, test } from 'node:test'

import { Window } from 'happy-dom'

import en from '@/i18n/locales/en.json'
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
const { act } = await import('react')
const { createRoot } = await import('react-dom/client')
const { createInstance } = await import('i18next')
const { I18nextProvider } = await import('react-i18next')
const { QueryClient, QueryClientProvider } =
  await import('@tanstack/react-query')
const { api } = await import('@/lib/api')
const { useAuthStore } = await import('@/stores/auth-store')
const { TaskPricingWarnings } = await import('./pricing-warnings')
const i18n = createInstance()
await i18n.init({ lng: 'zh', resources: { en, zh } })
const originalAdapter = api.defaults.adapter
const originalUser = useAuthStore.getState().auth.user
after(() => {
  api.defaults.adapter = originalAdapter
  useAuthStore.getState().auth.setUser(originalUser)
  dom.close()
})

async function settle() {
  await act(async () => {
    await new Promise((resolve) => setTimeout(resolve, 40))
  })
}

test('pricing diagnostics refresh after saved prices and plugin changes, expose errors, and retry', async () => {
  useAuthStore.getState().auth.setUser({ id: 1, username: 'root', role: 100 })
  let calls = 0
  let fail = false
  let warnings = [
    {
      model_name: 'public-model-alias',
      plugin_key: 'internal-plugin',
      plugin_name: 'Admin Plugin',
      channel_ids: [2, 4],
      reason: 'missing_price',
    },
  ]
  api.defaults.adapter = async (config) => {
    assert.equal(config.url, '/api/option/model_pricing/task_warnings')
    calls++
    return {
      status: 200,
      statusText: 'OK',
      headers: {},
      config,
      data: fail
        ? { success: false, message: 'database unavailable' }
        : {
            success: true,
            data: { enabled: true, warnings, skipped_channels: [8] },
          },
    }
  }
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  const host = document.createElement('div')
  document.body.append(host)
  const root = createRoot(host)
  try {
    await act(async () =>
      root.render(
        <QueryClientProvider client={client}>
          <I18nextProvider i18n={i18n}>
            <TaskPricingWarnings showPricingLink={false} />
          </I18nextProvider>
        </QueryClientProvider>
      )
    )
    await settle()
    assert.match(host.textContent || '', /public-model-alias/)
    assert.match(host.textContent || '', /未配置任务价格/)
    assert.match(host.textContent || '', /渠道 ID：8/)
    const initialCalls = calls
    warnings = []
    await act(async () => {
      client.setQueryData(['model-pricing-snapshot'], { version: 'saved' })
    })
    await settle()
    assert.ok(calls > initialCalls)
    assert.doesNotMatch(host.textContent || '', /public-model-alias/)
    assert.match(host.textContent || '', /未发现任务定价配置问题/)
    fail = true
    await act(async () => {
      client.setQueryData(['task-plugins'], [{ version: 'updated' }])
    })
    await settle()
    assert.match(host.textContent || '', /任务定价检查失败/)
    fail = false
    const refresh = [...host.querySelectorAll('button')].find((button) =>
      button.textContent?.includes('刷新')
    )
    assert.ok(refresh)
    await act(async () => {
      refresh.dispatchEvent(new MouseEvent('click', { bubbles: true }))
    })
    await settle()
    assert.doesNotMatch(host.textContent || '', /任务定价检查失败/)
    assert.match(host.textContent || '', /未发现任务定价配置问题/)
  } finally {
    await act(async () => root.unmount())
    client.clear()
    host.remove()
  }
})

test('ordinary users neither request nor render internal plugin diagnostics', async () => {
  useAuthStore.getState().auth.setUser({ id: 9, username: 'user', role: 1 })
  let calls = 0
  api.defaults.adapter = async () => {
    calls++
    throw new Error('must not request')
  }
  const client = new QueryClient()
  // Even cached root diagnostics must not render after a role change.
  client.setQueryData(['task-pricing-warnings'], {
    enabled: true,
    warnings: [{ plugin_name: 'Secret Plugin' }],
    skipped_channels: [],
  })
  const host = document.createElement('div')
  const root = createRoot(host)
  try {
    await act(async () =>
      root.render(
        <QueryClientProvider client={client}>
          <I18nextProvider i18n={i18n}>
            <TaskPricingWarnings />
          </I18nextProvider>
        </QueryClientProvider>
      )
    )
    await settle()
    assert.equal(calls, 0)
    assert.equal(host.textContent, '')
  } finally {
    await act(async () => root.unmount())
    client.clear()
  }
})
