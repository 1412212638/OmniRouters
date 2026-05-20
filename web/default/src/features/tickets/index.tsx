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
import {
  useEffect,
  useMemo,
  useState,
  type FormEvent,
  type ReactNode,
} from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  CheckCircle2,
  Clock3,
  FileText,
  Inbox,
  Lock,
  Mail,
  Plus,
  RefreshCcw,
  Reply,
  RotateCcw,
  Search,
  Send,
  ShieldCheck,
  Tag,
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
        'group grid w-full gap-2 border-b border-l-2 border-l-transparent px-3.5 py-3 text-left transition-colors last:border-b-0 hover:bg-muted/60',
        selected && 'border-l-primary bg-muted/70'
      )}
    >
      <div className='flex min-w-0 items-start justify-between gap-3'>
        <div className='min-w-0 space-y-1'>
          <div className='flex min-w-0 items-center gap-2'>
            {unreadCount > 0 && (
              <span
                className='bg-primary size-2 shrink-0 rounded-full'
                aria-hidden='true'
              />
            )}
            <span className='text-muted-foreground shrink-0 text-xs font-medium'>
              #{ticket.id}
            </span>
            <span className='truncate text-sm font-semibold'>
              {ticket.title}
            </span>
          </div>
          <div className='text-muted-foreground flex min-w-0 items-center gap-1.5 text-xs'>
            <Mail className='size-3.5 shrink-0' />
            <span className='truncate'>
              {admin ? getUserLabel(ticket) : t('Support')}
            </span>
          </div>
        </div>
        <span className='text-muted-foreground shrink-0 text-xs'>
          {formatRelativeTime(ticket.last_reply_at || ticket.created_at)}
        </span>
      </div>

      <div className='flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1'>
        <TicketStatusBadge status={ticket.status} />
        <TicketPriorityBadge priority={ticket.priority} />
        <span className='text-muted-foreground inline-flex min-w-0 items-center gap-1 text-xs'>
          <Tag className='size-3 shrink-0' />
          <span className='truncate'>
            {t(getTicketCategoryLabel(ticket.category))}
          </span>
        </span>
        {unreadCount > 0 && (
          <span className='bg-primary/10 text-primary ml-auto rounded px-1.5 py-0.5 text-xs font-medium'>
            {unreadCount} {t('Unread')}
          </span>
        )}
      </div>
    </button>
  )
}

function TicketMessageItem({ message }: { message: TicketMessage }) {
  const { t } = useTranslation()
  const isAdmin = message.sender_role === 'admin'
  const isSystem = message.sender_role === 'system'
  const senderLabel = message.internal
    ? t('Internal note')
    : message.sender_name || t(message.sender_role)

  return (
    <article
      className={cn(
        'grid gap-3 px-4 py-4 sm:grid-cols-[2.25rem_minmax(0,1fr)]',
        message.internal && 'bg-amber-50/70 dark:bg-amber-950/20'
      )}
    >
      <div
        className={cn(
          'flex size-9 items-center justify-center rounded-md border',
          isAdmin && 'border-primary/20 bg-primary/10 text-primary',
          isSystem && 'bg-muted text-muted-foreground',
          !isAdmin && !isSystem && 'bg-background text-muted-foreground'
        )}
      >
        {isAdmin ? (
          <ShieldCheck className='size-4' />
        ) : isSystem ? (
          <FileText className='size-4' />
        ) : (
          <UserRound className='size-4' />
        )}
      </div>

      <div className='min-w-0'>
        <div className='flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1'>
          <span className='truncate text-sm font-semibold'>
            {senderLabel}
          </span>
          <span className='text-muted-foreground text-xs'>
            {formatTimestamp(message.created_at)}
          </span>
          {message.internal && (
            <span className='rounded border border-amber-300 bg-amber-100 px-1.5 py-0.5 text-xs font-medium text-amber-900 dark:border-amber-800 dark:bg-amber-950 dark:text-amber-200'>
              {t('Internal note')}
            </span>
          )}
        </div>
        <div className='mt-2 whitespace-pre-wrap break-words text-sm leading-6'>
          {message.content}
        </div>
      </div>
    </article>
  )
}

