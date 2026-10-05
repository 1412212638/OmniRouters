import { formatBillingCurrencyFromUSD } from '@/lib/currency'

import { TOKEN_UNIT_DIVISORS } from '../constants'
import type { PricingModel, PricingUsableGroup, TokenUnit } from '../types'
import {
  BILLING_PRICING_VARS,
  parseTiersFromExpr,
  splitBillingExprAndRequestRules,
  tryParseRequestRuleExpr,
  type BillingVar,
  type ParsedTier,
  type RequestRuleGroup,
} from './billing-expr'
import { getDisplayGroupRatio } from './model-helpers'
import { getTaskPricingDisplayTiers } from './task-matrix-display'

export type DynamicPriceOptions = {
  tokenUnit: TokenUnit
  showRechargePrice?: boolean
  priceRate?: number
  usdExchangeRate?: number
  groupRatioMultiplier?: number
}

export type DynamicPriceEntry = {
  key: string
  field: string
  label: string
  shortLabel: string
  value: number
  formatted: string
  variable?: BillingVar
  displayUnit: 'token' | 'request' | 'custom'
  unit?: string
}

export type DynamicPriceRange = {
  key: string
  field: string
  minValue: number
  maxValue: number
  formatted: string
  displayUnit: 'token'
}

export type DynamicPricingSummary = {
  tiers: ParsedTier[]
  tier: ParsedTier | null
  tierCount: number
  hasRequestRules: boolean
  isSpecialExpression: boolean
  rawExpression: string
  entries: DynamicPriceEntry[]
  primaryEntries: DynamicPriceEntry[]
  secondaryEntries: DynamicPriceEntry[]
  primaryRanges: DynamicPriceRange[]
  requestPriceRange: string | null
}

const PRIMARY_DYNAMIC_FIELDS = new Set(['inputPrice', 'outputPrice'])

function isPrimaryDynamicField(field: string): boolean {
  return (
    PRIMARY_DYNAMIC_FIELDS.has(field) ||
    field === 'input_tokens' ||
    field === 'output_tokens'
  )
}

export function isDynamicPricingModel(model: PricingModel): boolean {
  return model.billing_mode === 'tiered_expr' && Boolean(model.billing_expr)
}

export function getDynamicDisplayGroupRatio(
  model: PricingModel,
  selectedGroup?: string,
  usableGroup?: PricingUsableGroup
): number {
  return getDisplayGroupRatio(model, selectedGroup, usableGroup)
}

function applyRechargeRate(
  price: number,
  showWithRecharge: boolean,
  priceRate: number,
  usdExchangeRate: number
): number {
  if (!showWithRecharge) return price
  return (price * priceRate) / usdExchangeRate
}

function formatTaskUnitPrice(
  value: number,
  unit: string | undefined,
  options: DynamicPriceOptions
): string {
  if (unit === 'token') return formatDynamicUnitPrice(value, options)
  const groupRatio = options.groupRatioMultiplier ?? 1
  const priceRate = options.priceRate ?? 1
  const usdExchangeRate = options.usdExchangeRate ?? 1
  return formatBillingCurrencyFromUSD(
    applyRechargeRate(
      value * groupRatio,
      options.showRechargePrice ?? false,
      priceRate,
      usdExchangeRate
    ),
    { digitsLarge: 4, digitsSmall: 6, abbreviate: false }
  )
}

function taskFieldLabel(field: string): string {
  if (field === 'input_tokens') return 'Input'
  if (field === 'output_tokens') return 'Output'
  if (field === 'cached_tokens') return 'Cached'
  return field
}

