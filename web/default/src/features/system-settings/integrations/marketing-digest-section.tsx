import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'

import {
  deleteMarketingDigestPromotion,
  generateMarketingDigest,
  listMarketingDigestPromotions,
  listMarketingDigestRecipients,
  listMarketingDigests,
  saveMarketingDigestPromotion,
  searchMarketingEmailUsers,
  sendMarketingDigest,
  updateEmailSettingOption,
  updateMarketingDigest,
} from '../api'
import { SettingsSection } from '../components/settings-section'
import type {
  MarketingDigest,
  MarketingDigestPromotion,
  MarketingDigestRecipient,
  MarketingEmailUser,
  SystemOption,
} from '../types'

type Props = { options: SystemOption[] }
type MarketingDigestPromotionForm = Omit<MarketingDigestPromotion, 'id'> & {
  id?: number
}

const emptyPromotion: Omit<MarketingDigestPromotion, 'id'> & { id?: number } = {
  model_name: '',
  description: '',
  discount: '',
  starts_at: 0,
  ends_at: 0,
  url: '',
  enabled: true,
}

function getDateValue(timestamp: number) {
  if (!timestamp) return ''
  const date = new Date(timestamp * 1000)
  date.setMinutes(date.getMinutes() - date.getTimezoneOffset())
  return date.toISOString().slice(0, 10)
}

function dateToTimestamp(value: string, endOfDay = false) {
  if (!value) return 0
  const date = new Date(`${value}T${endOfDay ? '23:59:59' : '00:00:00'}`)
  return Math.floor(date.getTime() / 1000)
}

function parseUserIDs(raw: string) {
  try {
    const parsed: unknown = JSON.parse(raw)
    return Array.isArray(parsed)
      ? parsed.filter((id): id is number => Number.isInteger(id) && id > 0)
      : []
  } catch {
    return []
  }
}

function formatDate(timestamp: number) {
  if (!timestamp) return '-'
  return new Date(timestamp * 1000).toLocaleDateString()
}

function digestStatusKey(status: MarketingDigest['status']) {
  const labels: Record<MarketingDigest['status'], string> = {
    pending_review: 'Awaiting review',
    sending: 'Sending',
    sent: 'Sent',
    sent_with_errors: 'Sent with errors',
    failed: 'Failed',
  }
  return labels[status]
}

function getOption(options: SystemOption[], key: string, fallback = '') {
  return options.find((option) => option.key === key)?.value ?? fallback
}

