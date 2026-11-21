import {
  Check,
  ChevronLeft,
  ChevronRight,
  Copy,
  Download,
  Search,
} from 'lucide-react'
import type { Ref } from 'react'
import { useTranslation } from 'react-i18next'

import { StatusBadge } from '@/components/status-badge'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import {
  InputGroup,
  InputGroupAddon,
  InputGroupInput,
} from '@/components/ui/input-group'
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
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { useCopyToClipboard } from '@/hooks/use-copy-to-clipboard'
import { formatCurrencyFromUSD } from '@/lib/currency'

import { useBillingHistory } from '../hooks/use-billing-history'
import {
  formatTimestamp,
  getPaymentMethodName,
  getStatusConfig,
} from '../lib/billing'

const pageSizes = ['10', '20', '50', '100']

export function RechargeHistory(props: {
  ref?: Ref<HTMLElement>
  revision: number
  onManageOrders?: () => void
}) {
  const { t } = useTranslation()
  const history = useBillingHistory({
    initialPageSize: 20,
    personalOnly: true,
    revision: props.revision,
  })
  const { copyToClipboard, copiedText } = useCopyToClipboard({ notify: false })
  const maxPage = Math.max(1, Math.ceil(history.total / history.pageSize))

  return (
    <section
      ref={props.ref}
      id='wallet-recharge-history'
      aria-labelledby='wallet-recharge-history-title'
      className='flex min-w-0 scroll-mt-4 flex-col gap-3'
    >
      <div className='flex flex-wrap items-center justify-between gap-2'>
        <h2
          id='wallet-recharge-history-title'
          className='text-base font-semibold'
        >
          {t('Recharge records')}
        </h2>
        <div className='flex max-w-full flex-wrap items-center gap-2'>
          {(history.total > 0 || history.keyword) && (
            <InputGroup className='h-8 w-full sm:w-56'>
              <InputGroupAddon>
                <Search />
              </InputGroupAddon>
              <InputGroupInput
                aria-label={t('Search by order number...')}
                placeholder={t('Search by order number...')}
                value={history.keyword}
                onChange={(event) => history.handleSearch(event.target.value)}
              />
            </InputGroup>
          )}
          {props.onManageOrders && (
            <Button
              type='button'
              variant='outline'
              size='sm'
              onClick={props.onManageOrders}
            >
              {t('Order History')}
            </Button>
          )}
        </div>
      </div>
      <Card
        data-card-hover='false'
        className='gap-0 overflow-hidden rounded-lg py-0'
      >
        <Table
          aria-labelledby='wallet-recharge-history-title'
          aria-busy={history.loading}
        >
          <TableHeader className='bg-muted/45'>
            <TableRow>
              <TableHead className='pl-4'>{t('Time')}</TableHead>
              <TableHead>{t('Payment Method')}</TableHead>
              <TableHead>{t('Amount')}</TableHead>
              <TableHead>{t('Order number')}</TableHead>
              <TableHead className='pr-4'>{t('Status')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {history.loading && (
              <TableRow>
                <TableCell colSpan={5} className='h-16 px-4'>
                  <Skeleton className='h-4 w-full' />
                </TableCell>
              </TableRow>
            )}
            {!history.loading && history.records.length === 0 && (
              <TableRow>
                <TableCell
                  colSpan={5}
                  className='text-muted-foreground h-16 px-4 text-center'
                >
                  {t('No billing records found')}
                </TableCell>
              </TableRow>
            )}
            {!history.loading &&
              history.records.map((record) => {
                const status = getStatusConfig(record.status)
                const copied = copiedText === record.trade_no
                return (
                  <TableRow key={record.id}>
                    <TableCell className='h-12 pl-4'>
                      {formatTimestamp(record.create_time)}
                    </TableCell>
                    <TableCell>
                      {getPaymentMethodName(record.payment_method, t)}
                    </TableCell>
                    <TableCell className='font-medium'>
                      {formatCurrencyFromUSD(record.amount, {
                        digitsLarge: 2,
                        digitsSmall: 2,
                        abbreviate: false,
                      })}
                    </TableCell>
                    <TableCell>
                      <div className='flex items-center gap-1.5'>
                        <code
                          className='text-muted-foreground max-w-40 truncate'
                          title={record.trade_no}
                        >
                          {record.trade_no}
                        </code>
                        <Tooltip>
                          <TooltipTrigger
                            render={
                              <Button
                                type='button'
                                variant='ghost'
                                size='icon-sm'
                                aria-label={copied ? t('Copied') : t('Copy')}
                                onClick={() => copyToClipboard(record.trade_no)}
                              />
                            }
                          >
                            {copied ? <Check /> : <Copy />}
                          </TooltipTrigger>
                          <TooltipContent>
                            {copied ? t('Copied') : t('Copy')}
                          </TooltipContent>
                        </Tooltip>
                        {record.invoice_available &&
                          record.status === 'success' && (
                            <Tooltip>
                              <TooltipTrigger
                                render={
                                  <Button
                                    type='button'
                                    variant='ghost'
                                    size='icon-sm'
                                    aria-label={t('View invoice')}
                                    onClick={() =>
                                      history.handleOpenInvoice(record.id)
                                    }
                                  />
                                }
                              >
                                <Download />
                              </TooltipTrigger>
                              <TooltipContent>
                                {t('View invoice')}
                              </TooltipContent>
                            </Tooltip>
                          )}
                      </div>
                    </TableCell>
                    <TableCell className='pr-4'>
                      <StatusBadge
                        label={t(status.label)}
                        variant={status.variant}
                        showDot
                        copyable={false}
                      />
                    </TableCell>
                  </TableRow>
                )
              })}
          </TableBody>
        </Table>
        <div className='bg-muted/25 flex flex-wrap items-center justify-between gap-2 border-t px-4 py-2'>
          <div className='text-muted-foreground flex flex-wrap items-center gap-2 text-xs'>
            <span>{t('Rows per page')}</span>
            <Select
              value={String(history.pageSize)}
              onValueChange={(value) => {
                if (value) history.handlePageSizeChange(Number(value))
              }}
            >
              <SelectTrigger
                aria-label={t('Rows per page')}
                className='h-8 w-16'
              >
                <SelectValue />
              </SelectTrigger>
              <SelectContent alignItemWithTrigger={false}>
                <SelectGroup>
                  {pageSizes.map((value) => (
                    <SelectItem key={value} value={value}>
                      {value}
                    </SelectItem>
                  ))}
                </SelectGroup>
              </SelectContent>
            </Select>
            <span>
              {t('Total')} {history.total}
            </span>
          </div>
          <div className='flex items-center gap-2'>
            <Button
              type='button'
              variant='ghost'
              size='icon-sm'
              aria-label={t('Previous page')}
              disabled={history.loading || history.page <= 1}
              onClick={() => history.handlePageChange(history.page - 1)}
            >
              <ChevronLeft />
            </Button>
            <span className='text-muted-foreground text-xs tabular-nums'>
              {history.page} / {maxPage}
            </span>
            <Button
              type='button'
              variant='ghost'
              size='icon-sm'
              aria-label={t('Next page')}
              disabled={history.loading || history.page >= maxPage}
              onClick={() => history.handlePageChange(history.page + 1)}
            >
              <ChevronRight />
            </Button>
          </div>
        </div>
      </Card>
    </section>
  )
}
