import {
  AlertCircleIcon,
  ArchiveArrowDownIcon,
  Clock01Icon,
  FileAddIcon,
  FileScriptIcon,
  FileSpreadsheetIcon,
  HistoryIcon,
  Invoice01Icon,
  RefreshIcon,
  Table02Icon,
} from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { SectionPageLayout } from '@/components/layout'
import { StatusBadge, type StatusVariant } from '@/components/status-badge'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '@/components/ui/empty'
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldTitle,
} from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import { Spinner } from '@/components/ui/spinner'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import { useActiveOrganizationId } from '@/features/organizations/workspace-store'
import { searchUsers } from '@/features/users/api'
import type { User } from '@/features/users/types'
import { useIsAdmin } from '@/hooks/use-admin'
import { formatTimestampToDate } from '@/lib/format'
import { cn } from '@/lib/utils'

import {
  billingExportErrorKey,
  createBillingExport,
  downloadBillingExport,
  listBillingExports,
} from './api'
import type {
  BillingExport,
  BillingExportLanguage,
  BillingExportLevel,
  BillingExportStatus,
} from './types'

const EXPORT_FORMATS = [
  {
    level: 'l0',
    title: 'Statement overview',
    description:
      'Billing period, account information and totals for reconciliation.',
    type: 'PDF',
    name: 'L0-statement.pdf',
    icon: Invoice01Icon,
  },
  {
    level: 'l1',
    title: 'Billing summary',
    description:
      'Totals by model and day, with a member breakdown for organizations.',
    type: 'XLSX',
    name: 'L1-summary.xlsx',
    icon: FileSpreadsheetIcon,
  },
  {
    level: 'l2',
    title: 'Resource breakdown',
    description: 'Compare usage and charges by model and API key.',
    type: 'CSV',
    name: 'L2-resources.csv',
    icon: Table02Icon,
  },
  {
    level: 'l3',
    title: 'Model billing details',
    description:
      'Usage, charges and refunds by model and billing unit for the selected period.',
    type: 'CSV',
    name: 'L3-details.csv',
    icon: FileScriptIcon,
  },
] as const

const EXPORT_LANGUAGES: Record<BillingExportLanguage, string> = {
  zh: '简体中文',
  en: 'English',
}

const STATUS: Record<
  BillingExportStatus,
  { label: string; variant: StatusVariant }
> = {
  pending: { label: 'Pending', variant: 'warning' },
  running: { label: 'Generating', variant: 'info' },
  succeeded: { label: 'Ready', variant: 'success' },
  failed: { label: 'Failed', variant: 'danger' },
}

const HISTORY_GRID =
  'grid grid-cols-[minmax(0,1fr)_auto] items-center gap-x-4 gap-y-2 px-4 sm:px-5 @3xl/history:grid-cols-[minmax(8.5rem,1fr)_5rem_minmax(5.5rem,0.65fr)_minmax(9rem,1fr)_minmax(13rem,1fr)]'

function billingMonth(offset: number) {
  const now = new Date(Date.now() + 8 * 60 * 60 * 1000)
  return new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth() + offset, 1))
    .toISOString()
    .slice(0, 7)
}

function monthNumber(value: string) {
  const [year, month] = value.split('-')
  return Number(year) * 12 + Number(month) - 1
}

function ExportFormats() {
  const { t } = useTranslation()
  return (
    <Card className='min-w-0'>
      <CardHeader className='px-5'>
        <CardTitle>
          <h3>{t('Export files')}</h3>
        </CardTitle>
        <CardDescription>
          {t('Each export includes all four files in one ZIP archive.')}
        </CardDescription>
      </CardHeader>
      <CardContent className='flex flex-1 px-5'>
        <div className='@container/formats w-full overflow-hidden rounded-lg border'>
          <ul className='bg-border grid h-full gap-px @md/formats:grid-cols-2 @md/formats:grid-rows-2'>
            {EXPORT_FORMATS.map((format) => (
              <li
                key={format.level}
                className='bg-card flex gap-3 p-3.5 sm:p-4'
              >
                <div className='bg-muted text-foreground grid size-9 shrink-0 place-items-center rounded-lg'>
                  <HugeiconsIcon
                    icon={format.icon}
                    className='size-[18px]'
                    aria-hidden='true'
                  />
                </div>
                <div className='flex min-w-0 flex-col gap-1.5'>
                  <div className='flex flex-wrap items-center gap-x-2 gap-y-1'>
                    <span className='font-medium'>{t(format.title)}</span>
                    <span className='text-muted-foreground text-xs font-medium whitespace-nowrap tabular-nums'>
                      {format.level.toUpperCase()} · {format.type}
                    </span>
                  </div>
                  <p className='text-muted-foreground text-xs/relaxed'>
                    {t(format.description)}
                  </p>
                </div>
              </li>
            ))}
          </ul>
        </div>
      </CardContent>
    </Card>
  )
}

