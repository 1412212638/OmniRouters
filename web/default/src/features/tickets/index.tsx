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
import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  CheckCircle2,
  Inbox,
  Lock,
  Plus,
  RefreshCcw,
  RotateCcw,
  Search,
  Send,
  ShieldCheck,
  Ticket as TicketIcon,
  UserRound,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { useAuthStore } from '@/stores/auth-store'
import dayjs from '@/lib/dayjs'
import { formatTimestamp } from '@/lib/format'
import { cn } from '@/lib/utils'
import { SectionPageLayout } from '@/components/layout'
import { EmptyState } from '@/components/empty-state'
import { StatusBadge } from '@/components/status-badge'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import {
  NativeSelect,
  NativeSelectOption,
} from '@/components/ui/native-select'
import { Textarea } from '@/components/ui/textarea'
import {
  TICKET_CATEGORIES,
  TICKET_PRIORITIES,
  TICKET_STATUSES,
  getTicketCategoryLabel,
  getTicketPriorityMeta,
  getTicketStatusMeta,
} from './constants'
import {
  assignTicket,
  closeTicket,
  createTicket,
  getTicket,
  listTickets,
  reopenTicket,
  sendTicketMessage,
  updateTicketStatus,
} from './api'
import type { CreateTicketPayload, Ticket, TicketMessage } from './types'

type TicketsPageProps = {
  admin?: boolean
}

const PAGE_SIZE = 20

function getUnreadCount(ticket: Ticket, admin?: boolean) {
  return admin ? ticket.admin_unread_count : ticket.user_unread_count
}

function getUserLabel(ticket: Ticket) {
  return ticket.user_display_name || ticket.username || `#${ticket.user_id}`
}

function formatRelativeTime(timestamp: number) {
  if (!timestamp) return '-'
  return dayjs(timestamp * 1000).fromNow()
}

function TicketStatusBadge({ status }: { status: string }) {
  const { t } = useTranslation()
  const meta = getTicketStatusMeta(status)
  return (
    <StatusBadge
      label={t(meta.label)}
      variant={meta.variant}
      copyable={false}
    />
  )
}

function TicketPriorityBadge({ priority }: { priority: string }) {
  const { t } = useTranslation()
  const meta = getTicketPriorityMeta(priority)
  return (
    <StatusBadge
      label={t(meta.label)}
      variant={meta.variant}
      copyable={false}
    />
  )
}

