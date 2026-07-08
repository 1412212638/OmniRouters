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
import { useCallback, useEffect, useRef, useState } from 'react'

import { DEFAULT_CONFIG, DEFAULT_PARAMETER_ENABLED } from '../constants'
import {
  saveConfig,
  loadModelParameterSettings,
  saveParameterEnabled,
  saveMessages,
  saveModelParameterSettings,
  applyMessageStateUpdate,
  getInitialParameterEnabled,
  getInitialPlaygroundConfig,
  isClaudeLikePlaygroundModel,
  isPlaygroundParameterKey,
  pickParameterConfig,
  resolveParameterEnabledForModel,
  loadMessages,
  type MessageStateUpdater,
} from '../lib'
import type {
  Message,
  PlaygroundConfig,
  ParameterEnabled,
  ModelOption,
  GroupOption,
} from '../types'

const MESSAGE_SAVE_DEBOUNCE_MS = 500

function getInitialState() {
  const config = getInitialPlaygroundConfig()
  const enabled = getInitialParameterEnabled()
  const modelParameterSettings = loadModelParameterSettings()
  const modelKey = config.model.trim()
  const savedForModel = modelParameterSettings[modelKey]
  const parameterEnabled = savedForModel
    ? resolveParameterEnabledForModel(modelKey, savedForModel)
    : {
        ...enabled,
        temperature: isClaudeLikePlaygroundModel(modelKey)
          ? false
          : enabled.temperature,
      }

  return {
    config: {
      ...config,
      ...savedForModel?.config,
    },
    parameterEnabled,
  }
}

function saveModelParameterSnapshot(
  model: string,
  config: PlaygroundConfig,
  parameterEnabled: ParameterEnabled
) {
  const normalizedModel = model.trim()
  if (!normalizedModel) return

  const settings = loadModelParameterSettings()
  settings[normalizedModel] = {
    config: pickParameterConfig(config),
    parameterEnabled,
  }
  saveModelParameterSettings(settings)
}

/**
 * Main state management hook for playground
 */
export function usePlaygroundState() {
  // Load initial state from localStorage
  const initialStateRef = useRef<ReturnType<typeof getInitialState> | null>(
    null
  )
  if (initialStateRef.current === null) {
    initialStateRef.current = getInitialState()
  }

  const [config, setConfig] = useState<PlaygroundConfig>(
    initialStateRef.current.config
  )

  const [parameterEnabled, setParameterEnabled] = useState<ParameterEnabled>(
    initialStateRef.current.parameterEnabled
  )

  const [messages, setMessages] = useState<Message[]>([])
  const [isLoadingMessages, setIsLoadingMessages] = useState(true)
  const messagesSaveTimerRef = useRef<number | null>(null)
  const latestMessagesRef = useRef<Message[]>(messages)
  const latestConfigRef = useRef<PlaygroundConfig>(config)
  const latestParameterEnabledRef = useRef<ParameterEnabled>(parameterEnabled)
  const hasLoadedMessagesRef = useRef(false)

  const [models, setModels] = useState<ModelOption[]>([])
  const [groups, setGroups] = useState<GroupOption[]>([])

  useEffect(() => {
    latestConfigRef.current = config
  }, [config])

  useEffect(() => {
    latestParameterEnabledRef.current = parameterEnabled
  }, [parameterEnabled])

  const persistMessages = useCallback((messagesToSave: Message[]) => {
    latestMessagesRef.current = messagesToSave

    if (!hasLoadedMessagesRef.current) {
      return
    }

    if (messagesSaveTimerRef.current !== null) {
      window.clearTimeout(messagesSaveTimerRef.current)
    }

    messagesSaveTimerRef.current = window.setTimeout(() => {
      messagesSaveTimerRef.current = null
      saveMessages(latestMessagesRef.current)
    }, MESSAGE_SAVE_DEBOUNCE_MS)
  }, [])

  useEffect(() => {
    let cancelled = false

    window.setTimeout(() => {
      const loadedMessages = loadMessages() ?? []
      if (cancelled) {
        return
      }

      latestMessagesRef.current = loadedMessages
      hasLoadedMessagesRef.current = true
      setMessages(loadedMessages)
      setIsLoadingMessages(false)
    }, 0)

    return () => {
      cancelled = true
    }
  }, [])

  useEffect(
    () => () => {
      if (messagesSaveTimerRef.current !== null) {
        window.clearTimeout(messagesSaveTimerRef.current)
        saveMessages(latestMessagesRef.current)
      }
    },
    []
  )

  // Update config with automatic save
  const updateConfig = useCallback(
    <K extends keyof PlaygroundConfig>(key: K, value: PlaygroundConfig[K]) => {
      if (key === 'model') {
        const nextModel = String(value).trim()
        saveModelParameterSnapshot(
          latestConfigRef.current.model,
          latestConfigRef.current,
          latestParameterEnabledRef.current
        )

        const settings = loadModelParameterSettings()
        const savedForModel = settings[nextModel]
        const nextParameterEnabled = resolveParameterEnabledForModel(
          nextModel,
          savedForModel
        )

        setParameterEnabled(nextParameterEnabled)
        saveParameterEnabled(nextParameterEnabled)

        setConfig((prev) => {
          const updated = {
            ...prev,
            model: nextModel,
            ...savedForModel?.config,
          }
          saveConfig(updated)
          latestConfigRef.current = updated
          latestParameterEnabledRef.current = nextParameterEnabled
          return updated
        })
        return
      }

      setConfig((prev) => {
        const updated = { ...prev, [key]: value }
        saveConfig(updated)
        latestConfigRef.current = updated

        if (isPlaygroundParameterKey(key)) {
          saveModelParameterSnapshot(
            updated.model,
            updated,
            latestParameterEnabledRef.current
          )
        }

        return updated
      })
    },
    []
  )

  // Update parameter enabled with automatic save
  const updateParameterEnabled = useCallback(
    (key: keyof ParameterEnabled, value: boolean) => {
      setParameterEnabled((prev) => {
        const updated = { ...prev, [key]: value }
        saveParameterEnabled(updated)
        latestParameterEnabledRef.current = updated
        saveModelParameterSnapshot(
          latestConfigRef.current.model,
          latestConfigRef.current,
          updated
        )
        return updated
      })
    },
    []
  )

  // Update messages with automatic save
  const updateMessages = useCallback(
    (updater: MessageStateUpdater) => {
      setMessages((prev) => {
        const newMessages = applyMessageStateUpdate(prev, updater)
        persistMessages(newMessages)
        return newMessages
      })
    },
    [persistMessages]
  )

  // Clear all messages
  const clearMessages = useCallback(() => {
    updateMessages([])
  }, [updateMessages])

  // Reset config to defaults
  const resetConfig = useCallback(() => {
    setConfig(DEFAULT_CONFIG)
    setParameterEnabled(DEFAULT_PARAMETER_ENABLED)
    saveConfig(DEFAULT_CONFIG)
    saveParameterEnabled(DEFAULT_PARAMETER_ENABLED)
  }, [])

  return {
    // State
    config,
    parameterEnabled,
    messages,
    isLoadingMessages,
    models,
    groups,

    // Setters
    setModels,
    setGroups,

    // Actions
    updateConfig,
    updateParameterEnabled,
    updateMessages,
    clearMessages,
    resetConfig,
  }
}
