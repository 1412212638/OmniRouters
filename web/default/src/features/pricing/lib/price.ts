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
import { formatCurrencyFromUSD } from '@/lib/currency'
import { QUOTA_TYPE_VALUES, TOKEN_UNIT_DIVISORS } from '../constants'
import type {
  PricingModel,
  PriceType,
  SoraResolutionTier,
  TokenUnit,
} from '../types'

// ----------------------------------------------------------------------------
// Price Calculation Utilities
// ----------------------------------------------------------------------------

/**
 * Strip trailing zeros from formatted price string while preserving currency symbols
 */
export function stripTrailingZeros(formatted: string): string {
  // Match currency symbol at start, number, and potential 'k' suffix
  const match = formatted.match(/^([^\d-]*)([-\d,]+\.?\d*)(k?)$/)
  if (!match) return formatted

  const [, symbol, number, suffix] = match

  // Remove commas for processing
  const cleanNumber = number.replace(/,/g, '')

  // Convert to number and back to remove trailing zeros
  const parsed = parseFloat(cleanNumber)
  if (isNaN(parsed)) return formatted

  // Convert to string, which automatically removes trailing zeros
  let result = parsed.toString()

  // If the result is in scientific notation, format it properly
  if (result.includes('e')) {
    result = parsed.toFixed(20).replace(/\.?0+$/, '')
  }

  return `${symbol}${result}${suffix}`
}

/**
 * Find minimum group ratio from enabled groups
 */
export function getMinGroupRatio(
  enableGroups: string[],
  groupRatio: Record<string, number>
): number {
  if (enableGroups.length === 0) return 1

  let minRatio = Number.POSITIVE_INFINITY

  for (const group of enableGroups) {
    const ratio = groupRatio[group]
    if (ratio !== undefined && ratio < minRatio) {
      minRatio = ratio
    }
  }

  return minRatio === Number.POSITIVE_INFINITY ? 1 : minRatio
}

export type SoraPricingDisplayTier = {
  key: string
  label: string
  multiplier: number
  price: string
}

export type SoraPricingDisplay = {
  basePrice: string
  tierCount: number
  resolutionTiers: SoraPricingDisplayTier[]
  audioGenerationSurcharge?: string
}

type SoraPricingOptions = {
  showRechargePrice?: boolean
  priceRate?: number
  usdExchangeRate?: number
}

export function isSoraPerRequestPricingModel(model: PricingModel): boolean {
  return (
    model.quota_type === QUOTA_TYPE_VALUES.REQUEST &&
    Boolean(model.sora_per_request_pricing?.enabled)
  )
}

function getSoraDisplayGroupRatio(model: PricingModel): number {
  const enableGroups = Array.isArray(model.enable_groups)
    ? model.enable_groups
    : []
  const groupRatio = model.group_ratio || {}
  return getMinGroupRatio(enableGroups, groupRatio)
}

function getSoraBasePriceUSD(model: PricingModel): number {
  return (model.model_price || 0) * getSoraDisplayGroupRatio(model)
}

function normalizeSoraResolutionTiers(
  tiers: SoraResolutionTier[] | undefined | null
): SoraResolutionTier[] {
  if (!Array.isArray(tiers)) return []

  return tiers
    .map((tier) => ({
      value: String(tier?.value || '').trim(),
      multiplier: Number(tier?.multiplier),
    }))
    .filter(
      (tier) =>
        Boolean(tier.value) &&
        Number.isFinite(tier.multiplier) &&
        tier.multiplier > 0
    )
}

