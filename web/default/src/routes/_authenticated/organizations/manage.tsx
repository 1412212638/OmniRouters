import { createFileRoute, redirect } from '@tanstack/react-router'

import { Organizations } from '@/features/organizations'
import { getActiveOrganizationId } from '@/features/organizations/workspace-store'
import { getStatus } from '@/lib/api'

export const Route = createFileRoute('/_authenticated/organizations/manage')({
  beforeLoad: async () => {
    const status = await getStatus()
    if (status?.organization_enabled === false) {
      throw redirect({ to: '/dashboard' })
    }
    if (getActiveOrganizationId() === null) {
      throw redirect({ to: '/organizations' })
    }
  },
  component: () => <Organizations manageCurrent />,
})
