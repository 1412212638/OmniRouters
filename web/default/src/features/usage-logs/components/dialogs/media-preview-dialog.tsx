import { Copy, ExternalLink, FileImage, Music, Video } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Dialog } from '@/components/dialog'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'

import type { PreviewMedia } from '../../lib/task-media'

export type { PreviewMedia } from '../../lib/task-media'

interface MediaPreviewDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  media: PreviewMedia[]
  taskId?: string
}

function MediaCard({ item }: { item: PreviewMedia }) {
  const { t } = useTranslation()
  const [hasError, setHasError] = useState(false)
  const Icon =
    item.kind === 'video' ? Video : item.kind === 'audio' ? Music : FileImage

  const openExternal = () =>
    window.open(item.url, '_blank', 'noopener,noreferrer')
  const copyUrl = async () => {
    await navigator.clipboard.writeText(item.url)
    toast.success(t('Copied'))
  }

  return (
    <div className='bg-card overflow-hidden rounded-lg border'>
      <div className='text-muted-foreground flex items-center gap-2 border-b px-3 py-2 text-xs'>
        <Icon className='size-3.5' />
        <span className='truncate'>{item.title || t('Preview')}</span>
      </div>
      <div className='bg-muted/30 flex min-h-24 items-center justify-center p-3'>
        {hasError ? (
          <div className='flex flex-wrap items-center justify-center gap-2 py-5'>
            <span className='text-muted-foreground text-xs'>
              {t(
                item.kind === 'video'
                  ? 'Video playback failed'
                  : item.kind === 'audio'
                    ? 'Audio playback failed'
                    : 'Failed to load image'
              )}
            </span>
            <Button
              variant='outline'
              size='sm'
              className='h-7 gap-1 text-xs'
              onClick={openExternal}
            >
              <ExternalLink className='size-3' />
              {t('Open in new tab')}
            </Button>
          </div>
        ) : item.kind === 'video' ? (
          <video
            src={item.url}
            controls
            playsInline
            preload='metadata'
            className='max-h-[58vh] w-full rounded-md object-contain'
            onError={() => setHasError(true)}
          />
        ) : item.kind === 'audio' ? (
          <audio
            src={item.url}
            controls
            preload='metadata'
            className='w-full'
            onError={() => setHasError(true)}
          />
        ) : (
          <img
            src={item.url}
            alt={item.title || t('Generated image')}
            loading='lazy'
            className='max-h-[58vh] w-full rounded-md object-contain'
            onError={() => setHasError(true)}
          />
        )}
      </div>
      <div className='flex items-center justify-end gap-1 border-t px-2 py-1.5'>
        <Button
          variant='ghost'
          size='sm'
          className='h-7 gap-1 text-xs'
          onClick={openExternal}
        >
          <ExternalLink className='size-3' />
          {t('Open in new tab')}
        </Button>
        <Button
          variant='ghost'
          size='sm'
          className='h-7 gap-1 text-xs'
          onClick={copyUrl}
        >
          <Copy className='size-3' />
          {t('Copy Link')}
        </Button>
      </div>
    </div>
  )
}

export function MediaPreviewDialog({
  open,
  onOpenChange,
  media,
  taskId,
}: MediaPreviewDialogProps) {
  const { t } = useTranslation()

  return (
    <Dialog
      open={open}
      onOpenChange={onOpenChange}
      title={t('Preview')}
      description={taskId ? `${t('Task ID:')} ${taskId}` : undefined}
      contentClassName='sm:max-w-3xl'
      contentHeight='auto'
      bodyClassName='space-y-3'
    >
      <ScrollArea className='max-h-[70vh]'>
        <div className='space-y-3 pr-2'>
          {media.map((item, index) => (
            <MediaCard key={`${item.kind}:${item.url}:${index}`} item={item} />
          ))}
        </div>
      </ScrollArea>
    </Dialog>
  )
}
