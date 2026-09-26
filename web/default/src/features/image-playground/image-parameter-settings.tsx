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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'

import type { ImageModelCapabilities, ImageParameterCapability } from './api'
import type {
  ImageParameterEnabled,
  ImageParameterValues,
} from './image-request-builder'

type ImageParameterSettingsProps = {
  capabilities: ImageModelCapabilities | null
  disabled?: boolean
  modelValue: string
  parameterEnabled: ImageParameterEnabled
  parameterValues: ImageParameterValues
  onParameterEnabledChange: (key: string, value: boolean) => void
  onParameterValueChange: (key: string, value: unknown) => void
}

function getParameterValue(
  parameter: ImageParameterCapability,
  value: unknown
) {
  if (value !== undefined && value !== null) return value
  return parameter.default ?? ''
}

export function ImageParameterSettings({
  capabilities,
  disabled,
  modelValue,
  parameterEnabled,
  parameterValues,
  onParameterEnabledChange,
  onParameterValueChange,
}: ImageParameterSettingsProps) {
  const { t } = useTranslation()
  const parameters = capabilities?.parameters ?? []

  const reset = () => {
    parameters.forEach((parameter) => {
      onParameterValueChange(parameter.key, parameter.default ?? '')
      onParameterEnabledChange(parameter.key, parameter.enabled_by_default)
    })
  }

  const renderValue = (parameter: ImageParameterCapability) => {
    const value = getParameterValue(parameter, parameterValues[parameter.key])
    const controlDisabled =
      disabled || parameterEnabled[parameter.key] === false

    if (parameter.type === 'boolean') {
      return (
        <Switch
          aria-label={parameter.key}
          checked={value === true}
          disabled={controlDisabled}
          onCheckedChange={(checked) =>
            onParameterValueChange(parameter.key, checked === true)
          }
          size='sm'
        />
      )
    }
    if (parameter.type === 'enum') {
      return (
        <Select
          value={String(value)}
          onValueChange={(next) =>
            next && onParameterValueChange(parameter.key, next)
          }
          disabled={controlDisabled}
        >
          <SelectTrigger
            aria-label={parameter.key}
            className='h-8 w-40 max-w-full'
          >
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {(parameter.options ?? []).map((option) => (
              <SelectItem key={option} value={option}>
                {option}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      )
    }

    return (
      <Input
        aria-label={parameter.key}
        className='h-8 w-40 max-w-full'
        disabled={controlDisabled}
        min={parameter.min}
        max={parameter.max}
        step={parameter.step}
        type={parameter.type === 'string' ? 'text' : 'number'}
        value={String(value)}
        onChange={(event) => {
          if (parameter.type === 'string') {
            onParameterValueChange(parameter.key, event.target.value)
            return
          }
          const next = Number(event.target.value)
          if (!Number.isFinite(next)) return
          const bounded = Math.min(
            parameter.max ?? Number.MAX_SAFE_INTEGER,
            Math.max(parameter.min ?? Number.MIN_SAFE_INTEGER, next)
          )
          onParameterValueChange(
            parameter.key,
            parameter.type === 'integer' ? Math.round(bounded) : bounded
          )
        }}
      />
    )
  }

  return (
    <Popover>
      <Tooltip>
        <TooltipTrigger
          render={
            <PopoverTrigger
              render={
                <PromptInputButton
                  aria-label={t('Parameters')}
                  className='text-muted-foreground hover:bg-muted/70 hover:text-foreground font-medium'
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
      <PopoverContent align='start' className='w-[min(92vw,460px)] p-3'>
        <PopoverHeader className='flex-row items-center justify-between gap-3'>
          <div className='min-w-0'>
            <PopoverTitle>{t('Parameters')}</PopoverTitle>
            <p className='text-muted-foreground truncate text-xs'>
              {modelValue}
            </p>
          </div>
          <Button
            disabled={disabled || parameters.length === 0}
            onClick={reset}
            size='sm'
            type='button'
            variant='ghost'
          >
            {t('Reset')}
          </Button>
        </PopoverHeader>
        {parameters.length === 0 ? (
          <p className='text-muted-foreground py-5 text-center text-sm'>
            {capabilities
              ? t('No configurable parameters')
              : t('Loading parameters')}
          </p>
        ) : (
          <div className='divide-border mt-1 divide-y'>
            {parameters.map((parameter) => {
              const enabled =
                parameterEnabled[parameter.key] ?? parameter.enabled_by_default
              return (
                <div
                  className='flex min-w-0 items-center justify-between gap-3 py-3'
                  key={parameter.key}
                >
                  <div className='min-w-0'>
                    <div className='text-foreground truncate text-sm font-medium'>
                      {parameter.key}
                    </div>
                    {parameter.min !== undefined &&
                      parameter.max !== undefined && (
                        <div className='text-muted-foreground text-xs'>
                          {parameter.min} - {parameter.max}
                        </div>
                      )}
                  </div>
                  <div className='flex min-w-0 items-center gap-2'>
                    {renderValue(parameter)}
                    <Switch
                      aria-label={t('Enable {{parameter}}', {
                        parameter: parameter.key,
                      })}
                      checked={enabled}
                      disabled={disabled}
                      onCheckedChange={(checked) =>
                        onParameterEnabledChange(
                          parameter.key,
                          checked === true
                        )
                      }
                      size='sm'
                    />
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </PopoverContent>
    </Popover>
  )
}
