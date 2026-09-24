import { createFileRoute } from '@tanstack/react-router'

import { TicketsPage } from '@/features/tickets'

export const Route = createFileRoute('/_authenticated/tickets/')({
  component: () => <TicketsPage />,
})
