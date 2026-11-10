import { createFileRoute } from '@tanstack/react-router'

import { SubscriptionPlansPage } from '@/features/subscriptions/components/subscription-plans-page'

export const Route = createFileRoute('/_authenticated/subscription-plans/')({
  component: SubscriptionPlansPage,
})
