import { SlidersHorizontalIcon } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { PromptInputButton } from '@/components/ai-elements/prompt-input'
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
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'

type ImageParameterSettingsProps = {
  count: string
  maxCount: number
  size: string
  quality: string
  disabled?: boolean
  onCountChange: (value: string) => void
  onSizeChange: (value: string) => void
  onQualityChange: (value: string) => void
}

export function ImageParameterSettings({
  count,
  maxCount,
  size,
  quality,
  disabled,
  onCountChange,
  onSizeChange,
  onQualityChange,
}: ImageParameterSettingsProps) {
  const { t } = useTranslation()

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
      <PopoverContent align='start' className='w-72 p-3'>
        <PopoverHeader>
          <PopoverTitle>{t('Parameters')}</PopoverTitle>
        </PopoverHeader>
        <div className='grid gap-3 pt-2'>
          <label className='grid gap-1.5 text-sm'>
            <span className='text-muted-foreground'>{t('Images')}</span>
            <Input
              aria-label={t('Images')}
              className='h-8'
              disabled={disabled || maxCount === 1}
              max={maxCount}
              min={1}
              onChange={(event) =>
                onCountChange(
                  String(
                    Math.min(
                      maxCount,
                      Math.max(1, Number(event.target.value) || 1)
                    )
                  )
                )
              }
              type='number'
              value={count}
            />
          </label>
          <label className='grid gap-1.5 text-sm'>
            <span className='text-muted-foreground'>{t('Size')}</span>
            <Select
              value={size}
              onValueChange={(value) => value && onSizeChange(value)}
              disabled={disabled}
            >
              <SelectTrigger aria-label={t('Size')} className='h-8 w-full'>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {[
                  '1024x1024',
                  '1536x1024',
                  '1024x1536',
                  '1792x1024',
                  '1024x1792',
                ].map((value) => (
                  <SelectItem key={value} value={value}>
                    {value}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </label>
          <label className='grid gap-1.5 text-sm'>
            <span className='text-muted-foreground'>{t('Quality')}</span>
            <Select
              value={quality}
              onValueChange={(value) => value && onQualityChange(value)}
              disabled={disabled}
            >
              <SelectTrigger aria-label={t('Quality')} className='h-8 w-full'>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {['auto', 'low', 'medium', 'high'].map((value) => (
                  <SelectItem key={value} value={value}>
                    {value}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </label>
        </div>
      </PopoverContent>
    </Popover>
  )
}
