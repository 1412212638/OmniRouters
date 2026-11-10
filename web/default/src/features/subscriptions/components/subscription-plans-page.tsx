import { useTranslation } from 'react-i18next'

import { SectionPageLayout } from '@/components/layout'
import { SubscriptionPlansCard } from '@/features/wallet/components/subscription-plans-card'
import { useTopupInfo } from '@/features/wallet/hooks'
import { useAuthStore } from '@/stores/auth-store'

export function SubscriptionPlansPage() {
  const { t } = useTranslation()
  const userQuota = useAuthStore((state) => state.auth.user?.quota)
  const { topupInfo } = useTopupInfo()

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>
        {t('Subscription Plans')}
      </SectionPageLayout.Title>
      <SectionPageLayout.Description>
        {t('Choose a billing unit and subscribe to a plan for model access.')}
      </SectionPageLayout.Description>
      <SectionPageLayout.Content>
        <div className='mx-auto w-full max-w-6xl'>
          <SubscriptionPlansCard
            topupInfo={topupInfo}
            userQuota={userQuota}
            standalone
          />
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
