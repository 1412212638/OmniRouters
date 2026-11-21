import { Activity, BarChart3, WalletCards } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Skeleton } from '@/components/ui/skeleton'
import { formatQuota } from '@/lib/format'

import type { FundingSummary } from '../funding'
import type { UserWalletData } from '../types'

interface WalletStatsCardProps {
  user: UserWalletData | null
  loading?: boolean
  funding?: FundingSummary
  fundingLoading?: boolean
  fundingFailed?: boolean
  showFunding?: boolean
}

function formatFundingAmount(value?: string) {
  const numeric = Number(value ?? 0)
  return Number.isSafeInteger(numeric) ? formatQuota(numeric) : (value ?? '0')
}

export function WalletStatsCard(props: WalletStatsCardProps) {
  const { t } = useTranslation()
  if (props.loading) {
    return (
      <div className='overflow-hidden rounded-lg border'>
        <div className='divide-border/60 grid grid-cols-1 divide-y sm:grid-cols-3 sm:divide-x sm:divide-y-0'>
          {['balance', 'usage', 'requests'].map((key) => (
            <div key={key} className='px-3 py-3 sm:px-5 sm:py-4'>
              <Skeleton className='h-3.5 w-20' />
              <Skeleton className='mt-2 h-7 w-28' />
              <Skeleton className='mt-1.5 h-3.5 w-24' />
              {key === 'balance' && props.showFunding && (
                <Skeleton className='mt-1.5 h-3.5 w-full max-w-72' />
              )}
            </div>
          ))}
        </div>
      </div>
    )
  }

  const stats = [
    {
      label: t('Current Balance'),
      value: formatQuota(props.user?.quota ?? 0),
      description: t('Remaining quota'),
      icon: WalletCards,
    },
    {
      label: t('Total Usage'),
      value: formatQuota(props.user?.used_quota ?? 0),
      description: t('Total consumed quota'),
      icon: BarChart3,
    },
    {
      label: t('API Requests'),
      value: (props.user?.request_count ?? 0).toLocaleString(),
      description: t('Total requests made'),
      icon: Activity,
    },
  ]
  const fundingDetail = (() => {
    if (props.fundingFailed) {
      return <span role='alert'>{t('Failed to load funding records')}</span>
    }
    if (props.fundingLoading) {
      return <Skeleton className='h-3.5 w-36' />
    }
    return (
      <>
        <span>
          {t('Paid credit')} {formatFundingAmount(props.funding?.paid)}
        </span>
        <span aria-hidden='true'>·</span>
        <span>
          {t('Gift credit')} {formatFundingAmount(props.funding?.gift)}
        </span>
      </>
    )
  })()

  return (
    <div className='overflow-hidden rounded-lg border'>
      <div className='divide-border/60 grid grid-cols-1 divide-y sm:grid-cols-3 sm:divide-x sm:divide-y-0'>
        {stats.map((item) => (
          <div key={item.label} className='px-3 py-3 sm:px-5 sm:py-4'>
            <div className='flex items-center gap-2'>
              <item.icon className='text-muted-foreground/60 size-3.5 shrink-0' />
              <div className='text-muted-foreground truncate text-xs font-medium tracking-wider uppercase'>
                {item.label}
              </div>
            </div>

            <div className='text-foreground mt-1.5 font-mono text-base font-bold tracking-tight break-all tabular-nums sm:mt-2 sm:text-2xl'>
              {item.value}
            </div>
            {item.label === t('Current Balance') && props.showFunding ? (
              <div className='text-muted-foreground/60 mt-1 flex flex-wrap items-center gap-x-1.5 gap-y-1 text-xs'>
                <span>{item.description}</span>
                <span aria-hidden='true'>/</span>
                {fundingDetail}
              </div>
            ) : (
              <div className='text-muted-foreground/60 mt-1 hidden text-xs md:block'>
                {item.description}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
