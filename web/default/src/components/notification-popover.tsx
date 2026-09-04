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
import { useState } from 'react'
import type { TFunction } from 'i18next'
import { Bell, Megaphone } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { RichContent } from '@/components/rich-content'
import { getAnnouncementColorClass } from '@/lib/colors'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '@/components/ui/empty'
import {
  Popover,
  PopoverContent,
  PopoverHeader,
  PopoverTitle,
  PopoverTrigger,
} from '@/components/ui/popover'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

interface AnnouncementItem {
  id?: number | string
  type?: string
  content?: string
  extra?: string
  publishDate?: string | Date
}

interface NotificationPopoverProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  unreadCount: number
  activeTab: 'notice' | 'announcements'
  onTabChange: (tab: 'notice' | 'announcements') => void
  notice: string
  announcements: AnnouncementItem[]
  loading: boolean
  className?: string
}

/**
 * Announcement status dot indicator
 */
function AnnouncementDot({ type }: { type?: string }) {
  return (
    <span
      className={cn(
        'mt-1.5 inline-block size-2 shrink-0 rounded-full',
        getAnnouncementColorClass(type)
      )}
    />
  )
}

function getAnnouncementRenderKey(announcement: AnnouncementItem): string {
  if (announcement.id !== undefined && announcement.id !== null) {
    return `id:${announcement.id}`
  }

  return JSON.stringify({
    content: announcement.content ?? '',
    extra: announcement.extra ?? '',
    publishDate: announcement.publishDate ?? '',
    type: announcement.type ?? '',
  })
}

function getContentParts(content: string) {
  const lines = content
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean)

  return {
    title: lines[0] || content,
    body: lines.length > 1 ? lines.slice(1).join('\n') : '',
  }
}

function TimelineEntry({
  date,
  type,
  content,
  extra,
  expanded,
  onToggle,
  t,
}: {
  date?: string
  type?: string
  content: string
  extra?: string
  expanded: boolean
  onToggle: () => void
  t: TFunction
}) {
  const { title, body } = getContentParts(content)
  const hasMore = Boolean(body || extra)

  return (
    <div className='flex gap-3'>
      <div className='w-12 shrink-0 pt-0.5'>
        {date ? (
          <time className='text-muted-foreground rounded-full border px-1.5 py-0.5 text-[10px] font-medium'>
            {date}
          </time>
        ) : null}
      </div>
      <div className='relative min-w-0 flex-1 border-l border-dashed border-border/70 pb-5 pl-4'>
        <span className='bg-background absolute -left-[5px] top-1.5 size-2.5 rounded-full border-2 border-primary' />
        <div className='mb-1 flex items-start gap-2'>
          <AnnouncementDot type={type} />
          <p className='min-w-0 flex-1 text-sm font-semibold leading-5'>
            <RichContent breaks content={title} />
          </p>
        </div>
        {expanded ? (
          <div className='text-muted-foreground text-xs leading-5'>
            <RichContent breaks content={body || content} />
            {extra ? <RichContent breaks content={extra} /> : null}
          </div>
        ) : null}
        {hasMore ? (
          <button
            type='button'
            onClick={onToggle}
            className='text-muted-foreground mt-1 inline-flex items-center gap-1 text-xs underline-offset-2 hover:underline'
          >
            {expanded ? t('Collapse') : t('Expand')}
          </button>
        ) : null}
      </div>
    </div>
  )
}

/**
 * Empty state component
 */
function EmptyState({
  icon,
  title,
  description,
}: {
  icon: React.ReactNode
  title: string
  description?: string
}) {
  return (
    <Empty className='min-h-48 border-0 p-4'>
      <EmptyHeader>
        <EmptyMedia variant='icon'>{icon}</EmptyMedia>
        <EmptyTitle>{title}</EmptyTitle>
        {description ? (
          <EmptyDescription>{description}</EmptyDescription>
        ) : null}
      </EmptyHeader>
    </Empty>
  )
}

/**
 * Notice tab content
 */
function NoticeContent({
  notice,
  loading,
  t,
}: {
  notice: string
  loading: boolean
  t: TFunction
}) {
  if (loading) {
    return (
      <EmptyState
        icon={<Bell />}
        title={t('Loading...')}
        description={t('Latest platform updates and notices')}
      />
    )
  }

  if (!notice) {
    return (
      <EmptyState icon={<Bell />} title={t('No announcements at this time')} />
    )
  }

  return <NoticeTimeline notice={notice} t={t} />
}

