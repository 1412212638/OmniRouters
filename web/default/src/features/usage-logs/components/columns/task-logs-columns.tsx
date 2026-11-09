import type { ColumnDef } from '@tanstack/react-table'
import { FileImage, Music, Video } from 'lucide-react'
/* eslint-disable react-refresh/only-export-components */
import { useState, useMemo } from 'react'
import { useTranslation } from 'react-i18next'

import { DataTableColumnHeader } from '@/components/data-table'
import { StatusBadge } from '@/components/status-badge'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { api } from '@/lib/api'
import { getUserAvatarFallback, getUserAvatarStyle } from '@/lib/avatar'
import { formatTimestampToDate } from '@/lib/format'
import { cn } from '@/lib/utils'

import { TASK_ACTIONS, TASK_STATUS } from '../../constants'
import { taskActionMapper, taskStatusMapper } from '../../lib/mappers'
import {
  getDashboardTaskArtifactsURL,
  getTaskPreviewMedia,
} from '../../lib/task-media'
import type { TaskLog } from '../../types'
import { FailReasonDialog } from '../dialogs/fail-reason-dialog'
import {
  MediaPreviewDialog,
  type PreviewMedia,
} from '../dialogs/media-preview-dialog'
import { TaskDetailsDialog } from '../dialogs/task-details-dialog'
import { ModelBadge } from '../model-badge'
import { useUsageLogsContext } from '../usage-logs-provider'
import {
  createDurationColumn,
  createChannelColumn,
  createProgressColumn,
} from './column-helpers'

function isVideoTask(log: TaskLog): boolean {
  return (
    log.action === TASK_ACTIONS.GENERATE ||
    log.action === TASK_ACTIONS.TEXT_GENERATE ||
    log.action === TASK_ACTIONS.FIRST_TAIL_GENERATE ||
    log.action === TASK_ACTIONS.REFERENCE_GENERATE ||
    log.action === TASK_ACTIONS.REMIX_GENERATE
  )
}

function getTaskModelInfo(log: TaskLog): {
  name: string
  actualModel?: string
} {
  const properties = (() => {
    if (!log.properties) return null
    if (typeof log.properties === 'string') {
      try {
        const parsed = JSON.parse(log.properties)
        return parsed && typeof parsed === 'object' ? parsed : null
      } catch {
        return null
      }
    }
    return log.properties
  })()

  const originModel =
    properties && typeof properties.origin_model_name === 'string'
      ? properties.origin_model_name.trim()
      : ''
  const upstreamModel =
    properties && typeof properties.upstream_model_name === 'string'
      ? properties.upstream_model_name.trim()
      : ''

  return {
    name: originModel || upstreamModel,
    actualModel:
      originModel && upstreamModel && originModel !== upstreamModel
        ? upstreamModel
        : undefined,
  }
}

function DashboardTaskMediaPreviewCell({
  log,
  media,
}: {
  log: TaskLog
  media: PreviewMedia[]
}) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [resolvedMedia, setResolvedMedia] = useState<PreviewMedia[]>(media)
  const [loading, setLoading] = useState(false)

  const loadArtifacts = async () => {
    if (loading) return
    setLoading(true)
    try {
      const response = await api.get(getDashboardTaskArtifactsURL(log.task_id))
      const payload = response.data as {
        data?: { artifacts?: Array<{ type?: string; content_url?: string }> }
      }
      const artifacts = payload.data?.artifacts ?? []
      const fetched = artifacts
        .filter(
          (artifact) =>
            (artifact.type === 'image' ||
              artifact.type === 'video' ||
              artifact.type === 'audio') &&
            typeof artifact.content_url === 'string'
        )
        .map((artifact) => ({
          kind: artifact.type as PreviewMedia['kind'],
          url: artifact.content_url as string,
        }))
      if (fetched.length > 0) setResolvedMedia(fetched)
      setOpen(true)
    } catch {
      setOpen(true)
    } finally {
      setLoading(false)
    }
  }

  const primaryKind = resolvedMedia[0]?.kind
  const Icon =
    primaryKind === 'video'
      ? Video
      : primaryKind === 'audio'
        ? Music
        : FileImage
  const label =
    primaryKind === 'video'
      ? t('Click to preview video')
      : primaryKind === 'audio'
        ? t('Click to preview audio')
        : t('Preview')

  return (
    <>
      <button
        type='button'
        className='group flex items-center gap-1 text-left text-xs disabled:opacity-60'
        onClick={loadArtifacts}
        disabled={loading}
      >
        <Icon className='text-muted-foreground size-3' />
        <span className='text-foreground leading-snug group-hover:underline'>
          {loading ? t('Loading') : label}
        </span>
      </button>
      <MediaPreviewDialog
        open={open}
        onOpenChange={setOpen}
        media={resolvedMedia}
        taskId={log.task_id}
      />
    </>
  )
}

