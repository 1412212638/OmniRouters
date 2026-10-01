import { createFileRoute } from '@tanstack/react-router'

import { FriendsHub } from '@/features/friends-hub'

export const Route = createFileRoute('/_authenticated/friends-hub/')({
  component: FriendsHub,
})
