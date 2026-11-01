import { createFileRoute, redirect } from '@tanstack/react-router'

import { MemberTiers } from '@/features/member-tiers'
import { isSidebarModuleEnabled } from '@/lib/nav-modules'

export const Route = createFileRoute('/_authenticated/member-tiers/')({
  beforeLoad: () => {
    if (!isSidebarModuleEnabled('personal', 'tiers')) {
      throw redirect({ to: '/dashboard' })
    }
  },
  component: MemberTiers,
})
