import { ExternalLink, Megaphone } from 'lucide-react'
import { memo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { RichContent } from '@/components/rich-content'
import { ScrollArea } from '@/components/ui/scroll-area'
import { useAnnouncements } from '@/features/dashboard/hooks/use-status-data'
import type { AnnouncementItem } from '@/features/dashboard/types'
import { getAnnouncementColorClass } from '@/lib/colors'
import { cn } from '@/lib/utils'

import { PanelWrapper } from '../ui/panel-wrapper'
import { AnnouncementDetailModal } from './announcement-detail-dialog'

const AnnouncementStatusDot = memo(function AnnouncementStatusDot(props: {
  type?: string
}) {
  return (
    <span
      className={cn(
        'mt-1.5 inline-block size-2 shrink-0 rounded-full',
        getAnnouncementColorClass(props.type)
      )}
    />
  )
})

export function AnnouncementsPanel() {
  const { t } = useTranslation()
  const { items: list, loading } = useAnnouncements()
  const [selectedAnnouncement, setSelectedAnnouncement] =
    useState<AnnouncementItem | null>(null)
  const [isDialogOpen, setIsDialogOpen] = useState(false)

  const handleAnnouncementClick = (item: AnnouncementItem) => {
    setSelectedAnnouncement(item)
    setIsDialogOpen(true)
  }

  return (
    <PanelWrapper
      title={
        <span className='flex items-center gap-2'>
          <Megaphone className='text-muted-foreground/60 size-4' />
          {t('Announcements')}
        </span>
      }
      description={t('Latest platform updates and notices')}
      loading={loading}
      empty={!list.length}
      emptyMessage={t('No announcements at this time')}
      height='h-96'
      contentClassName='p-0'
    >
      <ScrollArea className='h-96'>
        <div className='px-3 py-5 sm:px-5'>
          {list.map((item: AnnouncementItem, idx: number) => {
            const key = item.id ?? `announcement-${idx}`
            const date = item.publishDate
              ? new Date(item.publishDate)
              : undefined
            return (
              <article
                key={key}
                className={cn(
                  'group flex w-full text-left',
                  idx < list.length - 1 && 'pb-6'
                )}
              >
                <div className='w-16 shrink-0 pt-0.5 sm:w-20'>
                  {date && (
                    <time className='text-muted-foreground text-xs font-medium'>
                      {date.toLocaleDateString('en-US', {
                        month: 'short',
                        day: '2-digit',
                      })}
                    </time>
                  )}
                </div>
                <div className='border-border/70 relative flex min-w-0 flex-1 gap-4 border-l border-dashed pl-5'>
                  <span className='bg-background border-primary absolute top-1.5 -left-[5px] size-2.5 rounded-full border-2' />
                  <div className='min-w-0 flex-1'>
                    <div className='mb-1 flex items-start gap-2'>
                      <AnnouncementStatusDot type={item.type} />
                      <RichContent
                        breaks
                        content={item.content}
                        className='min-w-0 flex-1 text-sm [&_img]:h-auto [&_img]:max-w-full'
                      />
                    </div>
                    <button
                      type='button'
                      onClick={() => handleAnnouncementClick(item)}
                      className='text-muted-foreground mt-2 inline-flex items-center gap-1 text-xs underline-offset-2 hover:underline'
                    >
                      {t('Click for details')}
                      <ExternalLink className='size-3' />
                    </button>
                  </div>
                </div>
              </article>
            )
          })}
        </div>
      </ScrollArea>

      <AnnouncementDetailModal
        open={isDialogOpen}
        onOpenChange={setIsDialogOpen}
        announcement={selectedAnnouncement}
      />
    </PanelWrapper>
  )
}
