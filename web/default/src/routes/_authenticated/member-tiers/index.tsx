import { createFileRoute } from '@tanstack/react-router'

import { MemberTiers } from '@/features/member-tiers'

export const Route = createFileRoute('/_authenticated/member-tiers/')({
  component: MemberTiers,
})
