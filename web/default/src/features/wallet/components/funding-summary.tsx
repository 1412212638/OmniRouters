import { useTranslation } from 'react-i18next'

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { formatQuota } from '@/lib/format'

import type { FundingSummary } from '../funding'

function formatFundingAmount(value?: string) {
  const numeric = Number(value ?? 0)
  return Number.isSafeInteger(numeric) ? formatQuota(numeric) : (value ?? '0')
}

export function WalletFundingSummary({
  summary,
  loading,
  failed,
}: {
  summary?: FundingSummary
  loading?: boolean
  failed?: boolean
}) {
  const { t } = useTranslation()
  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('Funding overview')}</CardTitle>
        <CardDescription>
          {t(
            'Cumulative credits recorded since funding tracking was enabled. These are not remaining paid or gift balances; historical credits are not inferred.'
          )}
        </CardDescription>
      </CardHeader>
      <CardContent>
        {failed ? (
          <p role='alert'>{t('Failed to load funding records')}</p>
        ) : (
          <dl className='grid grid-cols-1 gap-4 sm:grid-cols-2'>
            {[
              { label: t('Cumulative paid credits'), value: summary?.paid },
              { label: t('Cumulative gift credits'), value: summary?.gift },
            ].map((item) => (
              <div key={item.label} className='min-w-0'>
                <dt className='text-muted-foreground text-sm'>{item.label}</dt>
                <dd className='mt-1 text-xl font-semibold break-all tabular-nums'>
                  {loading ? (
                    <Skeleton className='h-7 w-28' />
                  ) : (
                    formatFundingAmount(item.value)
                  )}
                </dd>
              </div>
            ))}
          </dl>
        )}
      </CardContent>
    </Card>
  )
}
