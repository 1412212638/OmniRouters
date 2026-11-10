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
  taskUsageUnitLabel,
} from '../lib/task-price-display'
import type { BillingUsageSchema, PricingModel } from '../types'

type TaskPricingSectionProps = {
  model: PricingModel
  priceRate: number
  usdExchangeRate: number
  groupRatio?: number
  title?: string
}

function formatTaskPrice(
  value: number,
  props: Pick<TaskPricingSectionProps, 'priceRate' | 'usdExchangeRate'>
) {
  const usd = value
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
    return getTaskPricingDisplayTiers(billingExpr, schema).filter((tier) => {
      // Conditions describe a possible tier, but an unpriced tier should not
      // appear as a row full of dashes in the public pricing table.
      return (
        tier.constant > 0 ||
        Object.values(tier.unitPrices).some((price) => Number(price) > 0) ||
        Object.values(tier.freeQuantities || {}).some(
          (quantity) => Number(quantity) > 0
        )
      )
    })
  }, [expression, schema])
  const numberFields = useMemo(() => {
    return Object.entries(schema || {}).filter(
      ([field, definition]) =>
        definition.type === 'number' &&
        definition.unit &&
        tiers.some(
          (tier) =>
            Number(tier.unitPrices[field]) > 0 ||
            Number(tier.freeQuantities?.[field]) > 0
        )
    )
  }, [schema, tiers])

  if (
    !schema ||
    !expression ||
    tiers.length === 0 ||
    numberFields.length === 0
  ) {
    return null
  }

  const showFixedPrice = tiers.some((tier) => tier.constant > 0)
  const combinedVideoSecondsField = Object.entries(schema || {}).find(
    ([, definition]) =>
      definition.pricingDisplay === 'input_output_video_seconds'
  )?.[0]
  const ratio = props.groupRatio ?? 1
  const discountedPriceClassName = ratio < 1 ? 'text-success' : undefined
  const title = props.title || t('Task pricing')
  return (
    <section className='space-y-3'>
      <div className='flex items-center justify-between gap-3'>
        <h3 className='text-foreground text-sm font-semibold'>{title}</h3>
        <Badge variant='secondary'>{t('Usage based')}</Badge>
      </div>
      <p className='text-muted-foreground text-xs leading-5'>
        {combinedVideoSecondsField
          ? t(
              'With a reference video, its duration and the generated video duration are billed together.'
            )
          : t(
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
          ...(showFixedPrice
            ? [
                {
                  id: 'constant',
                  header: t('Fixed price'),
                  className: 'text-muted-foreground px-3 py-2 text-right',
                  cellClassName: 'px-3 py-2 text-right font-mono',
                  cell: (tier: (typeof tiers)[number]) => (
                    <span className={discountedPriceClassName}>
                      {tier.constant > 0
                        ? formatTaskPrice(tier.constant * ratio, props)
                        : '-'}
                    </span>
                  ),
                },
              ]
            : []),
          ...numberFields.flatMap(([field, definition]) => {
            const headers =
              field === combinedVideoSecondsField
                ? [
                    {
                      id: `${field}-input`,
                      label: t('Reference video input'),
                    },
                    {
                      id: `${field}-output`,
                      label: t('Generated video output'),
                    },
                  ]
                : [
                    {
                      id: field,
                      label: <TaskUsageHeader schema={schema} field={field} />,
                    },
                  ]
            return headers.map(({ id, label }) => ({
              id,
              header: label,
              className: 'text-muted-foreground px-3 py-2 text-right',
              cellClassName: 'px-3 py-2 text-right font-mono',
              cell: (tier: (typeof tiers)[number]) => {
                const value = Number(tier.unitPrices[field] || 0)
                const freeQuantity =
                  definition.unit === 'count'
                    ? Math.max(Number(tier.freeQuantities?.[field] || 0), 0)
                    : 0
                let unitLabel = t('credit')
                if (definition.unit === 'token') unitLabel = t('1M tokens')
                else if (definition.unit === 'second') unitLabel = t('seconds')
                else if (definition.unit === 'count') {
                  unitLabel = taskUsageUnitLabel(
                    definition,
                    i18n.language,
                    t('Count')
                  )
                }
                if (freeQuantity > 0 && value > 0) {
                  return (
                    <span className={discountedPriceClassName}>
                      {`${t('First')} ${freeQuantity} ${unitLabel} ${t('free')}; ${t('then')} ${formatTaskPrice(value * ratio, props)} / ${unitLabel}`}
                    </span>
                  )
                }
                if (freeQuantity > 0) {
                  return (
                    <span className={discountedPriceClassName}>
                      {t('Free')}
                    </span>
                  )
                }
                return (
                  <span className={discountedPriceClassName}>
                    {value > 0
                      ? `${formatTaskPrice(value * ratio, props)} / ${unitLabel}`
                      : '-'}
                  </span>
                )
              },
            }))
          }),
        ]}
      />
    </section>
  )
}
