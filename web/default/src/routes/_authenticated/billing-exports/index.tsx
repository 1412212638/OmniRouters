import { createFileRoute } from '@tanstack/react-router'

import { BillingExports } from '@/features/billing-exports'

export const Route = createFileRoute('/_authenticated/billing-exports/')({
  component: BillingExports,
})