export function MarketingDigestSection({ options }: Props) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [intervalDays, setIntervalDays] = useState('0')
  const [reviewEmail, setReviewEmail] = useState('')
  const [savingSettings, setSavingSettings] = useState(false)
  const [periodDays, setPeriodDays] = useState('10')
  const [selectedDigestID, setSelectedDigestID] = useState<number | null>(null)
  const [subject, setSubject] = useState('')
  const [content, setContent] = useState('')
  const [audienceType, setAudienceType] =
    useState<MarketingDigest['audience_type']>('all')
  const [audienceGroup, setAudienceGroup] = useState('')
  const [selectedUserIDs, setSelectedUserIDs] = useState<number[]>([])
  const [searchKeyword, setSearchKeyword] = useState('')
  const [searchResults, setSearchResults] = useState<MarketingEmailUser[]>([])
  const [searchingUsers, setSearchingUsers] = useState(false)
  const [savingDraft, setSavingDraft] = useState(false)
  const [sending, setSending] = useState(false)
  const [recipientStatus, setRecipientStatus] = useState<
    MarketingDigestRecipient['status'] | 'all'
  >('all')
  const [promotionForm, setPromotionForm] =
    useState<MarketingDigestPromotionForm>(emptyPromotion)
  const [promotionEditing, setPromotionEditing] = useState(false)
  const [savingPromotion, setSavingPromotion] = useState(false)

  const promotionsQuery = useQuery({
    queryKey: ['marketing-digest-promotions'],
    queryFn: listMarketingDigestPromotions,
  })
  const digestsQuery = useQuery({
    queryKey: ['marketing-digests'],
    queryFn: listMarketingDigests,
  })
  const recipientQuery = useQuery({
    queryKey: [
      'marketing-digest-recipients',
      selectedDigestID,
      recipientStatus,
    ],
    queryFn: () =>
      listMarketingDigestRecipients(
        selectedDigestID ?? 0,
        recipientStatus === 'all' ? undefined : recipientStatus
      ),
    enabled: selectedDigestID !== null,
  })

  const digests = useMemo(
    () => digestsQuery.data?.data ?? [],
    [digestsQuery.data?.data]
  )
  const selectedDigest = useMemo(
    () => digests.find((digest) => digest.id === selectedDigestID),
    [digests, selectedDigestID]
  )
  const currentRecipients = recipientQuery.data?.data?.items ?? []
  const promotions = promotionsQuery.data?.data ?? []

  useEffect(() => {
    setIntervalDays(getOption(options, 'MarketingDigestIntervalDays', '0'))
    setReviewEmail(getOption(options, 'MarketingDigestReviewEmail'))
  }, [options])

  useEffect(() => {
    if (selectedDigestID === null && digests.length > 0) {
      setSelectedDigestID(digests[0].id)
    }
  }, [digests, selectedDigestID])

  useEffect(() => {
    if (!selectedDigest) return
    setSubject(selectedDigest.subject)
    setContent(selectedDigest.content)
    setAudienceType(selectedDigest.audience_type)
    setAudienceGroup(selectedDigest.audience_group || '')
    setSelectedUserIDs(parseUserIDs(selectedDigest.audience_user_ids))
    setSearchResults([])
  }, [selectedDigest])

  const refreshDigests = async () => {
    await queryClient.invalidateQueries({ queryKey: ['marketing-digests'] })
  }

  const handleSaveSettings = async () => {
    const days = Number(intervalDays)
    if (!Number.isInteger(days) || days < 0 || days > 365) {
      toast.error(t('Interval must be between 0 and 365 days'))
      return
    }
    setSavingSettings(true)
    try {
      const updates: [string, string][] = [
        ['MarketingDigestIntervalDays', String(days)],
        ['MarketingDigestReviewEmail', reviewEmail.trim()],
      ]
      for (const [key, value] of updates) {
        const result = await updateEmailSettingOption({ key, value })
        if (!result.success) {
          throw new Error(result.message || t('Failed to update setting'))
        }
      }
      await queryClient.invalidateQueries({
        queryKey: ['email-settings-options'],
      })
      toast.success(t('Newsletter settings saved'))
    } catch (error) {
      toast.error((error as Error)?.message || t('Failed to update setting'))
    } finally {
      setSavingSettings(false)
    }
  }

  const handleGenerate = async () => {
    const days = Number(periodDays)
    if (!Number.isInteger(days) || days < 1 || days > 365) {
      toast.error(t('Period must be between 1 and 365 days'))
      return
    }
    try {
      const result = await generateMarketingDigest(days)
      if (!result.success || !result.data?.digest) {
        throw new Error(
          result.message || t('Failed to generate newsletter draft')
        )
      }
      await refreshDigests()
      setSelectedDigestID(result.data.digest.id)
      toast.success(
        result.data.created
          ? t('Draft generated and reviewer notified')
          : t('Existing draft opened')
      )
    } catch (error) {
      toast.error(
        (error as Error)?.message || t('Failed to generate newsletter draft')
      )
    }
  }

  const handleSavePromotion = async () => {
    if (!promotionForm.model_name.trim()) {
      toast.error(t('Model name is required'))
      return
    }
    setSavingPromotion(true)
    try {
      const result = await saveMarketingDigestPromotion(promotionForm)
      if (!result.success) throw new Error(t('Failed to save promotion'))
      await queryClient.invalidateQueries({
        queryKey: ['marketing-digest-promotions'],
      })
      setPromotionForm(emptyPromotion)
      setPromotionEditing(false)
      toast.success(t('Promotion saved'))
    } catch (error) {
      toast.error((error as Error)?.message || t('Failed to save promotion'))
    } finally {
      setSavingPromotion(false)
    }
  }

  const handleDeletePromotion = async (promotion: MarketingDigestPromotion) => {
    if (!window.confirm(t('Delete this promotion?'))) return
    try {
      await deleteMarketingDigestPromotion(promotion.id)
      await queryClient.invalidateQueries({
        queryKey: ['marketing-digest-promotions'],
      })
      toast.success(t('Promotion deleted'))
    } catch (error) {
      toast.error((error as Error)?.message || t('Failed to delete promotion'))
    }
  }

  const handleSearchUsers = async () => {
    setSearchingUsers(true)
    try {
      const result = await searchMarketingEmailUsers({
        keyword: searchKeyword.trim(),
        p: 1,
        page_size: 20,
      })
      if (!result.success) {
        throw new Error(result.message || t('Failed to search users'))
      }
      setSearchResults(result.data?.items || [])
    } catch (error) {
      toast.error((error as Error)?.message || t('Failed to search users'))
    } finally {
      setSearchingUsers(false)
    }
  }

  const toggleUser = (id: number) => {
    setSelectedUserIDs((current) =>
      current.includes(id)
        ? current.filter((value) => value !== id)
        : [...current, id]
    )
  }

  const saveDraft = async () => {
    if (!selectedDigest || !subject.trim() || !content.trim()) {
      throw new Error(t('Subject and content are required'))
    }
    if (audienceType === 'group' && !audienceGroup.trim()) {
      throw new Error(t('Enter an audience group'))
    }
    if (audienceType === 'manual' && selectedUserIDs.length === 0) {
      throw new Error(t('Select at least one recipient'))
    }
    const result = await updateMarketingDigest(selectedDigest.id, {
      subject,
      content,
      audience_type: audienceType,
      audience_group: audienceGroup,
      user_ids: selectedUserIDs,
    })
    if (!result.success) throw new Error(t('Failed to save newsletter draft'))
    await refreshDigests()
  }

  const handleSaveDraft = async () => {
    setSavingDraft(true)
    try {
      await saveDraft()
      toast.success(t('Newsletter draft saved'))
    } catch (error) {
      toast.error(
        (error as Error)?.message || t('Failed to save newsletter draft')
      )
    } finally {
      setSavingDraft(false)
    }
  }

  const handleSend = async (retry = false) => {
    if (!selectedDigest) return
    if (!retry) {
      try {
        await saveDraft()
      } catch (error) {
        toast.error(
          (error as Error)?.message || t('Failed to save newsletter draft')
        )
        return
      }
    }
    let confirmationMessage = t(
      'Send this reviewed newsletter to the selected audience? Recipients who opted out will be skipped.'
    )
    if (retry && selectedDigest.status === 'sending') {
      confirmationMessage = t(
        'Resume delivery for this newsletter? Previously sent users will be skipped.'
      )
    } else if (retry) {
      confirmationMessage = t(
        'Retry failed recipients for this newsletter? Previously sent users will be skipped.'
      )
    }
    const confirmed = window.confirm(confirmationMessage)
    if (!confirmed) return
    setSending(true)
    try {
      const result = await sendMarketingDigest(selectedDigest.id, retry)
      if (!result.success) {
        throw new Error(t('Failed to start newsletter delivery'))
      }
      await Promise.all([
        refreshDigests(),
        queryClient.invalidateQueries({
          queryKey: ['marketing-digest-recipients', selectedDigest.id],
        }),
      ])
      let successMessage = t('Newsletter delivery started')
      if (retry && selectedDigest.status === 'sending') {
        successMessage = t('Newsletter delivery is active or has resumed')
      } else if (retry) {
        successMessage = t('Failed recipients queued for retry')
      }
      toast.success(successMessage)
    } catch (error) {
      toast.error(
        (error as Error)?.message || t('Failed to start newsletter delivery')
      )
    } finally {
      setSending(false)
    }
  }

  const editable = selectedDigest?.status === 'pending_review'
  let promotionSaveLabel = t('Add promotion')
  if (promotionEditing) promotionSaveLabel = t('Save promotion')
  if (savingPromotion) promotionSaveLabel = t('Saving...')

  let retryButtonLabel = t('Retry failed recipients')
  if (selectedDigest?.status === 'sending') {
    retryButtonLabel = t('Resume delivery')
  }
  if (sending) retryButtonLabel = t('Starting...')

  const previewContent = content
    .replaceAll('{{system_name}}', 'OmniRouters')
    .replaceAll('{{display_name}}', 'Alex Chen')
    .replaceAll('{{username}}', 'alex')
    .replaceAll('{{unsubscribe_url}}', '#unsubscribe')

  return (
    <SettingsSection
      title={t('Product update newsletters')}
      description={t(
        'Generate product update drafts for review, then send them to eligible users.'
      )}
    >
      <div className='flex flex-col gap-8'>
        <section className='flex flex-col gap-4 rounded-lg border p-4'>
          <div>
            <h4 className='text-sm font-semibold'>
              {t('Generation schedule')}
            </h4>
            <p className='text-muted-foreground mt-1 text-sm'>
              {t(
                'Set 0 for manual generation only. Scheduled generation creates a draft and emails the reviewer; it never sends automatically.'
              )}
            </p>
          </div>
          <div className='grid gap-4 md:grid-cols-2'>
            <label className='flex flex-col gap-2 text-sm'>
              <span>{t('Generation interval (days)')}</span>
              <Input
                type='number'
                min={0}
                max={365}
                value={intervalDays}
                onChange={(event) => setIntervalDays(event.target.value)}
              />
            </label>
            <label className='flex flex-col gap-2 text-sm'>
              <span>{t('Reviewer email')}</span>
              <Input
                type='email'
                value={reviewEmail}
                onChange={(event) => setReviewEmail(event.target.value)}
                placeholder={t('Blank uses the root administrator email')}
              />
            </label>
          </div>
          <div>
            <Button
              type='button'
              variant='outline'
              onClick={handleSaveSettings}
              disabled={savingSettings}
            >
              {savingSettings ? t('Saving...') : t('Save newsletter settings')}
            </Button>
          </div>
        </section>

        <section className='flex flex-col gap-4 rounded-lg border p-4'>
          <div>
            <h4 className='text-sm font-semibold'>{t('Model promotions')}</h4>
            <p className='text-muted-foreground mt-1 text-sm'>
              {t(
                'Promotion copy is informational and does not change model pricing or billing.'
              )}
            </p>
          </div>
          <div className='grid gap-3 md:grid-cols-2'>
            <Input
              value={promotionForm.model_name}
              onChange={(event) =>
                setPromotionForm({
                  ...promotionForm,
                  model_name: event.target.value,
                })
              }
              placeholder={t('Model name')}
            />
            <Input
              value={promotionForm.discount}
              onChange={(event) =>
                setPromotionForm({
                  ...promotionForm,
                  discount: event.target.value,
                })
              }
              placeholder={t('Offer or discount text')}
            />
            <Input
              value={promotionForm.description}
              onChange={(event) =>
                setPromotionForm({
                  ...promotionForm,
                  description: event.target.value,
                })
              }
              placeholder={t('Description')}
            />
            <Input
              value={promotionForm.url}
              onChange={(event) =>
                setPromotionForm({ ...promotionForm, url: event.target.value })
              }
              placeholder={t('Campaign URL (optional)')}
            />
            <label className='flex flex-col gap-2 text-sm'>
              <span>{t('Starts on')}</span>
              <Input
                type='date'
                value={getDateValue(promotionForm.starts_at)}
                onChange={(event) =>
                  setPromotionForm({
                    ...promotionForm,
                    starts_at: dateToTimestamp(event.target.value),
                  })
                }
              />
            </label>
            <label className='flex flex-col gap-2 text-sm'>
              <span>{t('Ends on')}</span>
              <Input
                type='date'
                value={getDateValue(promotionForm.ends_at)}
                onChange={(event) =>
                  setPromotionForm({
                    ...promotionForm,
                    ends_at: dateToTimestamp(event.target.value, true),
                  })
                }
              />
            </label>
          </div>
          <div className='flex flex-wrap items-center gap-3'>
            <label className='flex items-center gap-2 text-sm'>
              <Checkbox
                checked={promotionForm.enabled}
                onCheckedChange={(checked) =>
                  setPromotionForm({
                    ...promotionForm,
                    enabled: checked === true,
                  })
                }
              />
              {t('Enabled')}
            </label>
            <Button
              type='button'
              onClick={handleSavePromotion}
              disabled={savingPromotion}
            >
              {promotionSaveLabel}
            </Button>
            {promotionEditing && (
              <Button
                type='button'
                variant='ghost'
                onClick={() => {
                  setPromotionForm(emptyPromotion)
                  setPromotionEditing(false)
                }}
              >
                {t('Cancel')}
              </Button>
            )}
          </div>
          {promotionsQuery.isLoading && (
            <p className='text-muted-foreground text-sm'>
              {t('Loading promotions...')}
            </p>
          )}
          {!promotionsQuery.isLoading && promotions.length === 0 && (
            <p className='text-muted-foreground text-sm'>
              {t('No promotions configured')}
            </p>
          )}
          {!promotionsQuery.isLoading && promotions.length > 0 && (
            <div className='flex flex-col divide-y rounded-md border'>
              {promotions.map((promotion) => (
                <div
                  key={promotion.id}
                  className='flex flex-col gap-3 p-3 sm:flex-row sm:items-center'
                >
                  <div className='min-w-0 flex-1'>
                    <div className='flex flex-wrap items-center gap-2'>
                      <span className='font-medium'>
                        {promotion.model_name}
                      </span>
                      {promotion.discount && (
                        <Badge variant='secondary'>{promotion.discount}</Badge>
                      )}
                      <Badge
                        variant={promotion.enabled ? 'outline' : 'secondary'}
                      >
                        {promotion.enabled ? t('Enabled') : t('Disabled')}
                      </Badge>
                    </div>
                    <p className='text-muted-foreground mt-1 text-sm'>
                      {[
                        promotion.description,
                        promotion.starts_at
                          ? formatDate(promotion.starts_at)
                          : '',
                        promotion.ends_at ? formatDate(promotion.ends_at) : '',
                      ]
                        .filter(Boolean)
                        .join(' · ') || t('No description')}
                    </p>
                  </div>
                  <div className='flex gap-2'>
                    <Button
                      type='button'
                      size='sm'
                      variant='outline'
                      onClick={() => {
                        setPromotionForm(promotion)
                        setPromotionEditing(true)
                      }}
                    >
                      {t('Edit')}
                    </Button>
                    <Button
                      type='button'
                      size='sm'
                      variant='ghost'
                      onClick={() => handleDeletePromotion(promotion)}
                    >
                      {t('Delete')}
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>

        <section className='flex flex-col gap-4'>
          <div className='flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between'>
            <div>
              <h4 className='text-sm font-semibold'>{t('Review and send')}</h4>
              <p className='text-muted-foreground mt-1 text-sm'>
                {t(
                  'Review the generated model list and offer copy, edit the message, choose an audience, and send when ready.'
                )}
              </p>
            </div>
            <div className='flex items-center gap-2'>
              <Input
                className='w-24'
                type='number'
                min={1}
                max={365}
                value={periodDays}
                onChange={(event) => setPeriodDays(event.target.value)}
                aria-label={t('Lookback days')}
              />
              <Button type='button' variant='outline' onClick={handleGenerate}>
                {t('Generate draft')}
              </Button>
            </div>
          </div>

          <div className='grid gap-4 lg:grid-cols-[minmax(220px,0.8fr)_minmax(0,2fr)]'>
            <div className='flex max-h-[34rem] flex-col overflow-y-auto rounded-md border'>
              {digestsQuery.isLoading && (
                <p className='text-muted-foreground p-4 text-sm'>
                  {t('Loading drafts...')}
                </p>
              )}
              {!digestsQuery.isLoading && digests.length === 0 && (
                <p className='text-muted-foreground p-4 text-sm'>
                  {t('No newsletter drafts yet')}
                </p>
              )}
              {!digestsQuery.isLoading &&
                digests.length > 0 &&
                digests.map((digest) => (
                  <button
                    key={digest.id}
                    type='button'
                    onClick={() => {
                      setSelectedDigestID(digest.id)
                      setRecipientStatus('all')
                    }}
                    className={`hover:bg-muted/50 flex flex-col gap-2 border-b p-3 text-left last:border-b-0 ${selectedDigestID === digest.id ? 'bg-muted/60' : ''}`}
                  >
                    <span className='truncate text-sm font-medium'>
                      {digest.subject}
                    </span>
                    <span className='text-muted-foreground text-xs'>
                      {formatDate(digest.period_start)} –{' '}
                      {formatDate(digest.period_end)}
                    </span>
                    <div className='flex flex-wrap items-center gap-2'>
                      <Badge variant='outline'>
                        {t(digestStatusKey(digest.status))}
                      </Badge>
                      {digest.recipient_count > 0 && (
                        <span className='text-muted-foreground text-xs'>
                          {t('Sent {{sent}} / {{total}}', {
                            sent: digest.sent_count,
                            total: digest.recipient_count,
                          })}
                        </span>
                      )}
                    </div>
                  </button>
                ))}
            </div>

            {selectedDigest ? (
              <div className='flex min-w-0 flex-col gap-4 rounded-md border p-4'>
                <div className='flex flex-wrap items-center justify-between gap-3'>
                  <div>
                    <h5 className='text-sm font-semibold'>
                      {t('Draft #{{id}}', { id: selectedDigest.id })}
                    </h5>
                    <p className='text-muted-foreground mt-1 text-xs'>
                      {formatDate(selectedDigest.period_start)} –{' '}
                      {formatDate(selectedDigest.period_end)}
                    </p>
                  </div>
                  <Badge variant='outline'>
                    {t(digestStatusKey(selectedDigest.status))}
                  </Badge>
                </div>
                <div className='grid gap-3 sm:grid-cols-2'>
                  <label className='flex flex-col gap-2 text-sm'>
                    <span>{t('Subject')}</span>
                    <Input
                      value={subject}
                      disabled={!editable}
                      onChange={(event) => setSubject(event.target.value)}
                    />
                  </label>
                  <label className='flex flex-col gap-2 text-sm'>
                    <span>{t('Audience')}</span>
                    <Select
                      value={audienceType}
                      disabled={!editable}
                      onValueChange={(value) =>
                        setAudienceType(
                          value as MarketingDigest['audience_type']
                        )
                      }
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectGroup>
                          <SelectItem value='all'>
                            {t('All eligible users')}
                          </SelectItem>
                          <SelectItem value='topup'>
                            {t('Users with successful top-ups')}
                          </SelectItem>
                          <SelectItem value='active'>
                            {t('Active users (last 30 days)')}
                          </SelectItem>
                          <SelectItem value='group'>
                            {t('Specific user group')}
                          </SelectItem>
                          <SelectItem value='manual'>
                            {t('Selected users')}
                          </SelectItem>
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                  </label>
                </div>
                {audienceType === 'group' && (
                  <label className='flex flex-col gap-2 text-sm'>
                    <span>{t('User group')}</span>
                    <Input
                      value={audienceGroup}
                      disabled={!editable}
                      onChange={(event) => setAudienceGroup(event.target.value)}
                      placeholder={t('Enter an exact group name')}
                    />
                  </label>
                )}
                {audienceType === 'manual' && (
                  <div className='flex flex-col gap-3 rounded-md border p-3'>
                    <div className='flex flex-col gap-2 sm:flex-row'>
                      <Input
                        value={searchKeyword}
                        disabled={!editable}
                        onChange={(event) =>
                          setSearchKeyword(event.target.value)
                        }
                        onKeyDown={(event) => {
                          if (event.key === 'Enter') {
                            event.preventDefault()
                            void handleSearchUsers()
                          }
                        }}
                        placeholder={t('Search username, email, or user ID')}
                      />
                      <Button
                        type='button'
                        variant='outline'
                        disabled={!editable || searchingUsers}
                        onClick={handleSearchUsers}
                      >
                        {searchingUsers ? t('Searching...') : t('Search users')}
                      </Button>
                    </div>
                    {selectedUserIDs.length > 0 && (
                      <p className='text-muted-foreground text-xs'>
                        {t('{{count}} users selected', {
                          count: selectedUserIDs.length,
                        })}
                        : {selectedUserIDs.map((id) => `#${id}`).join(', ')}
                      </p>
                    )}
                    <div className='max-h-40 overflow-y-auto rounded-md border'>
                      {searchResults.length === 0 ? (
                        <p className='text-muted-foreground p-3 text-sm'>
                          {t('Search users to add recipients')}
                        </p>
                      ) : (
                        searchResults.map((user) => (
                          <label
                            key={user.id}
                            className='flex items-center gap-3 border-b p-2 text-sm last:border-b-0'
                          >
                            <Checkbox
                              checked={selectedUserIDs.includes(user.id)}
                              disabled={!editable}
                              onCheckedChange={() => toggleUser(user.id)}
                            />
                            <span className='min-w-0 flex-1 truncate'>
                              {user.username || `#${user.id}`} ·{' '}
                              {user.email || t('No email')}
                            </span>
                          </label>
                        ))
                      )}
                    </div>
                  </div>
                )}
                <label className='flex flex-col gap-2 text-sm'>
                  <span>{t('Email content (HTML)')}</span>
                  <Textarea
                    rows={12}
                    className='font-mono text-xs'
                    value={content}
                    disabled={!editable}
                    onChange={(event) => setContent(event.target.value)}
                  />
                </label>
                <div className='flex flex-col gap-2'>
                  <span className='text-sm font-medium'>{t('Preview')}</span>
                  <iframe
                    title={t('Newsletter preview')}
                    sandbox=''
                    srcDoc={`<!doctype html><html><head><meta charset="utf-8"></head><body style="font-family:system-ui,sans-serif;color:#1f2937;padding:16px">${previewContent}</body></html>`}
                    className='h-64 w-full rounded-md border bg-white'
                  />
                </div>
                <div className='flex flex-wrap items-center gap-2'>
                  {editable && (
                    <>
                      <Button
                        type='button'
                        variant='outline'
                        disabled={savingDraft || sending}
                        onClick={handleSaveDraft}
                      >
                        {savingDraft ? t('Saving...') : t('Save draft')}
                      </Button>
                      <Button
                        type='button'
                        disabled={sending}
                        onClick={() => handleSend(false)}
                      >
                        {sending ? t('Starting...') : t('Approve and send')}
                      </Button>
                    </>
                  )}
                  {(selectedDigest.status === 'sending' ||
                    selectedDigest.status === 'sent_with_errors' ||
                    selectedDigest.status === 'failed') && (
                    <Button
                      type='button'
                      variant='outline'
                      disabled={sending}
                      onClick={() => handleSend(true)}
                    >
                      {retryButtonLabel}
                    </Button>
                  )}
                </div>

                {selectedDigest.recipient_count > 0 && (
                  <div className='flex flex-col gap-3 border-t pt-4'>
                    <div className='flex flex-wrap items-center justify-between gap-2'>
                      <h5 className='text-sm font-semibold'>
                        {t('Delivery results')}
                      </h5>
                      <Select
                        value={recipientStatus}
                        onValueChange={(value) =>
                          setRecipientStatus(value as typeof recipientStatus)
                        }
                      >
                        <SelectTrigger className='w-40'>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectGroup>
                            <SelectItem value='all'>
                              {t('All recipients')}
                            </SelectItem>
                            <SelectItem value='sent'>{t('Sent')}</SelectItem>
                            <SelectItem value='skipped'>
                              {t('Skipped')}
                            </SelectItem>
                            <SelectItem value='failed'>
                              {t('Failed')}
                            </SelectItem>
                          </SelectGroup>
                        </SelectContent>
                      </Select>
                    </div>
                    <div className='text-muted-foreground flex flex-wrap gap-x-4 gap-y-1 text-xs'>
                      <span>
                        {t('Recipients')}: {selectedDigest.recipient_count}
                      </span>
                      <span>
                        {t('Sent')}: {selectedDigest.sent_count}
                      </span>
                      <span>
                        {t('Skipped')}: {selectedDigest.skipped_count}
                      </span>
                      <span>
                        {t('Failed')}: {selectedDigest.failed_count}
                      </span>
                    </div>
                    <div className='max-h-52 overflow-auto rounded-md border'>
                      {recipientQuery.isLoading && (
                        <p className='text-muted-foreground p-3 text-sm'>
                          {t('Loading recipients...')}
                        </p>
                      )}
                      {!recipientQuery.isLoading &&
                        currentRecipients.length === 0 && (
                          <p className='text-muted-foreground p-3 text-sm'>
                            {t('No recipient records')}
                          </p>
                        )}
                      {!recipientQuery.isLoading &&
                        currentRecipients.length > 0 &&
                        currentRecipients.map((recipient) => (
                          <div
                            key={recipient.id}
                            className='flex flex-wrap items-center gap-2 border-b p-2 text-sm last:border-b-0'
                          >
                            <span className='font-mono text-xs'>
                              #{recipient.user_id}
                            </span>
                            <span className='text-muted-foreground min-w-0 flex-1 truncate'>
                              {recipient.email_masked || t('No email')}
                            </span>
                            <Badge
                              variant={
                                recipient.status === 'failed'
                                  ? 'destructive'
                                  : 'outline'
                              }
                            >
                              {t(recipient.status)}
                            </Badge>
                            {recipient.reason && (
                              <span className='text-muted-foreground basis-full text-xs'>
                                {recipient.reason}
                              </span>
                            )}
                          </div>
                        ))}
                    </div>
                  </div>
                )}
              </div>
            ) : (
              <div className='text-muted-foreground flex min-h-64 items-center justify-center rounded-md border p-6 text-sm'>
                {t('Generate a draft or select a newsletter to review')}
              </div>
            )}
          </div>
        </section>
      </div>
    </SettingsSection>
  )
}
