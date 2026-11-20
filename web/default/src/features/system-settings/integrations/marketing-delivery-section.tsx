import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Loader2, MailCheck } from 'lucide-react'
import { useId, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Separator } from '@/components/ui/separator'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'

import {
  getMarketingDeliverySettings,
  saveMarketingDeliverySettings,
  testMarketingDelivery,
  type MarketingDeliverySettings,
} from '../api'

const queryKey = ['marketing-delivery-settings']

const senderFields = [
  {
    key: 'server',
    label: 'SMTP server',
    placeholder: 'gz-smtp.qcloudmail.com',
  },
  { key: 'account', label: 'SMTP account', placeholder: 'noreply@example.com' },
  { key: 'from', label: 'Marketing sender address', type: 'email' },
  { key: 'display_name', label: 'Sender name' },
  { key: 'reply_to', label: 'Reply-To address', type: 'email' },
  { key: 'password', label: 'SMTP password', type: 'password' },
] as const

export function MarketingDeliverySection() {
  const { t } = useTranslation()
  const id = useId()
  const client = useQueryClient()
  const query = useQuery({
    queryKey,
    queryFn: getMarketingDeliverySettings,
    staleTime: 60_000,
  })
  const saved = query.data?.data
  const [edits, setEdits] = useState<Partial<MarketingDeliverySettings>>({})
  const draft = saved && { ...saved.settings, ...edits }
  const [recipient, setRecipient] = useState('')
  const [action, setAction] = useState<'save' | 'test'>()

  const update = <K extends keyof MarketingDeliverySettings>(
    key: K,
    value: MarketingDeliverySettings[K]
  ) => setEdits((current) => ({ ...current, [key]: value }))

  const save = async () => {
    if (!draft) return
    setAction('save')
    try {
      const response = await saveMarketingDeliverySettings(draft)
      if (!response.success || !response.data) throw new Error(response.message)
      client.setQueryData(queryKey, response)
      setEdits({})
      toast.success(t('Marketing delivery settings saved'))
    } catch (error) {
      toast.error(
        (error as Error).message ||
          t('Failed to save marketing delivery settings')
      )
    } finally {
      setAction(undefined)
    }
  }

  const test = async () => {
    setAction('test')
    try {
      const response = await testMarketingDelivery(recipient.trim())
      if (!response.success) throw new Error(response.message)
      toast.success(t('Test email submitted'))
    } catch (error) {
      toast.error((error as Error).message || t('Failed to send test email'))
    } finally {
      setAction(undefined)
    }
  }

  const dirty = JSON.stringify(draft) !== JSON.stringify(saved?.settings)
  const configuredPassword =
    saved?.password_configured &&
    draft?.server === saved.settings.server &&
    draft?.port === saved.settings.port &&
    draft?.account === saved.settings.account

  return (
    <section
      className='mb-6 flex flex-col gap-4'
      aria-labelledby={`${id}-title`}
    >
      <h3 id={`${id}-title`} className='text-sm font-semibold'>
        {t('Marketing email delivery')}
      </h3>
      {!draft && query.isPending && (
        <Loader2
          className='size-4 animate-spin'
          role='status'
          aria-label={t('Loading email settings...')}
        />
      )}
      {!draft && !query.isPending && (
        <div className='flex flex-wrap items-center gap-2'>
          <p className='text-muted-foreground text-sm'>
            {t('Failed to load marketing delivery settings')}
          </p>
          <Button
            type='button'
            variant='outline'
            size='sm'
            onClick={() => void query.refetch()}
          >
            {t('Retry')}
          </Button>
        </div>
      )}
      {draft && (
        <FieldGroup className='gap-4'>
          <ToggleGroup
            aria-label={t('Marketing email delivery')}
            variant='outline'
            size='sm'
            value={[draft.mode]}
            onValueChange={(values) => {
              const mode = values[0]
              if (mode === 'inherit' || mode === 'smtp') update('mode', mode)
            }}
          >
            <ToggleGroupItem value='inherit'>
              {t('Use system delivery')}
            </ToggleGroupItem>
            <ToggleGroupItem value='smtp'>
              {t('Independent SMTP')}
            </ToggleGroupItem>
          </ToggleGroup>
          {draft.mode === 'smtp' && (
            <FieldGroup className='grid gap-4 sm:grid-cols-2'>
              {senderFields.map((field) => {
                let placeholder: string | undefined =
                  'placeholder' in field ? field.placeholder : undefined
                if (field.key === 'password' && configuredPassword) {
                  placeholder = t('Configured · leave blank to keep')
                }
                return (
                  <Field key={field.key}>
                    <FieldLabel htmlFor={`${id}-${field.key}`}>
                      {t(field.label)}
                    </FieldLabel>
                    <Input
                      id={`${id}-${field.key}`}
                      type={'type' in field ? field.type : 'text'}
                      autoComplete={
                        field.key === 'password' ? 'new-password' : 'off'
                      }
                      placeholder={placeholder}
                      value={draft[field.key] ?? ''}
                      onChange={(event) =>
                        update(field.key, event.target.value)
                      }
                    />
                  </Field>
                )
              })}
              <Field>
                <FieldLabel htmlFor={`${id}-port`}>
                  {t('SMTP port (SSL/TLS)')}
                </FieldLabel>
                <Input
                  id={`${id}-port`}
                  type='number'
                  min={1}
                  max={65535}
                  value={draft.port || ''}
                  onChange={(event) =>
                    update('port', Number(event.target.value))
                  }
                />
              </Field>
              <Field>
                <FieldLabel htmlFor={`${id}-callback`}>
                  {t('Tencent Cloud SES callback URL')}
                </FieldLabel>
                <Input
                  id={`${id}-callback`}
                  value={saved?.callback_url ?? ''}
                  readOnly
                />
              </Field>
            </FieldGroup>
          )}
          <Field className='max-w-xs'>
            <FieldLabel htmlFor={`${id}-rate`}>
              {t('Emails per second')}
            </FieldLabel>
            <Input
              id={`${id}-rate`}
              type='number'
              min={1}
              max={20}
              value={draft.per_second || ''}
              onChange={(event) =>
                update('per_second', Number(event.target.value))
              }
            />
          </Field>
          <div className='flex flex-wrap items-end gap-2'>
            <Button
              type='button'
              disabled={!dirty || !!action}
              onClick={() => void save()}
            >
              {action === 'save' && <Loader2 className='size-4 animate-spin' />}
              {t('Save marketing delivery')}
            </Button>
            <Field className='w-full max-w-xs'>
              <FieldLabel htmlFor={`${id}-recipient`}>
                {t('Test recipient address')}
              </FieldLabel>
              <Input
                id={`${id}-recipient`}
                type='email'
                value={recipient}
                onChange={(event) => setRecipient(event.target.value)}
              />
            </Field>
            <Button
              type='button'
              variant='outline'
              disabled={dirty || !!action || !recipient.trim()}
              onClick={() => void test()}
            >
              {action === 'test' ? (
                <Loader2 className='size-4 animate-spin' />
              ) : (
                <MailCheck className='size-4' />
              )}
              {t('Send test email')}
            </Button>
          </div>
        </FieldGroup>
      )}
      <Separator />
    </section>
  )
}
