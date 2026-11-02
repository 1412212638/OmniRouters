import type { TaskPluginOption } from '../../channels/api'
import type { BillingUsageSchema } from '../types'

// Match the backend's UsageForModel: a model profile replaces the default
// schema entirely, including an explicitly empty profile.
export function getTaskPluginUsageSchema(
  plugin: Pick<TaskPluginOption, 'usageSchema' | 'usageProfiles'>,
  model: string
): BillingUsageSchema | undefined {
  const folded = model.toLowerCase()
  const profile = plugin.usageProfiles?.find((candidate) =>
    candidate.models.some((name) => name.toLowerCase() === folded)
  )
  return profile ? profile.schema : plugin.usageSchema
}
