import { Check, Copy, Loader2 } from 'lucide-react'
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
import { useState, useCallback } from 'react'
import { useTranslation } from 'react-i18next'

import { BadgeCell } from '@/components/data-table'
import { StatusBadge } from '@/components/status-badge'
import { Button } from '@/components/ui/button'
import { Progress } from '@/components/ui/progress'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { copyToClipboard } from '@/lib/copy-to-clipboard'
import {
  formatCompactNumber,
  formatQuota,
  formatTimestampToDate,
} from '@/lib/format'
import { cn } from '@/lib/utils'

import {
  API_KEY_STATUS,
  API_KEY_STATUSES,
  SUBSCRIPTION_KEY_STATUSES,
} from '../constants'
import { type ApiKey } from '../types'
import { useApiKeys } from './api-keys-provider'

export function ApiKeyStatusCell({ apiKey }: { apiKey: ApiKey }) {
  const { t } = useTranslation()
  const subscriptionStatus =
    apiKey.key_type === 'subscription' && apiKey.status === API_KEY_STATUS.ENABLED
      ? SUBSCRIPTION_KEY_STATUSES[apiKey.effective_status]
      : undefined
  if (subscriptionStatus) {
    return (
      <StatusBadge
        label={t(subscriptionStatus.label)}
        variant={subscriptionStatus.variant}
        copyable={false}
        className='-ml-1.5'
      />
    )
  }
  const statusConfig = API_KEY_STATUSES[apiKey.status]
  if (!statusConfig) return null
  return (
    <StatusBadge
      label={t(statusConfig.label)}
      variant={statusConfig.variant}
      copyable={false}
      className='-ml-1.5'
    />
  )
}

export function ApiKeyTypeCell({ apiKey }: { apiKey: ApiKey }) {
  const { t } = useTranslation()
  if (apiKey.key_type === 'subscription') {
    return (
      <div className='min-w-0 space-y-0.5'>
        <StatusBadge
          label={t('Subscription key')}
          variant='info'
          copyable={false}
          className='-ml-1.5'
        />
        <div className='text-muted-foreground max-w-[180px] truncate text-xs'>
          {apiKey.subscription?.plan_title ||
            (apiKey.subscription
              ? `${t('Plan')} #${apiKey.subscription.plan_id}`
              : t('Subscription unavailable'))}
        </div>
      </div>
    )
  }
  const label =
    apiKey.key_type === 'campaign' ? t('Campaign key') : t('General key')
  return (
    <StatusBadge
      label={label}
      variant='neutral'
      copyable={false}
      className='-ml-1.5'
    />
  )
}

function formatSubscriptionAmount(
  value: number,
  billingMode: string,
  tokenLabel: string
): string {
  return billingMode === 'token'
    ? `${formatCompactNumber(value)} ${tokenLabel}`
    : formatQuota(value)
}

