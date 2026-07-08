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
import { SlidersHorizontalIcon } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { PromptInputButton } from '@/components/ai-elements/prompt-input'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Popover,
  PopoverContent,
  PopoverHeader,
  PopoverTitle,
  PopoverTrigger,
} from '@/components/ui/popover'
import { Slider } from '@/components/ui/slider'
import { Switch } from '@/components/ui/switch'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'

import { DEFAULT_CONFIG } from '../../constants'
import { getDefaultParameterEnabledForModel } from '../../lib'
import type {
  ParameterEnabled,
  PlaygroundConfig,
  PlaygroundParameterKey,
} from '../../types'

type PlaygroundParameterSettingsProps = {
  config: PlaygroundConfig
  disabled?: boolean
  modelValue: string
  parameterEnabled: ParameterEnabled
  onConfigChange: <K extends keyof PlaygroundConfig>(
    key: K,
    value: PlaygroundConfig[K]
  ) => void
  onParameterEnabledChange: (
    key: PlaygroundParameterKey,
    value: boolean
  ) => void
}

type RangeParameter = {
  key: Exclude<PlaygroundParameterKey, 'seed' | 'max_tokens'>
  min: number
  max: number
  step: number
  decimals: number
}

const RANGE_PARAMETERS: RangeParameter[] = [
  {
    key: 'temperature',
    min: 0,
    max: 2,
    step: 0.1,
    decimals: 1,
  },
  {
    key: 'top_p',
    min: 0,
    max: 1,
    step: 0.05,
    decimals: 2,
  },
  {
    key: 'frequency_penalty',
    min: -2,
    max: 2,
    step: 0.1,
    decimals: 1,
  },
  {
    key: 'presence_penalty',
    min: -2,
    max: 2,
    step: 0.1,
    decimals: 1,
  },
]

const PARAMETER_KEYS: PlaygroundParameterKey[] = [
  'temperature',
  'top_p',
  'max_tokens',
  'frequency_penalty',
  'presence_penalty',
  'seed',
]

function clampNumber(value: number, min: number, max: number) {
  if (!Number.isFinite(value)) return min
  return Math.min(max, Math.max(min, value))
}

function roundNumber(value: number, decimals: number) {
  const multiplier = 10 ** decimals
  return Math.round(value * multiplier) / multiplier
}

function getSliderValue(value: number | readonly number[]) {
  return Array.isArray(value) ? value[0] : value
}

