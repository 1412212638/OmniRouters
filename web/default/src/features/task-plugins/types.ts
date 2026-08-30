export type LocalizedText = Record<string, string>

export type TaskPlugin = {
  key: string
  name: string
  version: string
  icon?: string
  description?: LocalizedText
  author: { name: string; url?: string }
  channel_types?: number[]
  models?: string[]
  enabled: boolean
  runtime_enabled: boolean
}

export type TaskPluginsResponse = {
  success: boolean
  data: { runtime_enabled: boolean; plugins: TaskPlugin[] }
}
