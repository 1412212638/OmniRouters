import { useQuery } from '@tanstack/react-query'
import { getRouteApi } from '@tanstack/react-router'
import type { ColumnDef } from '@tanstack/react-table'
import { useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import {
  DataTablePage,
  DataTableRow,
  useDataTable,
} from '@/components/data-table'
import { useActiveOrganizationId } from '@/features/organizations/workspace-store'
import { useMediaQuery } from '@/hooks'
import { useIsAdmin } from '@/hooks/use-admin'
import { useTableUrlState } from '@/hooks/use-table-url-state'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth-store'

import {
  DEFAULT_LOGS_DATA,
  LOG_TYPE_ALL_VALUE,
  LOG_TYPE_ENUM,
} from '../constants'
import type { UsageLog } from '../data/schema'
import { useColumnsByCategory } from '../lib/columns'
import { parseLogOther } from '../lib/format'
import { fetchLogsByCategory } from '../lib/utils'
import type { LogCategory } from '../types'
import { CommonLogsFilterBar } from './common-logs-filter-bar'
import { UsageLogDetailsSheet } from './dialogs/details-dialog'
import { TaskLogsFilterBar } from './task-logs-filter-bar'
import { UsageLogsMobileList } from './usage-logs-mobile-card'

const route = getRouteApi('/_authenticated/usage-logs/$section')

const logTypeRowTint: Record<number, string> = {
  [LOG_TYPE_ENUM.ERROR]: 'bg-rose-50/40 dark:bg-rose-950/20',
  [LOG_TYPE_ENUM.REFUND]: 'bg-blue-50/30 dark:bg-blue-950/15',
}

// Warning tint for logs where a quota conversion saturated (admin-only marker).
// Takes precedence over the per-type tint since it flags a billing anomaly.
const quotaSaturationRowTint = 'bg-amber-50/60 dark:bg-amber-950/25'

function getColumnVisibilityStorageKey(
  logCategory: LogCategory,
  isAdmin: boolean
): string {
  return `usage-logs:${logCategory}:${isAdmin ? 'admin' : 'user'}:column-visibility`
}

function deserializeLogTypeFilter(value: unknown): unknown[] {
  const values = Array.isArray(value) ? value : [value].filter(Boolean)
  return values.filter((item) => String(item) !== LOG_TYPE_ALL_VALUE)
}

interface UsageLogsTableProps {
  logCategory: LogCategory
}

interface LogDetailsSelection {
  logs: UsageLog[]
  index: number
}

export function UsageLogsTable({ logCategory }: UsageLogsTableProps) {
  const { t } = useTranslation()
  const organizationId = useActiveOrganizationId()
  const userId = useAuthStore((state) => state.auth.user?.id)
  const isAdmin = useIsAdmin() && !organizationId
  const isMobile = useMediaQuery('(max-width: 640px)')
  const searchParams = route.useSearch()
  const [detailsSelection, setDetailsSelection] =
    useState<LogDetailsSelection | null>(null)
  const [detailsOpen, setDetailsOpen] = useState(false)
  const detailsTriggerId = useRef<number | null>(null)

  const {
    columnFilters,
    onColumnFiltersChange,
    pagination,
    onPaginationChange,
    ensurePageInRange,
  } = useTableUrlState({
    search: route.useSearch(),
    navigate: route.useNavigate(),
    pagination: { defaultPage: 1, defaultPageSize: isMobile ? 20 : 100 },
    globalFilter: { enabled: false },
    columnFilters: [
      {
        columnId: 'created_at',
        searchKey: 'type',
        type: 'array' as const,
        deserialize: deserializeLogTypeFilter,
      },
      { columnId: 'model_name', searchKey: 'model', type: 'string' as const },
      { columnId: 'token_name', searchKey: 'token', type: 'string' as const },
      { columnId: 'group', searchKey: 'group', type: 'string' as const },
      ...(isAdmin
        ? [
            {
              columnId: 'channel',
              searchKey: 'channel',
              type: 'string' as const,
            },
            {
              columnId: 'username',
              searchKey: 'username',
              type: 'string' as const,
            },
          ]
        : []),
    ],
  })

  const { data, isLoading, isFetching } = useQuery({
    queryKey: [
      'logs',
      userId,
      organizationId,
      logCategory,
      isAdmin,
      pagination.pageIndex + 1,
      pagination.pageSize,
      columnFilters,
      searchParams,
      t,
    ],
    queryFn: async () => {
      const result = await fetchLogsByCategory({
        logCategory,
        isAdmin,
        page: pagination.pageIndex + 1,
        pageSize: pagination.pageSize,
        searchParams,
        columnFilters,
      })

      if (!result?.success) {
        toast.error(result?.message || t('Failed to load logs'))
        return DEFAULT_LOGS_DATA
      }

      return result.data || DEFAULT_LOGS_DATA
    },
    placeholderData: (previousData, previousQuery) => {
      if (previousQuery?.queryKey[1] === logCategory) {
        return previousData
      }
      return undefined
    },
  })

  const logs = data?.items || []
  const isCommon = logCategory === 'common'
  const commonLogs = isCommon ? (logs as UsageLog[]) : []
  const selectedLog = detailsSelection?.logs[detailsSelection.index]
  const columns = useColumnsByCategory(logCategory, isAdmin, (log) => {
    const index = commonLogs.indexOf(log)
    if (index < 0) return
    detailsTriggerId.current = log.id
    // Personal log IDs can change position when the list refreshes.
    setDetailsSelection({ logs: commonLogs, index })
    setDetailsOpen(true)
  })
  const isLoadingData = isLoading || (isFetching && !data)

  const { table } = useDataTable({
    data: logs as Record<string, unknown>[],
    columns: columns as ColumnDef<Record<string, unknown>>[],
    columnFilters,
    pagination,
    enableRowSelection: false,
    onPaginationChange,
    onColumnFiltersChange,
    manualPagination: true,
    manualFiltering: true,
    totalCount: data?.total || 0,
    ensurePageInRange,
    columnVisibilityStorageKey: getColumnVisibilityStorageKey(
      logCategory,
      isAdmin
    ),
  })

  const tableContent = (
    <DataTablePage
      table={table}
      columns={columns as ColumnDef<Record<string, unknown>>[]}
      isLoading={isLoadingData}
      isFetching={isFetching}
      emptyTitle={t('No Logs Found')}
      emptyDescription={t(
        'No usage logs available. Logs will appear here once API calls are made.'
      )}
      skeletonKeyPrefix='usage-log-skeleton'
      applyHeaderSize
      tableClassName={cn(
        '[&_[data-slot=table]]:text-[13px] [&_[data-slot=table]_td]:text-[13px] [&_[data-slot=table]_td_*]:text-[13px] [&_[data-slot=table]_th]:text-[13px] [&_[data-slot=table]_th_*]:text-[13px]'
      )}
      mobile={
        <UsageLogsMobileList
          table={table}
          isLoading={isLoadingData}
          logCategory={logCategory}
        />
      }
      toolbar={
        isCommon ? (
          <CommonLogsFilterBar table={table} />
        ) : (
          <TaskLogsFilterBar table={table} logCategory={logCategory} />
        )
      }
      renderRow={(row) => {
        const logType = (row.original as Record<string, unknown>).type as
          | number
          | undefined
        let tintClass =
          isCommon && logType != null ? (logTypeRowTint[logType] ?? '') : ''
        if (isCommon && isAdmin) {
          const other = parseLogOther(
            ((row.original as Record<string, unknown>).other as string) ?? ''
          )
          if (other?.admin_info?.quota_saturation) {
            tintClass = quotaSaturationRowTint
          }
        }

        return (
          <DataTableRow
            key={row.id}
            row={row}
            className={cn('transition-colors', tintClass)}
            getColumnClassName={() => (isCommon ? 'py-2' : 'py-3.5')}
          />
        )
      }}
    />
  )

  return (
    <>
      {tableContent}
      {isCommon && detailsSelection && selectedLog && (
        <UsageLogDetailsSheet
          log={selectedLog}
          isAdmin={isAdmin}
          open={detailsOpen}
          onOpenChange={setDetailsOpen}
          onPrevious={
            detailsSelection.index > 0
              ? () =>
                  setDetailsSelection({
                    ...detailsSelection,
                    index: detailsSelection.index - 1,
                  })
              : undefined
          }
          onNext={
            detailsSelection.index < detailsSelection.logs.length - 1
              ? () =>
                  setDetailsSelection({
                    ...detailsSelection,
                    index: detailsSelection.index + 1,
                  })
              : undefined
          }
          finalFocus={() => {
            const triggers = document.querySelectorAll<HTMLButtonElement>(
              `[data-usage-log-details="${detailsTriggerId.current}"]`
            )
            return (
              [...triggers].find((trigger) => trigger.offsetWidth > 0) ?? true
            )
          }}
        />
      )}
    </>
  )
}
