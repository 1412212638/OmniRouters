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
import { useState, useEffect, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'

import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { formatLocalCurrencyAmount } from '@/lib/currency'
import { cn } from '@/lib/utils'

import {
  getPaymentIcon,
  getMinTopupAmount,
  normalizeTopupFeeRate,
} from '../lib'
import type {
  PaymentMethod,
  PresetAmount,
  TopupInfo,
  CreemProduct,
  WaffoPayMethod,
  CorporateTransferInfo,
} from '../types'
import { CorporateTransferCard } from './corporate-transfer-card'
import { CreemProductsSection } from './creem-products-section'
import { RechargeCarousel } from './recharge-carousel'

interface RechargeFormCardProps {
  paymentType?: string
  topupInfo: TopupInfo | null
  presetAmounts: PresetAmount[]
  selectedPreset: number | null
  onSelectPreset: (preset: PresetAmount) => void
  topupAmount: number
  onTopupAmountChange: (amount: number) => void
  onPaymentMethodSelect: (method: PaymentMethod) => void
  paymentLoading: string | null
  loading?: boolean
  usdExchangeRate?: number
  corporateTransfer?: CorporateTransferInfo
  creemProducts?: CreemProduct[]
  enableCreemTopup?: boolean
  onCreemProductSelect?: (product: CreemProduct) => void
  enableWaffoTopup?: boolean
  waffoPayMethods?: WaffoPayMethod[]
  waffoMinTopup?: number
  onWaffoMethodSelect?: (method: WaffoPayMethod, index: number) => void
  enableWaffoPancakeTopup?: boolean
}

export function RechargeFormCard({
  paymentType,
  topupInfo,
  presetAmounts,
  selectedPreset,
  onSelectPreset,
  topupAmount,
  onTopupAmountChange,
  onPaymentMethodSelect,
  paymentLoading,
  loading,
  usdExchangeRate = 1,
  corporateTransfer,
  creemProducts,
  enableCreemTopup,
  onCreemProductSelect,
  enableWaffoTopup,
  waffoPayMethods,
  waffoMinTopup,
  onWaffoMethodSelect,
  enableWaffoPancakeTopup,
}: RechargeFormCardProps) {
  const { t } = useTranslation()
  const [localAmount, setLocalAmount] = useState(topupAmount.toString())

  useEffect(() => {
    setLocalAmount((previousAmount) =>
      previousAmount === '' && topupAmount === 0
        ? previousAmount
        : topupAmount.toString()
    )
  }, [topupAmount])

  const handleAmountChange = (value: string) => {
    setLocalAmount(value)
    const numValue = Number.parseInt(value, 10) || 0
    if (numValue >= 0) {
      onTopupAmountChange(numValue)
    }
  }

  const hasConfigurableTopup =
    topupInfo?.enable_online_topup ||
    topupInfo?.enable_stripe_topup ||
    enableWaffoTopup ||
    enableWaffoPancakeTopup
  const hasAnyTopup = hasConfigurableTopup || enableCreemTopup
  const minTopup = getMinTopupAmount(topupInfo)
  const standardMethods = topupInfo?.pay_methods ?? []
  const waffoMethods = enableWaffoTopup ? (waffoPayMethods ?? []) : []
  const rowClass =
    'grid gap-2 sm:grid-cols-[8rem_minmax(0,1fr)] sm:items-start sm:gap-4'
  const labelClass = 'text-foreground text-sm font-medium sm:pt-2.5'

  const onlineContent = loading ? (
    <div aria-busy='true' className='flex max-w-2xl flex-col gap-4'>
      <Skeleton className='h-10 w-full' />
      <Skeleton className='h-10 w-3/4' />
      <Skeleton className='h-10 w-32' />
    </div>
  ) : (
    <div className='flex flex-col gap-5'>
      <RechargeCarousel items={topupInfo?.recharge_carousel} />
      {hasConfigurableTopup && (
        <>
          {presetAmounts.length > 0 && (
            <div className={rowClass}>
              <span id='wallet-preset-label' className={labelClass}>
                {t('Topup Amount')}
              </span>
              <div
                role='group'
                aria-labelledby='wallet-preset-label'
                className='flex flex-wrap gap-2'
              >
                {presetAmounts.map((preset) => {
                  const selected = preset.value === selectedPreset
                  return (
                    <Button
                      key={preset.value}
                      type='button'
                      variant='outline'
                      aria-pressed={selected}
                      disabled={Boolean(paymentLoading)}
                      onClick={() => onSelectPreset(preset)}
                      className={cn(
                        'h-auto min-h-10 min-w-20 px-4 py-2 tabular-nums',
                        selected && 'border-primary bg-primary/5 text-primary'
                      )}
                    >
                      {formatLocalCurrencyAmount(preset.value * usdExchangeRate, {
                        abbreviate: false,
                        digitsLarge: 0,
                        digitsSmall: 2,
                      })}
                    </Button>
                  )
                })}
              </div>
            </div>
          )}

          <div className={rowClass}>
            <Label htmlFor='topup-amount' className={labelClass}>
              {t('Custom Amount')}
            </Label>
            <Input
              id='topup-amount'
              type='number'
              min={minTopup}
              step={1}
              disabled={Boolean(paymentLoading)}
              value={localAmount}
              onChange={(event) => handleAmountChange(event.target.value)}
              placeholder={t('Minimum topup amount: {{amount}}', {
                amount: minTopup,
              })}
              className='h-10 w-full max-w-lg tabular-nums'
            />
          </div>

          <div className={rowClass}>
            <span id='wallet-payment-label' className={labelClass}>
              {t('Payment Methods')}
            </span>
            <div
              role='group'
              aria-labelledby='wallet-payment-label'
              className='flex min-w-0 flex-wrap gap-2'
            >
              {standardMethods.map((method) => {
                const minimum = Math.max(minTopup, method.min_topup || 0)
                const fee = normalizeTopupFeeRate(
                  method.fee_rate ?? topupInfo?.fee_rate
                )
                return (
                  <PaymentChoice
                    key={method.type}
                    name={method.name}
                    icon={getPaymentIcon(
                      method.type,
                      'size-5 shrink-0',
                      method.icon
                    )}
                    selected={paymentType === method.type}
                    minimum={topupAmount < minimum ? minimum : undefined}
                    fee={fee}
                    busy={Boolean(paymentLoading)}
                    onSelect={() => onPaymentMethodSelect(method)}
                  />
                )
              })}
              {waffoMethods.map((method, index) => (
                <PaymentChoice
                  key={`${method.payMethodType ?? ''}-${method.payMethodName ?? method.name}`}
                  name={method.name}
                  icon={getPaymentIcon('waffo', 'size-5 shrink-0', method.icon)}
                  selected={paymentType === `waffo-${index}`}
                  minimum={
                    topupAmount < (waffoMinTopup ?? 0)
                      ? waffoMinTopup
                      : undefined
                  }
                  busy={Boolean(paymentLoading)}
                  onSelect={() => onWaffoMethodSelect?.(method, index)}
                />
              ))}
              {standardMethods.length === 0 && waffoMethods.length === 0 && (
                <Alert>
                  <AlertDescription>
                    {t(
                      'No payment methods available. Please contact administrator.'
                    )}
                  </AlertDescription>
                </Alert>
              )}
            </div>
          </div>

        </>
      )}
      {!hasAnyTopup && (
        <Alert>
          <AlertDescription>
            {t(
              'Online topup is not enabled. Please use redemption code or contact administrator.'
            )}
          </AlertDescription>
        </Alert>
      )}

      {/* Creem Products Section */}
      {enableCreemTopup &&
        Array.isArray(creemProducts) &&
        creemProducts.length > 0 &&
        onCreemProductSelect && (
          <div className='space-y-2.5 border-t pt-4 sm:space-y-3 sm:pt-6'>
            <Label className='text-muted-foreground text-xs font-medium tracking-wider uppercase'>
              {t('Creem Payment')}
            </Label>
            <CreemProductsSection
              products={creemProducts}
              onProductSelect={onCreemProductSelect}
            />
          </div>
        )}

    </div>
  )

  return (
    <Card
      data-card-hover='false'
      className='gap-0 overflow-hidden rounded-lg py-0'
    >
      <Tabs defaultValue='online' className='gap-0'>
        <TabsList
          variant='line'
          className='h-11 w-full gap-0 rounded-none border-b p-0'
        >
          <TabsTrigger
            value='online'
            className='h-full rounded-none after:bottom-0'
          >
            {t('Online recharge')}
          </TabsTrigger>
          {corporateTransfer?.enabled && (
            <TabsTrigger
              value='corporate'
              className='h-full rounded-none after:bottom-0'
            >
              {t('Corporate Bank Transfer')}
            </TabsTrigger>
          )}
        </TabsList>
        <TabsContent value='online' className='min-w-0 p-4 sm:p-6'>
          {onlineContent}
        </TabsContent>
        {corporateTransfer?.enabled && (
          <TabsContent value='corporate' className='min-w-0 p-4 sm:p-6'>
            <CorporateTransferCard config={corporateTransfer} />
          </TabsContent>
        )}
      </Tabs>
    </Card>
  )
}

function PaymentChoice(props: {
  name: string
  icon: ReactNode
  selected: boolean
  minimum?: number
  fee?: number
  busy: boolean
  onSelect: () => void
}) {
  const { t } = useTranslation()
  const minimumLabel =
    props.minimum === undefined
      ? undefined
      : t('Minimum topup amount: {{amount}}', { amount: props.minimum })

  return (
    <Button
      type='button'
      variant='outline'
      disabled={props.busy || minimumLabel !== undefined}
      aria-pressed={props.selected}
      title={minimumLabel}
      onClick={props.onSelect}
      className={cn(
        'h-auto min-h-11 min-w-28 max-w-full justify-start px-3 py-2 text-left whitespace-normal',
        props.selected && 'border-primary bg-primary/5 text-primary'
      )}
    >
      {props.icon}
      <span className='flex min-w-0 flex-col gap-0.5 break-words'>
        {props.name}
        {props.fee !== undefined && props.fee > 0 && (
          <span className='text-muted-foreground text-xs font-normal'>
            {t('Handling fee')} {(props.fee * 100).toFixed(2)}%
          </span>
        )}
        {minimumLabel && (
          <span className='text-muted-foreground text-xs font-normal'>
            {minimumLabel}
          </span>
        )}
      </span>
    </Button>
  )
}
