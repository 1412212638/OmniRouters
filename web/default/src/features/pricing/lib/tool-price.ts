import { formatBillingCurrencyFromUSD } from '@/lib/currency'

import type { PricingModel, PricingToolPrice } from '../types'

const TOOL_PRICE_ORDER = [
  'web_search_preview',
  'web_search',
  'file_search',
  'google_search',
  'google_search_grounded_prompt',
  'image_generation',
  'deepseek_search',
  'deepseek_fetch',
]

export function getPricingToolPrices(model: PricingModel): PricingToolPrice[] {
  return (model.tool_prices ?? [])
    .filter(
      (tool) =>
        tool &&
        typeof tool.label === 'string' &&
        Number.isFinite(Number(tool.price_per_1k)) &&
        Number(tool.price_per_1k) >= 0
    )
    .sort(
      (left, right) =>
        (TOOL_PRICE_ORDER.indexOf(left.name) + 1 || Number.MAX_SAFE_INTEGER) -
        (TOOL_PRICE_ORDER.indexOf(right.name) + 1 || Number.MAX_SAFE_INTEGER)
    )
}

export function formatToolPrice(
  pricePerThousand: number,
  groupRatio: number,
  options: {
    showRechargePrice?: boolean
    priceRate?: number
    usdExchangeRate?: number
  } = {}
): string {
  const ratio = Number.isFinite(groupRatio) && groupRatio >= 0 ? groupRatio : 1
  const priceRate = options.priceRate ?? 1
  const usdExchangeRate = options.usdExchangeRate ?? 1
  const rechargeMultiplier = options.showRechargePrice
    ? priceRate / Math.max(usdExchangeRate, 0.001)
    : 1
  const value = Number(pricePerThousand) * ratio * rechargeMultiplier
  if (!Number.isFinite(value) || value < 0) return '-'
  return formatBillingCurrencyFromUSD(value, {
    digitsLarge: 4,
    digitsSmall: 6,
    abbreviate: false,
  })
}

export function getToolPriceSummary(
  model: PricingModel,
  groupRatio: number,
  options: Parameters<typeof formatToolPrice>[2] = {}
): string[] {
  return getPricingToolPrices(model)
    .slice(0, 2)
    .map(
      (tool) =>
        `${tool.label} ${formatToolPrice(tool.price_per_1k, groupRatio, options)}/1K`
    )
}