export function useTaskLogsColumns(isAdmin: boolean): ColumnDef<TaskLog>[] {
  const { t } = useTranslation()
  const columns: ColumnDef<TaskLog>[] = [
    {
      accessorKey: 'submit_time',
      header: t('Submit Time'),
      cell: ({ row }) => {
        const log = row.original
        const submitTime = row.getValue('submit_time') as number

        return (
          <div className='flex min-w-0 flex-col gap-0.5'>
            <span className='truncate font-mono text-xs tabular-nums'>
              {formatTimestampToDate(submitTime, 'seconds')}
            </span>
            {log.finish_time ? (
              <span className='text-muted-foreground/60 truncate font-mono text-[11px] tabular-nums'>
                {formatTimestampToDate(log.finish_time, 'seconds')}
              </span>
            ) : (
              <span className='text-muted-foreground/50 text-[11px]'>-</span>
            )}
          </div>
        )
      },
      size: 180,
    },
  ]

  if (isAdmin) {
    columns.push(createChannelColumn<TaskLog>({ headerLabel: t('Channel') }), {
      id: 'user',
      header: t('User'),
      accessorFn: (row) => row.username || row.user_id,
      cell: function UserCell({ row }) {
        const { sensitiveVisible, setSelectedUserId, setUserInfoDialogOpen } =
          useUsageLogsContext()
        const log = row.original
        const displayName = log.username || String(log.user_id || '?')

        return (
          <button
            type='button'
            className='flex items-center gap-1.5 text-left'
            onClick={(e) => {
              e.stopPropagation()
              setSelectedUserId(log.user_id)
              setUserInfoDialogOpen(true)
            }}
          >
            <Avatar className='ring-border/60 size-6 ring-1 max-sm:hidden'>
              <AvatarFallback
                className={cn(
                  'text-[11px] font-semibold',
                  !sensitiveVisible && 'bg-muted text-muted-foreground'
                )}
                style={
                  sensitiveVisible ? getUserAvatarStyle(displayName) : undefined
                }
              >
                {sensitiveVisible ? getUserAvatarFallback(displayName) : '•'}
              </AvatarFallback>
            </Avatar>
            <span className='text-muted-foreground truncate text-sm hover:underline'>
              {sensitiveVisible ? displayName : '••••'}
            </span>
          </button>
        )
      },
    })
  }

  columns.push(
    {
      accessorKey: 'task_id',
      header: t('Task ID'),
      cell: ({ row }) => {
        const log = row.original
        const taskId = row.getValue('task_id') as string
        if (!taskId) {
          return <span className='text-muted-foreground/60 text-xs'>-</span>
        }
        return (
          <div className='flex max-w-[170px] flex-col gap-0.5'>
            <StatusBadge
              label={taskId}
              copyText={taskId}
              variant='neutral'
              size='sm'
              className='border-border/60 bg-muted/30 !text-foreground max-w-full truncate rounded-md border px-1.5 py-0.5 font-mono'
            />
            <span className='text-muted-foreground/60 truncate text-[11px]'>
              {t(log.platform)} · {t(taskActionMapper.getLabel(log.action))}
            </span>
          </div>
        )
      },
      meta: { mobileTitle: true },
    },
    {
      accessorKey: 'properties',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Model')} />
      ),
      cell: ({ row }) => {
        const modelInfo = getTaskModelInfo(row.original)
        if (!modelInfo.name) {
          return <span className='text-muted-foreground/60 text-xs'>-</span>
        }

        return (
          <div className='flex max-w-[220px] flex-col gap-0.5'>
            <ModelBadge
              modelName={modelInfo.name}
              actualModel={isAdmin ? modelInfo.actualModel : undefined}
            />
          </div>
        )
      },
      meta: { label: t('Model') },
    },
    createDurationColumn<TaskLog>({
      submitTimeKey: 'submit_time',
      finishTimeKey: 'finish_time',
      unit: 'seconds',
      headerLabel: t('Duration'),
      warningThresholdSec: 300,
    }),
    {
      accessorKey: 'status',
      header: t('Status'),
      cell: ({ row }) => {
        const status = row.getValue('status') as string
        return (
          <StatusBadge
            label={t(taskStatusMapper.getLabel(status, status || 'Submitting'))}
            variant={taskStatusMapper.getVariant(status)}
            size='sm'
            copyable={false}
            className='-ml-1.5'
          />
        )
      },
    },
    createProgressColumn<TaskLog>({ headerLabel: t('Progress') }),
    {
      accessorKey: 'fail_reason',
      header: t('Details'),
      cell: function DetailsCell({ row }) {
        const log = row.original
        const failReason = row.getValue('fail_reason') as string
        const status = log.status
        const [dialogOpen, setDialogOpen] = useState(false)
        const previewMedia = useMemo(
          () => getTaskPreviewMedia(log, TASK_STATUS.SUCCESS, isVideoTask(log)),
          [log]
        )

        const isSuccess = status === TASK_STATUS.SUCCESS

        if (isSuccess && previewMedia.length > 0) {
          return (
            <DashboardTaskMediaPreviewCell log={log} media={previewMedia} />
          )
        }

        if (!failReason) {
          return (
            <>
              <button
                type='button'
                className='text-foreground text-xs hover:underline'
                onClick={() => setDialogOpen(true)}
              >
                {t('View details')}
              </button>
              <TaskDetailsDialog
                log={log}
                isAdmin={isAdmin}
                open={dialogOpen}
                onOpenChange={setDialogOpen}
              />
            </>
          )
        }

        return (
          <>
            <button
              type='button'
              className='group flex max-w-[200px] items-center gap-1 text-left text-xs'
              onClick={() => setDialogOpen(true)}
              title={t('Click to view full error message')}
            >
              <span className='truncate leading-snug text-red-600 group-hover:underline dark:text-red-400'>
                {failReason}
              </span>
            </button>
            <FailReasonDialog
              failReason={failReason}
              open={dialogOpen}
              onOpenChange={setDialogOpen}
            />
          </>
        )
      },
      size: 200,
      maxSize: 220,
    }
  )

  return columns
}