function getTaskPriceEntries(
  model: PricingModel,
  options: DynamicPriceOptions,
  taskTiers?: ReturnType<typeof getTaskPricingDisplayTiers>
): DynamicPriceEntry[] {
  const schema = model.billing_usage_schema
  if (!schema || !model.billing_expr) return []
  const { billingExpr } = splitBillingExprAndRequestRules(model.billing_expr)
  const tiers = taskTiers ?? getTaskPricingDisplayTiers(billingExpr, schema)
  if (tiers.length === 0) return []

  const fields = Object.keys(schema).filter((field) =>
    tiers.some((tier) => Number(tier.unitPrices[field]) > 0)
  )
  return fields.flatMap((field) => {
    const definition = schema[field]
    const values = tiers
      .map((tier) => Number(tier.unitPrices[field]))
      .filter((value) => Number.isFinite(value) && value > 0)
    if (!definition?.unit || values.length === 0) return []
    const min = Math.min(...values)
    const max = Math.max(...values)
    const formattedMin = formatTaskUnitPrice(min, definition.unit, options)
    const formattedMax = formatTaskUnitPrice(max, definition.unit, options)
    const displayUnit = definition.unit === 'token' ? 'token' : 'custom'
    return [
      {
        key: `task:${field}`,
        field,
        label: taskFieldLabel(field),
        shortLabel: taskFieldLabel(field),
        value: min,
        formatted:
          min === max ? formattedMin : `${formattedMin}-${formattedMax}`,
        displayUnit,
        unit: definition.unit,
      },
    ]
  })
}

export function formatDynamicUnitPrice(
  valuePerMillionTokens: number,
  options: DynamicPriceOptions
): string {
  const groupRatio = options.groupRatioMultiplier ?? 1
  const priceRate = options.priceRate ?? 1
  const usdExchangeRate = options.usdExchangeRate ?? 1
  const priceUSD =
    (valuePerMillionTokens * groupRatio) /
    TOKEN_UNIT_DIVISORS[options.tokenUnit]
  const displayPrice = applyRechargeRate(
    priceUSD,
    options.showRechargePrice ?? false,
    priceRate,
    usdExchangeRate
  )

  return formatBillingCurrencyFromUSD(displayPrice, {
    digitsLarge: 4,
    digitsSmall: 6,
    abbreviate: false,
  })
}

export function getDynamicPricingTiers(model: PricingModel): ParsedTier[] {
  if (!isDynamicPricingModel(model)) return []
  const { billingExpr } = splitBillingExprAndRequestRules(
    model.billing_expr || ''
  )
  return parseTiersFromExpr(billingExpr)
}

export function hasDynamicRequestRules(model: PricingModel): boolean {
  if (!isDynamicPricingModel(model)) return false
  const { requestRuleExpr } = splitBillingExprAndRequestRules(
    model.billing_expr || ''
  )
  return Boolean(tryParseRequestRuleExpr(requestRuleExpr || '')?.length)
}

export function getDynamicPriceEntries(
  tier: ParsedTier | null,
  options: DynamicPriceOptions
): DynamicPriceEntry[] {
  if (!tier) return []

  const entries: DynamicPriceEntry[] = BILLING_PRICING_VARS.flatMap(
    (variable) => {
      if (!variable.field) return []
      const value = Number(tier[variable.field])
      if (!Number.isFinite(value) || value <= 0) return []

      return [
        {
          key: variable.key,
          field: variable.field,
          label: variable.label,
          shortLabel: variable.shortLabel,
          value,
          formatted: formatDynamicUnitPrice(value, options),
          variable,
          displayUnit: 'token' as const,
        },
      ]
    }
  ).sort((a, b) => {
    const aPrimary = PRIMARY_DYNAMIC_FIELDS.has(a.field)
    const bPrimary = PRIMARY_DYNAMIC_FIELDS.has(b.field)
    if (aPrimary !== bPrimary) return aPrimary ? -1 : 1
    return 0
  })

  const fixedPrice = Number(tier.fixedPrice || 0)
  if (fixedPrice > 0) {
    const groupRatio = options.groupRatioMultiplier ?? 1
    const priceRate = options.priceRate ?? 1
    const usdExchangeRate = options.usdExchangeRate ?? 1
    const fixedPriceUSD = applyRechargeRate(
      (fixedPrice * groupRatio) / 1_000_000,
      options.showRechargePrice ?? false,
      priceRate,
      usdExchangeRate
    )
    const fixedEntry: DynamicPriceEntry = {
      key: 'fixedPrice',
      field: 'fixedPrice',
      label: 'Fixed price',
      shortLabel: 'Fixed price',
      value: fixedPrice,
      formatted: formatBillingCurrencyFromUSD(fixedPriceUSD, {
        digitsLarge: 4,
        digitsSmall: 6,
        abbreviate: false,
      }),
      displayUnit: 'request',
    }
    entries.push(fixedEntry)
  }

  return entries
}

