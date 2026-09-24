import { Trans } from 'react-i18next'

import { cn } from '@/lib/utils'

import type { SystemStatus } from '../types'

interface TermsFooterProps {
  variant?: 'sign-in' | 'sign-up'
  className?: string
  status?: SystemStatus | null
}

export function TermsFooter({
  variant = 'sign-in',
  className,
  status,
}: TermsFooterProps) {
  const hasUserAgreement = Boolean(status?.user_agreement_enabled)
  const hasPrivacyPolicy = Boolean(status?.privacy_policy_enabled)

  if (!hasUserAgreement && !hasPrivacyPolicy) {
    return null
  }

  const footerKey =
    variant === 'sign-in'
      ? hasUserAgreement && hasPrivacyPolicy
        ? 'By clicking sign in, you agree to our <agreement>User Agreement</agreement> and <privacy>Privacy Policy</privacy>.'
        : hasUserAgreement
          ? 'By clicking sign in, you agree to our <agreement>User Agreement</agreement>.'
          : 'By clicking sign in, you agree to our <privacy>Privacy Policy</privacy>.'
      : hasUserAgreement && hasPrivacyPolicy
        ? 'By creating an account, you agree to our <agreement>User Agreement</agreement> and <privacy>Privacy Policy</privacy>.'
        : hasUserAgreement
          ? 'By creating an account, you agree to our <agreement>User Agreement</agreement>.'
          : 'By creating an account, you agree to our <privacy>Privacy Policy</privacy>.'

  return (
    <p className={cn('text-muted-foreground text-center text-xs', className)}>
      <Trans
        i18nKey={footerKey}
        components={{
          agreement: (
            <a
              href='/user-agreement'
              className='hover:text-primary underline underline-offset-4'
            />
          ),
          privacy: (
            <a
              href='/privacy-policy'
              className='hover:text-primary underline underline-offset-4'
            />
          ),
        }}
      />
    </p>
  )
}