function DetailField({
  label,
  value,
}: {
  label: string
  value: ReactNode
}) {
  return (
    <div className='grid gap-1'>
      <dt className='text-muted-foreground text-xs'>{label}</dt>
      <dd className='min-w-0 text-sm font-medium'>{value}</dd>
    </div>
  )
}

function TicketProperties({
  ticket,
  admin,
}: {
  ticket: Ticket
  admin?: boolean
}) {
  const { t } = useTranslation()

  return (
    <aside className='hidden w-72 shrink-0 border-l bg-muted/20 xl:block'>
      <div className='border-b px-4 py-3'>
        <h3 className='text-sm font-semibold'>{t('Ticket details')}</h3>
      </div>
      <dl className='grid gap-4 p-4'>
        <DetailField
          label={t('Status')}
          value={<TicketStatusBadge status={ticket.status} />}
        />
        <DetailField
          label={t('Priority')}
          value={<TicketPriorityBadge priority={ticket.priority} />}
        />
        <DetailField
          label={t('Category')}
          value={t(getTicketCategoryLabel(ticket.category))}
        />
        {admin && (
          <DetailField label={t('Requester')} value={getUserLabel(ticket)} />
        )}
        <DetailField
          label={t('Assignee')}
          value={ticket.assigned_admin_name || t('Unassigned')}
        />
        <DetailField
          label={t('Created At')}
          value={formatTimestamp(ticket.created_at)}
        />
        <DetailField
          label={t('Last Reply')}
          value={formatTimestamp(ticket.last_reply_at || ticket.created_at)}
        />
      </dl>
    </aside>
  )
}