export function getSoraPricingDisplay(
  model: PricingModel,
  options: SoraPricingOptions = {}
): SoraPricingDisplay | null {
  if (!isSoraPerRequestPricingModel(model)) return null

  const priceRate = options.priceRate ?? 1
  const usdExchangeRate = options.usdExchangeRate ?? 1
  const showRechargePrice = options.showRechargePrice ?? false

  let basePriceInUSD = getSoraBasePriceUSD(model)
  basePriceInUSD = applyRechargeRate(
    basePriceInUSD,
    showRechargePrice,
    priceRate,
    usdExchangeRate
  )

  const basePrice = formatCurrencyFromUSD(basePriceInUSD, {
    digitsLarge: 4,
    digitsSmall: 4,
    abbreviate: false,
  })
  const displayGroupRatio = getSoraDisplayGroupRatio(model)
  let audioGenerationSurchargeInUSD =
    Number(model.sora_per_request_pricing?.audio_generation_surcharge) *
    displayGroupRatio
  const hasAudioGenerationSurcharge =
    Number.isFinite(audioGenerationSurchargeInUSD) &&
    audioGenerationSurchargeInUSD > 0
  if (hasAudioGenerationSurcharge) {
    audioGenerationSurchargeInUSD = applyRechargeRate(
      audioGenerationSurchargeInUSD,
      showRechargePrice,
      priceRate,
      usdExchangeRate
    )
  }
  const formatSoraPrice = (priceInUSD: number) =>
    formatCurrencyFromUSD(priceInUSD, {
      digitsLarge: 4,
      digitsSmall: 4,
      abbreviate: false,
    })
  const resolutionTiers = normalizeSoraResolutionTiers(
    model.sora_per_request_pricing?.resolution_tiers
  ).map((tier) => {
    const tierPriceInUSD = basePriceInUSD * tier.multiplier
    return {
      key: tier.value,
      label: tier.value,
      multiplier: tier.multiplier,
      price: formatSoraPrice(tierPriceInUSD),
    }
  })

  return {
    basePrice,
    tierCount: resolutionTiers.length,
    resolutionTiers,
    ...(hasAudioGenerationSurcharge
      ? {
          audioGenerationSurcharge: formatSoraPrice(
            audioGenerationSurchargeInUSD
          ),
        }
      : {}),
  }
}

/**
 * Calculate token price in USD.
 *
 * Returns NaN when the required ratio field is missing/null so callers can
 * skip rendering that price type.
 */
function calculateTokenPrice(
  model: PricingModel,
  type: PriceType,
  ratio: number
): number {
  const base = model.model_ratio * 2 * ratio

  switch (type) {
    case 'input':
      return base
    case 'output':
      return base * model.completion_ratio
    case 'cache':
      return hasRatio(model.cache_ratio)
        ? base * Number(model.cache_ratio)
        : NaN
    case 'create_cache':
      return hasRatio(model.create_cache_ratio)
        ? base * Number(model.create_cache_ratio)
        : NaN
    case 'image':
      return hasRatio(model.image_ratio)
        ? base * Number(model.image_ratio)
        : NaN
    case 'audio_input':
      return hasRatio(model.audio_ratio)
        ? base * Number(model.audio_ratio)
        : NaN
    case 'audio_output':
      return hasRatio(model.audio_ratio) &&
        hasRatio(model.audio_completion_ratio)
        ? base *
            Number(model.audio_ratio) *
            Number(model.audio_completion_ratio)
        : NaN
  }
}

function hasRatio(value: number | null | undefined): boolean {
  return value !== undefined && value !== null && Number.isFinite(Number(value))
}

/**
 * Apply recharge rate to price
 *
 * priceRate represents how much users need to recharge (in the display currency)
 * to get 1 USD credit. usdExchangeRate is the real exchange rate.
 *
 * The returned value will be formatted by formatCurrencyFromUSD, which will
 * multiply by the display currency's exchange rate.
 *
 * Examples:
 *
 * 1. Display currency = USD:
 *    - Model: 1 USD
 *    - priceRate = 0.5 (recharge $0.5 to get $1 credit)
 *    - usdExchangeRate = 1
 *    - Return: 1 × 0.5 / 1 = 0.5
 *    - formatCurrencyFromUSD(0.5) → $0.5 ✓
 *
 * 2. Display currency = CNY:
 *    - Model: 1 USD
 *    - priceRate = 4 (recharge ¥4 to get $1 credit)
 *    - usdExchangeRate = 7 (real rate: 1 USD = ¥7)
 *    - Return: 1 × 4 / 7 = 0.571
 *    - formatCurrencyFromUSD(0.571) → 0.571 × 7 = ¥4 ✓
 *    - Normal price: ¥7, Recharge price: ¥4 (cheaper!)
 */