export function ApiKeyQuotaCell({ apiKey }: { apiKey: ApiKey }) {
  const { t } = useTranslation()
  const subscription = apiKey.subscription
  if (apiKey.key_type !== 'subscription' || !subscription) {
    if (apiKey.unlimited_quota) {
      return <UnlimitedQuotaBadge used={apiKey.used_quota} />
    }
    return null
  }

  const total = subscription.amount_total
  const used = Math.max(0, subscription.amount_used)
  const remaining = Math.max(0, subscription.amount_remaining)
  const isUnlimited = total <= 0
  const percentage = !isUnlimited && total > 0 ? (remaining / total) * 100 : 0
  const formatAmount = (value: number) =>
    formatSubscriptionAmount(value, subscription.billing_mode, t('tokens'))

  if (isUnlimited) {
    return (
      <Popover>
        <PopoverTrigger
          render={
            <button
              type='button'
              className='focus-visible:ring-ring/50 -ml-1.5 cursor-help rounded-4xl focus-visible:ring-[3px] focus-visible:outline-none'
              aria-label={`${t('Unlimited')}; ${t('Used:')} ${formatAmount(used)}`}
            />
          }
        >
          <StatusBadge
            label={t('Unlimited')}
            variant='neutral'
            copyable={false}
          />
        </PopoverTrigger>
        <PopoverContent className='w-auto p-2' side='top'>
          <div className='space-y-1 text-xs'>
            <div>{t('Shared subscription quota')}</div>
            <div>{t('Used:')} {formatAmount(used)}</div>
          </div>
        </PopoverContent>
      </Popover>
    )
  }

  return (
    <Tooltip>
      <TooltipTrigger render={<div className='w-[150px] space-y-1' />}>
        <div className='flex justify-between text-xs'>
          <span className='font-medium tabular-nums'>{formatAmount(remaining)}</span>
          <span className='text-muted-foreground tabular-nums'>{formatAmount(total)}</span>
        </div>
        <Progress
          value={percentage}
          className={cn(
            'h-1.5',
            percentage <= 10
              ? '[&_[data-slot=progress-indicator]]:bg-rose-500'
              : percentage <= 30
                ? '[&_[data-slot=progress-indicator]]:bg-amber-500'
                : '[&_[data-slot=progress-indicator]]:bg-emerald-500'
          )}
        />
      </TooltipTrigger>
      <TooltipContent>
        <div className='space-y-1 text-xs'>
          <div>{t('Shared subscription quota')}</div>
          <div>{t('Used:')} {formatAmount(used)}</div>
          <div>{t('Remaining:')} {formatAmount(remaining)} ({percentage.toFixed(1)}%)</div>
          <div>{t('Total:')} {formatAmount(total)}</div>
          {subscription.next_reset_time > 0 && (
            <div>{t('Next reset:')} {formatTimestampToDate(subscription.next_reset_time)}</div>
          )}
        </div>
      </TooltipContent>
    </Tooltip>
  )
}

export function ApiKeyCell({ apiKey }: { apiKey: ApiKey }) {
  const { t } = useTranslation()
  const {
    resolveRealKey,
    resolvedKeys,
    loadingKeys,
    copiedKeyId,
    markKeyCopied,
  } = useApiKeys()
  const [popoverOpen, setPopoverOpen] = useState(false)

  const isLoading = !!loadingKeys[apiKey.id]
  const resolvedFullKey = resolvedKeys[apiKey.id]
  const isCopied = copiedKeyId === apiKey.id
  const maskedKey = `sk-${apiKey.key}`

  const handlePopoverOpen = useCallback(
    (open: boolean) => {
      setPopoverOpen(open)
      if (open && !resolvedFullKey) {
        resolveRealKey(apiKey.id)
      }
    },
    [resolvedFullKey, resolveRealKey, apiKey.id]
  )

  const handleCopy = useCallback(async () => {
    const realKey = resolvedFullKey || (await resolveRealKey(apiKey.id))
    if (realKey) {
      const ok = await copyToClipboard(realKey)
      if (ok) markKeyCopied(apiKey.id)
    }
  }, [resolvedFullKey, resolveRealKey, apiKey.id, markKeyCopied])

  return (
    <div className='flex max-w-full min-w-0 items-center'>
      <Popover open={popoverOpen} onOpenChange={handlePopoverOpen}>
        <PopoverTrigger
          render={
            <Button
              variant='ghost'
              size='sm'
              className='text-muted-foreground h-7 max-w-full min-w-0 justify-start truncate px-0 font-mono text-xs hover:bg-transparent aria-expanded:bg-transparent'
            />
          }
        >
          <span className='truncate'>{maskedKey}</span>
        </PopoverTrigger>
        <PopoverContent
          className='w-auto max-w-[min(90vw,28rem)]'
          align='start'
        >
          <div className='space-y-2'>
            <p className='text-muted-foreground text-xs'>{t('Full API Key')}</p>
            {isLoading ? (
              <div className='flex items-center gap-2 py-2'>
                <Loader2 className='size-3.5 animate-spin' />
                <span className='text-muted-foreground text-xs'>
                  {t('Loading...')}
                </span>
              </div>
            ) : (
              <input
                readOnly
                value={resolvedFullKey || maskedKey}
                autoFocus
                onFocus={(e) => e.target.select()}
                className='bg-muted/50 w-full min-w-[280px] rounded-md border px-3 py-2 font-mono text-xs outline-none'
              />
            )}
          </div>
        </PopoverContent>
      </Popover>
      <Tooltip>
        <TooltipTrigger
          render={
            <Button
              variant='ghost'
              size='icon'
              className='size-7 shrink-0'
              onClick={handleCopy}
              disabled={isLoading}
            />
          }
        >
          {isLoading ? (
            <Loader2 className='size-3.5 animate-spin' />
          ) : isCopied ? (
            <Check className='size-3.5 text-green-600' />
          ) : (
            <Copy className='size-3.5' />
          )}
        </TooltipTrigger>
        <TooltipContent>
          {isLoading
            ? t('Loading...')
            : isCopied
              ? t('Copied!')
              : t('Copy API key')}
        </TooltipContent>
      </Tooltip>
    </div>
  )
}

