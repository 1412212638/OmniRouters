import { createFileRoute, Navigate, redirect } from '@tanstack/react-router'

import { SubscriptionPlansPage } from '@/features/subscriptions/components/subscription-plans-page'
import {
  isSidebarRouteVisible,
  useIsSidebarModuleVisible,
} from '@/hooks/use-sidebar-config'
import { systemStatusQueryOptions } from '@/hooks/use-status'
import { useAuthStore } from '@/stores/auth-store'

const path = '/subscription-plans'

export const Route = createFileRoute('/_authenticated/subscription-plans/')({
  beforeLoad: async ({ context }) => {
    const status = await context.queryClient.fetchQuery({
      ...systemStatusQueryOptions(),
      staleTime: 0,
    })
    const user = useAuthStore.getState().auth.user
    if (!isSidebarRouteVisible(path, status, user)) {
      throw redirect({ to: '/dashboard', replace: true })
    }
  },
  component: RouteComponent,
})

function RouteComponent() {
  const visible = useIsSidebarModuleVisible(path)
  return visible ? (
    <SubscriptionPlansPage />
  ) : (
    <Navigate to='/dashboard' replace />
  )
}
