import { useQuery } from '@tanstack/react-query'
import {
  ChevronLeft,
  ChevronRight,
  Eye,
  RefreshCw,
  Search,
  X,
} from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { SectionPageLayout } from '@/components/layout'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { NativeSelect, NativeSelectOption } from '@/components/ui/native-select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Tooltip,
  TooltipTrigger,
  TooltipContent,
} from '@/components/ui/tooltip'
import { renderAuditContent } from '@/features/usage-logs/lib/format'

import { getAuditLogs, type AuditFilters, type AuditLog } from './api'
import { actionLabels, auditValue, readAuditDetails } from './details'

const emptyFilters = {
  username: '',
  user_id: '',
  action: '',
  ip: '',
  request_id: '',
  outcome: '',
  start: '',
  end: '',
}
const tabs = {
  important: 'Important operations',
  security: 'Security and login',
  access: 'Access records',
  all: 'All',
}

export function AuditLogs() {
  const { t } = useTranslation()
  const [view, setView] = useState('important')
  const [draft, setDraft] = useState(emptyFilters)
  const [filters, setFilters] = useState<Omit<AuditFilters, 'view'>>({})
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [selected, setSelected] = useState<AuditLog | null>(null)
  const { data, isLoading, isError, isFetching, refetch } = useQuery({
    queryKey: ['audit-logs', view, filters, page, pageSize],
    queryFn: () => getAuditLogs(page, pageSize, { ...filters, view }),
    retry: false,
  })
  const details = selected ? readAuditDetails(selected) : {}
  const outcomeLabels: Record<string, string> = {
    success: 'Success',
    failed: 'Failed',
    pending: 'Pending',
    unknown: 'Unknown',
    legacy: 'Historical result',
  }
  const authLabels: Record<string, string> = {
    session: 'Browser session',
    personal_token: 'Personal access token',
    anonymous: 'Anonymous',
    password: 'Password',
    two_factor: 'Two-factor authentication',
    passkey: 'Passkey',
    oauth: 'OAuth',
  }
  const actionName = (item: AuditLog) => {
    if (actionLabels[item.action]) return t(actionLabels[item.action])
    const meta = readAuditDetails(item)
    const params = {
      id: '-',
      name: '-',
      username: '-',
      role: '-',
      count: '-',
      type: '-',
      quota: '-',
      from: '-',
      to: '-',
      key: '-',
      tag: '-',
      plan_id: '-',
      target_user_id: '-',
      sourceId: '-',
      bindingType: '-',
      action: '-',
      method: meta.method || '',
      route: meta.route || '',
      ...meta.params,
    }
    return (
      renderAuditContent(
        { op: { action: item.action, params } },
        (key, options) => t(key, options)
      ) || t('Operation')
    )
  }
  const resultName = (item: AuditLog) =>
    item.outcome
      ? t(outcomeLabels[item.outcome] || 'Unknown')
      : `${t('Historical result')}: ${t(item.success ? 'Success' : 'Failed')}`
  const iconButton = (
    label: string,
    icon: React.ReactNode,
    onClick: () => void,
    disabled = false
  ) => (
    <Tooltip>
      <TooltipTrigger
        render={
          <Button
            type='button'
            variant='outline'
            size='icon'
            aria-label={label}
            disabled={disabled}
            onClick={onClick}
          />
        }
      >
        {icon}
      </TooltipTrigger>
      <TooltipContent>{label}</TooltipContent>
    </Tooltip>
  )
  const pages = Math.max(1, Math.ceil((data?.total ?? 0) / pageSize))
  let emptyLabel = 'No records'
  if (isLoading) emptyLabel = 'Loading...'
  else if (isError) emptyLabel = 'Failed to load audit logs'

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('Audit Logs')}</SectionPageLayout.Title>
      <SectionPageLayout.Content>
        <div className='flex min-w-0 flex-col gap-4'>
          <Tabs
            value={view}
            onValueChange={(value) => {
              setView(String(value))
              setPage(1)
            }}
          >
            <TabsList variant='line' className='max-w-full overflow-x-auto'>
              {Object.entries(tabs).map(([value, label]) => (
                <TabsTrigger key={value} value={value}>
                  {t(label)}
                </TabsTrigger>
              ))}
            </TabsList>
          </Tabs>
          <form
            onSubmit={(event) => {
              event.preventDefault()
              const { start, end, ...text } = draft
              if (start && end && start > end) return
              setFilters({
                ...text,
                start: start
                  ? Math.floor(new Date(start).getTime() / 1000)
                  : undefined,
                end: end
                  ? Math.floor(new Date(end).getTime() / 1000)
                  : undefined,
              })
              setPage(1)
            }}
          >
            <FieldGroup className='grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4'>
              {(
                [
                  'username',
                  'user_id',
                  'action',
                  'ip',
                  'request_id',
                  'start',
                  'end',
                ] as const
              ).map((key) => {
                const label = {
                  username: 'Username',
                  user_id: 'User ID',
                  action: 'Action code',
                  ip: 'IP',
                  request_id: 'Request ID',
                  start: 'Start time',
                  end: 'End time',
                }[key]
                let inputType = 'text'
                let minimum: number | string | undefined
                if (key === 'start' || key === 'end') {
                  inputType = 'datetime-local'
                }
                if (key === 'user_id') {
                  inputType = 'number'
                  minimum = 0
                }
                if (key === 'end') minimum = draft.start
                return (
                  <Field key={key} className='min-w-0'>
                    <FieldLabel htmlFor={`audit-${key}`}>{t(label)}</FieldLabel>
                    <Input
                      id={`audit-${key}`}
                      type={inputType}
                      min={minimum}
                      maxLength={256}
                      value={draft[key]}
                      autoComplete='off'
                      onChange={(e) =>
                        setDraft({ ...draft, [key]: e.target.value })
                      }
                    />
                  </Field>
                )
              })}
              <Field>
                <FieldLabel htmlFor='audit-outcome'>{t('Status')}</FieldLabel>
                <NativeSelect
                  id='audit-outcome'
                  className='w-full'
                  value={draft.outcome}
                  onChange={(e) =>
                    setDraft({ ...draft, outcome: e.target.value })
                  }
                >
                  <NativeSelectOption value=''>{t('All')}</NativeSelectOption>
                  {Object.entries(outcomeLabels).map(([value, label]) => (
                    <NativeSelectOption key={value} value={value}>
                      {t(label)}
                    </NativeSelectOption>
                  ))}
                </NativeSelect>
              </Field>
            </FieldGroup>
            <div className='mt-3 flex items-center gap-2'>
              <Button type='submit'>
                <Search data-icon='inline-start' />
                {t('Search')}
              </Button>
              {iconButton(t('Reset'), <X />, () => {
                setDraft(emptyFilters)
                setFilters({})
                setPage(1)
              })}
              {iconButton(
                t('Refresh'),
                <RefreshCw />,
                () => {
                  void refetch()
                },
                isFetching
              )}
            </div>
          </form>
          <div
            className='min-w-0 overflow-auto rounded-md border'
            aria-busy={isFetching}
          >
            <Table className='min-w-[900px] table-fixed'>
              <TableHeader>
                <TableRow>
                  {[
                    'Time',
                    'User',
                    'Action',
                    'Target',
                    'Status',
                    'IP',
                    'Details',
                  ].map((label) => (
                    <TableHead
                      key={label}
                      className={label === 'Details' ? 'w-20' : ''}
                    >
                      {t(label)}
                    </TableHead>
                  ))}
                </TableRow>
              </TableHeader>
              <TableBody>
                {isLoading || isError || !data?.items?.length ? (
                  <TableRow>
                    <TableCell colSpan={7} className='h-32 text-center'>
                      {t(emptyLabel)}
                    </TableCell>
                  </TableRow>
                ) : (
                  data.items.map((item) => {
                    const meta = readAuditDetails(item)
                    return (
                      <TableRow key={item.id}>
                        <TableCell className='whitespace-normal'>
                          {new Date(item.created_at * 1000).toLocaleString()}
                        </TableCell>
                        <TableCell>
                          <div className='truncate' title={item.username}>
                            {item.username ||
                              (item.user_id
                                ? `#${item.user_id}`
                                : t('Anonymous'))}
                          </div>
                          {item.user_id > 0 && (
                            <div className='text-muted-foreground text-xs'>
                              ID: {item.user_id}
                            </div>
                          )}
                        </TableCell>
                        <TableCell>
                          <div className='truncate' title={actionName(item)}>
                            {actionName(item)}
                          </div>
                          <div
                            className='text-muted-foreground truncate text-xs'
                            title={item.action}
                          >
                            {item.action}
                          </div>
                        </TableCell>
                        <TableCell
                          className='truncate'
                          title={meta.target || meta.route}
                        >
                          {meta.target || meta.route || '-'}
                        </TableCell>
                        <TableCell>
                          <Badge
                            variant={
                              item.outcome === 'failed'
                                ? 'destructive'
                                : 'secondary'
                            }
                            className='max-w-full whitespace-normal'
                          >
                            {resultName(item)}
                          </Badge>
                        </TableCell>
                        <TableCell className='truncate' title={item.ip}>
                          {item.ip || '-'}
                        </TableCell>
                        <TableCell>
                          {iconButton(t('Details'), <Eye />, () =>
                            setSelected(item)
                          )}
                        </TableCell>
                      </TableRow>
                    )
                  })
                )}
              </TableBody>
            </Table>
          </div>
          <div className='flex flex-wrap items-center justify-between gap-3'>
            <span className='text-muted-foreground text-sm'>
              {t('Total')}: {data?.total ?? 0}
            </span>
            <div className='flex items-center gap-2'>
              <NativeSelect
                aria-label={t('Rows per page')}
                value={pageSize}
                onChange={(e) => {
                  setPageSize(Number(e.target.value))
                  setPage(1)
                }}
              >
                {[20, 50, 100].map((size) => (
                  <NativeSelectOption key={size} value={size}>
                    {size}
                  </NativeSelectOption>
                ))}
              </NativeSelect>
              {iconButton(
                t('Previous page'),
                <ChevronLeft />,
                () => setPage(page - 1),
                page <= 1 || isFetching
              )}
              <span className='text-sm'>
                {page} / {pages}
              </span>
              {iconButton(
                t('Next page'),
                <ChevronRight />,
                () => setPage(page + 1),
                page >= pages || isFetching
              )}
            </div>
          </div>
        </div>
        <Dialog
          open={!!selected}
          onOpenChange={(open) => {
            if (!open) setSelected(null)
          }}
        >
          <DialogContent className='max-h-[85dvh] overflow-y-auto sm:max-w-2xl'>
            <DialogTitle>{t('Audit details')}</DialogTitle>
            <DialogDescription>
              {selected ? actionName(selected) : ''}
            </DialogDescription>
            {selected && (
              <>
                <dl className='grid grid-cols-1 gap-3 sm:grid-cols-2'>
                  {Object.entries({
                    Time: new Date(selected.created_at * 1000).toLocaleString(),
                    User: `${selected.username || t('Anonymous')} (#${selected.user_id})`,
                    'Action code': selected.action,
                    Category: t(
                      (
                        {
                          operation: 'Important operations',
                          security: 'Security and login',
                          access: 'Access records',
                          access_token: 'Access token (legacy)',
                        } as Record<string, string>
                      )[selected.category] || 'Unknown'
                    ),
                    Target: details.target || '-',
                    Status: resultName(selected),
                    IP: selected.ip || '-',
                    'Authentication method': t(
                      authLabels[details.auth_method || ''] || 'Unknown'
                    ),
                    Method: details.method || '-',
                    Route: details.route || '-',
                    'HTTP status': details.http_status || '-',
                    'Request ID': selected.request_id || '-',
                    'Event ID': selected.event_id || '-',
                    'Failure reason':
                      details.failure === 'business_rejected'
                        ? t('Business operation rejected')
                        : details.failure || '-',
                  }).map(([label, value]) => (
                    <div key={label} className='min-w-0'>
                      <dt className='text-muted-foreground text-xs'>
                        {t(label)}
                      </dt>
                      <dd className='mt-1 text-sm break-all'>{value}</dd>
                    </div>
                  ))}
                </dl>
                <h3 className='text-sm font-medium'>{t('Changes')}</h3>
                {details.changes && Object.keys(details.changes).length ? (
                  <Table className='table-fixed'>
                    <TableHeader>
                      <TableRow>
                        {['Field', 'Before', 'After'].map((label) => (
                          <TableHead key={label}>{t(label)}</TableHead>
                        ))}
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {Object.entries(details.changes).map(
                        ([field, change]) => (
                          <TableRow key={field}>
                            <TableCell className='break-all whitespace-pre-wrap'>
                              {field}
                            </TableCell>
                            {(['before', 'after'] as const).map((side) => (
                              <TableCell
                                key={side}
                                className='break-all whitespace-pre-wrap'
                              >
                                {change[side] === '[REDACTED]'
                                  ? t('Redacted')
                                  : auditValue(change[side])}
                              </TableCell>
                            ))}
                          </TableRow>
                        )
                      )}
                    </TableBody>
                  </Table>
                ) : (
                  <p className='text-muted-foreground text-sm'>
                    {t('No change details recorded')}
                  </p>
                )}
              </>
            )}
          </DialogContent>
        </Dialog>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
