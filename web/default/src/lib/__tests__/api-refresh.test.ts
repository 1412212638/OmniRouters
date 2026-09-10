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
import { after, beforeEach, test } from 'node:test'

import axios, { AxiosError, type InternalAxiosRequestConfig } from 'axios'
import { Window } from 'happy-dom'

const domWindow = new Window()
Object.defineProperty(globalThis, 'window', {
  configurable: true,
  value: domWindow,
})
const { api, setDashboardAccessToken } = await import('../api')
const { useAuthStore } = await import('../../stores/auth-store')
const user = { id: 1, username: 'test', role: 1 }
const originalAdapter = axios.defaults.adapter
const originalAPIAdapter = api.defaults.adapter
let refreshCalls = 0

function httpError(status: number, config: InternalAxiosRequestConfig) {
  return new AxiosError(`HTTP ${status}`, undefined, config, undefined, {
    status,
    statusText: '',
    data: {},
    headers: { 'retry-after': '60' },
    config,
  })
}

beforeEach(() => {
  refreshCalls = 0
  useAuthStore.getState().auth.setUser(user)
  setDashboardAccessToken('expired-token')
  api.defaults.adapter = async (config) => {
    throw httpError(401, config)
  }
})

after(() => {
  axios.defaults.adapter = originalAdapter
  api.defaults.adapter = originalAPIAdapter
  delete (globalThis as { window?: unknown }).window
  domWindow.happyDOM.abort()
})

for (const status of [429, 500, 503, 409]) {
  test(`refresh ${status} preserves user/token and propagates the real status`, async () => {
    axios.defaults.adapter = async (config) => {
      refreshCalls++
      throw httpError(status, config)
    }
    await assert.rejects(
      api.get('/api/user/self'),
      (error: unknown) =>
        axios.isAxiosError(error) && error.response?.status === status
    )
    assert.equal(useAuthStore.getState().auth.user?.id, 1)
    assert.equal(
      domWindow.localStorage.getItem('dashboard_access_token'),
      'expired-token'
    )
    await assert.rejects(api.get('/api/user/models'))
    assert.equal(refreshCalls, 1, 'cooldown prevents another refresh')
  })
}

test('network failure during refresh does not become an unauthorized error', async () => {
  axios.defaults.adapter = async () => {
    throw new AxiosError('Network Error')
  }
  await assert.rejects(
    api.get('/api/user/self'),
    (error: unknown) => axios.isAxiosError(error) && !error.response
  )
  assert.equal(useAuthStore.getState().auth.user?.id, 1)
  assert.equal(
    domWindow.localStorage.getItem('dashboard_access_token'),
    'expired-token'
  )
})

test('refresh 401 clears the invalid session', async () => {
  axios.defaults.adapter = async (config) => {
    throw httpError(401, config)
  }
  await assert.rejects(api.get('/api/user/self'))
  assert.equal(useAuthStore.getState().auth.user, null)
  assert.equal(domWindow.localStorage.getItem('dashboard_access_token'), null)
})

test('direct API throttling does not trigger refresh or reset', async () => {
  api.defaults.adapter = async (config) => {
    throw httpError(429, config)
  }
  axios.defaults.adapter = async () => {
    refreshCalls++
    throw new Error('unexpected refresh')
  }
  await assert.rejects(api.get('/api/user/self', { skipErrorHandler: true }))
  assert.equal(refreshCalls, 0)
  assert.equal(useAuthStore.getState().auth.user?.id, 1)
})

test('malformed refresh response preserves the session for retry', async () => {
  axios.defaults.adapter = async (config) => ({
    status: 200,
    statusText: 'OK',
    headers: {},
    config,
    data: {},
  })
  await assert.rejects(api.get('/api/user/self', { skipErrorHandler: true }))
  assert.equal(useAuthStore.getState().auth.user?.id, 1)
})

test('successful refresh is shared by concurrent requests and retries with the new token', async () => {
  axios.defaults.adapter = async (config) => {
    refreshCalls++
    return {
      status: 200,
      statusText: 'OK',
      headers: {},
      config,
      data: { success: true, data: { access_token: 'new-token', user } },
    }
  }
  api.defaults.adapter = async (config) => {
    if (config.headers.Authorization !== 'Bearer new-token') {
      throw httpError(401, config)
    }
    return {
      status: 200,
      statusText: 'OK',
      headers: {},
      config,
      data: { success: true },
    }
  }
  const responses = await Promise.all([
    api.get('/api/user/self'),
    api.get('/api/user/models'),
  ])
  assert.ok(responses.every((response) => response.status === 200))
  assert.equal(refreshCalls, 1)
  assert.equal(
    domWindow.localStorage.getItem('dashboard_access_token'),
    'new-token'
  )
})
