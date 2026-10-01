import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Check,
  ChevronLeft,
  ChevronRight,
  ClipboardCopy,
  Gift,
  Link2,
  UserPlus,
  Users,
  WalletCards,
} from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useCopyToClipboard } from '@/hooks/use-copy-to-clipboard'
import { getUserAvatarFallback, getUserAvatarStyle } from '@/lib/avatar'
import { formatPercent, formatQuota, formatTimestampToDate } from '@/lib/format'
import { useAuthStore } from '@/stores/auth-store'

import { getUserProfile } from '../profile/api'
import { getAffiliateCode, transferAffiliateQuota } from '../wallet/api'
import { generateAffiliateLink } from '../wallet/lib'
import {
  getFriendRewards,
  getFriends,
  getFriendsSummary,
  transferFriendRewards,
} from './api'
import type { FriendRecord, FriendsSummary, RewardRecord } from './types'

const PAGE_SIZE = 20

function unwrapList<T>(value: { items?: T[]; list?: T[] } | undefined): T[] {
  return value?.items ?? value?.list ?? []
}

function numberValue(value: unknown): number {
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : 0
}

function displayDate(value: unknown): string {
  if (typeof value === 'number') return formatTimestampToDate(value)
  if (typeof value === 'string' && value) {
    const asNumber = Number(value)
    return Number.isFinite(asNumber) ? formatTimestampToDate(asNumber) : value
  }
  return '-'
}

function rewardStatusLabel(
  status: unknown,
  t: (key: string) => string
): string {
  switch (String(status ?? '').toLowerCase()) {
    case 'transferred':
    case 'completed':
    case 'credited':
      return t('Transferred')
    case 'cancelled':
    case 'reversed':
      return t('Cancelled')
    case 'no_reward':
      return t('No cashback')
    case 'pending':
      return t('Pending')
    default:
      return status ? String(status) : t('Pending')
  }
}

function rewardTypeLabel(reward: RewardRecord, t: (key: string) => string) {
  return String(reward.reward_type ?? '').toLowerCase() === 'registration'
    ? t('Registration rewards')
    : t('First top-up cashback')
}

