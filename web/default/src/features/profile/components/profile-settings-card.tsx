import { useQuery } from '@tanstack/react-query'
import { Link2, Settings, ShieldCheck } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { TitledCard } from '@/components/ui/titled-card'
import { cn } from '@/lib/utils'

import { getIdentityVerification } from '../api'
import type { UserProfile } from '../types'
import { AccountBindingsTab } from './tabs/account-bindings-tab'
import { NotificationTab } from './tabs/notification-tab'
import { IdentityVerificationTab } from './tabs/identity-verification-tab'

// ============================================================================
// Profile Settings Card Component
// ============================================================================

interface ProfileSettingsCardProps {
  profile: UserProfile | null
  loading: boolean
  onProfileUpdate: () => void
}

export function ProfileSettingsCard({
  profile,
  loading,
  onProfileUpdate,
}: ProfileSettingsCardProps) {
  const { t } = useTranslation()
  const [activeTab, setActiveTab] = useState('bindings')
  const identityVerification = useQuery({
    queryKey: ['identity-verification'],
    queryFn: getIdentityVerification,
    enabled: !loading && !!profile,
  })
  const showIdentityVerification =
    identityVerification.data?.success === true &&
    identityVerification.data.data?.enabled === true
  const visibleTab =
    activeTab === 'identity' && !showIdentityVerification
      ? 'bindings'
      : activeTab

  if (loading) {
    return (
      <Card data-card-hover='false' className='gap-0 overflow-hidden py-0'>
        <CardHeader className='border-b p-3 !pb-3 sm:p-5 sm:!pb-5'>
          <Skeleton className='h-6 w-32' />
          <Skeleton className='mt-2 h-4 w-48' />
        </CardHeader>
        <CardContent className='space-y-4 p-3 sm:p-5'>
          <Skeleton className='h-10 w-full' />
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className='h-20 w-full' />
          ))}
        </CardContent>
      </Card>
    )
  }

  return (
    <TitledCard
      title={t('Settings')}
      description={t('Configure your account preferences and integrations')}
      icon={<Settings className='h-4 w-4' />}
      disableHoverEffect
    >
      <Tabs value={visibleTab} onValueChange={setActiveTab}>
        <TabsList
          className={cn(
            'grid w-full items-stretch gap-1 rounded-xl p-1 group-data-horizontal/tabs:h-10',
            showIdentityVerification ? 'grid-cols-3' : 'grid-cols-2'
          )}
        >
          <TabsTrigger
            value='bindings'
            className='h-full gap-2 rounded-lg px-3 py-0 leading-none'
          >
            <Link2 className='h-4 w-4' />
            <span className='hidden sm:inline'>{t('Account Bindings')}</span>
            <span className='sm:hidden'>{t('Bindings')}</span>
          </TabsTrigger>
          <TabsTrigger
            value='settings'
            className='h-full gap-2 rounded-lg px-3 py-0 leading-none'
          >
            <Settings className='h-4 w-4' />
            <span className='hidden sm:inline'>
              {t('Settings & Preferences')}
            </span>
            <span className='sm:hidden'>{t('Settings')}</span>
          </TabsTrigger>
          {showIdentityVerification && (
            <TabsTrigger
              value='identity'
              className='h-full gap-2 rounded-lg px-3 py-0 leading-none'
            >
              <ShieldCheck className='h-4 w-4' />
              <span className='hidden sm:inline'>{t('Real-name Verification')}</span>
              <span className='sm:hidden'>{t('Identity')}</span>
            </TabsTrigger>
          )}
        </TabsList>

        <TabsContent value='bindings' className='mt-4 sm:mt-6'>
          <AccountBindingsTab profile={profile} onUpdate={onProfileUpdate} />
        </TabsContent>

        <TabsContent value='settings' className='mt-4 sm:mt-6'>
          <NotificationTab profile={profile} onUpdate={onProfileUpdate} />
        </TabsContent>
        {showIdentityVerification && (
          <TabsContent value='identity' className='mt-4 sm:mt-6'>
            <IdentityVerificationTab />
          </TabsContent>
        )}
      </Tabs>
    </TitledCard>
  )
}
