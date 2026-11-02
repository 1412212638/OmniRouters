import { createFileRoute, redirect } from '@tanstack/react-router'

import { Organizations } from '@/features/organizations'
import { getStatus } from '@/lib/api'

export const Route = createFileRoute('/_authenticated/organizations/')({
  beforeLoad: async () => {
    const status = await getStatus()
    if (status?.organization_enabled === false) {
      throw redirect({ to: '/dashboard' })
    }
  },
  component: Organizations,
})
