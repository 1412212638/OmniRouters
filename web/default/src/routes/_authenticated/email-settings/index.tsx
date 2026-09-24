import { createFileRoute, redirect } from '@tanstack/react-router'

import { EmailSettings } from '@/features/email-settings'
import { ROLE } from '@/lib/roles'
import { useAuthStore } from '@/stores/auth-store'

export const Route = createFileRoute('/_authenticated/email-settings/')({
  beforeLoad: () => {
    const { auth } = useAuthStore.getState()

    if (!auth.user || auth.user.role < ROLE.ADMIN) {
      throw redirect({
        to: '/403',
      })
    }
  },
  component: EmailSettings,
})
