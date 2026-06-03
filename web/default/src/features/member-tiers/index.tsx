/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { useEffect, type ElementType } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  BadgeCheck,
  CheckCircle2,
  Info,
  Layers3,
  Loader2,
  TrendingUp,
  WalletCards,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { useAuthStore } from '@/stores/auth-store'
import { formatQuota } from '@/lib/format'
import { cn } from '@/lib/utils'
import { SectionPageLayout } from '@/components/layout'
import { EmptyState } from '@/components/empty-state'
import { Badge } from '@/components/ui/badge'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { getMemberTierSelf } from './api'
import type { MemberTierEvaluation, MemberTierProgress } from './types'

function ratioLabel(value: number | null | undefined) {
  if (value == null || Number.isNaN(value)) return '1'
  return Number(value).toFixed(4).replace(/0+$/, '').replace(/\.$/, '')
}

function progressPercent(total: number, target: number) {
  if (!target || target <= 0) return 100
  return Math.max(0, Math.min(100, (total / target) * 100))
}

function TierStatusBadge({ item }: { item: MemberTierProgress }) {
  const { t } = useTranslation()
  if (item.current) {
    return (
      <Badge variant='default'>
        <CheckCircle2 data-icon='inline-start' />
        {t('Current')}
      </Badge>
    )
  }
  if (item.next) {
    return <Badge variant='secondary'>{t('Next')}</Badge>
  }
  if (item.qualified) {
    return <Badge variant='outline'>{t('Qualified')}</Badge>
  }
  return <Badge variant='outline'>{t('Locked')}</Badge>
}

function SummaryCard({
  title,
  value,
  description,
  icon: Icon,
}: {
  title: string
  value: string
  description?: string
  icon: ElementType
}) {
  return (
    <Card size='sm'>
      <CardHeader className='grid grid-cols-[1fr_auto] items-start gap-3'>
        <div className='min-w-0'>
          <CardDescription>{title}</CardDescription>
          <CardTitle className='mt-1 truncate text-lg'>{value}</CardTitle>
        </div>
        <div className='bg-muted text-muted-foreground flex size-8 items-center justify-center rounded-lg'>
          <Icon className='size-4' />
        </div>
      </CardHeader>
      {description && (
        <CardContent>
          <p className='text-muted-foreground truncate text-xs'>{description}</p>
        </CardContent>
      )}
    </Card>
  )
}