export function PlaygroundParameterSettings({
  config,
  disabled,
  modelValue,
  parameterEnabled,
  onConfigChange,
  onParameterEnabledChange,
}: PlaygroundParameterSettingsProps) {
  const { t } = useTranslation()

  const handleReset = () => {
    onConfigChange('temperature', DEFAULT_CONFIG.temperature)
    onConfigChange('top_p', DEFAULT_CONFIG.top_p)
    onConfigChange('max_tokens', DEFAULT_CONFIG.max_tokens)
    onConfigChange('frequency_penalty', DEFAULT_CONFIG.frequency_penalty)
    onConfigChange('presence_penalty', DEFAULT_CONFIG.presence_penalty)
    onConfigChange('seed', DEFAULT_CONFIG.seed)

    const defaultEnabled = getDefaultParameterEnabledForModel(modelValue)
    PARAMETER_KEYS.forEach((key) => {
      onParameterEnabledChange(key, defaultEnabled[key])
    })
  }

  const renderParameterSwitch = (key: PlaygroundParameterKey) => (
    <Switch
      aria-label={key}
      checked={parameterEnabled[key]}
      disabled={disabled}
      onCheckedChange={(checked) => {
        onParameterEnabledChange(key, checked === true)
      }}
      size='sm'
    />
  )

  return (
    <Popover>
      <Tooltip>
        <TooltipTrigger
          render={
            <PopoverTrigger
              render={
                <PromptInputButton
                  aria-label={t('Parameters')}
                  className='text-muted-foreground hover:text-foreground hover:bg-muted/70 font-medium'
                  disabled={disabled}
                  variant='ghost'
                />
              }
            >
              <SlidersHorizontalIcon size={16} />
            </PopoverTrigger>
          }
        />
        <TooltipContent>
          <p>{t('Parameters')}</p>
        </TooltipContent>
      </Tooltip>

      <PopoverContent align='start' className='w-[min(92vw,420px)] p-3'>
        <PopoverHeader className='flex-row items-center justify-between gap-3'>
          <div className='min-w-0'>
            <PopoverTitle>{t('Parameters')}</PopoverTitle>
            <p className='text-muted-foreground truncate text-xs'>
              {modelValue}
            </p>
          </div>
          <Button
            disabled={disabled}
            onClick={handleReset}
            size='sm'
            type='button'
            variant='ghost'
          >
            {t('Reset')}
          </Button>
        </PopoverHeader>

        <div className='divide-border mt-1 divide-y'>
          {RANGE_PARAMETERS.map((item) => {
            const value = config[item.key]
            const enabled = parameterEnabled[item.key]

            return (
              <div className='grid gap-2 py-3' key={item.key}>
                <div className='flex items-center justify-between gap-3'>
                  <div className='min-w-0'>
                    <div className='text-foreground truncate text-sm font-medium'>
                      {item.key}
                    </div>
                    <div className='text-muted-foreground text-xs'>
                      {item.min} - {item.max}
                    </div>
                  </div>
                  <div className='flex items-center gap-2'>
                    <Input
                      aria-label={item.key}
                      className='h-7 w-20 text-right tabular-nums'
                      disabled={disabled || !enabled}
                      max={item.max}
                      min={item.min}
                      onChange={(event) => {
                        const nextValue = roundNumber(
                          clampNumber(
                            Number(event.target.value),
                            item.min,
                            item.max
                          ),
                          item.decimals
                        )
                        onConfigChange(item.key, nextValue)
                      }}
                      step={item.step}
                      type='number'
                      value={value}
                    />
                    {renderParameterSwitch(item.key)}
                  </div>
                </div>
                <Slider
                  disabled={disabled || !enabled}
                  max={item.max}
                  min={item.min}
                  onValueChange={(nextValue) => {
                    const next = roundNumber(
                      clampNumber(
                        getSliderValue(nextValue),
                        item.min,
                        item.max
                      ),
                      item.decimals
                    )
                    onConfigChange(item.key, next)
                  }}
                  step={item.step}
                  value={[value]}
                />
              </div>
            )
          })}

          <div className='grid gap-2 py-3'>
            <div className='flex items-center justify-between gap-3'>
              <div className='min-w-0'>
                <div className='text-foreground truncate text-sm font-medium'>
                  max_tokens
                </div>
                <div className='text-muted-foreground text-xs'>1 - 200000</div>
              </div>
              <div className='flex items-center gap-2'>
                <Input
                  aria-label='max_tokens'
                  className='h-7 w-24 text-right tabular-nums'
                  disabled={disabled || !parameterEnabled.max_tokens}
                  max={200000}
                  min={1}
                  onChange={(event) => {
                    onConfigChange(
                      'max_tokens',
                      Math.trunc(
                        clampNumber(Number(event.target.value), 1, 200000)
                      )
                    )
                  }}
                  step={1}
                  type='number'
                  value={config.max_tokens}
                />
                {renderParameterSwitch('max_tokens')}
              </div>
            </div>
          </div>

          <div className='grid gap-2 py-3'>
            <div className='flex items-center justify-between gap-3'>
              <div className='min-w-0'>
                <div className='text-foreground truncate text-sm font-medium'>
                  seed
                </div>
                <div className='text-muted-foreground text-xs'>
                  {t('Deterministic sampling seed (best-effort)')}
                </div>
              </div>
              <div className='flex items-center gap-2'>
                <Input
                  aria-label='seed'
                  className={cn(
                    'h-7 w-24 text-right tabular-nums',
                    config.seed === null && 'text-muted-foreground'
                  )}
                  disabled={disabled || !parameterEnabled.seed}
                  min={0}
                  onChange={(event) => {
                    const rawValue = event.target.value.trim()
                    onConfigChange(
                      'seed',
                      rawValue === ''
                        ? null
                        : Math.max(0, Math.trunc(Number(rawValue) || 0))
                    )
                  }}
                  placeholder='auto'
                  step={1}
                  type='number'
                  value={config.seed ?? ''}
                />
                {renderParameterSwitch('seed')}
              </div>
            </div>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  )
}
