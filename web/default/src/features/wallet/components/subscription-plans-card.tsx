import { Crown, RefreshCw, Sparkles, Check } from 'lucide-react'
import { useState, useEffect, useMemo, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import {
  StatusBadge,
  dotColorMap,
  textColorMap,
} from '@/components/status-badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Progress } from '@/components/ui/progress'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { TitledCard } from '@/components/ui/titled-card'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import {
  getPublicPlans,
  getSelfSubscriptionFull,
  createSubscriptionToken,
} from '@/features/subscriptions/api'
import { SubscriptionPurchaseDialog } from '@/features/subscriptions/components/dialogs/subscription-purchase-dialog'
import { formatDuration, formatResetPeriod } from '@/features/subscriptions/lib'
import type {
  PlanRecord,
  UserSubscriptionRecord,
} from '@/features/subscriptions/types'
import { cn } from '@/lib/utils'

import type { PaymentMethod, TopupInfo } from '../types'

interface SubscriptionPlansCardProps {
  topupInfo: TopupInfo | null
  onAvailabilityChange?: (available: boolean) => void
  userQuota?: number
  onPurchaseSuccess?: () => void | Promise<void>
  standalone?: boolean
}

type SubscriptionBillingMode = 'quota' | 'token'

function getEpayMethods(payMethods: PaymentMethod[] = []): PaymentMethod[] {
  return payMethods.filter(
    (m) => m?.type && m.type !== 'stripe' && m.type !== 'creem'
  )
}

