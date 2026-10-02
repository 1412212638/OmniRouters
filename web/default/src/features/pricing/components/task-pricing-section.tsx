import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'

import { StaticDataTable } from '@/components/data-table'
import { Badge } from '@/components/ui/badge'
import { formatBillingCurrencyFromUSD } from '@/lib/currency'

import { splitBillingExprAndRequestRules } from '../lib/billing-expr'
import { getTaskPricingDisplayTiers } from '../lib/task-matrix-display'
import {
  taskPriceLabel,
  taskPricingConditions,
} from '../lib/task-price-display'
import type { BillingUsageSchema, PricingModel } from '../types'

const TOKEN_SCALE = 1_000_000

type TaskPricingSectionProps = {
  model: PricingModel
  priceRate: number
  usdExchangeRate: number
  groupRatio?: number
  title?: string
}

function formatTaskPrice(
  value: number,
  unit: string | undefined,
  props: Pick<TaskPricingSectionProps, 'priceRate' | 'usdExchangeRate'>
) {
  const usd = unit === 'token' ? value / TOKEN_SCALE : value
  const display = (usd * props.priceRate) / props.usdExchangeRate
  return formatBillingCurrencyFromUSD(display, {
    digitsLarge: 4,
    digitsSmall: 6,
    abbreviate: false,
  })
}

function TaskUsageHeader({
  schema,
  field,
}: {
  schema: BillingUsageSchema
  field: string
}) {
  const { i18n } = useTranslation()
  const definition = schema[field]
  return taskPriceLabel(definition?.description, field, i18n.language)
}

export function TaskPricingSection(props: TaskPricingSectionProps) {
  const { t, i18n } = useTranslation()
  const schema = props.model.billing_usage_schema
  const expression = props.model.billing_expr
  const tiers = useMemo(() => {
    if (!schema || !expression) return []
    const { billingExpr } = splitBillingExprAndRequestRules(expression)
    return getTaskPricingDisplayTiers(billingExpr, schema)
  }, [expression, schema])
  const numberFields = useMemo(
    () =>
      Object.entries(schema || {}).filter(
        ([, definition]) => definition.type === 'number' && definition.unit
      ),
    [schema]
  )

  if (
    !schema ||
    !expression ||
    tiers.length === 0 ||
    numberFields.length === 0
  ) {
    return null
  }

  const ratio = props.groupRatio ?? 1
  const title = props.title || t('Task pricing')
  return (
    <section className='space-y-3'>
      <div className='flex items-center justify-between gap-3'>
        <h3 className='text-foreground text-sm font-semibold'>{title}</h3>
        <Badge variant='secondary'>{t('Usage based')}</Badge>
      </div>
      <p className='text-muted-foreground text-xs leading-5'>
        {t(
          'Prices are charged according to the task usage fields shown below.'
        )}
      </p>
      <StaticDataTable
        className='rounded-lg border'
        tableClassName='text-xs'
        data={tiers}
        getRowKey={(tier, index) => `${tier.label}-${index}`}
        columns={[
          {
            id: 'condition',
            header: t('Task condition'),
            className: 'text-muted-foreground px-3 py-2',
            cellClassName: 'max-w-[220px] whitespace-normal px-3 py-2',
            cell: (tier) =>
              taskPricingConditions(
                tier.conditions,
                schema,
                i18n.language,
                t
              ) || t('Default'),
          },
          {
            id: 'constant',
            header: t('Fixed price'),
            className: 'text-muted-foreground px-3 py-2 text-right',
            cellClassName: 'px-3 py-2 text-right font-mono',
            cell: (tier) =>
              tier.constant > 0
                ? formatTaskPrice(tier.constant * ratio, undefined, props)
                : '-',
          },
          ...numberFields.map(([field, definition]) => ({
            id: field,
            header: <TaskUsageHeader schema={schema} field={field} />,
            className: 'text-muted-foreground px-3 py-2 text-right',
            cellClassName: 'px-3 py-2 text-right font-mono',
            cell: (tier: (typeof tiers)[number]) => {
              const value = Number(tier.unitPrices[field] || 0)
              return value > 0
                ? `${formatTaskPrice(value * ratio, definition.unit, props)} / ${definition.unit}`
                : '-'
            },
          })),
        ]}
      />
    </section>
  )
}
