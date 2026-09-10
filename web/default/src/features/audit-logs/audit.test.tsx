import assert from 'node:assert/strict'
import { after, test } from 'node:test'

import { Window } from 'happy-dom'

import en from '@/i18n/locales/en.json'
import zh from '@/i18n/locales/zh.json'

import type { AuditLog } from './api'
import { auditValue, readAuditDetails } from './details'

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
const { AuditLogs } = await import('./index')
const instance = createInstance()
await instance.init({ lng: 'zh', resources: { en, zh } })
const originalAdapter = api.defaults.adapter
const requests: Record<string, unknown>[] = []
const item: AuditLog = {
  id: 1,
  user_id: 7,
  username: 'admin',
  created_at: 10,
  category: 'operation',
  action: 'option.update',
  outcome: 'failed',
  success: false,
  event_id: 'event',
  request_id: 'request-123',
  other: JSON.stringify({
    version: 1,
    method: 'PUT',
    route: '/api/option/',
    target: 'ModelRatio',
    outcome: 'failed',
    failure: 'business_rejected',
    changes: { key: { before: '[REDACTED]', after: '[REDACTED]' } },
  }),
}

after(() => {
  api.defaults.adapter = originalAdapter
  dom.close()
})

test('old and malformed metadata never crash the details parser', () => {
  assert.deepEqual(readAuditDetails({ ...item, other: '/api/user/self' }), {
    route: '/api/user/self',
  })
  assert.deepEqual(readAuditDetails({ ...item, other: '{bad' }), {})
  assert.deepEqual(readAuditDetails({ ...item, other: 'null' }), {})
  assert.equal(auditValue(null), '-')
})

test('business API failure is shown as an error instead of an empty log list', async () => {
  api.defaults.adapter = async (config) => ({
    status: 200,
    statusText: 'OK',
    headers: {},
    config,
    data: { success: false, message: 'denied' },
  })
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  const host = document.createElement('div')
  document.body.append(host)
  const root = createRoot(host)
  await act(async () =>
    root.render(
      <QueryClientProvider client={client}>
        <I18nextProvider i18n={instance}>
          <AuditLogs />
        </I18nextProvider>
      </QueryClientProvider>
    )
  )
  await act(async () => {
    await new Promise((resolve) => setTimeout(resolve, 50))
  })
  assert.match(host.textContent || '', /加载审计日志失败/)
  await act(async () => root.unmount())
  client.clear()
  host.remove()
})

test('audit page filters, pagination, details and language switching', async () => {
  api.defaults.adapter = async (config) => {
    requests.push(config.params)
    return {
      status: 200,
      statusText: 'OK',
      headers: {},
      config,
      data: {
        success: true,
        data: { items: [item], total: 42, page: 1, page_size: 20 },
      },
    }
  }
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  const host = document.createElement('div')
  document.body.append(host)
  const root = createRoot(host)
  const flush = async () => {
    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 40))
    })
  }
  await act(async () =>
    root.render(
      <QueryClientProvider client={client}>
        <I18nextProvider i18n={instance}>
          <AuditLogs />
        </I18nextProvider>
      </QueryClientProvider>
    )
  )
  await flush()
  assert.equal(requests[0].view, 'important')
  assert.match(host.textContent || '', /重要操作/)
  assert.match(host.textContent || '', /ModelRatio/)
  await act(async () =>
    (
      host.querySelector('button[aria-label="Next page"]') ||
      host.querySelector('button[aria-label="下一页"]')
    )?.dispatchEvent(new MouseEvent('click', { bubbles: true }))
  )
  await flush()
  assert.equal(requests.at(-1)?.p, 2)
  const accessTab = [
    ...host.querySelectorAll<HTMLButtonElement>('[role="tab"]'),
  ].find((element) => element.textContent === '访问记录')
  assert.ok(accessTab)
  await act(async () => accessTab.click())
  await flush()
  assert.equal(requests.at(-1)?.view, 'access')
  assert.equal(requests.at(-1)?.p, 1)
  const outcome = host.querySelector<HTMLSelectElement>('#audit-outcome')
  assert.ok(outcome)
  await act(async () => {
    outcome.value = 'failed'
    outcome.dispatchEvent(new Event('change', { bubbles: true }))
  })
  await act(async () => {
    host
      .querySelector('form')
      ?.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }))
  })
  await flush()
  assert.equal(requests.at(-1)?.outcome, 'failed')
  const details =
    host.querySelector<HTMLButtonElement>('button[aria-label="详情"]') ||
    host.querySelector<HTMLButtonElement>('button[aria-label="Details"]')
  assert.ok(details)
  await act(async () => details.click())
  await flush()
  assert.match(document.body.textContent || '', /request-123/)
  assert.match(document.body.textContent || '', /已脱敏/)
  await act(async () => instance.changeLanguage('en'))
  await flush()
  assert.match(document.body.textContent || '', /Audit details/)
  assert.match(document.body.textContent || '', /Redacted/)
  await act(async () => root.unmount())
  client.clear()
  host.remove()
})
