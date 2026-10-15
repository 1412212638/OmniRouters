import { api, setDashboardAccessToken } from '@/lib/api'

import type {
  LoginPayload,
  LoginResponse,
  Login2FAResponse,
  TwoFAPayload,
  RegisterPayload,
  ApiResponse,
} from './types'

function persistAuthUser(
  data: { user?: { id?: number | string | null } } | null | undefined
) {
  if (data?.user?.id != null && typeof window !== 'undefined') {
    window.localStorage.setItem('uid', String(data.user.id))
  }
}

// ============================================================================
// Authentication APIs
// ============================================================================

// ----------------------------------------------------------------------------
// Login & Logout
// ----------------------------------------------------------------------------

// User login with username and password
export async function login(payload: LoginPayload) {
  const turnstile = payload.turnstile ?? ''
  const res = await api.post<LoginResponse>(
    `/api/user/login?turnstile=${turnstile}`,
    {
      username: payload.username,
      password: payload.password,
    }
  )
  if (res.data?.data?.access_token) {
    setDashboardAccessToken(res.data.data.access_token)
  }
  persistAuthUser(res.data?.data)
  return res.data
}

// Two-factor authentication login
export async function login2fa(payload: TwoFAPayload) {
  const res = await api.post<Login2FAResponse>('/api/user/login/2fa', payload)
  if (res.data?.data?.access_token) {
    setDashboardAccessToken(res.data.data.access_token)
  }
  persistAuthUser(res.data?.data)
  return res.data
}

// User logout
export async function logout(): Promise<ApiResponse> {
  const res = await api.post('/api/user/auth/logout', undefined, {
    skipErrorHandler: true,
  })
  setDashboardAccessToken()
  return res.data
}

export async function refreshDashboardSession(): Promise<ApiResponse> {
  const res = await api.post('/api/user/auth/refresh', undefined, {
    skipErrorHandler: true,
  })
  const token = res.data?.data?.access_token
  if (token) setDashboardAccessToken(token)
  persistAuthUser(res.data?.data)
  return res.data
}

// ----------------------------------------------------------------------------
// Password Management
// ----------------------------------------------------------------------------

// Send password reset email
export async function sendPasswordResetEmail(
  email: string,
  turnstile?: string
): Promise<ApiResponse> {
  const res = await api.get('/api/reset_password', {
    params: { email, turnstile },
  })
  return res.data
}

// ----------------------------------------------------------------------------
// OAuth
// ----------------------------------------------------------------------------

// Start GitHub OAuth flow
export async function githubOAuthStart(clientId: string, state: string) {
  const url = `https://github.com/login/oauth/authorize?client_id=${clientId}&state=${state}&scope=user:email`
  window.open(url)
}

// Get OAuth state for CSRF protection
export async function getOAuthState(): Promise<string> {
  const aff =
    typeof window !== 'undefined' ? (localStorage.getItem('aff') ?? '') : ''
  const res = await api.get('/api/oauth/state', { params: { aff } })
  if (res.data?.success) return res.data.data
  return ''
}

export async function startTelegramOAuth(
  intent: 'login' | 'bind' = 'login',
  affCode = ''
) {
  const res = await api.get('/api/oauth/telegram/start', {
    params: { intent, ...(intent === 'login' ? { aff: affCode } : {}) },
  })
  return res.data as { success: boolean; data?: { authorization_url?: string } }
}

// WeChat login by authorization code
export async function wechatLoginByCode(
  code: string,
  affCode = ''
): Promise<ApiResponse> {
  const res = await api.get('/api/oauth/wechat', {
    params: { code, aff: affCode },
  })
  return res.data
}

export type WeChatOfficialQrResponse = {
  success: boolean
  message: string
  data?: {
    code: string
    image_url: string
    expires_in: number
    name?: string
  }
}

export async function createWeChatOfficialQr(): Promise<WeChatOfficialQrResponse> {
  const res = await api.get<WeChatOfficialQrResponse>(
    '/api/oauth/wechat/official/qr'
  )
  return res.data
}

export async function pollWeChatOfficialLogin(code: string) {
  const res = await api.get<{
    success: boolean
    message: string
    data?: { status?: 'pending' | 'ready' | 'consumed'; expires_at?: number }
  }>('/api/oauth/wechat/official/poll', {
    params: { code },
    skipErrorHandler: true,
  })
  return res.data
}

export async function loginWithWeChatOfficial(
  code: string,
  affCode = ''
): Promise<ApiResponse> {
  const res = await api.get('/api/oauth/wechat/official/login', {
    params: { code, aff: affCode },
  })
  if (res.data?.data?.access_token) {
    setDashboardAccessToken(res.data.data.access_token)
  }
  persistAuthUser(res.data?.data)
  return res.data
}

export type WeChatOfficialBindQrResponse = WeChatOfficialQrResponse

export async function createWeChatOfficialBindQr(): Promise<WeChatOfficialBindQrResponse> {
  const res = await api.get<WeChatOfficialBindQrResponse>(
    '/api/user/oauth/wechat/official/bind/qr'
  )
  return res.data
}

export async function pollWeChatOfficialBind(code: string) {
  const res = await api.get<{
    success: boolean
    message: string
    data?: { status?: 'pending' | 'ready' | 'consumed'; expires_at?: number }
  }>('/api/user/oauth/wechat/official/bind/poll', {
    params: { code },
    skipErrorHandler: true,
  })
  return res.data
}

export async function bindWeChatOfficial(code: string): Promise<ApiResponse> {
  const res = await api.post('/api/user/oauth/wechat/official/bind', { code })
  return res.data
}

// ----------------------------------------------------------------------------
// Registration
// ----------------------------------------------------------------------------

// User registration
export async function register(payload: RegisterPayload): Promise<ApiResponse> {
  const res = await api.post(`/api/user/register`, payload, {
    params: { turnstile: payload.turnstile ?? '' },
  })
  return res.data
}

// Send email verification code
export async function sendEmailVerification(
  email: string,
  turnstile?: string
): Promise<ApiResponse> {
  const res = await api.get('/api/verification', {
    params: { email, turnstile },
  })
  return res.data
}

// Bind email to OAuth account
export async function bindEmail(
  email: string,
  code: string
): Promise<ApiResponse> {
  const res = await api.post('/api/oauth/email/bind', {
    email,
    code,
  })
  return res.data
}
