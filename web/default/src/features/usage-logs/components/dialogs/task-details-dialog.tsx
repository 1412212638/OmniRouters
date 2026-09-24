import { Copy, Check } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Dialog } from '@/components/dialog'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { ScrollArea } from '@/components/ui/scroll-area'
import { useCopyToClipboard } from '@/hooks/use-copy-to-clipboard'

import type { TaskLog } from '../../types'

function stringify(value: unknown): string {
  if (value == null || value === '') return '-'
  if (typeof value === 'string') return value
  try {
    return JSON.stringify(value, null, 2)
  } catch {
    return String(value)
  }
}

export function TaskDetailsDialog({
  log,
  isAdmin,
  open,
  onOpenChange,
}: {
  log: TaskLog
  isAdmin: boolean
  open: boolean
  onOpenChange: (open: boolean) => void
}) {
  const { t } = useTranslation()
  const { copiedText, copyToClipboard } = useCopyToClipboard({ notify: false })
  const properties = log.properties ?? {}
  const root = log.root_info
  const detail = stringify(log.data)
  const copyValue = detail === '-' ? log.task_id : detail
  return (
    <Dialog
      open={open}
      onOpenChange={onOpenChange}
      title={t('Task Details')}
      description={t('View task status, plugin, and provider details')}
      contentClassName='sm:max-w-2xl'
      contentHeight='auto'
      bodyClassName='space-y-4'
    >
      <ScrollArea className='max-h-[560px] pr-4'>
        <div className='space-y-4 py-2'>
          <div className='grid gap-2 sm:grid-cols-2'>
            <div>
              <Label>{t('Task ID')}</Label>
              <p className='font-mono text-xs break-all'>{log.task_id}</p>
            </div>
            <div>
              <Label>{t('Status')}</Label>
              <p className='text-sm'>{t(log.status)}</p>
            </div>
            <div>
              <Label>{t('Action')}</Label>
              <p className='text-sm'>{t(log.action)}</p>
            </div>
            <div>
              <Label>{t('Progress')}</Label>
              <p className='text-sm'>{log.progress || '-'}</p>
            </div>
            <div>
              <Label>{t('Request Model')}</Label>
              <p className='font-mono text-xs break-all'>
                {properties.origin_model_name || '-'}
              </p>
            </div>
            {isAdmin && (
              <div>
                <Label>{t('Upstream Model')}</Label>
                <p className='font-mono text-xs break-all'>
                  {properties.upstream_model_name || '-'}
                </p>
              </div>
            )}
          </div>
          {isAdmin && (root?.upstream_task_id || root?.task_plugin) && (
            <div className='space-y-1'>
              <Label>{t('Plugin and Provider')}</Label>
              <pre className='bg-muted/30 overflow-x-auto rounded-md border p-3 text-xs'>
                {stringify({
                  plugin: root.task_plugin,
                  upstream_task_id: root.upstream_task_id,
                })}
              </pre>
            </div>
          )}
          <div className='space-y-1'>
            <Label>{t('Task Data')}</Label>
            <div className='bg-muted/30 relative rounded-md border p-3'>
              <Button
                variant='ghost'
                size='sm'
                className='absolute top-1 right-1 h-7 w-7 p-0'
                onClick={() => copyToClipboard(copyValue)}
                title={t('Copy to clipboard')}
                aria-label={t('Copy to clipboard')}
              >
                {copiedText === copyValue ? (
                  <Check className='size-3 text-green-600' />
                ) : (
                  <Copy className='size-3' />
                )}
              </Button>
              <pre className='max-h-72 overflow-auto pr-7 text-xs break-all whitespace-pre-wrap'>
                {detail}
              </pre>
            </div>
          </div>
          {log.fail_reason && (
            <div className='space-y-1'>
              <Label className='text-red-600'>{t('Failure Details')}</Label>
              <p className='text-sm break-all whitespace-pre-wrap text-red-600'>
                {log.fail_reason}
              </p>
            </div>
          )}
        </div>
      </ScrollArea>
    </Dialog>
  )
}
