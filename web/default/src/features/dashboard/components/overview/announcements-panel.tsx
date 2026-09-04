/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { memo, useState } from 'react'
import { ExternalLink, Megaphone } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { getAnnouncementColorClass } from '@/lib/colors'
import { cn } from '@/lib/utils'
import { ScrollArea } from '@/components/ui/scroll-area'
import { useAnnouncements } from '@/features/dashboard/hooks/use-status-data'
import type { AnnouncementItem } from '@/features/dashboard/types'
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

function getAnnouncementParts(content: string) {
  const lines = content
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean)

  return {
    title: lines[0] || content,
    body: lines.length > 1 ? lines.slice(1).join('\n') : '',
  }
}

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
            const { title, body } = getAnnouncementParts(item.content)
            const date = item.publishDate
              ? new Date(item.publishDate)
              : undefined
            return (
              <button
                key={key}
                type='button'
                onClick={() => handleAnnouncementClick(item)}
                className={cn(
                  'group flex w-full text-left',
                  idx < list.length - 1 && 'pb-6'
                )}
              >
                <div className='w-16 shrink-0 pt-0.5 sm:w-20'>
                  {date && (
                    <time className='text-muted-foreground text-xs font-medium'>
                      {date.toLocaleDateString(undefined, {
                        month: '2-digit',
                        day: '2-digit',
                      })}
                    </time>
                  )}
                </div>
                <div className='relative flex min-w-0 flex-1 gap-4 border-l border-dashed border-border/70 pl-5'>
                  <span className='bg-background absolute -left-[5px] top-1.5 size-2.5 rounded-full border-2 border-primary' />
                  <div className='min-w-0 flex-1'>
                    <div className='mb-1 flex items-start gap-2'>
                      <AnnouncementStatusDot type={item.type} />
                      <p className='line-clamp-2 text-sm font-semibold leading-5'>
                        {title}
                      </p>
                    </div>
                    {body && (
                      <p className='text-muted-foreground line-clamp-3 whitespace-pre-line text-xs leading-5'>
                        {body}
                      </p>
                    )}
                    <span className='text-muted-foreground/60 mt-2 inline-flex items-center gap-1 text-xs opacity-0 transition-opacity group-hover:opacity-100'>
                      {t('Click for details')}
                      <ExternalLink className='size-3' />
                    </span>
                  </div>
                </div>
              </button>
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
