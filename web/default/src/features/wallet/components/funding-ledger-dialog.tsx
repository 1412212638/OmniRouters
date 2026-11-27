import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Dialog } from '@/components/dialog'
import { Button } from '@/components/ui/button'
import { Field, FieldLabel } from '@/components/ui/field'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { toIntlLocale } from '@/i18n/languages'
import { formatNumber, formatQuota } from '@/lib/format'
import { useAuthStore } from '@/stores/auth-store'

import { fundingKinds, getFundingLedger } from '../funding'
import { WalletFundingSummary } from './funding-summary'

export function FundingLedgerDialog({ userId }: { userId: number }) {
  const { t, i18n } = useTranslation()
  const viewerId = useAuthStore((state) => state.auth.user?.id)
  const [open, setOpen] = useState(false)
  const [page, setPage] = useState(1)
  const [kind, setKind] = useState('')
  const query = useQuery({
    queryKey: ['wallet-funding-ledger', viewerId, userId, page, kind],
    queryFn: () => getFundingLedger(userId, page, kind),
    enabled: open,
    staleTime: 0,
  })
  const labels: Record<string, string> = {
    paid: t('Paid credit'),
    gift: t('Gift credit'),
    adjustment: t('Balance adjustment'),
    unclassified: t('Unclassified credit'),
    topup: t('Online top-up'),
    admin: t('Administrator'),
    registration: t('Registration reward'),
    invitation: t('Invitation reward'),
    checkin: t('Check-in reward'),
    redemption: t('Redemption code'),
    referral_cashback: t('Referral cashback'),
    affiliate_transfer: t('Affiliate transfer'),
  }
  return (
    <Dialog
      open={open}
      onOpenChange={setOpen}
      title={t('Funding records')}
      contentClassName='sm:max-w-6xl'
      trigger={
        <Button type='button' variant='outline'>
          {t('View funding records')}
        </Button>
      }
      bodyClassName='flex flex-col gap-4'
    >
      <WalletFundingSummary
        summary={query.data?.summary}
        loading={query.isPending}
        failed={query.isError}
      />
      <Field>
        <FieldLabel htmlFor='funding-filter'>
          {t('Funding classification')}
        </FieldLabel>
        <Select
          value={kind}
          onValueChange={(value) => {
            setKind(value ?? '')
            setPage(1)
          }}
        >
          <SelectTrigger id='funding-filter'>
            <SelectValue>{kind ? labels[kind] : t('All')}</SelectValue>
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value=''>{t('All')}</SelectItem>
              {fundingKinds.map((value) => (
                <SelectItem key={value} value={value}>
                  {labels[value]}
                </SelectItem>
              ))}
            </SelectGroup>
          </SelectContent>
        </Select>
      </Field>
      {query.isPending && <Skeleton className='h-32 w-full' />}
      {query.isError && (
        <div role='alert' className='flex items-center gap-2'>
          {t('Failed to load funding records')}
          <Button variant='outline' onClick={() => void query.refetch()}>
            {t('Retry')}
          </Button>
        </div>
      )}
      {query.data && (
        <>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t('Time')}</TableHead>
                <TableHead>{t('Funding classification')}</TableHead>
                <TableHead>{t('Source')}</TableHead>
                <TableHead className='text-right'>
                  {t('Credited amount')}
                </TableHead>
                <TableHead className='text-right'>{t('Amount paid')}</TableHead>
                <TableHead>{t('Operator ID')}</TableHead>
                <TableHead>{t('Reference')}</TableHead>
                <TableHead>{t('Note')}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {query.data.items.map((entry) => (
                <TableRow key={entry.id}>
                  <TableCell className='whitespace-nowrap'>
                    {new Date(entry.created_at * 1000).toLocaleString()}
                  </TableCell>
                  <TableCell>{labels[entry.kind]}</TableCell>
                  <TableCell>{labels[entry.source] ?? entry.source}</TableCell>
                  <TableCell className='text-right'>
                    {Number(entry.amount) > 0 ? '+' : ''}
                    {formatQuota(Number(entry.amount))}
                  </TableCell>
                  <TableCell className='text-right'>
                    {entry.paid_amount == null
                      ? '—'
                      : formatNumber(
                          Number(entry.paid_amount),
                          toIntlLocale(i18n.resolvedLanguage)
                        )}
                  </TableCell>
                  <TableCell>{entry.operator_id || t('System')}</TableCell>
                  <TableCell className='max-w-52 break-all whitespace-normal'>
                    {entry.reference || '—'}
                  </TableCell>
                  <TableCell className='max-w-64 break-words whitespace-normal'>
                    {entry.note || '—'}
                  </TableCell>
                </TableRow>
              ))}
              {query.data.items.length === 0 && (
                <TableRow>
                  <TableCell colSpan={8}>{t('No funding records')}</TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
          <div className='flex flex-wrap items-center justify-end gap-2'>
            <span className='text-muted-foreground text-sm'>
              {t('Total records: {{count}}', { count: query.data.total })}
            </span>
            <Button
              variant='outline'
              disabled={page <= 1 || query.isFetching}
              onClick={() => setPage(page - 1)}
            >
              {t('Previous')}
            </Button>
            <span>{page}</span>
            <Button
              variant='outline'
              disabled={page * 20 >= query.data.total || query.isFetching}
              onClick={() => setPage(page + 1)}
            >
              {t('Next')}
            </Button>
          </div>
          <p className='text-muted-foreground text-sm'>
            {t(
              'Adjustments and unclassified credits are excluded from paid and gift totals.'
            )}
          </p>
        </>
      )}
    </Dialog>
  )
}