function TicketReplyBox({
  admin,
  closed,
  sending,
  onSend,
}: {
  admin?: boolean
  closed: boolean
  sending?: boolean
  onSend: (content: string, internal: boolean) => void
}) {
  const { t } = useTranslation()
  const [content, setContent] = useState('')
  const [internal, setInternal] = useState(false)

  const submitReply = (event: FormEvent) => {
    event.preventDefault()
    if (!content.trim()) return
    onSend(content, internal)
    setContent('')
  }

  return (
    <form onSubmit={submitReply} className='shrink-0 border-t bg-background p-3'>
      <div className='mb-2 flex flex-wrap items-center justify-between gap-2'>
        <div className='flex items-center gap-2 text-sm font-semibold'>
          <Reply className='size-4' />
          {t('Reply to ticket')}
        </div>
        {admin && (
          <label className='text-muted-foreground flex items-center gap-2 text-xs font-medium'>
            <input
              type='checkbox'
              checked={internal}
              onChange={(event) => setInternal(event.target.checked)}
              className='accent-primary size-3.5'
            />
            {t('Internal note')}
          </label>
        )}
      </div>

      <div className='flex flex-col gap-2 sm:flex-row'>
        <Textarea
          value={content}
          onChange={(event) => setContent(event.target.value)}
          disabled={closed || sending}
          placeholder={closed ? t('Ticket is closed') : t('Write a reply...')}
          className='min-h-24 flex-1 resize-y'
        />
        <Button
          type='submit'
          disabled={closed || sending || !content.trim()}
          className='sm:self-end'
        >
          <Send />
          {sending ? t('Sending...') : t('Post Reply')}
        </Button>
      </div>
    </form>
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

  if (!ticket) {
    return (
      <div className='flex min-h-[640px] items-center justify-center rounded-lg border bg-background'>
        <EmptyState
          icon={Inbox}
          title={t('No ticket selected')}
          description={t('Select a ticket to view its details.')}
        />
      </div>
    )
  }

  const closed = ticket.status === 'closed'

  return (
    <div className='flex min-h-[640px] overflow-hidden rounded-lg border bg-background'>
      <div className='flex min-w-0 flex-1 flex-col'>
        <div className='shrink-0 border-b px-4 py-3'>
          <div className='flex flex-wrap items-start justify-between gap-3'>
            <div className='min-w-0'>
              <div className='text-muted-foreground mb-1 flex flex-wrap items-center gap-2 text-xs'>
                <span>#{ticket.id}</span>
                <span>{t(getTicketCategoryLabel(ticket.category))}</span>
                <span>{formatTimestamp(ticket.created_at)}</span>
                {admin && <span>{getUserLabel(ticket)}</span>}
              </div>
              <h3 className='break-words text-base font-semibold sm:text-lg'>
                {ticket.title}
              </h3>
              <div className='mt-2 flex flex-wrap items-center gap-2 xl:hidden'>
                <TicketStatusBadge status={ticket.status} />
                <TicketPriorityBadge priority={ticket.priority} />
                <span className='text-muted-foreground text-xs'>
                  {t('Assignee')}:{' '}
                  {ticket.assigned_admin_name || t('Unassigned')}
                </span>
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
                      <NativeSelectOption
                        key={status.value}
                        value={status.value}
                      >
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

        <div className='min-h-0 flex-1 overflow-auto bg-muted/20 p-3 sm:p-4'>
          <div className='mx-auto max-w-5xl overflow-hidden rounded-lg border bg-background'>
            <div className='flex flex-wrap items-center justify-between gap-2 border-b px-4 py-3'>
              <div className='flex items-center gap-2 text-sm font-semibold'>
                <Mail className='size-4' />
                {t('Correspondence')}
              </div>
              <span className='text-muted-foreground text-xs'>
                {messages.length} {t('messages')}
              </span>
            </div>
            {messages.length === 0 ? (
              <div className='flex min-h-48 items-center justify-center p-4'>
                <EmptyState
                  icon={FileText}
                  title={t('No messages yet')}
                  description={t('Replies will appear here.')}
                />
              </div>
            ) : (
              <div className='divide-y'>
                {messages.map((message) => (
                  <TicketMessageItem key={message.id} message={message} />
                ))}
              </div>
            )}
          </div>
        </div>

        <TicketReplyBox
          admin={admin}
          closed={closed}
          sending={sending}
          onSend={onSend}
        />
      </div>

      <TicketProperties ticket={ticket} admin={admin} />
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
    if (tickets.length === 0) {
      setSelectedId(null)
      return
    }
    if (!selectedId || !tickets.some((ticket) => ticket.id === selectedId)) {
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
          <div className='grid min-h-[640px] gap-3 xl:grid-cols-[360px_minmax(0,1fr)]'>
            <div className='flex min-h-0 flex-col overflow-hidden rounded-lg border bg-background'>
              <div className='shrink-0 border-b px-3.5 py-3'>
                <div className='mb-3 flex items-center justify-between gap-3'>
                  <div className='min-w-0'>
                    <div className='flex items-center gap-2 text-sm font-semibold'>
                      <Inbox className='size-4' />
                      {t('Ticket Inbox')}
                    </div>
                    <div className='text-muted-foreground mt-0.5 text-xs'>
                      {total} {t('tickets')}
                    </div>
                  </div>
                  <Button
                    type='button'
                    variant='outline'
                    size='icon-sm'
                    onClick={() => listQuery.refetch()}
                  >
                    <RefreshCcw />
                    <span className='sr-only'>{t('Refresh')}</span>
                  </Button>
                </div>

                <form onSubmit={runSearch} className='mb-2 flex gap-2'>
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

              <div className='text-muted-foreground flex shrink-0 items-center justify-between border-t px-3 py-2.5 text-xs'>
                <span className='inline-flex items-center gap-1'>
                  <Clock3 className='size-3.5' />
                  {t('Page')} {page} / {pageCount}
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
