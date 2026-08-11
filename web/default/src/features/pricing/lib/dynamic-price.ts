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
} from './billing-expr'
import { getDisplayGroupRatio } from './model-helpers'

type DynamicPriceOptions = {
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
  displayUnit: 'token' | 'request'
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
}

const PRIMARY_DYNAMIC_FIELDS = new Set(['inputPrice', 'outputPrice'])

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
      fixedPrice * groupRatio,
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
      minValue === maxValue
        ? minFormatted
        : `${minFormatted}-${maxFormatted}`,
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

export function getDynamicPricingSummary(
  model: PricingModel,
  options: DynamicPriceOptions
): DynamicPricingSummary | null {
  if (!isDynamicPricingModel(model)) return null

  const tiers = getDynamicPricingTiers(model)
  const tier = tiers[0] || null
  const entries = getDynamicPriceEntries(tier, options)
  const rawExpression = model.billing_expr || ''
  const hasRequestRules = hasDynamicRequestRules(model)

  return {
    tiers,
    tier,
    tierCount: tiers.length,
    hasRequestRules,
    isSpecialExpression: rawExpression.trim().length > 0 && tiers.length === 0,
    rawExpression,
    entries,
    primaryEntries: entries.filter((entry) =>
      PRIMARY_DYNAMIC_FIELDS.has(entry.field)
    ),
    secondaryEntries: entries.filter(
      (entry) => !PRIMARY_DYNAMIC_FIELDS.has(entry.field)
    ),
    primaryRanges: hasRequestRules
      ? []
      : getDynamicPrimaryPriceRanges(tiers, options),
  }
}
