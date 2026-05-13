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
import { useMemo, useState } from 'react'
import type { ReactNode } from 'react'
import {
  Building2,
  Check,
  Copy,
  Landmark,
  Mail,
  Phone,
  UserRound,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'
import { TitledCard } from '@/components/ui/titled-card'
import type { CorporateTransferInfo } from '../types'

type CorporateTransferCardProps = {
  config?: CorporateTransferInfo
  username?: string
}

type CopyableRowProps = {
  label: string
  value?: string
  icon: ReactNode
}

function maskPhone(value?: string) {
  const digits = value?.replace(/\D/g, '') ?? ''
  if (digits.length < 7) return value || ''
  return `${digits.slice(0, 3)}****${digits.slice(-4)}`
}

function CopyableRow({ label, value, icon }: CopyableRowProps) {
  const { t } = useTranslation()
  const [copied, setCopied] = useState(false)
  const displayValue = value?.trim() || t('Not configured')
  const canCopy = Boolean(value?.trim())

  const handleCopy = async () => {
    if (!canCopy) return
    await navigator.clipboard.writeText(value!.trim())
    setCopied(true)
    window.setTimeout(() => setCopied(false), 1200)
  }

  return (
    <div className='grid gap-1.5 rounded-md border bg-background px-3 py-2.5 sm:grid-cols-[8rem_minmax(0,1fr)_auto] sm:items-center'>
      <div className='text-muted-foreground flex items-center gap-2 text-xs font-medium'>
        {icon}
        {label}
      </div>
      <div
        className={cn(
          'min-w-0 break-all text-sm font-medium',
          !canCopy && 'text-muted-foreground font-normal'
        )}
      >
        {displayValue}
      </div>
      <Button
        type='button'
        variant='ghost'
        size='icon'
        className='h-7 w-7 justify-self-end'
        disabled={!canCopy}
        onClick={handleCopy}
        title={copied ? t('Copied') : t('Copy')}
      >
        {copied ? (
          <Check className='h-3.5 w-3.5' />
        ) : (
          <Copy className='h-3.5 w-3.5' />
        )}
        <span className='sr-only'>{copied ? t('Copied') : t('Copy')}</span>
      </Button>
    </div>
  )
}

export function CorporateTransferCard({
  config,
  username,
}: CorporateTransferCardProps) {
  const { t } = useTranslation()
  const [acknowledged, setAcknowledged] = useState(false)

  const enabled = Boolean(config?.enabled)
  const title = config?.title?.trim() || t('Corporate Bank Transfer')
  const maskedPhone = maskPhone(config?.operator_phone)

  const noticeItems = useMemo(() => {
    const configured = (config?.notice ?? '')
      .split(/\r?\n/)
      .map((line) => line.trim())
      .filter(Boolean)

    if (configured.length > 0) return configured

    const fallback = [
      t('Please include your account username in the transfer note: {{username}}', {
        username: username || '-',
      }),
      t('Use a bank account with the same verified entity name.'),
      t('After transfer, contact support with the payment voucher.'),
    ]

    if (config?.support_email) {
      fallback.push(
        t('If you have questions, contact {{email}}.', {
          email: config.support_email,
        })
      )
    }

    return fallback
  }, [config?.notice, config?.support_email, t, username])

  if (!enabled) return null

  return (
    <TitledCard
      title={title}
      description={t('Corporate remittance instructions')}
      icon={<Building2 className='h-4 w-4' />}
      contentClassName='space-y-4'
    >
      <div className='rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-950 dark:border-amber-900/70 dark:bg-amber-950/30 dark:text-amber-100'>
        <div className='mb-2 font-medium'>{t('Notice')}</div>
        <ul className='list-disc space-y-1 pl-5'>
          {noticeItems.map((item, index) => (
            <li key={`${item}-${index}`}>{item}</li>
          ))}
        </ul>
        {(username || maskedPhone) && (
          <div className='mt-3 flex flex-wrap gap-2 text-xs'>
            {username && (
              <span className='rounded-md border border-amber-300/70 bg-white/70 px-2 py-1 font-medium text-amber-900 dark:border-amber-700 dark:bg-amber-950/40 dark:text-amber-100'>
                {t('Transfer note')}: {t('Your account')} {username}
              </span>
            )}
            {maskedPhone && (
              <span className='rounded-md border border-amber-300/70 bg-white/70 px-2 py-1 font-medium text-amber-900 dark:border-amber-700 dark:bg-amber-950/40 dark:text-amber-100'>
                {t('Contact phone')}: {maskedPhone}
              </span>
            )}
          </div>
        )}
      </div>

      <div className='flex items-start gap-2'>
        <Checkbox
          id='corporate-transfer-ack'
          checked={acknowledged}
          onCheckedChange={(checked) => setAcknowledged(checked === true)}
          className='mt-0.5'
        />
        <Label
          htmlFor='corporate-transfer-ack'
          className='text-muted-foreground text-sm leading-5'
        >
          {t('I have read and understood the notes above.')}
        </Label>
      </div>

      <div className='space-y-2'>
        <div className='text-sm font-semibold'>
          {t('Receiving account information')}
        </div>
        <div className='relative overflow-hidden rounded-lg border border-dashed p-3'>
          <div
            className={cn(
              'space-y-2 transition duration-200',
              !acknowledged &&
                'pointer-events-none select-none blur-[5px] opacity-50'
            )}
            aria-hidden={!acknowledged}
          >
            <CopyableRow
              label={t('Account name')}
              value={config?.account_name}
              icon={<UserRound className='h-3.5 w-3.5' />}
            />
            <CopyableRow
              label={t('Opening bank')}
              value={config?.bank_name}
              icon={<Landmark className='h-3.5 w-3.5' />}
            />
            <CopyableRow
              label={t('Bank account')}
              value={config?.bank_account}
              icon={<Building2 className='h-3.5 w-3.5' />}
            />
            <CopyableRow
              label={t('Contact phone')}
              value={config?.operator_phone}
              icon={<Phone className='h-3.5 w-3.5' />}
            />
            <CopyableRow
              label={t('Support email')}
              value={config?.support_email}
              icon={<Mail className='h-3.5 w-3.5' />}
            />
          </div>
          {!acknowledged && (
            <div className='absolute inset-0 flex items-center justify-center bg-background/45 px-4 text-center text-sm font-medium backdrop-blur-[1px]'>
              {t(
                'Check the acknowledgement above to reveal the receiving account.'
              )}
            </div>
          )}
        </div>
      </div>
    </TitledCard>
  )
}
