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
import axios, { type AxiosRequestConfig } from 'axios'
import { t } from 'i18next'
import { toast } from 'sonner'

import { useAuthStore } from '@/stores/auth-store'

declare module 'axios' {
  export interface AxiosRequestConfig {
    skipBusinessError?: boolean
    skipErrorHandler?: boolean
    disableDuplicate?: boolean
  }
}

export type ApiRequestConfig = AxiosRequestConfig

// ============================================================================
// Axios Instance Configuration
// ============================================================================

// Base URL: empty string for same-origin API requests
const baseURL = ''

const accessTokenKey = 'dashboard_access_token'

export function setDashboardAccessToken(token?: string) {
  refreshRetryAt = 0
  refreshFailure = undefined
  if (typeof window === 'undefined') return
  if (token) window.localStorage.setItem(accessTokenKey, token)
  else window.localStorage.removeItem(accessTokenKey)
}

function getDashboardAccessToken() {
  try {
    return typeof window !== 'undefined'
      ? window.localStorage.getItem(accessTokenKey)
      : null
  } catch {
    return null
  }
}

// Create axios instance with default config
export const api = axios.create({
  baseURL,
  withCredentials: true, // Include cookies in cross-origin requests
  headers: {
    'Cache-Control': 'no-store', // Prevent caching
  },
})

let refreshPromise: Promise<string> | null = null
let refreshRetryAt = 0
let refreshFailure: unknown

function refreshDashboardToken(): Promise<string> {
  if (refreshPromise) return refreshPromise
  if (Date.now() < refreshRetryAt) return Promise.reject(refreshFailure)
  refreshPromise = axios
    .post('/api/user/auth/refresh', undefined, {
      withCredentials: true,
      headers: { 'Cache-Control': 'no-store' },
    })
    .then((response) => {
      const token = response.data?.data?.access_token
      if (typeof token !== 'string' || token.length === 0) {
        throw new Error(t('Request failed'))
      }
      setDashboardAccessToken(token)
      const userId = response.data?.data?.user?.id
      if (userId != null && typeof window !== 'undefined') {
        window.localStorage.setItem('uid', String(userId))
      }
      return token
    })
    .catch((error: unknown) => {
      // Throttling, refresh races and outages do not invalidate the session.
      // Share a cooldown across requests instead of repeatedly refreshing.
      if (!axios.isAxiosError(error) || error.response?.status !== 401) {
        const status = axios.isAxiosError(error)
          ? error.response?.status
          : undefined
        const retryAfter = axios.isAxiosError(error)
          ? Number(error.response?.headers?.['retry-after'])
          : Number.NaN
        let delay = status === 429 ? 60 : 5
        if (Number.isFinite(retryAfter) && retryAfter > 0) {
          delay = Math.min(retryAfter, 1200)
        }
        refreshFailure = error
        refreshRetryAt = Date.now() + delay * 1000
      }
      throw error
    })
    .finally(() => {
      refreshPromise = null
    })
  return refreshPromise
}

// ============================================================================
// Request Deduplication
// ============================================================================

// Deduplicate concurrent GET requests to the same URL
// Prevents multiple identical requests from being sent simultaneously
const inFlightGet = new Map<string, Promise<unknown>>()
const originalGet = api.get.bind(api)

api.get = ((url: string, config: ApiRequestConfig = {}) => {
  const disableDuplicate = config.disableDuplicate
  if (disableDuplicate) return originalGet(url, config)

  const params = config.params ? JSON.stringify(config.params) : '{}'
  const key = `${url}?${params}`

  // Return existing in-flight request if available
  if (inFlightGet.has(key)) return inFlightGet.get(key)!

  // Create new request and clean up after completion
  const req = originalGet(url, config).finally(() => inFlightGet.delete(key))
  inFlightGet.set(key, req)
  return req
}) as typeof api.get

// ============================================================================
// Response Interceptor
// ============================================================================