function ExportDownloads({
  task,
  organizationId,
  userId,
}: {
  task: BillingExport
  organizationId: number | null
  userId: number | null
}) {
  const { t } = useTranslation()
  const download = useMutation({
    mutationFn: async (level: BillingExportLevel) => {
      const blob = await downloadBillingExport(
        task.task_id,
        level,
        organizationId,
        userId
      )
      const name =
        EXPORT_FORMATS.find((format) => format.level === level)?.name ??
        'billing-statements.zip'
      const objectUrl = URL.createObjectURL(blob)
      const anchor = document.createElement('a')
      anchor.href = objectUrl
      anchor.download = `OmniRouters-${task.start_month}-${task.end_month}-${task.language ?? 'en'}-${name}`
      document.body.append(anchor)
      anchor.click()
      anchor.remove()
      window.setTimeout(() => URL.revokeObjectURL(objectUrl), 10_000)
    },
    onError: (error) => {
      toast.error(
        t(billingExportErrorKey(error, 'Could not download the bill.'))
      )
    },
  })
  return (
    <div className='flex flex-wrap items-center gap-1.5 @3xl/history:justify-end'>
      <Button
        size='sm'
        aria-label={t('Download ZIP')}
        disabled={download.isPending}
        onClick={() => download.mutate('zip')}
      >
        {download.isPending && download.variables === 'zip' ? (
          <Spinner data-icon='inline-start' />
        ) : (
          <HugeiconsIcon icon={ArchiveArrowDownIcon} data-icon='inline-start' />
        )}
        ZIP
      </Button>
      {EXPORT_FORMATS.map(({ level, title }) => (
        <Button
          key={level}
          size='sm'
          variant='outline'
          aria-label={`${t('Download')} ${level.toUpperCase()} ${t(title)}`}
          title={t(title)}
          disabled={download.isPending}
          onClick={() => download.mutate(level)}
        >
          {level.toUpperCase()}
        </Button>
      ))}
    </div>
  )
}