function CreateTicketDialog({
  open,
  onOpenChange,
  onCreated,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  onCreated: (ticket: Ticket) => void
}) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [form, setForm] = useState<CreateTicketPayload>({
    title: '',
    category: 'general',
    priority: 'normal',
    content: '',
  })

  const createMutation = useMutation({
    mutationFn: createTicket,
    onSuccess: (res) => {
      if (!res.success || !res.data) {
        toast.error(res.message || t('Create failed'))
        return
      }
      toast.success(t('Ticket created'))
      queryClient.invalidateQueries({ queryKey: ['tickets'] })
      onCreated(res.data.ticket)
      onOpenChange(false)
      setForm({
        title: '',
        category: 'general',
        priority: 'normal',
        content: '',
      })
    },
  })

  const onSubmit = (event: FormEvent) => {
    event.preventDefault()
    createMutation.mutate(form)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-xl'>
        <form onSubmit={onSubmit} className='contents'>
          <DialogHeader>
            <DialogTitle>{t('New Ticket')}</DialogTitle>
            <DialogDescription>
              {t('Send the issue details to the support queue.')}
            </DialogDescription>
          </DialogHeader>

          <div className='grid gap-3'>
            <div className='grid gap-1.5'>
              <label className='text-sm font-medium' htmlFor='ticket-title'>
                {t('Title')}
              </label>
              <Input
                id='ticket-title'
                value={form.title}
                onChange={(event) =>
                  setForm((current) => ({
                    ...current,
                    title: event.target.value,
                  }))
                }
                maxLength={255}
                required
              />
            </div>

            <div className='grid gap-3 sm:grid-cols-2'>
              <div className='grid gap-1.5'>
                <label
                  className='text-sm font-medium'
                  htmlFor='ticket-category'
                >
                  {t('Category')}
                </label>
                <NativeSelect
                  id='ticket-category'
                  className='w-full'
                  value={form.category}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      category: event.target.value,
                    }))
                  }
                >
                  {TICKET_CATEGORIES.map((category) => (
                    <NativeSelectOption
                      key={category.value}
                      value={category.value}
                    >
                      {t(category.label)}
                    </NativeSelectOption>
                  ))}
                </NativeSelect>
              </div>

              <div className='grid gap-1.5'>
                <label
                  className='text-sm font-medium'
                  htmlFor='ticket-priority'
                >
                  {t('Priority')}
                </label>
                <NativeSelect
                  id='ticket-priority'
                  className='w-full'
                  value={form.priority}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      priority: event.target.value,
                    }))
                  }
                >
                  {TICKET_PRIORITIES.map((priority) => (
                    <NativeSelectOption
                      key={priority.value}
                      value={priority.value}
                    >
                      {t(priority.label)}
                    </NativeSelectOption>
                  ))}
                </NativeSelect>
              </div>
            </div>

            <div className='grid gap-1.5'>
              <label className='text-sm font-medium' htmlFor='ticket-content'>
                {t('Message')}
              </label>
              <Textarea
                id='ticket-content'
                value={form.content}
                onChange={(event) =>
                  setForm((current) => ({
                    ...current,
                    content: event.target.value,
                  }))
                }
                className='min-h-32'
                required
              />
            </div>
          </div>

          <DialogFooter>
            <Button
              type='button'
              variant='outline'
              onClick={() => onOpenChange(false)}
            >
              {t('Cancel')}
            </Button>
            <Button type='submit' disabled={createMutation.isPending}>
              <Plus />
              {createMutation.isPending ? t('Creating...') : t('Create')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

function TicketListItem({
  ticket,
  selected,
  admin,
  onSelect,
}: {
  ticket: Ticket
  selected: boolean
  admin?: boolean
  onSelect: () => void
}) {
  const { t } = useTranslation()
  const unreadCount = getUnreadCount(ticket, admin)

  return (
    <button
      type='button'
      onClick={onSelect}
      className={cn(
        'hover:bg-muted/60 flex w-full flex-col gap-2 border-b px-3 py-3 text-left transition-colors last:border-b-0',
        selected && 'bg-muted'
      )}
    >
      <div className='flex min-w-0 items-start justify-between gap-3'>
        <div className='min-w-0'>
          <div className='flex min-w-0 items-center gap-2'>
            {unreadCount > 0 && (
              <span className='bg-primary size-2 shrink-0 rounded-full' />
            )}
            <span className='truncate text-sm font-medium'>
              #{ticket.id} {ticket.title}
            </span>
          </div>
          {admin && (
            <div className='text-muted-foreground mt-1 flex items-center gap-1 text-xs'>
              <UserRound className='size-3' />
              <span className='truncate'>{getUserLabel(ticket)}</span>
            </div>
          )}
        </div>
        <span className='text-muted-foreground shrink-0 text-xs'>
          {formatRelativeTime(ticket.last_reply_at || ticket.created_at)}
        </span>
      </div>
      <div className='flex flex-wrap items-center gap-2'>
        <TicketStatusBadge status={ticket.status} />
        <TicketPriorityBadge priority={ticket.priority} />
        <span className='text-muted-foreground text-xs'>
          {t(getTicketCategoryLabel(ticket.category))}
        </span>
        {unreadCount > 0 && (
          <span className='bg-primary/10 text-primary rounded-full px-1.5 py-0.5 text-xs font-medium'>
            {unreadCount}
          </span>
        )}
      </div>
    </button>
  )
}

function TicketMessageItem({
  message,
  currentUserId,
}: {
  message: TicketMessage
  currentUserId?: number
}) {
  const { t } = useTranslation()
  const isMine = message.sender_id > 0 && message.sender_id === currentUserId
  const isAdmin = message.sender_role === 'admin'

  return (
    <div
      className={cn(
        'flex gap-3',
        isMine && !message.internal ? 'justify-end' : 'justify-start'
      )}
    >
      {!isMine && (
        <div
          className={cn(
            'mt-1 flex size-8 shrink-0 items-center justify-center rounded-lg border',
            isAdmin ? 'bg-primary/10 text-primary' : 'bg-muted'
          )}
        >
          {isAdmin ? (
            <ShieldCheck className='size-4' />
          ) : (
            <UserRound className='size-4' />
          )}
        </div>
      )}
      <div
        className={cn(
          'max-w-[min(760px,85%)] rounded-lg border px-3 py-2',
          isMine && !message.internal
            ? 'bg-primary text-primary-foreground border-primary'
            : 'bg-background',
          message.internal && 'border-amber-300 bg-amber-50 text-amber-950'
        )}
      >
        <div className='mb-1 flex flex-wrap items-center gap-2 text-xs opacity-80'>
          <span className='font-medium'>
            {message.internal
              ? t('Internal note')
              : message.sender_name || t(message.sender_role)}
          </span>
          <span>{formatTimestamp(message.created_at)}</span>
        </div>
        <div className='whitespace-pre-wrap break-words text-sm leading-6'>
          {message.content}
        </div>
      </div>
    </div>
  )
}

function TicketDetailPanel({
  ticket,
  messages,
  admin,
  onSend,
  onStatusChange,
  onAssignToMe,
  onClose,
  onReopen,
  sending,
  statusChanging,
}: {
  ticket?: Ticket
  messages: TicketMessage[]
  admin?: boolean
  onSend: (content: string, internal: boolean) => void
  onStatusChange: (status: string) => void
  onAssignToMe: () => void
  onClose: () => void
  onReopen: () => void
  sending?: boolean
  statusChanging?: boolean
}) {
  const { t } = useTranslation()
  const currentUserId = useAuthStore((state) => state.auth.user?.id)
  const [content, setContent] = useState('')
  const [internal, setInternal] = useState(false)

  useEffect(() => {
    setContent('')
    setInternal(false)
  }, [ticket?.id])

  if (!ticket) {
    return (
      <div className='flex h-full min-h-80 items-center justify-center rounded-lg border'>
        <EmptyState
          icon={Inbox}
          title={t('No ticket selected')}
          description={t('Select a ticket to view the conversation.')}
        />
      </div>
    )
  }

  const closed = ticket.status === 'closed'

  const submitReply = (event: FormEvent) => {
    event.preventDefault()
    if (!content.trim()) return
    onSend(content, internal)
    setContent('')
  }

  return (
    <div className='flex h-full min-h-[560px] flex-col rounded-lg border'>
      <div className='flex shrink-0 flex-col gap-3 border-b p-3 sm:p-4'>
        <div className='flex flex-wrap items-start justify-between gap-3'>
          <div className='min-w-0'>
            <div className='flex flex-wrap items-center gap-2'>
              <h3 className='truncate text-base font-semibold'>
                #{ticket.id} {ticket.title}
              </h3>
              <TicketStatusBadge status={ticket.status} />
              <TicketPriorityBadge priority={ticket.priority} />
            </div>
            <div className='text-muted-foreground mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs'>
              <span>{t(getTicketCategoryLabel(ticket.category))}</span>
              <span>{formatTimestamp(ticket.created_at)}</span>
              {admin && <span>{getUserLabel(ticket)}</span>}
              {ticket.assigned_admin_name && (
                <span>
                  {t('Assignee')}: {ticket.assigned_admin_name}
                </span>
              )}
            </div>
          </div>

          <div className='flex flex-wrap items-center gap-2'>
            {admin && (
              <>
                <NativeSelect
                  size='sm'
                  value={ticket.status}
                  disabled={statusChanging}
                  onChange={(event) => onStatusChange(event.target.value)}
                >
                  {TICKET_STATUSES.map((status) => (
                    <NativeSelectOption key={status.value} value={status.value}>
                      {t(status.label)}
                    </NativeSelectOption>
                  ))}
                </NativeSelect>
                <Button
                  size='sm'
                  variant='outline'
                  onClick={onAssignToMe}
                  disabled={statusChanging}
                >
                  <CheckCircle2 />
                  {t('Assign to me')}
                </Button>
              </>
            )}
            {closed ? (
              !admin && (
                <Button size='sm' variant='outline' onClick={onReopen}>
                  <RotateCcw />
                  {t('Reopen')}
                </Button>
              )
            ) : (
              <Button size='sm' variant='outline' onClick={onClose}>
                <Lock />
                {t('Close')}
              </Button>
            )}
          </div>
        </div>
      </div>

      <div className='min-h-0 flex-1 space-y-4 overflow-auto p-3 sm:p-4'>
        {messages.map((message) => (
          <TicketMessageItem
            key={message.id}
            message={message}
            currentUserId={currentUserId}
          />
        ))}
      </div>

      <form
        onSubmit={submitReply}
        className='bg-background shrink-0 space-y-2 border-t p-3 sm:p-4'
      >
        {admin && (
          <label className='flex w-fit items-center gap-2 text-xs font-medium'>
            <input
              type='checkbox'
              checked={internal}
              onChange={(event) => setInternal(event.target.checked)}
              className='accent-primary size-3.5'
            />
            {t('Internal note')}
          </label>
        )}
        <div className='flex flex-col gap-2 sm:flex-row'>
          <Textarea
            value={content}
            onChange={(event) => setContent(event.target.value)}
            disabled={closed || sending}
            placeholder={closed ? t('Ticket is closed') : t('Write a reply...')}
            className='min-h-20 flex-1'
          />
          <Button
            type='submit'
            disabled={closed || sending || !content.trim()}
            className='sm:self-end'
          >
            <Send />
            {sending ? t('Sending...') : t('Send')}
          </Button>
        </div>
      </form>
    </div>
  )
}

export function TicketsPage({ admin = false }: TicketsPageProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const currentUserId = useAuthStore((state) => state.auth.user?.id)
  const [createOpen, setCreateOpen] = useState(false)
  const [page, setPage] = useState(1)
  const [status, setStatus] = useState('all')
  const [category, setCategory] = useState('all')
  const [priority, setPriority] = useState('all')
  const [keywordDraft, setKeywordDraft] = useState('')
  const [keyword, setKeyword] = useState('')
  const [selectedId, setSelectedId] = useState<number | null>(null)

  const listQueryKey = useMemo(
    () => [
      'tickets',
      admin ? 'admin' : 'self',
      page,
      status,
      category,
      priority,
      keyword,
    ],
    [admin, page, status, category, priority, keyword]
  )

  const listQuery = useQuery({
    queryKey: listQueryKey,
    queryFn: async () => {
      const res = await listTickets({
        admin,
        p: page,
        page_size: PAGE_SIZE,
        status,
        category,
        priority,
        keyword,
      })
      if (!res.success || !res.data) {
        toast.error(res.message || t('Failed to load tickets'))
        return { items: [], total: 0, page, page_size: PAGE_SIZE }
      }
      return res.data
    },
    placeholderData: (previousData) => previousData,
  })

  const tickets = listQuery.data?.items ?? []
  const total = listQuery.data?.total ?? 0
  const pageCount = Math.max(1, Math.ceil(total / PAGE_SIZE))

  useEffect(() => {
    if (!selectedId && tickets.length > 0) {
      setSelectedId(tickets[0].id)
    }
  }, [selectedId, tickets])

  useEffect(() => {
    if (page > pageCount) {
      setPage(pageCount)
    }
  }, [page, pageCount])

  const detailQuery = useQuery({
    queryKey: ['ticket-detail', admin ? 'admin' : 'self', selectedId],
    enabled: selectedId != null,
    queryFn: async () => {
      const res = await getTicket(selectedId!, admin)
      if (!res.success || !res.data) {
        toast.error(res.message || t('Failed to load ticket'))
        return null
      }
      queryClient.invalidateQueries({ queryKey: ['tickets'] })
      return res.data
    },
  })

  const sendMutation = useMutation({
    mutationFn: ({
      id,
      content,
      internal,
    }: {
      id: number
      content: string
      internal: boolean
    }) => sendTicketMessage(id, { content, internal }, admin),
    onSuccess: (res) => {
      if (!res.success) {
        toast.error(res.message || t('Send failed'))
        return
      }
      queryClient.invalidateQueries({ queryKey: ['tickets'] })
      queryClient.invalidateQueries({ queryKey: ['ticket-detail'] })
    },
  })

  const statusMutation = useMutation({
    mutationFn: ({ id, nextStatus }: { id: number; nextStatus: string }) =>
      updateTicketStatus(id, nextStatus),
    onSuccess: (res) => {
      if (!res.success) {
        toast.error(res.message || t('Update failed'))
        return
      }
      queryClient.invalidateQueries({ queryKey: ['tickets'] })
      queryClient.invalidateQueries({ queryKey: ['ticket-detail'] })
    },
  })

  const assignMutation = useMutation({
    mutationFn: ({ id, adminId }: { id: number; adminId: number }) =>
      assignTicket(id, adminId),
    onSuccess: (res) => {
      if (!res.success) {
        toast.error(res.message || t('Assign failed'))
        return
      }
      toast.success(t('Assigned'))
      queryClient.invalidateQueries({ queryKey: ['tickets'] })
      queryClient.invalidateQueries({ queryKey: ['ticket-detail'] })
    },
  })

  const closeMutation = useMutation({
    mutationFn: ({ id }: { id: number }) => closeTicket(id, admin),
    onSuccess: (res) => {
      if (!res.success) {
        toast.error(res.message || t('Close failed'))
        return
      }
      queryClient.invalidateQueries({ queryKey: ['tickets'] })
      queryClient.invalidateQueries({ queryKey: ['ticket-detail'] })
    },
  })

  const reopenMutation = useMutation({
    mutationFn: ({ id }: { id: number }) => reopenTicket(id),
    onSuccess: (res) => {
      if (!res.success) {
        toast.error(res.message || t('Reopen failed'))
        return
      }
      queryClient.invalidateQueries({ queryKey: ['tickets'] })
      queryClient.invalidateQueries({ queryKey: ['ticket-detail'] })
    },
  })

  const runSearch = (event: FormEvent) => {
    event.preventDefault()
    setPage(1)
    setKeyword(keywordDraft.trim())
  }

  const detail = detailQuery.data
  const selectedTicket = detail?.ticket
  const selectedMessages = detail?.messages ?? []

  return (
    <>
      <SectionPageLayout>
        <SectionPageLayout.Title>
          {admin ? t('Tickets') : t('Support Tickets')}
        </SectionPageLayout.Title>
        <SectionPageLayout.Description>
          {admin
            ? t('Review and resolve user support requests.')
            : t('Track your support requests and replies.')}
        </SectionPageLayout.Description>
        <SectionPageLayout.Actions>
          {!admin && (
            <Button onClick={() => setCreateOpen(true)}>
              <Plus />
              {t('New Ticket')}
            </Button>
          )}
        </SectionPageLayout.Actions>
        <SectionPageLayout.Content>
          <div className='grid min-h-[640px] gap-3 lg:grid-cols-[380px_minmax(0,1fr)]'>
            <div className='flex min-h-0 flex-col rounded-lg border'>
              <div className='shrink-0 space-y-3 border-b p-3'>
                <form onSubmit={runSearch} className='flex gap-2'>
                  <div className='relative min-w-0 flex-1'>
                    <Search className='text-muted-foreground pointer-events-none absolute top-1/2 left-2.5 size-4 -translate-y-1/2' />
                    <Input
                      value={keywordDraft}
                      onChange={(event) => setKeywordDraft(event.target.value)}
                      placeholder={t('Search tickets...')}
                      className='pl-8'
                    />
                  </div>
                  <Button type='submit' variant='outline' size='icon'>
                    <Search />
                    <span className='sr-only'>{t('Search')}</span>
                  </Button>
                  <Button
                    type='button'
                    variant='outline'
                    size='icon'
                    onClick={() => listQuery.refetch()}
                  >
                    <RefreshCcw />
                    <span className='sr-only'>{t('Refresh')}</span>
                  </Button>
                </form>

                <div className='grid grid-cols-3 gap-2'>
                  <NativeSelect
                    size='sm'
                    className='w-full'
                    value={status}
                    onChange={(event) => {
                      setPage(1)
                      setStatus(event.target.value)
                    }}
                  >
                    <NativeSelectOption value='all'>
                      {t('All Statuses')}
                    </NativeSelectOption>
                    {TICKET_STATUSES.map((item) => (
                      <NativeSelectOption key={item.value} value={item.value}>
                        {t(item.label)}
                      </NativeSelectOption>
                    ))}
                  </NativeSelect>

                  <NativeSelect
                    size='sm'
                    className='w-full'
                    value={category}
                    onChange={(event) => {
                      setPage(1)
                      setCategory(event.target.value)
                    }}
                  >
                    <NativeSelectOption value='all'>
                      {t('All Categories')}
                    </NativeSelectOption>
                    {TICKET_CATEGORIES.map((item) => (
                      <NativeSelectOption key={item.value} value={item.value}>
                        {t(item.label)}
                      </NativeSelectOption>
                    ))}
                  </NativeSelect>

                  <NativeSelect
                    size='sm'
                    className='w-full'
                    value={priority}
                    onChange={(event) => {
                      setPage(1)
                      setPriority(event.target.value)
                    }}
                  >
                    <NativeSelectOption value='all'>
                      {t('All Priorities')}
                    </NativeSelectOption>
                    {TICKET_PRIORITIES.map((item) => (
                      <NativeSelectOption key={item.value} value={item.value}>
                        {t(item.label)}
                      </NativeSelectOption>
                    ))}
                  </NativeSelect>
                </div>
              </div>

              <div className='min-h-0 flex-1 overflow-auto'>
                {listQuery.isLoading ? (
                  <div className='space-y-2 p-3'>
                    {Array.from({ length: 6 }).map((_, index) => (
                      <div
                        key={index}
                        className='bg-muted h-20 animate-pulse rounded-lg'
                      />
                    ))}
                  </div>
                ) : tickets.length === 0 ? (
                  <div className='flex h-full min-h-80 items-center justify-center p-4'>
                    <EmptyState
                      icon={TicketIcon}
                      title={t('No Tickets Found')}
                      description={t('No tickets match the current filters.')}
                    />
                  </div>
                ) : (
                  tickets.map((ticket) => (
                    <TicketListItem
                      key={ticket.id}
                      ticket={ticket}
                      selected={selectedId === ticket.id}
                      admin={admin}
                      onSelect={() => setSelectedId(ticket.id)}
                    />
                  ))
                )}
              </div>

              <div className='text-muted-foreground flex shrink-0 items-center justify-between border-t p-3 text-xs'>
                <span>
                  {total} {t('tickets')}
                </span>
                <div className='flex items-center gap-2'>
                  <Button
                    size='xs'
                    variant='outline'
                    disabled={page <= 1}
                    onClick={() => setPage((current) => current - 1)}
                  >
                    {t('Previous')}
                  </Button>
                  <span>
                    {page} / {pageCount}
                  </span>
                  <Button
                    size='xs'
                    variant='outline'
                    disabled={page >= pageCount}
                    onClick={() => setPage((current) => current + 1)}
                  >
                    {t('Next')}
                  </Button>
                </div>
              </div>
            </div>

            <TicketDetailPanel
              ticket={selectedTicket}
              messages={selectedMessages}
              admin={admin}
              sending={sendMutation.isPending}
              statusChanging={
                statusMutation.isPending ||
                assignMutation.isPending ||
                closeMutation.isPending ||
                reopenMutation.isPending
              }
              onSend={(content, internal) => {
                if (!selectedId) return
                sendMutation.mutate({ id: selectedId, content, internal })
              }}
              onStatusChange={(nextStatus) => {
                if (!selectedId) return
                statusMutation.mutate({ id: selectedId, nextStatus })
              }}
              onAssignToMe={() => {
                if (!selectedId || !currentUserId) return
                assignMutation.mutate({ id: selectedId, adminId: currentUserId })
              }}
              onClose={() => {
                if (!selectedId) return
                closeMutation.mutate({ id: selectedId })
              }}
              onReopen={() => {
                if (!selectedId) return
                reopenMutation.mutate({ id: selectedId })
              }}
            />
          </div>
        </SectionPageLayout.Content>
      </SectionPageLayout>

      <CreateTicketDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        onCreated={(ticket) => setSelectedId(ticket.id)}
      />
    </>
  )
}
