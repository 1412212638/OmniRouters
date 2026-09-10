import type { AuditLog } from './api'

export type AuditDetails = {
  version?: number
  method?: string
  route?: string
  target?: string
  auth_method?: string
  http_status?: number
  outcome?: string
  failure?: string
  changes?: Record<string, { before: unknown; after: unknown }>
  params?: Record<string, string | number | boolean>
}

export function readAuditDetails(item: AuditLog): AuditDetails {
  if (!item.other) return {}
  try {
    const parsed = JSON.parse(item.other)
    if (!parsed || parsed.version !== 1) return {}
    const result: AuditDetails = { version: 1 }
    if (
      parsed.params &&
      typeof parsed.params === 'object' &&
      !Array.isArray(parsed.params)
    ) {
      result.params = Object.fromEntries(
        Object.entries(parsed.params).filter(([, value]) =>
          ['string', 'number', 'boolean'].includes(typeof value)
        )
      ) as AuditDetails['params']
    }
    for (const key of [
      'method',
      'route',
      'target',
      'auth_method',
      'outcome',
      'failure',
    ] as const) {
      if (typeof parsed[key] === 'string') result[key] = parsed[key]
    }
    if (typeof parsed.http_status === 'number') {
      result.http_status = parsed.http_status
    }
    if (
      parsed.changes &&
      typeof parsed.changes === 'object' &&
      !Array.isArray(parsed.changes)
    ) {
      result.changes = Object.fromEntries(
        Object.entries(parsed.changes).filter(
          ([, value]) =>
            value &&
            typeof value === 'object' &&
            'before' in value &&
            'after' in value
        )
      ) as AuditDetails['changes']
    }
    return result
  } catch {
    return item.other.startsWith('/api/') ? { route: item.other } : {}
  }
}

export function auditValue(value: unknown): string {
  if (value == null) return '-'
  return typeof value === 'string' ? value : JSON.stringify(value)
}

export const actionLabels: Record<string, string> = {
  request: 'Request',
  generic: 'Operation',
  'login.success': 'Login successful',
  'login.attempt': 'Login attempt',
  'oauth.callback': 'OAuth callback',
  'token.view': 'View API key',
  'session.logout': 'Sign out',
  'password.change': 'Change password',
  'email.bind': 'Bind email',
  'oauth.unbind': 'Unbind OAuth',
  'oauth.unbind.admin': 'Unbind OAuth',
  'wechat.bind': 'Bind WeChat',
  'telegram.bind': 'Bind Telegram',
  '2fa.enable': 'Enable two-factor authentication',
  '2fa.disable': 'Disable two-factor authentication',
  '2fa.backup_codes.regenerate': 'Regenerate backup codes',
  'account.delete': 'Delete account',
  'user.passkey_register': 'Register passkey',
  'user.passkey_delete': 'Delete passkey',
  'option.update': 'Update settings',
  'channel.update': 'Update channel',
  'channel.create': 'Create channel',
  'channel.delete': 'Delete channel',
  'channel.key_view': 'View channel key',
  'user.update': 'Update user',
  'user.create': 'Create user',
  'user.delete': 'Delete user',
  'token.create': 'Create API key',
  'token.update': 'Update API key',
  'token.delete': 'Delete API key',
  'session.revoke': 'Revoke session',
  generate: 'Create access token',
  revoke: 'Revoke access token',
}
