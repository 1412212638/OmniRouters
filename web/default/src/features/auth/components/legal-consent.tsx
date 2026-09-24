import { Trans } from 'react-i18next'

import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'
import { cn } from '@/lib/utils'

import type { SystemStatus } from '../types'

interface LegalConsentProps {
  status: SystemStatus | null
  checked: boolean
  onCheckedChange: (nextValue: boolean) => void
  className?: string
}

export function LegalConsent({
  status,
  checked,
  onCheckedChange,
  className,
}: LegalConsentProps) {
  const hasUserAgreement = Boolean(status?.user_agreement_enabled)
  const hasPrivacyPolicy = Boolean(status?.privacy_policy_enabled)

  if (!hasUserAgreement && !hasPrivacyPolicy) {
    return null
  }

  const consentKey =
    hasUserAgreement && hasPrivacyPolicy
      ? 'I have read and agree to the <agreement>User Agreement</agreement> and <privacy>Privacy Policy</privacy>.'
      : hasUserAgreement
        ? 'I have read and agree to the <agreement>User Agreement</agreement>.'
        : 'I have read and agree to the <privacy>Privacy Policy</privacy>.'

  const handleChange = (value: boolean) => {
    onCheckedChange(value === true)
  }

  return (
    <div
      className={cn(
        'border-border/60 bg-muted/40 flex items-start gap-3 rounded-md border p-3',
        className
      )}
    >
      <Checkbox
        id='legal-consent'
        checked={checked}
        onCheckedChange={handleChange}
        className='mt-0.5'
      />
      <Label
        htmlFor='legal-consent'
        className='text-muted-foreground items-start gap-1 text-left text-xs leading-5 font-normal'
      >
        <Trans
          i18nKey={consentKey}
          components={{
            agreement: (
              <a
                href='/user-agreement'
                target='_blank'
                rel='noopener noreferrer'
                className='text-primary hover:underline'
              />
            ),
            privacy: (
              <a
                href='/privacy-policy'
                target='_blank'
                rel='noopener noreferrer'
                className='text-primary hover:underline'
              />
            ),
          }}
        />
      </Label>
    </div>
  )
}