function getDynamicPriceRange(
  tiers: ParsedTier[],
  variable: BillingVar,
  options: DynamicPriceOptions
): DynamicPriceRange | null {
  const field = variable.field
  if (!field || tiers.length < 2) return null

  const values = tiers.map((tier) => Number(tier[field]))
  if (values.some((value) => !Number.isFinite(value) || value < 0)) {
    return null
  }

  const minValue = Math.min(...values)
  const maxValue = Math.max(...values)
  const minFormatted = formatDynamicUnitPrice(minValue, options)
  const maxFormatted = formatDynamicUnitPrice(maxValue, options)

  return {
    key: variable.key,
    field,
    minValue,
    maxValue,
    formatted:
      minValue === maxValue ? minFormatted : `${minFormatted}-${maxFormatted}`,
    displayUnit: 'token',
  }
}

function getDynamicPrimaryPriceRanges(
  tiers: ParsedTier[],
  options: DynamicPriceOptions
): DynamicPriceRange[] {
  return BILLING_PRICING_VARS.flatMap((variable) => {
    if (!PRIMARY_DYNAMIC_FIELDS.has(variable.field || '')) return []
    const range = getDynamicPriceRange(tiers, variable, options)
    return range ? [range] : []
  })
}

function getTimeRuleInterval(
  condition: RequestRuleGroup['conditions'][number]
): [number, number] | null {
  if (condition.source !== 'time') return null

  if (condition.mode === 'eq') {
    const value = Number(condition.value)
    return Number.isFinite(value) ? [value, value + 1] : null
  }

  if (condition.mode === 'range') {
    const start = Number(condition.rangeStart)
    const end = Number(condition.rangeEnd)
    return Number.isFinite(start) && Number.isFinite(end) && start < end
      ? [start, end]
      : null
  }

  return null
}

function timeRuleGroupsAreDisjoint(
  left: RequestRuleGroup,
  right: RequestRuleGroup
): boolean {
  for (const leftCondition of left.conditions) {
    const leftInterval = getTimeRuleInterval(leftCondition)
    if (!leftInterval) continue

    for (const rightCondition of right.conditions) {
      if (
        rightCondition.source !== 'time' ||
        leftCondition.source !== 'time' ||
        rightCondition.timeFunc !== leftCondition.timeFunc ||
        rightCondition.timezone !== leftCondition.timezone
      ) {
        continue
      }
      const rightInterval = getTimeRuleInterval(rightCondition)
      if (!rightInterval) continue
      if (
        leftInterval[0] >= rightInterval[1] ||
        rightInterval[0] >= leftInterval[1]
      ) {
        return true
      }
    }
  }

  return false
}

/**
 * Returns a safe card-only price range for request-rule pricing. The full
 * request rules remain available to the detail view; this helper only handles
 * mutually exclusive time windows such as peak/off-peak pricing.
 */