type UnlimitedQuotaBadgeProps = {
  used: number
}

export function UnlimitedQuotaBadge(props: UnlimitedQuotaBadgeProps) {
  const { t } = useTranslation()
  const formattedUsed = formatQuota(props.used)

  return (
    <Popover>
      <PopoverTrigger
        render={
          <button
            type='button'
            className='focus-visible:ring-ring/50 -ml-1.5 cursor-help rounded-4xl focus-visible:ring-[3px] focus-visible:outline-none'
            aria-label={`${t('Unlimited')}; ${t('Used:')} ${formattedUsed}`}
          />
        }
      >
        <StatusBadge
          label={t('Unlimited')}
          variant='neutral'
          copyable={false}
        />
      </PopoverTrigger>
      <PopoverContent className='w-auto p-2' side='top'>
        <span className='text-xs'>
          {t('Used:')} {formattedUsed}
        </span>
      </PopoverContent>
    </Popover>
  )
}

export function ModelLimitsCell({ apiKey }: { apiKey: ApiKey }) {
  const { t } = useTranslation()

  if (!apiKey.model_limits_enabled || !apiKey.model_limits) {
    return (
      <StatusBadge
        label={t('Unlimited')}
        variant='neutral'
        copyable={false}
        className='-ml-1.5'
      />
    )
  }

  const models = apiKey.model_limits.split(',').filter(Boolean)

  return (
    <Tooltip>
      <TooltipTrigger render={<BadgeCell />}>
        <StatusBadge
          label={t('{{count}} model(s)', { count: models.length })}
          variant='neutral'
          copyable={false}
        />
      </TooltipTrigger>
      <TooltipContent side='top' className='max-w-xs'>
        <div className='max-h-[200px] space-y-0.5 overflow-y-auto text-xs'>
          {models.map((m) => (
            <div key={m} className='font-mono'>
              {m}
            </div>
          ))}
        </div>
      </TooltipContent>
    </Tooltip>
  )
}

export function IpRestrictionsCell({ apiKey }: { apiKey: ApiKey }) {
  const { t } = useTranslation()
  const allowIps = apiKey.allow_ips?.trim()

  if (!allowIps) {
    return (
      <StatusBadge
        label={t('No restriction')}
        variant='neutral'
        copyable={false}
        className='-ml-1.5'
      />
    )
  }

  const ips = allowIps
    .split('\n')
    .map((ip) => ip.trim())
    .filter(Boolean)

  return (
    <Tooltip>
      <TooltipTrigger render={<BadgeCell />}>
        <StatusBadge
          label={t('{{count}} IP(s)', { count: ips.length })}
          variant='neutral'
          copyable={false}
        />
      </TooltipTrigger>
      <TooltipContent side='top' className='max-w-xs'>
        <div className='max-h-[200px] space-y-0.5 overflow-y-auto text-xs'>
          {ips.map((ip) => (
            <div key={ip} className='font-mono'>
              {ip}
            </div>
          ))}
        </div>
      </TooltipContent>
    </Tooltip>
  )
}
