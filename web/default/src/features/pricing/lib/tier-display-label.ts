import type { TOptions } from 'i18next'

import { toIntlLocale } from '@/i18n/languages'

import type { ParsedTier, TierCondition } from './billing-expr'
import { formatBillingCondition } from './billing-expression/condition-display'

const conditionLabels = {
  p: 'Input Tokens',
  c: 'Output Tokens',
  len: 'Context Length',
} as const

const comparisonLabels = { '<': '<', '<=': '≤', '>': '>', '>=': '≥' }
const inverseComparison = {
  '<': '>=',
  '<=': '>',
  '>': '<=',
  '>=': '<',
} as const

function formatTokenCount(value: number, locale: string): string {
  if (value >= 1_000_000 && value % 1_000_000 === 0) {
    return `${value / 1_000_000}M`
  }
  if (value >= 1_000 && value % 1_000 === 0) return `${value / 1_000}K`
  // A rounded threshold can describe a different price range.
  return value.toLocaleString(toIntlLocale(locale), {
    maximumSignificantDigits: 21,
  })
}

function strongestTokenBounds(conditions: TierCondition[]): TierCondition[] {
  const bounds = new Map<string, TierCondition>()
  for (const condition of conditions) {
    const isUpper = condition.op.startsWith('<')
    const key = `${condition.var}:${isUpper}`
    const previous = bounds.get(key)
    if (
      !previous ||
      (isUpper && condition.value < previous.value) ||
      (!isUpper && condition.value > previous.value) ||
      (condition.value === previous.value && condition.op.length === 1)
    ) {
      bounds.set(key, condition)
    }
  }
  return [...bounds.values()]
}

/** Use tier conditions only for display. The original label identifies the tier in usage logs. */
export function formatTierDisplayLabel(
  tier: ParsedTier,
  tiers: ParsedTier[],
  t: (key: string, options?: TOptions) => string,
  locale = 'en'
): string {
  if (tier.conditionText) {
    if (tier.conditionText === 'Audio requests') return t('Audio requests')
    if (tier.conditionText === 'Text-only requests') {
      return t('Text-only requests')
    }
    return (
      formatBillingCondition(tier.conditionText, t, locale) ??
      (tier.label || t('Default'))
    )
  }

  const index = tiers.indexOf(tier)
  const preceding = index > 0 ? tiers.slice(0, index) : []
  const hasComplexPreceding = preceding.some(
    (branch) => branch.conditionText || branch.conditions.length !== 1
  )
  const inferred = hasComplexPreceding
    ? []
    : preceding.map(({ conditions: [condition] }) => ({
        ...condition,
        op: inverseComparison[condition.op],
      }))

  const condition = strongestTokenBounds([...inferred, ...tier.conditions])
    .map((bound) => {
      const label = t(conditionLabels[bound.var])
      return `${label} ${comparisonLabels[bound.op]} ${formatTokenCount(bound.value, locale)} Tokens`
    })
    .join(` ${t('and')} `)

  if (condition) {
    return hasComplexPreceding
      ? t('Otherwise, {{condition}}', {
          condition,
          interpolation: { escapeValue: false },
        })
      : condition
  }
  if (preceding.length > 0) return t('Other requests')
  return tier.label || t('Default')
}