function TierCard({
  item,
  totalTopup,
  usedQuota,
}: {
  item: MemberTierProgress
  totalTopup: number
  usedQuota: number
}) {
  const { t } = useTranslation()
  const rule = item.rule

  return (
    <Card
      size='sm'
      className={cn(
        item.current && 'ring-primary/40 bg-primary/[0.03]',
        item.next && !item.current && 'ring-info/40'
      )}
    >
      <CardHeader>
        <div className='flex items-start justify-between gap-3'>
          <div className='min-w-0'>
            <CardTitle className='truncate'>
              {rule.display_name || rule.group}
            </CardTitle>
            <CardDescription className='truncate'>
              {rule.group}
            </CardDescription>
          </div>
          <TierStatusBadge item={item} />
        </div>
      </CardHeader>
      <CardContent className='grid gap-4'>
        {rule.description && (
          <p className='text-muted-foreground text-sm leading-6'>
            {rule.description}
          </p>
        )}

        <div className='grid gap-3'>
          <div className='grid gap-1'>
            <div className='flex items-center justify-between gap-2 text-xs'>
              <span className='text-muted-foreground'>
                {t('Minimum top-up')}
              </span>
              <span className='font-medium'>
                {formatQuota(rule.min_topup_quota)}
              </span>
            </div>
            <Progress
              value={progressPercent(totalTopup, rule.min_topup_quota)}
            />
            {item.topup_remaining > 0 && (
              <div className='text-muted-foreground text-xs'>
                {t('Remaining')}: {formatQuota(item.topup_remaining)}
              </div>
            )}
          </div>

          <div className='grid gap-1'>
            <div className='flex items-center justify-between gap-2 text-xs'>
              <span className='text-muted-foreground'>
                {t('Minimum consumption')}
              </span>
              <span className='font-medium'>
                {formatQuota(rule.min_used_quota)}
              </span>
            </div>
            <Progress value={progressPercent(usedQuota, rule.min_used_quota)} />
            {item.used_remaining > 0 && (
              <div className='text-muted-foreground text-xs'>
                {t('Remaining')}: {formatQuota(item.used_remaining)}
              </div>
            )}
          </div>
        </div>

        <div className='grid grid-cols-2 gap-2 border-t pt-3 text-sm'>
          <div>
            <div className='text-muted-foreground text-xs'>
              {t('Model ratio')}
            </div>
            <div className='font-mono font-medium'>
              {ratioLabel(item.group_ratio)}
            </div>
          </div>
          <div>
            <div className='text-muted-foreground text-xs'>
              {t('Top-up ratio')}
            </div>
            <div className='font-mono font-medium'>
              {ratioLabel(item.topup_group_ratio)}
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

function MemberTierContent({ data }: { data: MemberTierEvaluation }) {
  const { t } = useTranslation()
  const currentTier = data.current_tier ?? data.target_tier
  const nextTier = data.next_tier

  if (!data.enabled || data.progress.length === 0) {
    return (
      <EmptyState
        icon={BadgeCheck}
        title={t('No member tiers configured')}
        description={t('Member tier benefits will appear here once enabled.')}
        bordered
      />
    )
  }

  return (
    <div className='mx-auto flex w-full max-w-7xl flex-col gap-4'>
      {data.skipped_by_active_subscription && (
        <Alert>
          <Info className='h-4 w-4' />
          <AlertDescription>
            {t(
              'Your active subscription group is currently taking priority over automatic tier upgrades.'
            )}
          </AlertDescription>
        </Alert>
      )}

      <div className='grid gap-3 md:grid-cols-2 xl:grid-cols-4'>
        <SummaryCard
          title={t('Current tier')}
          value={currentTier?.display_name || data.current_group}
          description={data.current_group}
          icon={BadgeCheck}
        />
        <SummaryCard
          title={t('Total top-up')}
          value={formatQuota(data.total_topup_quota)}
          icon={WalletCards}
        />
        <SummaryCard
          title={t('Total consumption')}
          value={formatQuota(data.used_quota)}
          icon={TrendingUp}
        />
        <SummaryCard
          title={t('Next tier')}
          value={nextTier?.display_name || t('Max tier reached')}
          description={nextTier?.group}
          icon={Layers3}
        />
      </div>

      <div className='grid gap-3 lg:grid-cols-2 xl:grid-cols-3'>
        {data.progress.map((item) => (
          <TierCard
            key={item.rule.group}
            item={item}
            totalTopup={data.total_topup_quota}
            usedQuota={data.used_quota}
          />
        ))}
      </div>
    </div>
  )
}

export function MemberTiers() {
  const { t } = useTranslation()
  const currentUser = useAuthStore((state) => state.auth.user)
  const setUser = useAuthStore((state) => state.auth.setUser)

  const query = useQuery({
    queryKey: ['member-tier-self'],
    queryFn: async () => {
      const res = await getMemberTierSelf()
      if (!res.success || !res.data) {
        toast.error(res.message || t('Failed to load member tiers'))
        return null
      }
      return res.data
    },
  })

  useEffect(() => {
    const data = query.data
    if (!data || !currentUser) return
    if (currentUser.group === data.current_group) return
    setUser({ ...currentUser, group: data.current_group })
  }, [currentUser, query.data, setUser])

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('Member Tiers')}</SectionPageLayout.Title>
      <SectionPageLayout.Description>
        {t('View your current tier and group-based benefits.')}
      </SectionPageLayout.Description>
      <SectionPageLayout.Content>
        {query.isLoading ? (
          <div className='text-muted-foreground flex min-h-80 items-center justify-center gap-2 text-sm'>
            <Loader2 className='size-4 animate-spin' />
            {t('Loading...')}
          </div>
        ) : query.data ? (
          <MemberTierContent data={query.data} />
        ) : (
          <EmptyState
            icon={BadgeCheck}
            title={t('Failed to load')}
            description={t('Please refresh and try again.')}
            bordered
          />
        )}
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
