/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
*/
import type { TaskPluginOption } from '../api'
import {
  CHANNEL_TYPE_NEW_API,
  CHANNEL_TYPE_OPTIONS,
  CHANNEL_TYPE_TASK_PLUGIN,
} from '../constants'

// Legacy task-only channels are represented by their built-in adapters and
// cannot be extended through a New API channel binding.
const LEGACY_TASK_PLUGIN_KEYS: Readonly<Partial<Record<number, string>>> = {
  36: 'sunoapi',
  50: 'kling',
  51: 'jimeng',
  52: 'vidu',
  54: 'doubao',
  55: 'sora',
}

export function supportsChannelPluginExtensions(channelType: number): boolean {
  return (
    channelType !== CHANNEL_TYPE_TASK_PLUGIN &&
    !LEGACY_TASK_PLUGIN_KEYS[channelType] &&
    CHANNEL_TYPE_OPTIONS.some((option) => option.value === channelType)
  )
}

export function readTaskExtendPluginKeys(
  channelType: number,
  setting: {
    task_plugin_key?: unknown
    task_extend_plugin_keys?: unknown
  } | null | undefined
): string[] {
  if (channelType !== CHANNEL_TYPE_NEW_API || !setting) return []
  const extensions = Array.isArray(setting.task_extend_plugin_keys)
    ? setting.task_extend_plugin_keys.filter(
        (key): key is string => typeof key === 'string' && key.trim() !== ''
      )
    : []
  const single =
    typeof setting.task_plugin_key === 'string'
      ? setting.task_plugin_key.trim()
      : ''
  return [...new Set(single ? [single, ...extensions] : extensions)]
}

export function supportsNewAPIUpstream(
  plugin: Pick<TaskPluginOption, 'upstreams'>
): boolean {
  return plugin.upstreams?.includes('new_api') ?? false
}
