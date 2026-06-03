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
import { useEffect, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  BadgeCheck,
  CheckCircle2,
  Circle,
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
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
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

function tierStatus(item: MemberTierProgress) {
  if (item.current) return 'current'
  if (item.next) return 'next'
  if (item.qualified) return 'qualified'
  return 'locked'
}

function TierStatusBadge({ item }: { item: MemberTierProgress }) {
  const { t } = useTranslation()

  if (item.current) {
    return (
      <Badge className='bg-foreground text-background hover:bg-foreground/90'>
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

function MetricCell({
  label,
  value,
  helper,
  icon,
}: {
  label: string
  value: string
  helper?: string
  icon: ReactNode
}) {
  return (
    <div className='min-h-20 border-b px-4 py-3 sm:border-r lg:border-b-0'>
      <div className='grid grid-cols-[1fr_auto] items-start gap-3'>
        <div className='min-w-0'>
          <div className='text-muted-foreground text-xs'>{label}</div>
          <div className='mt-1 truncate text-base font-semibold tracking-tight'>
            {value}
          </div>
          {helper && (
            <div className='text-muted-foreground mt-2 truncate text-xs'>
              {helper}
            </div>
          )}
        </div>
        <div className='bg-muted text-muted-foreground flex size-7 items-center justify-center rounded-md'>
          {icon}
        </div>
      </div>
    </div>
  )
}

function RequirementProgress({
  label,
  total,
  target,
  remaining,
}: {
  label: string
  total: number
  target: number
  remaining: number
}) {
  const { t } = useTranslation()

  return (
    <div className='grid gap-1.5'>
      <div className='flex items-center justify-between gap-3 text-xs'>
        <span className='text-muted-foreground'>{label}</span>
        <span className='font-medium'>{formatQuota(target)}</span>
      </div>
      <Progress
        value={progressPercent(total, target)}
        className='gap-1 [&_[data-slot=progress-track]]:h-1.5'
      />
      <div className='text-muted-foreground flex items-center justify-between gap-3 text-xs'>
        <span>{formatQuota(total)}</span>
        <span>
          {t('Remaining')}: {formatQuota(remaining)}
        </span>
      </div>
    </div>
  )
}

function TierRail({ progress }: { progress: MemberTierProgress[] }) {
  const { t } = useTranslation()

  return (
    <aside className='bg-muted/[0.18] hidden border-r p-4 lg:block'>
      <div className='grid gap-4'>
        {progress.map((item, index) => {
          const status = tierStatus(item)
          return (
            <div
              key={item.rule.group}
              className='relative grid gap-1 pl-5 text-sm'
            >
              {index < progress.length - 1 && (
                <span className='bg-border absolute top-4 bottom-[-18px] left-[5px] w-px' />
              )}
              <span
                className={cn(
                  'absolute top-1.5 left-0 size-2.5 rounded-full border-2 bg-background',
                  status === 'current' && 'border-foreground bg-foreground',
                  status === 'next' && 'border-primary bg-primary',
                  status === 'qualified' && 'border-emerald-500 bg-emerald-500',
                  status === 'locked' && 'border-muted-foreground/40'
                )}
              />
              <div className='flex min-w-0 items-center gap-2'>
                <span className='truncate font-medium'>
                  {item.rule.display_name || item.rule.group}
                </span>
                {item.current && (
                  <span className='bg-foreground text-background rounded-full px-1.5 py-0.5 text-[10px] font-medium'>
                    {t('Current')}
                  </span>
                )}
              </div>
              <div className='text-muted-foreground truncate text-xs'>
                {item.rule.group}
              </div>
              <div className='text-muted-foreground mt-1 grid gap-0.5 text-xs leading-5'>
                <span>
                  {t('Minimum top-up')}: {formatQuota(item.rule.min_topup_quota)}
                </span>
                <span>
                  {t('Minimum consumption')}:{' '}
                  {formatQuota(item.rule.min_used_quota)}
                </span>
              </div>
            </div>
          )
        })}
      </div>
    </aside>
  )
}

function RatioBlock({ label, value }: { label: string; value: number }) {
  return (
    <div className='bg-muted/40 rounded-md px-2.5 py-2'>
      <div className='text-muted-foreground truncate text-xs'>{label}</div>
      <div className='mt-1 font-mono text-sm font-semibold'>
        x{ratioLabel(value)}
      </div>
    </div>
  )
}

function TierMatrixRow({
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
    <TableRow
      className={cn(
        'hover:bg-muted/30',
        item.current && 'bg-foreground/[0.035] hover:bg-foreground/[0.05]',
        item.next && !item.current && 'bg-primary/[0.035]'
      )}
    >
      <TableCell className='w-60 px-4 py-3 align-top whitespace-normal'>
        <div className='flex items-start gap-3'>
          <div
            className={cn(
              'mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-md border',
              item.current && 'border-foreground bg-foreground text-background',
              item.next &&
                !item.current &&
                'border-primary/30 bg-primary/10 text-primary'
            )}
          >
            {item.current || item.qualified ? (
              <CheckCircle2 className='size-4' />
            ) : (
              <Circle className='size-4' />
            )}
          </div>
          <div className='min-w-0'>
            <div className='truncate font-semibold'>
              {rule.display_name || rule.group}
            </div>
            <div className='text-muted-foreground truncate text-xs'>
              {rule.group}
            </div>
            {rule.description && (
              <div className='text-muted-foreground mt-2 line-clamp-2 text-xs leading-5'>
                {rule.description}
              </div>
            )}
          </div>
        </div>
      </TableCell>
      <TableCell className='min-w-[20rem] px-4 py-3 align-top whitespace-normal'>
        <div className='grid gap-3'>
          <RequirementProgress
            label={t('Minimum top-up')}
            total={totalTopup}
            target={rule.min_topup_quota}
            remaining={item.topup_remaining}
          />
          <RequirementProgress
            label={t('Minimum consumption')}
            total={usedQuota}
            target={rule.min_used_quota}
            remaining={item.used_remaining}
          />
        </div>
      </TableCell>
      <TableCell className='w-52 px-4 py-3 align-top'>
        <div className='grid grid-cols-2 gap-2'>
          <RatioBlock label={t('Model ratio')} value={item.group_ratio} />
          <RatioBlock label={t('Top-up ratio')} value={item.topup_group_ratio} />
        </div>
      </TableCell>
      <TableCell className='w-28 px-4 py-3 text-right align-top'>
        <TierStatusBadge item={item} />
      </TableCell>
    </TableRow>
  )
}

function MemberTierContent({ data }: { data: MemberTierEvaluation }) {
  const { t } = useTranslation()
  const currentTier = data.current_tier ?? data.target_tier
  const nextTier = data.next_tier
  const currentItem =
    data.progress.find((item) => item.current) ??
    data.progress.find((item) => item.rule.group === data.current_group) ??
    data.progress.find((item) => item.rule.group === data.target_group) ??
    data.progress[0]
  const nextItem = nextTier
    ? data.progress.find((item) => item.rule.group === nextTier.group)
    : undefined

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
    <div className='flex w-full max-w-none flex-col gap-4'>
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

      <section className='bg-background overflow-hidden rounded-lg border shadow-xs'>
        <div className='flex flex-wrap items-center justify-between gap-3 border-b px-4 py-3'>
          <div className='flex min-w-0 items-center gap-2'>
            <h3 className='truncate text-lg font-semibold tracking-tight'>
              {t('Member Tiers')}
            </h3>
            <Badge variant='secondary'>
              {currentTier?.display_name || data.current_group}
            </Badge>
          </div>
          <div className='text-muted-foreground text-xs'>
            {t('View tier progress and group-based benefits.')}
          </div>
        </div>

        <div className='grid sm:grid-cols-2 lg:grid-cols-4 lg:[&>*:last-child]:border-r-0'>
          <MetricCell
            label={t('Current tier')}
            value={currentTier?.display_name || data.current_group}
            helper={data.current_group}
            icon={<BadgeCheck className='size-4' />}
          />
          <MetricCell
            label={t('Total top-up')}
            value={formatQuota(data.total_topup_quota)}
            icon={<WalletCards className='size-4' />}
          />
          <MetricCell
            label={t('Total consumption')}
            value={formatQuota(data.used_quota)}
            icon={<TrendingUp className='size-4' />}
          />
          <MetricCell
            label={t('Next tier')}
            value={nextTier?.display_name || t('Max tier reached')}
            helper={nextTier?.group}
            icon={<Layers3 className='size-4' />}
          />
        </div>

        <div className='grid lg:grid-cols-[260px_minmax(0,1fr)]'>
          <TierRail progress={data.progress} />
          <div className='min-w-0'>
            <Table className='min-w-[880px]'>
              <TableHeader>
                <TableRow className='bg-muted/35 hover:bg-muted/35'>
                  <TableHead className='h-11 px-4'>{t('Tier name')}</TableHead>
                  <TableHead className='h-11 px-4'>
                    {t('Member Tier Rules')}
                  </TableHead>
                  <TableHead className='h-11 px-4'>{t('Benefits')}</TableHead>
                  <TableHead className='h-11 px-4 text-right'>
                    {t('Status')}
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data.progress.map((item) => (
                  <TierMatrixRow
                    key={item.rule.group}
                    item={item}
                    totalTopup={data.total_topup_quota}
                    usedQuota={data.used_quota}
                  />
                ))}
              </TableBody>
            </Table>
          </div>
        </div>
      </section>

      <section className='bg-background overflow-hidden rounded-lg border'>
        <div className='flex flex-wrap items-center gap-2 border-b px-4 py-3'>
          <h3 className='text-lg font-semibold tracking-tight'>
            {t('Benefits')}
          </h3>
          <Info className='text-info size-4' />
          <span className='text-muted-foreground text-xs'>
            {t('Tier progress and group-based benefits')}
          </span>
        </div>
        <Table className='min-w-[640px]'>
          <TableHeader>
            <TableRow className='bg-muted/35 hover:bg-muted/35'>
              <TableHead className='px-4'>{t('Benefits')}</TableHead>
              <TableHead className='px-4'>{t('Current tier')}</TableHead>
              <TableHead className='px-4'>{t('Next tier')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow>
              <TableCell className='px-4 font-medium'>
                {t('Model ratio')}
              </TableCell>
              <TableCell className='px-4 font-mono'>
                x{ratioLabel(currentItem?.group_ratio)}
              </TableCell>
              <TableCell className='px-4 font-mono'>
                {nextItem
                  ? `x${ratioLabel(nextItem.group_ratio)}`
                  : t('Max tier reached')}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell className='px-4 font-medium'>
                {t('Top-up ratio')}
              </TableCell>
              <TableCell className='px-4 font-mono'>
                x{ratioLabel(currentItem?.topup_group_ratio)}
              </TableCell>
              <TableCell className='px-4 font-mono'>
                {nextItem
                  ? `x${ratioLabel(nextItem.topup_group_ratio)}`
                  : t('Max tier reached')}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell className='px-4 font-medium'>
                {t('Minimum top-up')}
              </TableCell>
              <TableCell className='px-4'>
                {formatQuota(currentItem?.rule.min_topup_quota ?? 0)}
              </TableCell>
              <TableCell className='px-4'>
                {nextTier ? formatQuota(nextTier.min_topup_quota) : '-'}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell className='px-4 font-medium'>
                {t('Minimum consumption')}
              </TableCell>
              <TableCell className='px-4'>
                {formatQuota(currentItem?.rule.min_used_quota ?? 0)}
              </TableCell>
              <TableCell className='px-4'>
                {nextTier ? formatQuota(nextTier.min_used_quota) : '-'}
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </section>
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