function applyRechargeRate(
  price: number,
  showWithRecharge: boolean,
  priceRate: number,
  usdExchangeRate: number
): number {
  if (!showWithRecharge) return price
  return (price * priceRate) / usdExchangeRate
}

/**
 * Format token-based price for display
 */
export function formatPrice(
  model: PricingModel,
  type: PriceType,
  tokenUnit: TokenUnit,
  showWithRecharge = false,
  priceRate = 1,
  usdExchangeRate = 1
): string {
  if (model.quota_type === QUOTA_TYPE_VALUES.REQUEST) {
    return '-'
  }

  const enableGroups = Array.isArray(model.enable_groups)
    ? model.enable_groups
    : []
  const groupRatio = model.group_ratio || {}
  const minRatio = getMinGroupRatio(enableGroups, groupRatio)

  let priceInUSD = calculateTokenPrice(model, type, minRatio)
  priceInUSD = applyRechargeRate(
    priceInUSD,
    showWithRecharge,
    priceRate,
    usdExchangeRate
  )

  const price = priceInUSD / TOKEN_UNIT_DIVISORS[tokenUnit]
  return formatCurrencyFromUSD(price, {
    digitsLarge: 4,
    digitsSmall: 6,
    abbreviate: false,
  })
}

/**
 * Format price for a specific group (token-based)
 */
export function formatGroupPrice(
  model: PricingModel,
  group: string,
  type: PriceType,
  tokenUnit: TokenUnit,
  showWithRecharge = false,
  priceRate = 1,
  usdExchangeRate = 1,
  groupRatio: Record<string, number>
): string {
  if (model.quota_type === QUOTA_TYPE_VALUES.REQUEST) {
    return '-'
  }

  const ratio = groupRatio[group] || 1
  let priceInUSD = calculateTokenPrice(model, type, ratio)

  priceInUSD = applyRechargeRate(
    priceInUSD,
    showWithRecharge,
    priceRate,
    usdExchangeRate
  )

  const price = priceInUSD / TOKEN_UNIT_DIVISORS[tokenUnit]
  return formatCurrencyFromUSD(price, {
    digitsLarge: 4,
    digitsSmall: 6,
    abbreviate: false,
  })
}

/**
 * Format fixed price for pay-per-request models (with specific group)
 */
export function formatFixedPrice(
  model: PricingModel,
  group: string,
  showWithRecharge = false,
  priceRate = 1,
  usdExchangeRate = 1,
  groupRatio: Record<string, number>
): string {
  if (model.quota_type !== QUOTA_TYPE_VALUES.REQUEST) {
    return '-'
  }

  const ratio = groupRatio[group] || 1
  let priceInUSD = (model.model_price || 0) * ratio

  priceInUSD = applyRechargeRate(
    priceInUSD,
    showWithRecharge,
    priceRate,
    usdExchangeRate
  )

  return formatCurrencyFromUSD(priceInUSD, {
    digitsLarge: 4,
    digitsSmall: 4,
    abbreviate: false,
  })
}

/**
 * Format fixed price for pay-per-request models (minimum price from all groups)
 */
export function formatRequestPrice(
  model: PricingModel,
  showWithRecharge = false,
  priceRate = 1,
  usdExchangeRate = 1
): string {
  if (model.quota_type !== QUOTA_TYPE_VALUES.REQUEST) {
    return '-'
  }

  const enableGroups = Array.isArray(model.enable_groups)
    ? model.enable_groups
    : []
  const groupRatio = model.group_ratio || {}
  const minRatio = getMinGroupRatio(enableGroups, groupRatio)

  let priceInUSD = (model.model_price || 0) * minRatio

  priceInUSD = applyRechargeRate(
    priceInUSD,
    showWithRecharge,
    priceRate,
    usdExchangeRate
  )

  return formatCurrencyFromUSD(priceInUSD, {
    digitsLarge: 4,
    digitsSmall: 4,
    abbreviate: false,
  })
}
