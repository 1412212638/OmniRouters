import { create } from 'zustand'

import type { AdminCapabilities } from '@/lib/admin-permissions'

export type UserPermissions = {
  sidebar_settings?: boolean
  sidebar_modules?: Record<string, unknown>
  admin_permissions?: AdminCapabilities
}

export interface AuthUser {
  id: number
  username: string
  display_name?: string
  email?: string
  role: number
  status?: number
  group?: string
  quota?: number
  used_quota?: number
  request_count?: number
  aff_code?: string
  aff_count?: number
  aff_quota?: number
  aff_history_quota?: number
  inviter_id?: number
  github_id?: string
  oidc_id?: string
  wechat_id?: string
  telegram_id?: string
  linux_do_id?: string
  setting?: Record<string, unknown> | string
  stripe_customer?: string
  sidebar_modules?: string
  permissions?: UserPermissions
}

function removeInvoiceProfile(user: AuthUser | null): AuthUser | null {
  if (!user) return null
  const safeUser = { ...user } as AuthUser & { invoice_profile?: unknown }
  delete safeUser.invoice_profile
  return safeUser
}

interface AuthState {
  auth: {
    user: AuthUser | null
    setUser: (user: AuthUser | null) => void
    reset: () => void
  }
}

export const useAuthStore = create<AuthState>()((set) => {
  // Restore user info from localStorage
  const initUser = (() => {
    try {
      if (typeof window !== 'undefined') {
        const saved = window.localStorage.getItem('user')
        if (!saved) return null
        const user = JSON.parse(saved) as AuthUser & {
          invoice_profile?: unknown
        }
        if (user && typeof user === 'object' && 'invoice_profile' in user) {
          delete user.invoice_profile
          window.localStorage.setItem('user', JSON.stringify(user))
        }
        return user
      }
    } catch {
      // Clear dirty data when parsing fails
      if (typeof window !== 'undefined') {
        window.localStorage.removeItem('user')
      }
    }
    return null
  })()

  return {
    auth: {
      user: initUser,
      setUser: (user) =>
        set((state) => {
          const safeUser = removeInvoiceProfile(user)
          // Persist user to localStorage
          if (typeof window !== 'undefined') {
            if (safeUser) {
              window.localStorage.setItem('user', JSON.stringify(safeUser))
            } else {
              window.localStorage.removeItem('user')
            }
          }
          return { ...state, auth: { ...state.auth, user: safeUser } }
        }),
      reset: () =>
        set((state) => {
          if (typeof window !== 'undefined') {
            window.localStorage.removeItem('user')
          }
          return {
            ...state,
            auth: { ...state.auth, user: null },
          }
        }),
    },
  }
})
