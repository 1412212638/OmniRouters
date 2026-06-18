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
import { useTranslation } from 'react-i18next'
import {
  ZENMUX_MODALITY_ICONS,
  type ZenMuxIconComponent,
} from '@/assets/custom/zenmux-modality-icons'
import { cn } from '@/lib/utils'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { StaticDataTable } from '@/components/data-table'
import type { Modality } from '../types'

const MODALITY_META: Record<
  Modality,
  { icon: ZenMuxIconComponent; labelKey: string }
> = {
  text: { icon: ZENMUX_MODALITY_ICONS.text, labelKey: 'Text' },
  image: { icon: ZENMUX_MODALITY_ICONS.image, labelKey: 'Image' },
  audio: { icon: ZENMUX_MODALITY_ICONS.audio, labelKey: 'Audio' },
  video: { icon: ZENMUX_MODALITY_ICONS.video, labelKey: 'Video' },
  file: { icon: ZENMUX_MODALITY_ICONS.file, labelKey: 'File' },
  embedding: {
    icon: ZENMUX_MODALITY_ICONS.embedding,
    labelKey: 'Embedding',
  },
  rerank: { icon: ZENMUX_MODALITY_ICONS.rerank, labelKey: 'Rerank' },
  speech: { icon: ZENMUX_MODALITY_ICONS.speech, labelKey: 'Speech' },
  transcription: {
    icon: ZENMUX_MODALITY_ICONS.transcription,
    labelKey: 'Transcription',
  },
}

const ALL_MODALITIES: Modality[] = [
  'text',
  'image',
  'audio',
  'video',
  'file',
  'embedding',
  'rerank',
  'speech',
  'transcription',
]

/** Inline modality icons (used by the quick-stats flow). */
export function ModalityIcons(props: {
  modalities: Modality[]
  className?: string
}) {
  const { t } = useTranslation()
  if (props.modalities.length === 0) {
    return <span className='text-muted-foreground text-xs'>—</span>
  }
  return (
    <span className='inline-flex items-center gap-1'>
      {props.modalities.map((modality) => {
        const meta = MODALITY_META[modality]
        const Icon = meta.icon
        return (
          <Tooltip key={modality}>
            <TooltipTrigger
              render={
                <span
                  aria-label={t(meta.labelKey)}
                  className='text-foreground/80 inline-flex'
                />
              }
            >
              <Icon className={cn('size-3.5', props.className)} />
            </TooltipTrigger>
            <TooltipContent side='top' className='text-xs'>
              {t(meta.labelKey)}
            </TooltipContent>
          </Tooltip>
        )
      })}
    </span>
  )
}

/**
 * 2 × N matrix showing which modalities are supported as input vs output.
 * Cells with a checkmark indicate support; empty cells show a dash.
 */
export function ModalitiesMatrix(props: {
  input: Modality[]
  output: Modality[]
}) {
  const { t } = useTranslation()
  const inputSet = new Set(props.input)
  const outputSet = new Set(props.output)

  return (
    <StaticDataTable
      className='rounded-lg'
      tableClassName='text-sm'
      headerRowClassName='bg-muted/40'
      data={[
        { label: t('Input'), set: inputSet },
        { label: t('Output'), set: outputSet },
      ]}
      getRowKey={(row) => row.label}
      columns={[
        {
          id: 'modality',
          header: t('Modality'),
          className:
            'text-muted-foreground px-3 py-2 text-left text-[11px] font-medium tracking-wider uppercase',
          cellClassName:
            'text-muted-foreground bg-muted/30 px-3 py-2 text-left text-[11px] font-medium tracking-wider uppercase',
          cell: (row) => row.label,
        },
        ...ALL_MODALITIES.map((modality) => ({
          id: modality,
          header: t(MODALITY_META[modality].labelKey),
          className:
            'text-muted-foreground border-l px-3 py-2 text-center text-[11px] font-medium tracking-wider uppercase',
          cellClassName: (row: { label: string; set: Set<Modality> }) =>
            cn(
              'border-l px-3 py-2 text-center',
              row.set.has(modality)
                ? 'bg-emerald-50/40 dark:bg-emerald-500/10'
                : 'bg-background'
            ),
          cell: (row: { label: string; set: Set<Modality> }) => {
            const enabled = row.set.has(modality)
            const Icon = MODALITY_META[modality].icon
            return (
              <span
                className={cn(
                  'inline-flex items-center justify-center',
                  enabled
                    ? 'text-emerald-700 dark:text-emerald-300'
                    : 'text-muted-foreground/40'
                )}
                aria-label={
                  enabled
                    ? t('{{modality}} supported', {
                        modality: t(MODALITY_META[modality].labelKey),
                      })
                    : t('{{modality}} not supported', {
                        modality: t(MODALITY_META[modality].labelKey),
                      })
                }
              >
                <Icon className='size-4' />
              </span>
            )
          },
        })),
      ]}
    />
  )
}
