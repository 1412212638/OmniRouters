import { BracesIcon } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { PromptInputButton } from '@/components/ai-elements/prompt-input'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
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
import { Textarea } from '@/components/ui/textarea'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'

import type {
  ImageRequestMode,
  ImageRequestProfile,
  ImageReferenceField,
} from './api'

type Props = {
  profile: ImageRequestProfile
  disabled?: boolean
  onChange: (profile: ImageRequestProfile) => void
}
const modes: ImageRequestMode[] = ['auto', 'generations', 'edits', 'custom']

function modeLabel(mode: ImageRequestMode, t: (key: string) => string) {
  if (mode === 'auto') return t('Auto')
  if (mode === 'generations') return t('Generations')
  if (mode === 'edits') return t('Edits')
  return t('Custom')
}

export function ImageRequestSettings({ profile, disabled, onChange }: Props) {
  const { t } = useTranslation()
  const update = (patch: Partial<ImageRequestProfile>) =>
    onChange({ ...profile, ...patch })

  return (
    <Popover>
      <Tooltip>
        <TooltipTrigger
          render={
            <PopoverTrigger
              render={
                <PromptInputButton
                  aria-label={t('Request settings')}
                  className='text-muted-foreground hover:bg-muted/70 hover:text-foreground font-medium'
                  disabled={disabled}
                  variant='ghost'
                >
                  <BracesIcon size={16} />
                </PromptInputButton>
              }
            />
          }
        />
        <TooltipContent>
          <p>{t('Request settings')}</p>
        </TooltipContent>
      </Tooltip>
      <PopoverContent align='start' className='w-[min(92vw,480px)] p-4'>
        <PopoverHeader>
          <PopoverTitle>{t('Request settings')}</PopoverTitle>
        </PopoverHeader>
        <div className='flex flex-col gap-4'>
          <div className='flex flex-col gap-2'>
            <Label>{t('Request mode')}</Label>
            <Select
              value={profile.mode}
              onValueChange={(value) =>
                value && update({ mode: value as ImageRequestMode })
              }
              disabled={disabled}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {modes.map((mode) => (
                  <SelectItem key={mode} value={mode}>
                    {modeLabel(mode, t)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {profile.mode === 'custom' && (
            <div className='flex flex-col gap-2'>
              <Label>{t('Endpoint')}</Label>
              <Select
                value={profile.endpoint}
                onValueChange={(value) =>
                  value &&
                  update({ endpoint: value as ImageRequestProfile['endpoint'] })
                }
                disabled={disabled}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value='generations'>
                    /v1/images/generations
                  </SelectItem>
                  <SelectItem value='edits'>/v1/images/edits</SelectItem>
                  <SelectItem value='gemini'>
                    /v1beta/models/&lt;model&gt;:generateContent
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
          )}

          <div className='flex flex-col gap-2'>
            <Label>{t('Reference image field')}</Label>
            <Select
              value={profile.referenceField}
              onValueChange={(value) =>
                value &&
                update({ referenceField: value as ImageReferenceField })
              }
              disabled={disabled}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value='none'>{t('Do not send')}</SelectItem>
                <SelectItem value='image'>image</SelectItem>
                <SelectItem value='image_url'>image_url</SelectItem>
                <SelectItem value='input.image'>input.image</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className='flex flex-col gap-2'>
            <Label htmlFor='image-custom-parameters'>
              {t('Custom JSON parameters')}
            </Label>
            <Textarea
              id='image-custom-parameters'
              aria-label={t('Custom JSON parameters')}
              className='min-h-28 font-mono text-xs'
              disabled={disabled}
              onChange={(event) =>
                update({ customParameters: event.target.value })
              }
              placeholder='{\n  "seed": 42\n}'
              spellCheck={false}
              value={profile.customParameters}
            />
            <p className='text-muted-foreground text-xs'>
              {t(
                'Only JSON object fields are accepted. Model, prompt, group, n, and reference fields are controlled by Studio.'
              )}
            </p>
          </div>

          <Button
            className='self-end'
            onClick={() =>
              onChange({
                mode: 'auto',
                endpoint: 'generations',
                referenceField: 'image',
                customParameters: '',
              })
            }
            size='sm'
            type='button'
            variant='ghost'
          >
            {t('Reset')}
          </Button>
        </div>
      </PopoverContent>
    </Popover>
  )
}
