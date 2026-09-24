import {
  formatLocalCurrencyAmount,
  getCurrencyDisplay,
  type CurrencyFormatOptions,
} from '@/lib/currency'

import { DEFAULT_DISCOUNT_RATE } from '../constants'

// ============================================================================
// Wallet-specific Formatting Functions
// ============================================================================

/**
 * Format Creem price with currency symbol (USD/EUR)
 */
export function formatCreemPrice(
  price: number,
  currency: 'USD' | 'EUR'
): string {
  const symbol = currency === 'EUR' ? '€' : '$'
  return `${symbol}${price.toFixed(2)}`
}

/**
 * Format large quota numbers with K/M suffix
 */
export function formatQuotaShort(quota: number): string {
  if (quota >= 1000000) {
    return `${(quota / 1000000).toFixed(1)}M`
  }
  if (quota >= 1000) {
    return `${(quota / 1000).toFixed(1)}K`
  }
  return quota.toString()
}

/**
 * Format a payment amount returned by the top-up amount APIs.
 *
 * The classic frontend treats those API values as the base payment currency
 * amount. When admins display balances in USD or a custom currency, it converts
 * the payment amount back through the configured USD exchange rate before
 * formatting it. This keeps the wallet display aligned with the selected
 * currency while preserving the original amount sent to payment APIs.
 */
export function formatTopupPaymentAmount(
  amount: number | string | null | undefined,
  options?: CurrencyFormatOptions
): string {
  const numeric =
    typeof amount === 'number' ? amount : Number.parseFloat(String(amount ?? 0))
  if (!Number.isFinite(numeric)) return '-'

  const { config, meta } = getCurrencyDisplay()
  const usdExchangeRate =
    config.usdExchangeRate && config.usdExchangeRate > 0
      ? config.usdExchangeRate
      : 1

  let displayAmount = numeric
  if (meta.kind === 'currency' && meta.currencyCode === 'USD') {
    displayAmount = numeric / usdExchangeRate
  } else if (meta.kind === 'custom') {
    displayAmount = (numeric / usdExchangeRate) * meta.exchangeRate
  }

  return formatLocalCurrencyAmount(displayAmount, options)
}

export function normalizeTopupFeeRate(feeRate: unknown): number {
  const numeric = Number(feeRate || 0)
  return Number.isFinite(numeric) && numeric > 0 && numeric <= 1 ? numeric : 0
}

// Pancake quotes are already USD, matching its checkout session currency.
export function formatPaymentQuote(
  amount: number,
  paymentType?: string
): string {
  if (paymentType === 'waffo_pancake') {
    return Number.isFinite(amount) ? `$${amount.toFixed(2)}` : '-'
  }
  return formatTopupPaymentAmount(amount)
}

/**
 * Get discount label for display (e.g., "20% OFF")
 */
export function getDiscountLabel(discount: number): string {
  if (discount >= DEFAULT_DISCOUNT_RATE) {
    return ''
  }
  const off = Math.round((1 - discount) * 100)
  return `${off}% OFF`
}

/**
 * Calculate pricing details for a preset amount
 */
export function calculatePresetPricing(
  presetValue: number,
  priceRatio: number,
  discount: number,
  usdExchangeRate: number = 1
) {
  const originalPrice = presetValue * priceRatio
  const actualPrice = originalPrice * discount
  const savedAmount = originalPrice - actualPrice
  const hasDiscount = discount < 1.0
  const displayValue = presetValue * usdExchangeRate

  return {
    displayValue,
    originalPrice,
    actualPrice,
    savedAmount,
    hasDiscount,
  }
}
