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
import { DEFAULT_CONFIG, DEFAULT_PARAMETER_ENABLED } from '../../constants'
import type {
  Message,
  ParameterEnabled,
  PlaygroundConfig,
  PlaygroundModelParameterSettings,
  PlaygroundParameterConfig,
  PlaygroundParameterKey,
} from '../../types'
import {
  loadConfig,
  loadMessages,
  loadParameterEnabled,
} from '../storage/storage'

export type MessageStateUpdater =
  | Message[]
  | ((previousMessages: Message[]) => Message[])

export function getInitialPlaygroundConfig(): PlaygroundConfig {
  return { ...DEFAULT_CONFIG, ...loadConfig() }
}

export function getInitialParameterEnabled(): ParameterEnabled {
  return { ...DEFAULT_PARAMETER_ENABLED, ...loadParameterEnabled() }
}

export const PLAYGROUND_PARAMETER_KEYS: PlaygroundParameterKey[] = [
  'temperature',
  'top_p',
  'max_tokens',
  'frequency_penalty',
  'presence_penalty',
  'seed',
]

export function isClaudeLikePlaygroundModel(model: string): boolean {
  const normalized = model.trim().toLowerCase()
  return normalized.includes('claude') || normalized.includes('anthropic')
}

export function getDefaultParameterEnabledForModel(
  model: string
): ParameterEnabled {
  const defaults = { ...DEFAULT_PARAMETER_ENABLED }

  if (isClaudeLikePlaygroundModel(model)) {
    defaults.temperature = false
  }

  return defaults
}

export function resolveParameterEnabledForModel(
  model: string,
  saved?: PlaygroundModelParameterSettings
): ParameterEnabled {
  return {
    ...getDefaultParameterEnabledForModel(model),
    ...saved?.parameterEnabled,
  }
}

export function pickParameterConfig(
  config: PlaygroundConfig
): PlaygroundParameterConfig {
  return {
    temperature: config.temperature,
    top_p: config.top_p,
    max_tokens: config.max_tokens,
    frequency_penalty: config.frequency_penalty,
    presence_penalty: config.presence_penalty,
    seed: config.seed,
  }
}

export function isPlaygroundParameterKey(
  key: keyof PlaygroundConfig
): key is PlaygroundParameterKey {
  return PLAYGROUND_PARAMETER_KEYS.includes(key as PlaygroundParameterKey)
}

export function getInitialMessages(): Message[] {
  return loadMessages() || []
}

export function applyMessageStateUpdate(
  previousMessages: Message[],
  updater: MessageStateUpdater
): Message[] {
  return typeof updater === 'function' ? updater(previousMessages) : updater
}