function formatTokenCount(value: number): string {
  if (value >= 1_000_000_000) return `${(value / 1_000_000_000).toFixed(2)}B`
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(2)}M`
  if (value >= 1_000) return `${(value / 1_000).toFixed(1)}K`
  return String(Math.max(0, Math.round(value)))
}

export function SubscriptionPlansCard({
  topupInfo,
  onAvailabilityChange,
  userQuota,
  onPurchaseSuccess,
  standalone = false,
}: SubscriptionPlansCardProps) {
  const { t } = useTranslation()

  const [plans, setPlans] = useState<PlanRecord[]>([])
  const [activeSubscriptions, setActiveSubscriptions] = useState<
    UserSubscriptionRecord[]
  >([])
  const [allSubscriptions, setAllSubscriptions] = useState<
    UserSubscriptionRecord[]
  >([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)

  const [purchaseOpen, setPurchaseOpen] = useState(false)
  const [selectedPlan, setSelectedPlan] = useState<PlanRecord | null>(null)
  const [tokenSubscriptionId, setTokenSubscriptionId] = useState<number | null>(
    null
  )
  const [tokenName, setTokenName] = useState('')
  const [createdToken, setCreatedToken] = useState<string | null>(null)
  const [creatingToken, setCreatingToken] = useState(false)
  const [billingMode, setBillingMode] =
    useState<SubscriptionBillingMode>('quota')

  const enableStripe = !!topupInfo?.enable_stripe_topup
  const enableCreem = !!topupInfo?.enable_creem_topup
  const enableWaffoPancake = !!topupInfo?.enable_waffo_pancake_topup
  const enableOnlineTopUp = !!topupInfo?.enable_online_topup
  const epayMethods = useMemo(
    () => getEpayMethods(topupInfo?.pay_methods),
    [topupInfo?.pay_methods]
  )

  const fetchPlans = useCallback(async () => {
    try {
      const res = await getPublicPlans()
      if (res.success) {
        setPlans(res.data || [])
      }
    } catch {
      setPlans([])
    }
  }, [])

  const fetchSelfSubscription = useCallback(async () => {
    try {
      const res = await getSelfSubscriptionFull()
      if (res.success && res.data) {
        setActiveSubscriptions(res.data.subscriptions || [])
        setAllSubscriptions(res.data.all_subscriptions || [])
      }
    } catch {
      // ignore
    }
  }, [])

  useEffect(() => {
    const init = async () => {
      setLoading(true)
      await Promise.all([fetchPlans(), fetchSelfSubscription()])
      setLoading(false)
    }
    init()
  }, [fetchPlans, fetchSelfSubscription])

  const handleRefresh = async () => {
    setRefreshing(true)
    try {
      await fetchSelfSubscription()
    } finally {
      setRefreshing(false)
    }
  }

  const hasActive = activeSubscriptions.length > 0
  const hasAny = allSubscriptions.length > 0
  const isAvailable = loading || plans.length > 0 || hasAny

  const planPurchaseCountMap = useMemo(() => {
    const map = new Map<number, number>()
    for (const sub of allSubscriptions) {
      const planId = sub?.subscription?.plan_id
      if (!planId) continue
      map.set(planId, (map.get(planId) || 0) + 1)
    }
    return map
  }, [allSubscriptions])

  const hasQuotaPlans = useMemo(
    () => plans.some((item) => item?.plan?.billing_mode !== 'token'),
    [plans]
  )
  const hasTokenPlans = useMemo(
    () => plans.some((item) => item?.plan?.billing_mode === 'token'),
    [plans]
  )
  const visiblePlans = useMemo(
    () =>
      plans.filter((item) =>
        billingMode === 'token'
          ? item?.plan?.billing_mode === 'token'
          : item?.plan?.billing_mode !== 'token'
      ),
    [billingMode, plans]
  )

  useEffect(() => {
    if (billingMode === 'quota' && !hasQuotaPlans && hasTokenPlans) {
      setBillingMode('token')
    } else if (billingMode === 'token' && !hasTokenPlans && hasQuotaPlans) {
      setBillingMode('quota')
    }
  }, [billingMode, hasQuotaPlans, hasTokenPlans])

  useEffect(() => {
    onAvailabilityChange?.(isAvailable)
  }, [isAvailable, onAvailabilityChange])

  const planTitleMap = useMemo(() => {
    const map = new Map<number, string>()
    for (const p of plans) {
      if (p?.plan?.id) {
        map.set(p.plan.id, p.plan.title || '')
      }
    }
    return map
  }, [plans])

  const getRemainingDays = (sub: UserSubscriptionRecord) => {
    const endTime = sub?.subscription?.end_time || 0
    if (!endTime) return 0
    const now = Date.now() / 1000
    return Math.max(0, Math.ceil((endTime - now) / 86400))
  }

  const getUsagePercent = (sub: UserSubscriptionRecord) => {
    const total = Number(sub?.subscription?.amount_total || 0)
    const used = Number(sub?.subscription?.amount_used || 0)
    if (total <= 0) return 0
    return Math.round((used / total) * 100)
  }

  const formatCredits = (value: number): string =>
    new Intl.NumberFormat(undefined, { maximumFractionDigits: 0 }).format(
      Math.max(0, Math.round(value))
    )

  if (loading) {
    return (
      <Card data-card-hover='false' className='gap-0 overflow-hidden py-0'>
        <CardHeader className='border-b p-3 !pb-3 sm:p-5 sm:!pb-5'>
          <Skeleton className='h-6 w-32' />
        </CardHeader>
        <CardContent className='space-y-4 p-3 sm:p-5'>
          <Skeleton className='h-20 w-full' />
          <div className='grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3'>
            {[1, 2, 3].map((skeleton) => (
              <Skeleton
                key={`subscription-skeleton-${skeleton}`}
                className='h-48 w-full'
              />
            ))}
          </div>
        </CardContent>
      </Card>
    )
  }

  if (plans.length === 0 && !hasAny && !standalone) {
    return null
  }

  return (
    <>
      <TitledCard
        title={t('Subscription Plans')}
        description={t('Subscribe to a plan for model access')}
        icon={<Crown className='h-4 w-4' />}
        disableHoverEffect
        contentClassName='space-y-4 sm:space-y-5'
      >
        <div className='bg-muted/20 flex flex-wrap items-center justify-between gap-3 rounded-xl border p-3 sm:p-4'>
          <div className='min-w-0'>
            <p className='text-sm font-medium'>{t('Billing unit')}</p>
            <p className='text-muted-foreground mt-1 text-xs'>
              {t('1 quota = 1 Credit')}
            </p>
          </div>
          <ToggleGroup
            value={[billingMode]}
            onValueChange={(value) => {
              const nextMode = value[0] as SubscriptionBillingMode | undefined
              if (nextMode) setBillingMode(nextMode)
            }}
            variant='default'
            size='sm'
            aria-label={t('Billing unit')}
            className='bg-background h-8 rounded-[6px] border p-0.5'
          >
            <ToggleGroupItem
              value='quota'
              disabled={!hasQuotaPlans}
              className='aria-pressed:bg-muted h-7 rounded-[4px] px-3 aria-pressed:shadow-sm'
            >
              {t('Credits')}
            </ToggleGroupItem>
            <ToggleGroupItem
              value='token'
              disabled={!hasTokenPlans}
              className='aria-pressed:bg-muted h-7 rounded-[4px] px-3 aria-pressed:shadow-sm'
            >
              {t('Subscription Tokens')}
            </ToggleGroupItem>
          </ToggleGroup>
        </div>

        {/* Subscription balances are accessed through subscription keys only. */}
        <div className='rounded-xl border p-3 sm:p-4'>
          <div className='flex flex-wrap items-center justify-between gap-2.5 sm:gap-3'>
            <div className='flex min-w-0 flex-col gap-1'>
              <div className='flex flex-wrap items-center gap-2'>
                <span className='text-sm font-medium'>
                  {t('My Subscriptions')}
                </span>
                <span className='flex items-center gap-1.5 text-xs font-medium'>
                  <span
                    className={cn(
                      'size-1.5 shrink-0 rounded-full',
                      hasActive ? dotColorMap.success : dotColorMap.neutral
                    )}
                    aria-hidden='true'
                  />
                  {hasActive ? (
                    <span className={cn(textColorMap.success)}>
                      {activeSubscriptions.length} {t('active')}
                    </span>
                  ) : (
                    <span className='text-muted-foreground'>
                      {t('No Active')}
                    </span>
                  )}
                  {allSubscriptions.length > activeSubscriptions.length && (
                    <>
                      <span className='text-muted-foreground/30'>·</span>
                      <span className='text-muted-foreground'>
                        {allSubscriptions.length - activeSubscriptions.length}{' '}
                        {t('expired')}
                      </span>
                    </>
                  )}
                </span>
              </div>
              <p className='text-muted-foreground text-xs'>
                {t(
                  'Subscription plans are billed only through their dedicated keys. General API keys use wallet balance.'
                )}
              </p>
            </div>
            <div className='flex items-center gap-2'>
              <Button
                variant='ghost'
                size='icon'
                className='h-8 w-8'
                onClick={handleRefresh}
                disabled={refreshing}
              >
                <RefreshCw
                  className={`h-3.5 w-3.5 ${refreshing ? 'animate-spin' : ''}`}
                />
              </Button>
            </div>
          </div>

          {hasAny && (
            <>
              <Separator className='my-3' />
              <div className='max-h-64 space-y-3 overflow-y-auto pr-1'>
                {allSubscriptions.map((sub) => {
                  const subscription = sub.subscription
                  const totalAmount = Number(subscription?.amount_total || 0)
                  const usedAmount = Number(subscription?.amount_used || 0)
                  const remainAmount =
                    totalAmount > 0 ? Math.max(0, totalAmount - usedAmount) : 0
                  const planTitle =
                    planTitleMap.get(subscription?.plan_id) || ''
                  const remainDays = getRemainingDays(sub)
                  const usagePercent = getUsagePercent(sub)
                  const now = Date.now() / 1000
                  const isExpired = (subscription?.end_time || 0) < now
                  const isCancelled = subscription?.status === 'cancelled'
                  const isActive =
                    subscription?.status === 'active' && !isExpired
                  let statusLabel = t('Expired')
                  let timeLabel = t('Expired at')
                  if (isActive) {
                    statusLabel = t('Active')
                    timeLabel = t('Until')
                  } else if (isCancelled) {
                    statusLabel = t('Cancelled')
                    timeLabel = t('Cancelled at')
                  }

                  return (
                    <div
                      key={subscription?.id}
                      className='bg-background rounded-md border p-3 text-xs'
                    >
                      <div className='flex items-center justify-between'>
                        <div className='flex items-center gap-2'>
                          <span className='font-medium'>
                            {planTitle
                              ? `${planTitle} · ${t('Subscription')} #${subscription?.id}`
                              : `${t('Subscription')} #${subscription?.id}`}
                          </span>
                          <StatusBadge
                            label={statusLabel}
                            variant={isActive ? 'success' : 'neutral'}
                            copyable={false}
                          />
                        </div>
                        {isActive && (
                          <span className='text-muted-foreground'>
                            {t('{{count}} days remaining', {
                              count: remainDays,
                            })}
                          </span>
                        )}
                      </div>
                      <div className='text-muted-foreground mt-1.5'>
                        {timeLabel}{' '}
                        {new Date(
                          (subscription?.end_time || 0) * 1000
                        ).toLocaleString()}
                      </div>
                      {isActive && (subscription?.next_reset_time ?? 0) > 0 && (
                        <div className='text-muted-foreground mt-1'>
                          {t('Next reset')}:{' '}
                          {new Date(
                            (subscription?.next_reset_time ?? 0) * 1000
                          ).toLocaleString()}
                        </div>
                      )}
                      <div className='text-muted-foreground mt-1'>
                        {t('Allowed Groups')}:{' '}
                        {(subscription?.allowed_groups || []).length > 0
                          ? subscription?.allowed_groups?.join(', ')
                          : t('All groups')}
                      </div>
                      {subscription?.billing_mode === 'token' &&
                        (subscription.allowed_models || []).length > 0 && (
                          <div className='text-muted-foreground mt-1 wrap-anywhere'>
                            {t('Allowed Models')}:{' '}
                            {subscription.allowed_models?.join(', ')}
                          </div>
                        )}
                      {isActive && subscription && (
                        <div className='mt-2 flex items-center justify-between gap-2'>
                          <span className='text-muted-foreground'>
                            {t('Subscription key')}
                          </span>
                          {isActive && (
                            <Button
                              variant='outline'
                              size='sm'
                              onClick={() => {
                                setTokenSubscriptionId(subscription.id)
                                setTokenName('')
                                setCreatedToken(null)
                              }}
                            >
                              {t('Create subscription key')}
                            </Button>
                          )}
                        </div>
                      )}
                      <div className='text-muted-foreground mt-1'>
                        {subscription?.billing_mode === 'token'
                          ? t('Total Subscription Tokens')
                          : t('Credits')}
                        :{' '}
                        {totalAmount > 0 ? (
                          <Tooltip>
                            <TooltipTrigger
                              render={<span className='cursor-help' />}
                            >
                              {subscription?.billing_mode === 'token'
                                ? `${formatTokenCount(usedAmount)}/${formatTokenCount(totalAmount)} · ${t('Remaining')} ${formatTokenCount(remainAmount)}`
                                : `${formatCredits(usedAmount)}/${formatCredits(totalAmount)} · ${t('Remaining')} ${formatCredits(remainAmount)}`}
                            </TooltipTrigger>
                            <TooltipContent>
                              {subscription?.billing_mode === 'token'
                                ? t('Total Subscription Tokens')
                                : t('Credits')}
                              : {usedAmount}/{totalAmount} · {t('Remaining')}{' '}
                              {remainAmount}
                            </TooltipContent>
                          </Tooltip>
                        ) : (
                          t('Unlimited')
                        )}
                        {totalAmount > 0 && (
                          <span className='ml-2'>
                            {t('Used')} {usagePercent}%
                          </span>
                        )}
                      </div>
                      {totalAmount > 0 && isActive && (
                        <Progress value={usagePercent} className='mt-2 h-1.5' />
                      )}
                    </div>
                  )
                })}
              </div>
            </>
          )}

          {!hasAny && (
            <p className='text-muted-foreground mt-2 text-xs'>
              {t('Subscribe to a plan for model access')}
            </p>
          )}
        </div>

        {/* Available plans grid */}
        {visiblePlans.length > 0 ? (
          <div className='grid grid-cols-1 gap-3 2xl:grid-cols-2 2xl:gap-4'>
            {visiblePlans.map((p, index) => {
              const plan = p?.plan
              if (!plan) return null
              const totalAmount = Number(plan.total_amount || 0)
              const price = Number(plan.price_amount || 0).toFixed(2)
              const isPopular = index === 0 && visiblePlans.length > 1
              const limit = Number(plan.max_purchase_per_user || 0)
              const count = planPurchaseCountMap.get(plan.id) || 0
              const reached = limit > 0 && count >= limit

              const benefits = [
                `${t('Validity Period')}: ${formatDuration(plan, t)}`,
                formatResetPeriod(plan, t) !== t('No Reset')
                  ? `${t('Quota Reset')}: ${formatResetPeriod(plan, t)}`
                  : null,
                totalAmount > 0
                  ? `${plan.billing_mode === 'token' ? t('Total Subscription Tokens') : t('Credits')}: ${plan.billing_mode === 'token' ? formatTokenCount(totalAmount) : formatCredits(totalAmount)}`
                  : `${plan.billing_mode === 'token' ? t('Total Subscription Tokens') : t('Credits')}: ${t('Unlimited')}`,
                limit > 0 ? `${t('Purchase Limit')}: ${limit}` : null,
                plan.upgrade_group
                  ? `${t('Upgrade Group')}: ${plan.upgrade_group}`
                  : null,
                (plan.allowed_groups || []).length > 0
                  ? `${t('Allowed Groups')}: ${(plan.allowed_groups || []).join(', ')}`
                  : `${t('Allowed Groups')}: ${t('All groups')}`,
                plan.billing_mode === 'token' &&
                (plan.allowed_models || []).length > 0
                  ? `${t('Allowed Models')}: ${(plan.allowed_models || []).join(', ')}`
                  : null,
              ].filter(Boolean) as string[]

              return (
                <Card
                  key={plan.id}
                  data-card-hover='false'
                  className={cn(isPopular && 'border-primary/70 shadow-sm')}
                >
                  <CardContent className='flex h-full flex-col p-3.5 sm:p-4'>
                    <div className='mb-2 flex items-start justify-between gap-3'>
                      <div className='min-w-0'>
                        <h4 className='truncate font-semibold'>
                          {plan.title || t('Subscription Plans')}
                        </h4>
                        {plan.subtitle && (
                          <p className='text-muted-foreground truncate text-xs'>
                            {plan.subtitle}
                          </p>
                        )}
                      </div>
                      {isPopular && (
                        <StatusBadge
                          variant='info'
                          copyable={false}
                          className='shrink-0'
                        >
                          <Sparkles className='h-3 w-3' />
                          {t('Recommended')}
                        </StatusBadge>
                      )}
                    </div>

                    <div className='py-2'>
                      <span className='text-primary text-2xl font-bold'>
                        ${price}
                      </span>
                    </div>

                    <div className='flex-1 space-y-1.5 pb-3'>
                      {benefits.map((label) => (
                        <div
                          key={label}
                          className='text-muted-foreground flex items-center gap-2 text-xs'
                        >
                          <Check className='text-primary h-3 w-3 shrink-0' />
                          <span className='min-w-0 wrap-anywhere'>{label}</span>
                        </div>
                      ))}
                    </div>

                    <Separator className='mb-3' />

                    {reached ? (
                      <Tooltip>
                        <TooltipTrigger render={<div />}>
                          <Button variant='outline' className='w-full' disabled>
                            {t('Limit Reached')}
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent>
                          {t('Purchase limit reached')} ({count}/{limit})
                        </TooltipContent>
                      </Tooltip>
                    ) : (
                      <Button
                        variant='outline'
                        className='w-full'
                        onClick={() => {
                          setSelectedPlan(p)
                          setPurchaseOpen(true)
                        }}
                      >
                        {t('Subscribe Now')}
                      </Button>
                    )}
                  </CardContent>
                </Card>
              )
            })}
          </div>
        ) : (
          <p className='text-muted-foreground py-4 text-center text-sm'>
            {t('No plans available for this billing unit')}
          </p>
        )}
      </TitledCard>

      <Dialog
        open={tokenSubscriptionId !== null}
        onOpenChange={(open) => {
          if (!open) {
            setTokenSubscriptionId(null)
            setCreatedToken(null)
          }
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('Create subscription key')}</DialogTitle>
            <DialogDescription>
              {t('This key shares the subscription allowance.')}
            </DialogDescription>
          </DialogHeader>
          {createdToken ? (
            <div className='space-y-2'>
              <p className='text-muted-foreground text-sm'>
                {t('Copy this key now. It will not be shown again.')}
              </p>
              <Input
                value={createdToken}
                readOnly
                onFocus={(e) => e.currentTarget.select()}
              />
            </div>
          ) : (
            <Input
              value={tokenName}
              onChange={(e) => setTokenName(e.target.value)}
              placeholder={t('Key name')}
              maxLength={50}
            />
          )}
          <DialogFooter>
            <Button
              disabled={creatingToken || !tokenName.trim() || !!createdToken}
              onClick={async () => {
                if (!tokenSubscriptionId || !tokenName.trim()) return
                setCreatingToken(true)
                try {
                  const res = await createSubscriptionToken(
                    tokenSubscriptionId,
                    tokenName.trim()
                  )
                  if (res.success && res.data?.token) {
                    setCreatedToken(res.data.token)
                    toast.success(t('Created successfully'))
                  } else {
                    toast.error(res.message || t('Request failed'))
                  }
                } catch {
                  toast.error(t('Request failed'))
                } finally {
                  setCreatingToken(false)
                }
              }}
            >
              {t('Create')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <SubscriptionPurchaseDialog
        open={purchaseOpen}
        onOpenChange={(open) => {
          setPurchaseOpen(open)
          if (!open) {
            fetchSelfSubscription()
          }
        }}
        plan={selectedPlan}
        enableStripe={enableStripe}
        enableCreem={enableCreem}
        enableWaffoPancake={enableWaffoPancake}
        enableOnlineTopUp={enableOnlineTopUp}
        epayMethods={epayMethods}
        userQuota={userQuota}
        onPurchaseSuccess={onPurchaseSuccess}
        purchaseLimit={
          selectedPlan?.plan?.max_purchase_per_user
            ? Number(selectedPlan.plan.max_purchase_per_user)
            : undefined
        }
        purchaseCount={
          selectedPlan?.plan?.id
            ? planPurchaseCountMap.get(selectedPlan.plan.id)
            : undefined
        }
      />
    </>
  )
}
