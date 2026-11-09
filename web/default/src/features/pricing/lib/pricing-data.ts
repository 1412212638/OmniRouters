import type { PricingModel } from '../types'

export function getPricingQueryKey(
  userId: number | undefined,
  organizationId: number | null
) {
  return [
    'pricing',
    userId ?? 'anonymous',
    organizationId ?? 'personal',
  ] as const
}

export function resolveModelGroupRatios(
  model: PricingModel,
  groupRatios: Record<string, number>,
  groupModelRatios: Record<string, Record<string, number>>
): Record<string, number> {
  const resolved = Object.fromEntries(
    Object.entries(groupRatios).map(([group, ratio]) => [
      group,
      ratio * (groupModelRatios[group]?.[model.model_name] ?? 1),
    ])
  )

  // The pricing API provides the final per-model multiplier after applying
  // group and user-specific overrides. Prefer it to avoid multiplying an
  // override into the generic group ratio a second time.
  for (const [group, ratio] of Object.entries(model.model_group_ratio ?? {})) {
    if (Number.isFinite(ratio)) resolved[group] = ratio
  }

  return resolved
}
