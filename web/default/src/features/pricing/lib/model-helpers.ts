import { EXCLUDED_GROUPS, FILTER_ALL, QUOTA_TYPE_VALUES } from '../constants'
import type { PricingModel, PricingUsableGroup } from '../types'

// ----------------------------------------------------------------------------
// Model Helper Utilities
// ----------------------------------------------------------------------------

/**
 * Get available groups for a model
 */
export function getAvailableGroups(
  model: PricingModel,
  usableGroup: PricingUsableGroup
): string[] {
  const modelEnableGroups = Array.isArray(model.enable_groups)
    ? model.enable_groups
    : []

  return Object.keys(usableGroup)
    .filter((g) => !EXCLUDED_GROUPS.includes(g))
    .filter(
      (g) => modelEnableGroups.includes(g) || modelEnableGroups.includes('all')
    )
}

/**
 * Read a configured group ratio while preserving valid zero ratios.
 */
export function getConfiguredGroupRatio(
  groupRatio: Record<string, number>,
  group: string
): number {
  const ratio = groupRatio[group]
  return typeof ratio === 'number' && Number.isFinite(ratio) ? ratio : 1
}

/**
 * Resolve the group ratio used by model square summary prices.
 *
 * When no specific group is selected, the model square shows the best price
 * available to the viewer. When a group filter is active, it mirrors classic
 * and shows that group's price.
 */
export function getDisplayGroupRatio(
  model: PricingModel,
  selectedGroup?: string,
  usableGroup?: PricingUsableGroup
): number {
  const modelEnableGroups = Array.isArray(model.enable_groups)
    ? model.enable_groups
    : []
  const groupRatio = model.group_ratio || {}

  const availableGroups = usableGroup
    ? getAvailableGroups(model, usableGroup)
    : modelEnableGroups.filter((group) => group !== 'all')

  if (
    selectedGroup &&
    selectedGroup !== FILTER_ALL &&
    availableGroups.includes(selectedGroup) &&
    (modelEnableGroups.includes(selectedGroup) ||
      modelEnableGroups.includes('all'))
  ) {
    return getConfiguredGroupRatio(groupRatio, selectedGroup)
  }

  if (modelEnableGroups.length === 0) {
    return 1
  }

  let minRatio = Number.POSITIVE_INFINITY

  for (const group of availableGroups) {
    const ratio = groupRatio[group]
    if (
      typeof ratio === 'number' &&
      Number.isFinite(ratio) &&
      ratio < minRatio
    ) {
      minRatio = ratio
    }
  }

  return minRatio === Number.POSITIVE_INFINITY ? 1 : minRatio
}

/**
 * Replace model placeholder in endpoint path
 */
export function replaceModelInPath(path: string, modelName: string): string {
  return path.replaceAll('{model}', modelName)
}

/**
 * Check if model is token-based pricing
 */
export function isTokenBasedModel(model: PricingModel): boolean {
  return model.quota_type === QUOTA_TYPE_VALUES.TOKEN
}

export function isFreePricingModel(
  model: PricingModel,
  displayRatio: number
): boolean {
  if (displayRatio === 0) return true
  // Dynamic pricing can use zero placeholders for otherwise paid requests.
  if (
    model.billing_mode === 'tiered_expr' ||
    model.billing_expr?.trim() ||
    model.sora_per_request_pricing?.enabled
  )
    return false
  if (isTokenBasedModel(model)) {
    return (
      model.model_ratio === 0 &&
      Number.isFinite(model.completion_ratio) &&
      model.completion_ratio >= 0
    )
  }
  return (
    model.quota_type === QUOTA_TYPE_VALUES.REQUEST && model.model_price === 0
  )
}