// Handle business logic errors and HTTP errors globally
api.interceptors.response.use(
  (response) => {
    const skipBusiness = response.config.skipBusinessError

    // Unified business response format: { success, message, data }
    if (
      !skipBusiness &&
      response &&
      response.data &&
      typeof response.data.success === 'boolean'
    ) {
      if (!response.data.success) {
        // Show error toast for business failures
        const msg = response.data.message || t('Request failed')
        toast.error(msg)
      }
    }
    return response
  },
  (error) => {
    const skip = error?.config?.skipErrorHandler
    const status = error?.response?.status

    const request = error?.config as
      | (ApiRequestConfig & {
          _dashboardRefreshRetried?: boolean
        })
      | undefined
    const requestURL = request?.url ?? ''
    const canRefresh =
      status === 401 &&
      request &&
      !request._dashboardRefreshRetried &&
      !requestURL.includes('/api/user/auth/refresh') &&
      !requestURL.includes('/api/user/auth/logout') &&
      Boolean(getDashboardAccessToken())
    if (canRefresh) {
      request._dashboardRefreshRetried = true
      return refreshDashboardToken().then(
        (token) => {
          request.headers = request.headers ?? {}
          ;(request.headers as Record<string, string>).Authorization =
            `Bearer ${token}`
          return api.request(request)
        },
        (refreshError: unknown) => {
          if (
            axios.isAxiosError(refreshError) &&
            refreshError.response?.status === 401
          ) {
            setDashboardAccessToken()
            useAuthStore.getState().auth.reset()
          } else if (!skip) {
            const limited =
              axios.isAxiosError(refreshError) &&
              refreshError.response?.status === 429
            toast.error(
              limited ? t('Too many requests') : t('Please try again later.'),
              {
                id: 'dashboard-session-refresh',
              }
            )
          }
          // Propagate the refresh status, not the original access-token 401:
          // route guards and QueryCache must not log users out on a 429/5xx.
          throw refreshError
        }
      )
    }

    if (status === 401) {
      try {
        useAuthStore.getState().auth.reset()
      } catch {
        /* empty */
      }

      if (!skip) {
        toast.error(t('Session expired!'))
      }
    } else if (status === 429 && !skip) {
      toast.error(t('Too many requests'))
    } else if (!skip) {
      // Other errors: show error message from response or default
      const msg =
        error?.response?.data?.message || error?.message || t('Request failed')
      toast.error(msg)
    }
    return Promise.reject(error)
  }
)

// ============================================================================
// Common Headers Utility
// ============================================================================

/**
 * Get user ID from localStorage
 */
function getUserId(): string | null {
  try {
    if (typeof window !== 'undefined') {
      return window.localStorage.getItem('uid')
    }
  } catch {
    /* empty */
  }
  return null
}

/**
 * Get common request headers (for both axios and SSE requests)
 */
export function getCommonHeaders(): Record<string, string> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }

  const uid = getUserId()
  if (uid) {
    headers['New-Api-User'] = uid
  }

  return headers
}

// ============================================================================
// Request Interceptor
// ============================================================================

// Attach user ID header for all requests
api.interceptors.request.use((config) => {
  const uid = getUserId()
  if (uid) {
    // Custom header for user identification
    ;(config.headers as Record<string, string>)['New-Api-User'] = uid
  }
  const accessToken = getDashboardAccessToken()
  if (accessToken) {
    ;(config.headers as Record<string, string>)['Authorization'] =
      `Bearer ${accessToken}`
  }
  return config
})

// ============================================================================
// Common API Functions
// ============================================================================

// ----------------------------------------------------------------------------
// User APIs
// ----------------------------------------------------------------------------

// Get current user info
export async function getSelf() {
  const res = await api.get('/api/user/self', {
    // Avoid global 401 toast during guards/preloads
    skipErrorHandler: true,
  })
  return res.data
}

interface GetUserModelsOptions {
  endpointType?: string
}

// Get user available models
export async function getUserModels(
  options: GetUserModelsOptions = {}
): Promise<{
  success: boolean
  message?: string
  data?: string[]
}> {
  const res = await api.get('/api/user/models', {
    params: options.endpointType
      ? { endpoint_type: options.endpointType }
      : undefined,
  })
  return res.data
}

// Get user groups with descriptions and ratios
export async function getUserGroups(): Promise<{
  success: boolean
  message?: string
  data?: Record<string, { desc: string; ratio: number | string }>
}> {
  const res = await api.get('/api/user/self/groups')
  return res.data
}

// ----------------------------------------------------------------------------
// System APIs
// ----------------------------------------------------------------------------

// Get system status
export async function getStatus() {
  const res = await api.get('/api/status')
  return res.data?.data as Record<string, unknown>
}

// Get system notice
export async function getNotice(): Promise<{
  success: boolean
  message?: string
  data?: string
}> {
  const res = await api.get('/api/notice')
  return res.data
}

// ----------------------------------------------------------------------------
// 2FA Management APIs
// ----------------------------------------------------------------------------

// Get 2FA status
export async function get2FAStatus() {
  const res = await api.get('/api/user/2fa/status')
  return res.data
}

// Setup 2FA
export async function setup2FA() {
  const res = await api.post('/api/user/2fa/setup')
  return res.data
}

// Enable 2FA with verification code
export async function enable2FA(code: string) {
  const res = await api.post('/api/user/2fa/enable', { code })
  return res.data
}

// Disable 2FA with verification code
export async function disable2FA(code: string) {
  const res = await api.post('/api/user/2fa/disable', { code })
  return res.data
}

// Regenerate 2FA backup codes
export async function regenerate2FABackupCodes(code: string) {
  const res = await api.post('/api/user/2fa/backup_codes', { code })
  return res.data
}