export function FriendsHub() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const user = useAuthStore((state) => state.auth.user)
  const { copyToClipboard } = useCopyToClipboard()
  const [friendsPage, setFriendsPage] = useState(1)
  const [rewardsPage, setRewardsPage] = useState(1)
  const [copied, setCopied] = useState<'code' | 'link' | null>(null)

  const summaryQuery = useQuery({
    queryKey: ['friends-hub', 'summary'],
    queryFn: async () => (await getFriendsSummary()).data ?? {},
  })
  const profileQuery = useQuery({
    queryKey: ['friends-hub', 'profile'],
    queryFn: async () => (await getUserProfile()).data,
  })
  const codeQuery = useQuery({
    queryKey: ['friends-hub', 'affiliate-code'],
    queryFn: async () => (await getAffiliateCode()).data ?? '',
  })
  const friendsQuery = useQuery({
    queryKey: ['friends-hub', 'friends', friendsPage],
    queryFn: async () => (await getFriends(friendsPage, PAGE_SIZE)).data,
  })
  const rewardsQuery = useQuery({
    queryKey: ['friends-hub', 'rewards', rewardsPage],
    queryFn: async () => (await getFriendRewards(rewardsPage, PAGE_SIZE)).data,
  })
  const transferMutation = useMutation({
    mutationFn: () => transferFriendRewards(),
    onSuccess: (response) => {
      if (response.success === false) {
        toast.error(response.message || t('Transfer failed'))
        return
      }
      toast.success(response.message || t('Transfer successful'))
      void queryClient.invalidateQueries({ queryKey: ['friends-hub'] })
    },
  })
  const registrationTransferMutation = useMutation({
    mutationFn: () =>
      transferAffiliateQuota({
        quota: numberValue(summary.registration_pending_quota),
      }),
    onSuccess: (response) => {
      if (response.success === false) {
        toast.error(response.message || t('Transfer failed'))
        return
      }
      toast.success(response.message || t('Transfer successful'))
      void queryClient.invalidateQueries({ queryKey: ['friends-hub'] })
    },
  })

  const summary = (summaryQuery.data ?? {}) as FriendsSummary
  const inviteCode = String(
    summary.invite_code || summary.aff_code || codeQuery.data || ''
  )
  const inviteLink = String(
    summary.invite_link || (inviteCode ? generateAffiliateLink(inviteCode) : '')
  )
  const friends = unwrapList<FriendRecord>(friendsQuery.data)
  const rewards = unwrapList<RewardRecord>(rewardsQuery.data)
  const friendsTotal = numberValue(friendsQuery.data?.total)
  const rewardsTotal = numberValue(rewardsQuery.data?.total)
  const displayName = user?.display_name || user?.username || t('Anonymous')
  const avatarStyle = getUserAvatarStyle(displayName)
  const registrationPending = numberValue(summary.registration_pending_quota)
  const cashbackPending = numberValue(summary.cashback_pending_quota)
  const cashbackRate = numberValue(
    summary.referral_rate ?? summary.cashback_rate
  )
  const inviterRegistrationReward = numberValue(
    summary.inviter_registration_reward_quota
  )
  const inviteeRegistrationReward = numberValue(
    summary.invitee_registration_reward_quota
  )
  const totalRewards =
    numberValue(summary.registration_total_quota) +
    numberValue(summary.cashback_total_quota)

  const copy = async (kind: 'code' | 'link', value: string) => {
    if (!value) return
    await copyToClipboard(value)
    setCopied(kind)
    window.setTimeout(() => setCopied(null), 1200)
  }

  return (
    <div className='bg-muted/20 min-h-full overflow-auto'>
      <div className='mx-auto flex w-full max-w-7xl flex-col gap-5 p-3 sm:gap-6 sm:p-5 lg:p-6'>
        <div className='flex items-center justify-between gap-4 border-b pb-5'>
          <div className='flex min-w-0 items-center gap-3 sm:gap-4'>
            <Avatar size='lg' className='rounded-2xl'>
              <AvatarFallback
                className='rounded-2xl text-base font-semibold text-white'
                style={avatarStyle}
              >
                {getUserAvatarFallback(displayName)}
              </AvatarFallback>
            </Avatar>
            <div className='min-w-0'>
              <p className='text-muted-foreground text-xs font-medium tracking-[0.16em] uppercase'>
                {t('Friends Hub')}
              </p>
              <h1 className='truncate text-xl font-semibold tracking-tight sm:text-2xl'>
                {t('Hi')}, {displayName}
              </h1>
              {profileQuery.data?.created_time && (
                <p className='text-muted-foreground mt-0.5 text-xs'>
                  {t('Joined')}: {displayDate(profileQuery.data.created_time)}
                </p>
              )}
            </div>
          </div>
        </div>

        <Card className='overflow-hidden border-amber-200/80 bg-amber-50/70 dark:border-amber-900/60 dark:bg-amber-950/20'>
          <CardHeader className='gap-1 border-b border-amber-200/70 pb-4 dark:border-amber-900/50'>
            <CardTitle className='flex items-center gap-2 text-base'>
              <Gift className='size-4 text-amber-700 dark:text-amber-300' />
              {t('My referral invitation')}
            </CardTitle>
            <p className='text-muted-foreground text-sm'>
              {t(
                'Invite friends and earn rewards for their first successful top-up.'
              )}
            </p>
          </CardHeader>
          <CardContent className='grid gap-4 pt-5 md:grid-cols-[minmax(0,0.75fr)_minmax(0,1.25fr)]'>
            <div className='min-w-0'>
              <span className='text-muted-foreground text-xs'>
                {t('Invitation Code')}
              </span>
              <div className='mt-1 flex items-center gap-2'>
                <span className='min-w-0 truncate text-2xl font-semibold tracking-[0.12em]'>
                  {inviteCode || '-'}
                </span>
                <Button
                  variant='ghost'
                  size='icon'
                  className='size-8 shrink-0'
                  aria-label={t('Copy invitation code')}
                  title={t('Copy invitation code')}
                  onClick={() => void copy('code', inviteCode)}
                  disabled={!inviteCode}
                >
                  {copied === 'code' ? <Check /> : <ClipboardCopy />}
                </Button>
              </div>
            </div>
            <div className='min-w-0'>
              <span className='text-muted-foreground text-xs'>
                {t('Invitation link')}
              </span>
              <div className='mt-1 flex min-w-0 gap-2'>
                <div className='bg-background/80 flex min-h-9 min-w-0 flex-1 items-center truncate rounded-md border border-amber-200/80 px-3 text-xs dark:border-amber-900/60'>
                  {inviteLink || '-'}
                </div>
                <Button
                  variant='outline'
                  size='sm'
                  className='shrink-0 border-amber-300/80 bg-transparent'
                  onClick={() => void copy('link', inviteLink)}
                  disabled={!inviteLink}
                >
                  {copied === 'link' ? <Check /> : <Link2 />}
                  <span className='hidden sm:inline'>{t('Copy Link')}</span>
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className='gap-1 pb-4'>
            <CardTitle className='text-base'>{t('How rewards work')}</CardTitle>
            <p className='text-muted-foreground text-sm'>
              {t('Reward rules are configured by the administrator.')}
            </p>
          </CardHeader>
          <CardContent className='pt-0'>
            <div className='grid gap-5 md:grid-cols-3 md:divide-x'>
              <div className='border-l-2 border-amber-400 pl-4 md:border-l-0 md:pr-5'>
                <div className='flex flex-wrap items-center gap-2'>
                  <h3 className='text-sm font-semibold'>
                    {t('Inviter registration reward')}
                  </h3>
                  <Badge variant='outline'>{t('One-time')}</Badge>
                </div>
                <p className='mt-2 text-xl font-semibold tabular-nums'>
                  {formatQuota(inviterRegistrationReward)}
                </p>
                <p className='text-muted-foreground mt-1 text-xs leading-5'>
                  {t(
                    'Each invited friend can trigger this reward once after registering through your invitation.'
                  )}
                </p>
              </div>
              <div className='border-l-2 border-sky-400 pl-4 md:px-5'>
                <div className='flex flex-wrap items-center gap-2'>
                  <h3 className='text-sm font-semibold'>
                    {t('Invitee registration reward')}
                  </h3>
                  <Badge variant='outline'>{t('One-time')}</Badge>
                </div>
                <p className='mt-2 text-xl font-semibold tabular-nums'>
                  {formatQuota(inviteeRegistrationReward)}
                </p>
                <p className='text-muted-foreground mt-1 text-xs leading-5'>
                  {t(
                    'New users receive this one-time reward after registering with an invitation code.'
                  )}
                </p>
                <p className='text-muted-foreground mt-1 text-xs leading-5'>
                  {t(
                    'This reward is credited directly to the invited friend’s wallet.'
                  )}
                </p>
              </div>
              <div className='border-l-2 border-emerald-400 pl-4 md:pl-5'>
                <div className='flex flex-wrap items-center gap-2'>
                  <h3 className='text-sm font-semibold'>
                    {t('First top-up cashback')}
                  </h3>
                  <Badge variant='outline'>
                    {t('Eligible once per invited friend')}
                  </Badge>
                </div>
                <p className='mt-2 text-xl font-semibold tabular-nums'>
                  {cashbackRate > 0
                    ? formatPercent(cashbackRate)
                    : t('Not enabled')}
                </p>
                <p className='text-muted-foreground mt-1 text-xs leading-5'>
                  {t(
                    'Only the invited friend’s first successful top-up qualifies. Later top-ups do not generate cashback.'
                  )}
                </p>
                {cashbackRate > 0 && (
                  <p className='text-muted-foreground mt-1 text-xs leading-5'>
                    {t('Cashback is based on the final wallet credit.')}
                  </p>
                )}
              </div>
            </div>
          </CardContent>
        </Card>

        <div className='grid gap-3 sm:grid-cols-3'>
          {[
            {
              label: t('Invited friends'),
              value: numberValue(summary.invite_count),
              icon: Users,
            },
            {
              label: t('Completed first top-ups'),
              value: numberValue(
                summary.first_topup_count ?? summary.first_top_up_invite_count
              ),
              icon: UserPlus,
            },
            {
              label: t('Total rewards'),
              value: formatQuota(totalRewards),
              icon: WalletCards,
            },
          ].map((stat) => (
            <Card key={stat.label} className='border-border/70'>
              <CardContent className='flex items-center gap-3 p-4'>
                <div className='bg-muted flex size-9 shrink-0 items-center justify-center rounded-lg'>
                  <stat.icon className='text-muted-foreground size-4' />
                </div>
                <div className='min-w-0'>
                  <p className='text-muted-foreground truncate text-xs'>
                    {stat.label}
                  </p>
                  <p className='mt-1 truncate text-lg font-semibold tabular-nums'>
                    {stat.value}
                  </p>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>

        <Card>
          <CardHeader className='gap-1 pb-4'>
            <CardTitle className='text-base'>{t('Transfer Rewards')}</CardTitle>
            <p className='text-muted-foreground text-sm'>
              {t(
                'Registration rewards and first top-up cashback are tracked separately and can be transferred to your wallet.'
              )}
            </p>
          </CardHeader>
          <CardContent className='grid gap-3 pt-0 md:grid-cols-2'>
            <div className='bg-muted/20 flex items-center justify-between gap-4 rounded-lg border p-3'>
              <div className='min-w-0'>
                <p className='text-sm font-medium'>
                  {t('Registration rewards')}
                </p>
                <p className='text-muted-foreground mt-1 text-xs'>
                  {t('Registration pending')}
                </p>
                <p className='mt-0.5 text-lg font-semibold tabular-nums'>
                  {formatQuota(registrationPending)}
                </p>
              </div>
              <Button
                variant='outline'
                size='sm'
                className='shrink-0'
                disabled={
                  registrationTransferMutation.isPending ||
                  registrationPending <= 0
                }
                onClick={() => registrationTransferMutation.mutate()}
              >
                {t('Transfer to Balance')}
              </Button>
            </div>
            <div className='bg-muted/20 flex items-center justify-between gap-4 rounded-lg border p-3'>
              <div className='min-w-0'>
                <p className='text-sm font-medium'>
                  {t('First top-up cashback')}
                </p>
                <p className='text-muted-foreground mt-1 text-xs'>
                  {t('Cashback pending')}
                </p>
                <p className='mt-0.5 text-lg font-semibold tabular-nums'>
                  {formatQuota(cashbackPending)}
                </p>
              </div>
              <Button
                size='sm'
                className='shrink-0'
                disabled={transferMutation.isPending || cashbackPending <= 0}
                onClick={() => transferMutation.mutate()}
              >
                {t('Transfer to Balance')}
              </Button>
            </div>
          </CardContent>
        </Card>

        <Tabs defaultValue='friends' className='min-w-0'>
          <div className='flex flex-wrap items-center justify-between gap-3'>
            <TabsList variant='line'>
              <TabsTrigger value='friends'>
                <Users /> {t('My friends')}
                <Badge variant='secondary' className='ml-1'>
                  {friendsTotal || friends.length}
                </Badge>
              </TabsTrigger>
              <TabsTrigger value='rewards'>
                <Gift /> {t('Reward history')}
                <Badge variant='secondary' className='ml-1'>
                  {rewardsTotal || rewards.length}
                </Badge>
              </TabsTrigger>
            </TabsList>
          </div>
          <TabsContent value='friends' className='mt-3'>
            <Card>
              <CardContent className='p-0'>
                {friends.length === 0 ? (
                  <div className='text-muted-foreground px-4 py-12 text-center text-sm'>
                    {t('No invited friends yet')}
                  </div>
                ) : (
                  <div className='divide-y'>
                    <div className='text-muted-foreground hidden grid-cols-[minmax(0,1.5fr)_minmax(0,1fr)_minmax(0,1fr)_minmax(0,0.8fr)] gap-4 px-4 py-3 text-xs md:grid'>
                      <span>{t('Friend')}</span>
                      <span>{t('Registered')}</span>
                      <span>{t('First top-up')}</span>
                      <span>{t('Cashback')}</span>
                    </div>
                    {friends.map((friend, index) => (
                      <div
                        key={String(friend.id ?? friend.user_id ?? index)}
                        className='grid gap-2 px-4 py-3 md:grid-cols-[minmax(0,1.5fr)_minmax(0,1fr)_minmax(0,1fr)_minmax(0,0.8fr)] md:items-center md:gap-4'
                      >
                        <div className='min-w-0'>
                          <p className='truncate text-sm font-medium'>
                            {String(
                              friend.display_name ||
                                friend.username ||
                                friend.email ||
                                t('Anonymous')
                            )}
                          </p>
                          <p className='text-muted-foreground text-xs md:hidden'>
                            {t('Friend')}
                          </p>
                        </div>
                        <div>
                          <span className='text-muted-foreground mr-2 text-xs md:hidden'>
                            {t('Registered')}
                          </span>
                          <span className='text-sm'>
                            {displayDate(
                              friend.register_time ?? friend.created_at
                            )}
                          </span>
                        </div>
                        <div>
                          <span className='text-muted-foreground mr-2 text-xs md:hidden'>
                            {t('First top-up')}
                          </span>
                          <span className='text-sm'>
                            {rewardStatusLabel(
                              friend.first_topup_status ??
                                friend.first_top_up_status,
                              t
                            )}
                          </span>
                        </div>
                        <div>
                          <span className='text-muted-foreground mr-2 text-xs md:hidden'>
                            {t('Cashback')}
                          </span>
                          <span className='text-sm font-medium'>
                            {formatQuota(numberValue(friend.cashback_quota))}
                          </span>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
                {friendsTotal > PAGE_SIZE && (
                  <div className='flex justify-end gap-2 border-t p-3'>
                    <Button
                      size='icon'
                      variant='outline'
                      aria-label={t('Previous')}
                      title={t('Previous')}
                      disabled={friendsPage === 1}
                      onClick={() => setFriendsPage((page) => page - 1)}
                    >
                      <ChevronLeft />
                    </Button>
                    <Button
                      size='icon'
                      variant='outline'
                      aria-label={t('Next')}
                      title={t('Next')}
                      disabled={friendsPage * PAGE_SIZE >= friendsTotal}
                      onClick={() => setFriendsPage((page) => page + 1)}
                    >
                      <ChevronRight />
                    </Button>
                  </div>
                )}
              </CardContent>
            </Card>
          </TabsContent>
          <TabsContent value='rewards' className='mt-3'>
            <Card>
              <CardContent className='p-0'>
                {rewards.length === 0 ? (
                  <div className='text-muted-foreground px-4 py-12 text-center text-sm'>
                    {t('No rewards yet')}
                  </div>
                ) : (
                  <div className='divide-y'>
                    <div className='text-muted-foreground hidden grid-cols-[minmax(0,1.35fr)_minmax(0,1fr)_minmax(0,0.75fr)_minmax(0,0.75fr)_minmax(0,0.8fr)] gap-4 px-4 py-3 text-xs md:grid'>
                      <span>{t('Type')}</span>
                      <span>{t('Friend')}</span>
                      <span>{t('Reward')}</span>
                      <span>{t('Status')}</span>
                      <span>{t('Date')}</span>
                    </div>
                    {rewards.map((reward, index) => (
                      <div
                        key={String(reward.id ?? index)}
                        className='grid gap-2 px-4 py-3 md:grid-cols-[minmax(0,1.35fr)_minmax(0,1fr)_minmax(0,0.75fr)_minmax(0,0.75fr)_minmax(0,0.8fr)] md:items-center md:gap-4'
                      >
                        <div className='text-sm font-medium'>
                          {rewardTypeLabel(reward, t)}
                        </div>
                        <div className='text-sm'>
                          <span className='text-muted-foreground mr-2 text-xs md:hidden'>
                            {t('Friend')}
                          </span>
                          {String(
                            reward.invitee_name || reward.invitee_id || '-'
                          )}
                        </div>
                        <div className='text-sm font-medium'>
                          <span className='text-muted-foreground mr-2 text-xs md:hidden'>
                            {t('Reward')}
                          </span>
                          {formatQuota(numberValue(reward.reward_quota))}
                        </div>
                        <div>
                          <Badge
                            variant={
                              String(reward.status).toLowerCase() === 'pending'
                                ? 'warning'
                                : 'outline'
                            }
                          >
                            {rewardStatusLabel(reward.status, t)}
                          </Badge>
                        </div>
                        <div className='text-muted-foreground text-xs'>
                          <span className='mr-2 md:hidden'>{t('Date')}</span>
                          {displayDate(reward.created_at)}
                        </div>
                      </div>
                    ))}
                  </div>
                )}
                {rewardsTotal > PAGE_SIZE && (
                  <div className='flex justify-end gap-2 border-t p-3'>
                    <Button
                      size='icon'
                      variant='outline'
                      aria-label={t('Previous')}
                      title={t('Previous')}
                      disabled={rewardsPage === 1}
                      onClick={() => setRewardsPage((page) => page - 1)}
                    >
                      <ChevronLeft />
                    </Button>
                    <Button
                      size='icon'
                      variant='outline'
                      aria-label={t('Next')}
                      title={t('Next')}
                      disabled={rewardsPage * PAGE_SIZE >= rewardsTotal}
                      onClick={() => setRewardsPage((page) => page + 1)}
                    >
                      <ChevronRight />
                    </Button>
                  </div>
                )}
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}
