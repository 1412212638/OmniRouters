import { Link, useSearch } from '@tanstack/react-router'
import { ArrowLeft } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { useStatus } from '@/hooks/use-status'
import { useSystemConfig } from '@/hooks/use-system-config'

import { AuthLayout } from '../auth-layout'
import { TermsFooter } from '../components/terms-footer'
import { UserAuthForm } from './components/user-auth-form'

export function SignIn() {
  const { t } = useTranslation()
  const { redirect } = useSearch({ from: '/(auth)/sign-in' })
  const { status } = useStatus()
  const { systemName } = useSystemConfig()
  const [emailLoginExpanded, setEmailLoginExpanded] = useState(false)

  return (
    <AuthLayout variant='sign-in'>
      <div className='w-full space-y-8'>
        <div className='space-y-2'>
          {emailLoginExpanded && (
            <button
              type='button'
              onClick={() => setEmailLoginExpanded(false)}
              className='text-muted-foreground hover:text-foreground mb-6 inline-flex items-center gap-2 text-sm transition-colors'
            >
              <ArrowLeft className='h-4 w-4' />
              {t('Back')}
            </button>
          )}
          <h2 className='text-center text-2xl font-semibold tracking-tight sm:text-left'>
            {t('Welcome to {{name}}', { name: systemName })}
          </h2>
          {emailLoginExpanded &&
            !status?.self_use_mode_enabled &&
            status?.register_enabled !== false && (
              <p className='text-muted-foreground text-left text-sm sm:text-base'>
                {t("Don't have an account?")}{' '}
                <Link
                  to='/sign-up'
                  className='hover:text-primary font-medium underline underline-offset-4'
                >
                  {t('Sign up')}
                </Link>
                .
              </p>
            )}
        </div>

        <UserAuthForm
          redirectTo={redirect}
          emailLoginExpanded={emailLoginExpanded}
          onEmailLoginExpandedChange={setEmailLoginExpanded}
        />

        <TermsFooter
          variant='sign-in'
          status={status}
          className='auth-terms-footer'
        />
      </div>
    </AuthLayout>
  )
}
