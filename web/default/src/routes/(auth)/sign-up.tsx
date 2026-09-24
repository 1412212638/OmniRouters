import { createFileRoute, redirect } from '@tanstack/react-router'

import { SignUp } from '@/features/auth/sign-up'
import { useAuthStore } from '@/stores/auth-store'

export const Route = createFileRoute('/(auth)/sign-up')({
  component: SignUp,
  beforeLoad: async () => {
    const { auth } = useAuthStore.getState()

    // Authenticated users do not need the sign-up page.
    if (auth.user) {
      throw redirect({ to: '/dashboard' })
    }
  },
})