function NoticeTimeline({ notice, t }: { notice: string; t: TFunction }) {
  const [expanded, setExpanded] = useState(true)
  return (
    <ScrollArea className='h-[min(52vh,28rem)] pr-3'>
      <div className='py-3'>
        <TimelineEntry
          content={notice}
          expanded={expanded}
          onToggle={() => setExpanded((value) => !value)}
          t={t}
        />
      </div>
    </ScrollArea>
  )
}

/**
 * Announcements tab content
 */
function AnnouncementsContent({
  announcements,
  loading,
  t,
}: {
  announcements: AnnouncementItem[]
  loading: boolean
  t: TFunction
}) {
  const [expandedKeys, setExpandedKeys] = useState<Set<string>>(
    () => new Set(announcements.map(getAnnouncementRenderKey))
  )
  if (loading) {
    return (
      <EmptyState
        icon={<Megaphone />}
        title={t('Loading...')}
        description={t('Latest platform updates and notices')}
      />
    )
  }

  if (announcements.length === 0) {
    return (
      <EmptyState icon={<Megaphone />} title={t('No system announcements')} />
    )
  }

  return (
    <ScrollArea className='h-[min(52vh,28rem)] pr-3'>
      <div className='flex flex-col py-3'>
        {announcements.map((item) => {
          const announcementKey = getAnnouncementRenderKey(item)
          const publishDate = item.publishDate
            ? new Date(item.publishDate)
            : null
          const dateLabel = publishDate
            ? publishDate.toLocaleDateString(undefined, {
                month: '2-digit',
                day: '2-digit',
              })
            : undefined
          const expanded = expandedKeys.has(announcementKey)

          return (
            <TimelineEntry
              key={announcementKey}
              date={dateLabel}
              type={item.type}
              content={item.content || ''}
              extra={item.extra}
              expanded={expanded}
              onToggle={() =>
                setExpandedKeys((current) => {
                  const next = new Set(current)
                  if (next.has(announcementKey)) next.delete(announcementKey)
                  else next.add(announcementKey)
                  return next
                })
              }
              t={t}
            />
          )
        })}
      </div>
    </ScrollArea>
  )
}

/**
 * Notification popover with Notice and Announcements tabs
 */
export function NotificationPopover({
  open,
  onOpenChange,
  unreadCount,
  activeTab,
  onTabChange,
  notice,
  announcements,
  loading,
  className,
}: NotificationPopoverProps) {
  const { t } = useTranslation()
  return (
    <Popover open={open} onOpenChange={onOpenChange}>
      <PopoverTrigger
        render={
          <Button
            variant='ghost'
            size='icon'
            className={cn('relative size-9', className)}
            aria-label={t('Notifications')}
          />
        }
      >
        <Bell className='size-[1.2rem]' />
        {unreadCount > 0 ? (
          <Badge
            variant='destructive'
            className='absolute -top-1 -right-1 flex h-5 min-w-5 items-center justify-center px-1 text-[10px] font-semibold tabular-nums'
          >
            {unreadCount > 99 ? '99+' : unreadCount}
          </Badge>
        ) : null}
      </PopoverTrigger>

      <PopoverContent
        align='end'
        sideOffset={8}
        className='w-[min(26rem,calc(100vw-1rem))] gap-3 p-3'
      >
        <PopoverHeader className='gap-1 px-1'>
          <PopoverTitle>{t('System Announcements')}</PopoverTitle>
          <p className='text-muted-foreground text-xs'>
            {t('Latest platform updates and notices')}
          </p>
        </PopoverHeader>

        <Tabs
          value={activeTab}
          onValueChange={onTabChange as (value: string) => void}
        >
          <TabsList className='grid w-full grid-cols-2'>
            <TabsTrigger value='notice' className='gap-1.5'>
              <Bell className='size-3.5' />
              {t('Notice')}
            </TabsTrigger>
            <TabsTrigger value='announcements' className='gap-1.5'>
              <Megaphone className='size-3.5' />
              {t('Timeline')}
            </TabsTrigger>
          </TabsList>

          <TabsContent value='notice' className='mt-2'>
            <NoticeContent notice={notice} loading={loading} t={t} />
          </TabsContent>

          <TabsContent value='announcements' className='mt-2'>
            <AnnouncementsContent
              announcements={announcements}
              loading={loading}
              t={t}
            />
          </TabsContent>
        </Tabs>

        <div className='flex justify-end'>
          <Button size='sm' onClick={() => onOpenChange(false)}>
            {t('Close')}
          </Button>
        </div>
      </PopoverContent>
    </Popover>
  )
}