function ExportHistory({
  items,
  organizationId,
  userId,
}: {
  items: BillingExport[]
  organizationId: number | null
  userId: number | null
}) {
  const { t } = useTranslation()
  return (
    <div className='@container/history'>
      <div
        className={cn(
          HISTORY_GRID,
          'bg-muted/50 text-muted-foreground hidden border-y py-2 text-xs font-medium @3xl/history:grid'
        )}
      >
        <span>{t('Billing period')}</span>
        <span>{t('Export language')}</span>
        <span>{t('Status')}</span>
        <span>{t('Created')}</span>
        <span className='text-right'>{t('Downloads')}</span>
      </div>
      <ul className='divide-y border-t @3xl/history:border-t-0'>
        {items.map((item) => {
          const status = STATUS[item.status]
          const language = item.language === 'zh' ? 'zh' : 'en'
          return (
            <li key={item.task_id} className={cn(HISTORY_GRID, 'py-3.5')}>
              <div className='font-medium whitespace-nowrap tabular-nums'>
                <time dateTime={item.start_month}>{item.start_month}</time>
                {item.start_month !== item.end_month && (
                  <>
                    {' – '}
                    <time dateTime={item.end_month}>{item.end_month}</time>
                  </>
                )}
              </div>
              <div className='text-muted-foreground col-start-1 row-start-2 text-xs @3xl/history:col-start-auto @3xl/history:row-start-auto @3xl/history:text-sm'>
                <span lang={language === 'zh' ? 'zh-CN' : 'en'}>
                  {EXPORT_LANGUAGES[language]}
                </span>
              </div>
              <div className='col-start-2 row-start-1 flex flex-wrap items-center justify-end gap-1 @3xl/history:col-start-auto @3xl/history:row-start-auto @3xl/history:justify-start'>
                <StatusBadge
                  type='text'
                  variant={status.variant}
                  showDot
                  copyable={false}
                  label={t(status.label)}
                />
              </div>
              <div className='text-muted-foreground col-span-2 flex flex-col gap-0.5 text-xs tabular-nums @3xl/history:col-span-1'>
                <span>
                  <span className='@3xl/history:sr-only'>
                    {t('Created')} ·{' '}
                  </span>
                  <time
                    dateTime={new Date(item.created_at * 1000).toISOString()}
                  >
                    {formatTimestampToDate(item.created_at)}
                  </time>
                </span>
                {item.status === 'succeeded' && item.expires_at > 0 && (
                  <span>
                    {t('Available until {{date}}', {
                      date: formatTimestampToDate(item.expires_at),
                    })}
                  </span>
                )}
              </div>
              {item.status === 'succeeded' && (
                <div className='col-span-2 @3xl/history:col-span-1'>
                  <ExportDownloads
                    task={item}
                    organizationId={organizationId}
                    userId={userId}
                  />
                </div>
              )}
              {item.status === 'failed' && (
                <p className='text-destructive col-span-2 text-xs @3xl/history:col-span-1 @3xl/history:text-right'>
                  {t(
                    billingExportErrorKey(
                      item.error,
                      'Could not generate the bill.'
                    )
                  )}
                </p>
              )}
            </li>
          )
        })}
      </ul>
    </div>
  )
}