export function getDynamicCatalogPriceRanges(
  model: PricingModel,
  options: DynamicPriceOptions
): DynamicPriceRange[] | null {
  if (!isDynamicPricingModel(model)) return null

  const tiers = getDynamicPricingTiers(model)
  if (tiers.length !== 1) return null

  const { requestRuleExpr } = splitBillingExprAndRequestRules(
    model.billing_expr || ''
  )
  const groups = tryParseRequestRuleExpr(requestRuleExpr || '')
  if (!groups || groups.length === 0) return null

  const multipliers = groups.map((group) => Number(group.multiplier))
  if (
    multipliers.some(
      (multiplier) => !Number.isFinite(multiplier) || multiplier < 0
    )
  ) {
    return null
  }

  const allTimeRules = groups.every((group) =>
    group.conditions.every((condition) => condition.source === 'time')
  )
  const disjoint =
    allTimeRules &&
    groups.every((group, index) =>
      groups
        .slice(index + 1)
        .every((other) => timeRuleGroupsAreDisjoint(group, other))
    )
  if (!disjoint) return null

  const multiplierRange = {
    min: Math.min(1, ...multipliers),
    max: Math.max(1, ...multipliers),
  }

  return BILLING_PRICING_VARS.flatMap((variable) => {
    if (!PRIMARY_DYNAMIC_FIELDS.has(variable.field || '')) return []
    const field = variable.field
    if (!field) return []
    const baseValue = Number(tiers[0][field])
    if (!Number.isFinite(baseValue) || baseValue < 0) return []

    const minValue = baseValue * multiplierRange.min
    const maxValue = baseValue * multiplierRange.max
    const minFormatted = formatDynamicUnitPrice(minValue, options)
    const maxFormatted = formatDynamicUnitPrice(maxValue, options)
    return [
      {
        key: variable.key,
        field,
        minValue,
        maxValue,
        formatted:
          minValue === maxValue
            ? minFormatted
            : `${minFormatted}-${maxFormatted}`,
        displayUnit: 'token' as const,
      },
    ]
  })
}

function getDynamicRequestPriceRange(
  tiers: ParsedTier[],
  options: DynamicPriceOptions
): string | null {
  const values = tiers
    .map((tier) => Number(tier.fixedPrice))
    .filter((value) => Number.isFinite(value) && value > 0)
  if (values.length === 0) return null

  const groupRatio = options.groupRatioMultiplier ?? 1
  const priceRate = options.priceRate ?? 1
  const usdExchangeRate = options.usdExchangeRate ?? 1
  const formatted = (value: number) =>
    formatBillingCurrencyFromUSD(
      applyRechargeRate(
        (value * groupRatio) / 1_000_000,
        options.showRechargePrice ?? false,
        priceRate,
        usdExchangeRate
      ),
      { digitsLarge: 4, digitsSmall: 6, abbreviate: false }
    )
  const min = Math.min(...values)
  const max = Math.max(...values)
  return min === max ? formatted(min) : `${formatted(min)}-${formatted(max)}`
}

export function getDynamicPricingSummary(
  model: PricingModel,
  options: DynamicPriceOptions
): DynamicPricingSummary | null {
  if (!isDynamicPricingModel(model)) return null

  const tiers = getDynamicPricingTiers(model)
  const tier = tiers[0] || null
  const taskTiers = model.billing_usage_schema
    ? getTaskPricingDisplayTiers(
        splitBillingExprAndRequestRules(model.billing_expr || '').billingExpr,
        model.billing_usage_schema
      )
    : []
  const taskEntries = model.billing_usage_schema
    ? getTaskPriceEntries(model, options, taskTiers)
    : []
  const entries =
    taskEntries.length > 0 ? taskEntries : getDynamicPriceEntries(tier, options)
  const rawExpression = model.billing_expr || ''
  const hasRequestRules = hasDynamicRequestRules(model)

  return {
    tiers,
    tier,
    tierCount: tiers.length || taskTiers.length,
    hasRequestRules,
    isSpecialExpression:
      rawExpression.trim().length > 0 &&
      tiers.length === 0 &&
      taskEntries.length === 0,
    rawExpression,
    entries,
    primaryEntries: entries.filter((entry) =>
      isPrimaryDynamicField(entry.field)
    ),
    secondaryEntries: entries.filter(
      (entry) => !isPrimaryDynamicField(entry.field)
    ),
    primaryRanges: hasRequestRules
      ? []
      : getDynamicPrimaryPriceRanges(tiers, options),
    requestPriceRange: getDynamicRequestPriceRange(tiers, options),
  }
}
