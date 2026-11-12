import { useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Dialog } from '@/components/dialog'
import { Button } from '@/components/ui/button'
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import { FundingKindSelect } from '@/features/wallet/components/funding-kind-select'
import type { FundingKind } from '@/features/wallet/funding'
import { getCurrencyDisplay, getCurrencyLabel } from '@/lib/currency'
import { formatQuota, parseQuotaFromDollars } from '@/lib/format'

import { adjustUserQuota } from '../api'
import type { QuotaAdjustMode } from '../types'

interface UserQuotaDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  userId: number
  currentQuota: number
  onSuccess: () => void
}

export function UserQuotaDialog(props: UserQuotaDialogProps) {
  const { t } = useTranslation()
  const [mode, setMode] = useState<QuotaAdjustMode>('add')
  const queryClient = useQueryClient()
  const [kind, setKind] = useState<FundingKind>('adjustment')
  const [note, setNote] = useState('')
  const [amount, setAmount] = useState('')
  const [loading, setLoading] = useState(false)

  const { meta: currencyMeta } = getCurrencyDisplay()
  const currencyLabel = getCurrencyLabel()
  const tokensOnly = currencyMeta.kind === 'tokens'

  const amountValue = Number.parseFloat(amount) || 0
  const quotaValue = parseQuotaFromDollars(Math.abs(amountValue))

  const getPreviewText = () => {
    const current = props.currentQuota
    const val = quotaValue
    switch (mode) {
      case 'add':
        return `${t('Current quota')}: ${formatQuota(current)}  +${formatQuota(val)} = ${formatQuota(current + val)}`
      case 'subtract':
        return `${t('Current quota')}: ${formatQuota(current)}  -${formatQuota(val)} = ${formatQuota(current - val)}`
      case 'override': {
        const overrideQuota = parseQuotaFromDollars(amountValue)
        return `${t('Current quota')}: ${formatQuota(current)} → ${formatQuota(overrideQuota)}`
      }
      default:
        return ''
    }
  }

  const handleConfirm = async () => {
    if (!Number.isFinite(Number(amount)) || Number(amount) < 0) return
    if (!amount && mode !== 'override') return
    if (quotaValue <= 0 && mode !== 'override') return

    setLoading(true)
    try {
      const value =
        mode === 'override' ? parseQuotaFromDollars(amountValue) : quotaValue
      const result = await adjustUserQuota({
        id: props.userId,
        action: 'add_quota',
        mode,
        funding_kind:
          mode === 'add' && kind !== 'unclassified' ? kind : 'adjustment',
        note,
        value: mode === 'override' ? value : Math.abs(value),
      })
      if (result.success) {
        toast.success(t('Quota adjusted successfully'))
        setAmount('')
        setMode('add')
        setKind('adjustment')
        setNote('')
        void queryClient.invalidateQueries({
          queryKey: ['wallet-funding-ledger'],
        })
        void queryClient.invalidateQueries({
          queryKey: ['wallet-funding-summary'],
        })
        props.onOpenChange(false)
        props.onSuccess()
      } else {
        toast.error(result.message || t('Failed to adjust quota'))
      }
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : t('Failed to adjust quota'))
    } finally {
      setLoading(false)
    }
  }

  const handleCancel = () => {
    setAmount('')
    setMode('add')
    setKind('adjustment')
    setNote('')
    props.onOpenChange(false)
  }

  const placeholder = tokensOnly
    ? t('Enter amount in tokens')
    : t('Enter amount in {{currency}}', { currency: currencyLabel })

  return (
    <Dialog
      open={props.open}
      onOpenChange={(open) => {
        if (!loading) {
          if (open) props.onOpenChange(true)
          else handleCancel()
        }
      }}
      title={t('Adjust Quota')}
      description={t('Select an operation mode and enter the amount')}
      contentHeight='auto'
      bodyClassName='flex flex-col gap-4'
      footer={
        <>
          <Button variant='outline' onClick={handleCancel}>
            {t('Cancel')}
          </Button>
          <Button onClick={handleConfirm} disabled={loading}>
            {loading ? t('Processing...') : t('Confirm')}
          </Button>
        </>
      }
    >
      <FieldGroup>
        <p className='text-muted-foreground text-sm'>{getPreviewText()}</p>
        <Field>
          <FieldLabel>{t('Mode')}</FieldLabel>
          <ToggleGroup
            value={[mode]}
            onValueChange={(values) => {
              const next = values[0] as QuotaAdjustMode | undefined
              if (next) {
                setMode(next)
                setAmount('')
              }
            }}
            variant='outline'
            disabled={loading}
            aria-label={t('Mode')}
          >
            <ToggleGroupItem value='add'>{t('Add')}</ToggleGroupItem>
            <ToggleGroupItem value='subtract'>{t('Subtract')}</ToggleGroupItem>
            <ToggleGroupItem value='override'>{t('Override')}</ToggleGroupItem>
          </ToggleGroup>
        </Field>
        <Field>
          <FieldLabel htmlFor='quota-adjust-amount'>
            {t('Amount')} ({currencyLabel})
          </FieldLabel>
          <Input
            id='quota-adjust-amount'
            type='number'
            step={tokensOnly ? 1 : 0.000001}
            min={0}
            placeholder={placeholder}
            value={amount}
            disabled={loading}
            onChange={(event) => setAmount(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === 'Enter') void handleConfirm()
            }}
          />
        </Field>
        <Field>
          <FieldLabel htmlFor='quota-funding-kind'>
            {t('Funding classification')}
          </FieldLabel>
          <FundingKindSelect
            id='quota-funding-kind'
            value={mode === 'add' ? kind : 'adjustment'}
            onChange={setKind}
            disabled={mode !== 'add' || loading}
          />
          <FieldDescription>
            {t(
              'Only additions classified as paid or gift affect cumulative funding totals. Subtractions and replacements are adjustments.'
            )}
          </FieldDescription>
        </Field>
        <Field>
          <FieldLabel htmlFor='quota-funding-note'>{t('Note')}</FieldLabel>
          <Textarea
            id='quota-funding-note'
            value={note}
            onChange={(event) => setNote(event.target.value)}
            maxLength={500}
            disabled={loading}
          />
        </Field>
      </FieldGroup>
    </Dialog>
  )
}