export function BillingExports() {
  const { t, i18n } = useTranslation()
  const isAdmin = useIsAdmin()
  const organizationId = useActiveOrganizationId()
  const queryClient = useQueryClient()
  const [selectedUser, setSelectedUser] = useState<User | null>(null)
  const [userSearch, setUserSearch] = useState('')
  const currentMonth = billingMonth(0)
  const [startMonth, setStartMonth] = useState(() => billingMonth(-1))
  const [endMonth, setEndMonth] = useState(() => billingMonth(-1))
  const [language, setLanguage] = useState<BillingExportLanguage>(() =>
    i18n.resolvedLanguage?.toLowerCase().startsWith('zh') ? 'zh' : 'en'
  )
  const userSearchQuery = useQuery({
    queryKey: ['billing-export-users', userSearch],
    enabled: isAdmin && userSearch.trim().length >= 2,
    queryFn: async () => {
      const result = await searchUsers({
        keyword: userSearch.trim(),
        p: 1,
        page_size: 10,
      })
      return result.success ? (result.data?.items ?? []) : []
    },
  })
  const targetUserId = isAdmin && !organizationId ? selectedUser?.id : undefined
  const invalidPeriod =
    Boolean(startMonth && endMonth) &&
    (startMonth > endMonth ||
      monthNumber(endMonth) - monthNumber(startMonth) >= 12)
  const exportsQuery = useQuery({
    queryKey: ['billing-exports', organizationId, targetUserId],
    queryFn: () => listBillingExports(organizationId, targetUserId),
    refetchInterval: (query) =>
      query.state.data?.some(
        (item) => item.status === 'pending' || item.status === 'running'
      )
        ? 5000
        : false,
  })
  const createMutation = useMutation({
    mutationFn: createBillingExport,
    onSuccess: async (_data, filters) => {
      toast.success(t('Billing export queued'))
      await queryClient.invalidateQueries({
        queryKey: [
          'billing-exports',
          filters.organizationId,
          filters.userId ?? undefined,
        ],
      })
    },
    onError: (error) => {
      toast.error(
        t(billingExportErrorKey(error, 'Could not generate the bill.'))
      )
    },
  })
  const items = exportsQuery.data ?? []
  let exportContent
  if (exportsQuery.isPending) {
    exportContent = (
      <div
        role='status'
        aria-label={t('Loading...')}
        className='grid gap-3 border-t p-5'
      >
        <Skeleton className='h-9 w-full' />
        <Skeleton className='h-9 w-full' />
      </div>
    )
  } else if (exportsQuery.isError || items.length === 0) {
    exportContent = (
      <Empty className='border-t border-solid'>
        <EmptyHeader>
          <EmptyMedia variant='icon'>
            <HugeiconsIcon
              icon={exportsQuery.isError ? AlertCircleIcon : HistoryIcon}
            />
          </EmptyMedia>
          <EmptyTitle>
            {exportsQuery.isError
              ? t('Failed to load billing exports')
              : t('No billing exports yet')}
          </EmptyTitle>
          {exportsQuery.isError && (
            <EmptyDescription>
              {t(
                billingExportErrorKey(
                  exportsQuery.error,
                  'Please try again later.'
                )
              )}
            </EmptyDescription>
          )}
        </EmptyHeader>
      </Empty>
    )
  } else {
    exportContent = (
      <ExportHistory
        items={items}
        organizationId={organizationId}
        userId={targetUserId ?? null}
      />
    )
  }
  const billingScopeLabel = organizationId
    ? t('Organization')
    : (selectedUser?.username ??
      (isAdmin ? t('Select a billing account') : t('Personal space')))

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>
        {t('Reconciliation Center')}
      </SectionPageLayout.Title>
      <SectionPageLayout.Description>
        {t('Review charges, usage, and refunds for a selected billing period.')}
      </SectionPageLayout.Description>
      <SectionPageLayout.Content>
        <div className='@container mx-auto flex max-w-[90rem] flex-col gap-4 px-px pb-1 sm:gap-5'>
          <div className='grid items-stretch gap-4 sm:gap-5 @4xl:grid-cols-[minmax(19rem,2fr)_minmax(0,3fr)]'>
            <Card className='min-w-0'>
              <CardHeader className='gap-2 px-5'>
                <CardTitle>
                  <h3>{t('New export')}</h3>
                </CardTitle>
                <CardAction>
                  <Badge variant='outline'>{billingScopeLabel}</Badge>
                </CardAction>
              </CardHeader>
              <CardContent className='px-5'>
                {isAdmin && !organizationId && (
                  <Field className='mb-4'>
                    <FieldLabel htmlFor='billing-export-user'>
                      {t('Billing account')}
                    </FieldLabel>
                    <Input
                      id='billing-export-user'
                      value={selectedUser?.username ?? userSearch}
                      placeholder={t('Search users by username or email')}
                      onChange={(event) => {
                        setSelectedUser(null)
                        setUserSearch(event.target.value)
                      }}
                      autoComplete='off'
                    />
                    {userSearch.trim().length >= 2 && !selectedUser && (
                      <ul className='max-h-52 overflow-y-auto rounded-md border'>
                        {(userSearchQuery.data ?? []).map((user) => (
                          <li key={user.id}>
                            <button
                              type='button'
                              className='hover:bg-muted flex w-full flex-col items-start px-3 py-2 text-left text-sm'
                              onClick={() => {
                                setSelectedUser(user)
                                setUserSearch('')
                              }}
                            >
                              <span className='font-medium'>
                                {user.username}
                              </span>
                              <span className='text-muted-foreground text-xs'>
                                {user.display_name ||
                                  user.email ||
                                  `#${user.id}`}
                              </span>
                            </button>
                          </li>
                        ))}
                        {!userSearchQuery.isFetching &&
                          userSearchQuery.data?.length === 0 && (
                            <li className='text-muted-foreground px-3 py-2 text-sm'>
                              {t('No users found')}
                            </li>
                          )}
                      </ul>
                    )}
                  </Field>
                )}
                <form
                  className='flex flex-col gap-4'
                  onSubmit={(event) => {
                    event.preventDefault()
                    if (!startMonth || !endMonth || invalidPeriod) {
                      return
                    }
                    if (isAdmin && !organizationId && !selectedUser) {
                      toast.error(t('Select a billing account.'))
                      return
                    }
                    createMutation.mutate({
                      startMonth,
                      endMonth,
                      language,
                      organizationId,
                      userId: targetUserId ?? null,
                    })
                  }}
                >
                  <FieldGroup className='gap-4'>
                    <div className='grid grid-cols-2 gap-3'>
                      <Field data-invalid={invalidPeriod}>
                        <FieldLabel htmlFor='billing-start-month'>
                          {t('Start month')}
                        </FieldLabel>
                        <Input
                          id='billing-start-month'
                          type='month'
                          min='2000-01'
                          max={currentMonth}
                          value={startMonth}
                          onChange={(event) =>
                            setStartMonth(event.target.value)
                          }
                          aria-invalid={invalidPeriod}
                          aria-describedby='billing-period-info'
                          required
                        />
                      </Field>
                      <Field data-invalid={invalidPeriod}>
                        <FieldLabel htmlFor='billing-end-month'>
                          {t('End month')}
                        </FieldLabel>
                        <Input
                          id='billing-end-month'
                          type='month'
                          min='2000-01'
                          max={currentMonth}
                          value={endMonth}
                          onChange={(event) => setEndMonth(event.target.value)}
                          aria-invalid={invalidPeriod}
                          aria-describedby='billing-period-info'
                          required
                        />
                      </Field>
                      <div id='billing-period-info' className='col-span-2'>
                        {invalidPeriod ? (
                          <FieldError>
                            {t('Select a billing period of up to 12 months.')}
                          </FieldError>
                        ) : (
                          <FieldDescription>
                            {t(
                              'Up to 12 months per export. Billing periods use UTC+08:00.'
                            )}
                          </FieldDescription>
                        )}
                      </div>
                    </div>
                    <Field
                      orientation='horizontal'
                      className='flex-wrap justify-between'
                    >
                      <FieldTitle id='billing-export-language'>
                        {t('Export language')}
                      </FieldTitle>
                      <ToggleGroup
                        aria-labelledby='billing-export-language'
                        variant='outline'
                        value={[language]}
                        onValueChange={(value) => {
                          const nextLanguage = value[0]
                          if (nextLanguage === 'zh' || nextLanguage === 'en') {
                            setLanguage(nextLanguage)
                          }
                        }}
                      >
                        <ToggleGroupItem
                          value='zh'
                          type='button'
                          lang='zh-CN'
                          className='text-muted-foreground aria-pressed:bg-primary aria-pressed:text-primary-foreground aria-pressed:hover:text-primary-foreground min-w-20'
                        >
                          {EXPORT_LANGUAGES.zh}
                        </ToggleGroupItem>
                        <ToggleGroupItem
                          value='en'
                          type='button'
                          lang='en'
                          className='text-muted-foreground aria-pressed:bg-primary aria-pressed:text-primary-foreground aria-pressed:hover:text-primary-foreground min-w-20'
                        >
                          {EXPORT_LANGUAGES.en}
                        </ToggleGroupItem>
                      </ToggleGroup>
                    </Field>
                  </FieldGroup>
                  <Button
                    type='submit'
                    size='lg'
                    className='w-full sm:w-auto sm:self-start'
                    disabled={
                      !startMonth ||
                      !endMonth ||
                      invalidPeriod ||
                      createMutation.isPending
                    }
                  >
                    {createMutation.isPending ? (
                      <Spinner data-icon='inline-start' />
                    ) : (
                      <HugeiconsIcon
                        icon={FileAddIcon}
                        data-icon='inline-start'
                      />
                    )}
                    {t('Generate billing')}
                  </Button>
                </form>
              </CardContent>
            </Card>
            <ExportFormats />
          </div>

          <Card className='gap-3.5 pb-0'>
            <CardHeader className='px-5'>
              <CardTitle>
                <h3>{t('Recent exports')}</h3>
              </CardTitle>
              <CardDescription className='flex items-start gap-1.5'>
                <HugeiconsIcon
                  icon={Clock01Icon}
                  className='mt-0.5 size-4 shrink-0'
                  aria-hidden='true'
                />
                {t('Files are available for 7 days.')}
              </CardDescription>
              <CardAction>
                <Button
                  variant='outline'
                  onClick={() => void exportsQuery.refetch()}
                  disabled={exportsQuery.isFetching}
                >
                  <HugeiconsIcon
                    icon={RefreshIcon}
                    data-icon='inline-start'
                    className={cn(exportsQuery.isFetching && 'animate-spin')}
                  />
                  {t('Refresh')}
                </Button>
              </CardAction>
            </CardHeader>
            {exportContent}
          </Card>
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
