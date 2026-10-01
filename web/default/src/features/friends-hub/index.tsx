import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import {
  ArrowRight,
  Check,
  ClipboardCopy,
  Gift,
  Link2,
  Users,
  WalletCards,
} from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { SectionPageLayout } from '@/components/layout'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { useCopyToClipboard } from '@/hooks/use-copy-to-clipboard'
import { formatPercent, formatQuota, formatTimestampToDate } from '@/lib/format'

import { getAffiliateCode, transferAffiliateQuota } from '../wallet/api'
import { generateAffiliateLink } from '../wallet/lib'
import { getFriendRewards, getFriends, getFriendsSummary, transferFriendRewards } from './api'
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

function rewardStatusLabel(status: unknown, t: (key: string) => string): string {
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

export function FriendsHub() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { copyToClipboard } = useCopyToClipboard()
  const [friendsPage, setFriendsPage] = useState(1)
  const [rewardsPage, setRewardsPage] = useState(1)
  const [copied, setCopied] = useState<'code' | 'link' | null>(null)

  const summaryQuery = useQuery({
    queryKey: ['friends-hub', 'summary'],
    queryFn: async () => (await getFriendsSummary()).data ?? {},
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
  const inviteCode = String(summary.invite_code || summary.aff_code || codeQuery.data || '')
  const inviteLink = String(summary.invite_link || (inviteCode ? generateAffiliateLink(inviteCode) : ''))
  const friends = unwrapList<FriendRecord>(friendsQuery.data)
  const rewards = unwrapList<RewardRecord>(rewardsQuery.data)
  const friendsTotal = numberValue(friendsQuery.data?.total)
  const rewardsTotal = numberValue(rewardsQuery.data?.total)
  const cards = [
    [t('Referral rate'), formatPercent(numberValue(summary.referral_rate ?? summary.cashback_rate))],
    [t('Invited friends'), numberValue(summary.invite_count)],
    [t('Completed first top-ups'), numberValue(summary.first_topup_count ?? summary.first_top_up_invite_count)],
    [t('Registration pending'), formatQuota(numberValue(summary.registration_pending_quota))],
    [t('Cashback pending'), formatQuota(numberValue(summary.cashback_pending_quota))],
    [t('Total rewards'), formatQuota(numberValue(summary.registration_total_quota) + numberValue(summary.cashback_total_quota))],
    [t('Wallet balance'), formatQuota(numberValue(summary.wallet_quota))],
  ]

  const copy = async (kind: 'code' | 'link', value: string) => {
    if (!value) return
    await copyToClipboard(value)
    setCopied(kind)
    window.setTimeout(() => setCopied(null), 1200)
  }

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('Friends Hub')}</SectionPageLayout.Title>
      <SectionPageLayout.Description>
        {t('Invite friends and earn rewards for their first successful top-up.')}
      </SectionPageLayout.Description>
      <SectionPageLayout.Content>
        <div className='mx-auto flex w-full max-w-7xl flex-col gap-4 sm:gap-5'>
          <div className='grid gap-3 sm:grid-cols-2 lg:grid-cols-5'>
            {cards.map(([label, value], index) => (
              <Card key={label} size='sm'>
                <CardContent className='flex min-h-20 flex-col justify-between gap-2'>
                  <span className='text-muted-foreground text-xs'>{label}</span>
                  <span className={index === 0 ? 'text-primary text-xl font-semibold' : 'text-xl font-semibold tabular-nums'}>{value}</span>
                </CardContent>
              </Card>
            ))}
          </div>

          <Card className='border-primary/20 bg-primary/[0.04]'>
            <CardHeader>
              <CardTitle className='flex items-center gap-2'><Gift className='size-4' />{t('My referral invitation')}</CardTitle>
              <p className='text-muted-foreground text-sm'>{t('Friends who register through your link earn you registration rewards and first top-up cashback.')}</p>
            </CardHeader>
            <CardContent className='grid gap-3 lg:grid-cols-2'>
              <div className='space-y-1.5'>
                <label className='text-muted-foreground text-xs'>{t('Invitation code')}</label>
                <div className='flex gap-2'>
                  <div className='bg-background flex min-h-9 min-w-0 flex-1 items-center rounded-md border px-3 font-mono text-sm'>{inviteCode || '-'}</div>
                  <Button variant='outline' size='icon' aria-label={t('Copy invitation code')} onClick={() => void copy('code', inviteCode)} disabled={!inviteCode}>
                    {copied === 'code' ? <Check /> : <ClipboardCopy />}
                  </Button>
                </div>
              </div>
              <div className='space-y-1.5'>
                <label className='text-muted-foreground text-xs'>{t('Invitation link')}</label>
                <div className='flex gap-2'>
                  <div className='bg-background flex min-h-9 min-w-0 flex-1 items-center truncate rounded-md border px-3 font-mono text-xs'>{inviteLink || '-'}</div>
                  <Button variant='outline' size='icon' aria-label={t('Copy invitation link')} onClick={() => void copy('link', inviteLink)} disabled={!inviteLink}>
                    {copied === 'link' ? <Check /> : <Link2 />}
                  </Button>
                </div>
              </div>
              <div className='text-muted-foreground flex items-start gap-2 text-xs lg:col-span-2'>
                <WalletCards className='mt-0.5 size-4 shrink-0' />
                {t('Registration rewards and first top-up cashback are tracked separately and can be transferred to your wallet.')}
              </div>
            </CardContent>
          </Card>

          <div className='grid gap-4 xl:grid-cols-2'>
            <Card>
              <CardHeader className='flex-row items-center justify-between'><CardTitle className='flex items-center gap-2'><Users className='size-4' />{t('My friends')}</CardTitle><Badge variant='outline'>{friendsTotal || friends.length}</Badge></CardHeader>
              <CardContent>
                <Table>
                  <TableHeader><TableRow><TableHead>{t('Friend')}</TableHead><TableHead>{t('Registered')}</TableHead><TableHead>{t('First top-up')}</TableHead><TableHead>{t('Cashback')}</TableHead></TableRow></TableHeader>
                  <TableBody>{friends.length === 0 ? <TableRow><TableCell colSpan={4} className='text-muted-foreground py-8 text-center'>{t('No invited friends yet')}</TableCell></TableRow> : friends.map((friend, index) => <TableRow key={String(friend.id ?? friend.user_id ?? index)}><TableCell className='font-medium'>{String(friend.display_name || friend.username || friend.email || t('Anonymous'))}</TableCell><TableCell>{displayDate(friend.register_time ?? friend.created_at)}</TableCell><TableCell>{rewardStatusLabel(friend.first_topup_status ?? friend.first_top_up_status, t)}</TableCell><TableCell>{formatQuota(numberValue(friend.cashback_quota))}</TableCell></TableRow>)}</TableBody>
                </Table>
                {friendsTotal > PAGE_SIZE && <div className='mt-3 flex justify-end gap-2'><Button size='sm' variant='outline' disabled={friendsPage === 1} onClick={() => setFriendsPage((page) => page - 1)}>{t('Previous')}</Button><Button size='sm' variant='outline' disabled={friendsPage * PAGE_SIZE >= friendsTotal} onClick={() => setFriendsPage((page) => page + 1)}>{t('Next')}</Button></div>}
              </CardContent>
            </Card>

            <Card>
              <CardHeader className='flex-row items-center justify-between gap-3'><CardTitle className='flex items-center gap-2'><Gift className='size-4' />{t('Reward history')}</CardTitle><div className='flex flex-wrap items-center justify-end gap-2'><Badge variant='outline'>{rewardsTotal || rewards.length}</Badge><Button size='sm' variant='outline' disabled={registrationTransferMutation.isPending || numberValue(summary.registration_pending_quota) <= 0} onClick={() => registrationTransferMutation.mutate()}>{t('Registration rewards')} · {formatQuota(numberValue(summary.registration_pending_quota))}</Button><Button size='sm' disabled={transferMutation.isPending || numberValue(summary.cashback_pending_quota) <= 0} onClick={() => transferMutation.mutate()}>{t('Cashback')} · {formatQuota(numberValue(summary.cashback_pending_quota))}<ArrowRight /></Button></div></CardHeader>
              <CardContent>
                <Table>
                  <TableHeader><TableRow><TableHead>{t('Type')}</TableHead><TableHead>{t('Friend')}</TableHead><TableHead>{t('Reward')}</TableHead><TableHead>{t('Status')}</TableHead><TableHead>{t('Date')}</TableHead></TableRow></TableHeader>
                  <TableBody>{rewards.length === 0 ? <TableRow><TableCell colSpan={5} className='text-muted-foreground py-8 text-center'>{t('No rewards yet')}</TableCell></TableRow> : rewards.map((reward, index) => <TableRow key={String(reward.id ?? index)}><TableCell>{t('First top-up cashback')}</TableCell><TableCell>{String(reward.invitee_name || reward.invitee_id || '-')}</TableCell><TableCell className='font-medium'>{formatQuota(numberValue(reward.reward_quota))}</TableCell><TableCell><Badge variant={String(reward.status).toLowerCase() === 'pending' ? 'warning' : 'outline'}>{rewardStatusLabel(reward.status, t)}</Badge></TableCell><TableCell>{displayDate(reward.created_at)}</TableCell></TableRow>)}</TableBody>
                </Table>
                {rewardsTotal > PAGE_SIZE && <div className='mt-3 flex justify-end gap-2'><Button size='sm' variant='outline' disabled={rewardsPage === 1} onClick={() => setRewardsPage((page) => page - 1)}>{t('Previous')}</Button><Button size='sm' variant='outline' disabled={rewardsPage * PAGE_SIZE >= rewardsTotal} onClick={() => setRewardsPage((page) => page + 1)}>{t('Next')}</Button></div>}
              </CardContent>
            </Card>
          </div>

          <div className='text-muted-foreground flex justify-end text-sm'><Button variant='link' size='sm' render={<Link to='/wallet' />}>{t('Go to wallet')}<ArrowRight /></Button></div>
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
