import type { ModelPricingSnapshotResponse } from '../api'

export type LegacyModelPricingDefaults = {
  ModelPrice: string
  ModelRatio: string
  CacheRatio: string
  CreateCacheRatio: string
  CompletionRatio: string
  ImageRatio: string
  AudioRatio: string
  AudioCompletionRatio: string
  BillingMode: string
  BillingExpr: string
  PluginBillingExpr: string
  SoraPerRequestPricing: string
}

const optionKeys: Array<[keyof LegacyModelPricingDefaults, string]> = [
  ['ModelPrice', 'ModelPrice'],
  ['ModelRatio', 'ModelRatio'],
  ['CacheRatio', 'CacheRatio'],
  ['CreateCacheRatio', 'CreateCacheRatio'],
  ['CompletionRatio', 'CompletionRatio'],
  ['ImageRatio', 'ImageRatio'],
  ['AudioRatio', 'AudioRatio'],
  ['AudioCompletionRatio', 'AudioCompletionRatio'],
  ['BillingMode', 'billing_setting.billing_mode'],
  ['BillingExpr', 'billing_setting.billing_expr'],
  ['PluginBillingExpr', 'billing_setting.plugin_billing_expr'],
  ['SoraPerRequestPricing', 'billing_setting.sora_per_request_pricing'],
]

export function adaptModelPricingSnapshot(
  snapshot: ModelPricingSnapshotResponse | undefined,
  fallback: LegacyModelPricingDefaults
): LegacyModelPricingDefaults {
  const options = snapshot?.success ? snapshot.data?.options : undefined
  if (!options) return fallback

  const result = { ...fallback }
  for (const [field, key] of optionKeys) {
    const value = options[key]
    if (typeof value === 'string' && value.trim() !== '') {
      result[field] = value
    }
  }
  return result
}
