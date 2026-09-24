import { useTranslation } from 'react-i18next'

import { ZENMUX_MODALITY_ICONS } from '@/assets/custom/zenmux-modality-icons'
import { StatusBadge } from '@/components/status-badge'

import { MODEL_MODALITIES, type Model, type ModelModality } from '../types'

const MODEL_MODALITY_LABELS = {
  text: 'Text',
  image: 'Image',
  file: 'File',
  audio: 'Audio',
  video: 'Video',
  embedding: 'Embedding',
  rerank: 'Rerank',
  speech: 'Speech',
  transcription: 'Transcription',
  decisions: 'Decisions',
} satisfies Record<ModelModality, string>

function normalizeModalities(
  modalities: ModelModality[] | undefined
): ModelModality[] {
  if (!Array.isArray(modalities)) return []
  return MODEL_MODALITIES.filter((modality) => modalities.includes(modality))
}

function ModelModalityBadges(props: {
  modalities: ModelModality[] | undefined
}) {
  const { t } = useTranslation()
  const modalities = normalizeModalities(props.modalities)

  if (modalities.length === 0) {
    return <span className='text-muted-foreground text-xs'>-</span>
  }

  return (
    <div className='flex min-w-0 flex-wrap items-center gap-1'>
      {modalities.map((modality) => {
        const Icon = ZENMUX_MODALITY_ICONS[modality]
        return (
          <StatusBadge
            key={modality}
            autoColor={modality}
            copyable={false}
            size='sm'
          >
            <Icon className='size-3.5 shrink-0' />
            <span className='truncate'>
              {t(MODEL_MODALITY_LABELS[modality])}
            </span>
          </StatusBadge>
        )
      })}
    </div>
  )
}

export function ModelModalitiesCell(props: { model: Model }) {
  const { t } = useTranslation()

  return (
    <div className='flex min-w-0 flex-col gap-1.5'>
      <div className='grid min-w-0 grid-cols-[3.25rem_minmax(0,1fr)] items-center gap-1.5'>
        <span className='text-muted-foreground text-xs font-medium'>
          {t('Input')}
        </span>
        <ModelModalityBadges modalities={props.model.input_modalities} />
      </div>
      <div className='grid min-w-0 grid-cols-[3.25rem_minmax(0,1fr)] items-center gap-1.5'>
        <span className='text-muted-foreground text-xs font-medium'>
          {t('Output')}
        </span>
        <ModelModalityBadges modalities={props.model.output_modalities} />
      </div>
    </div>
  )
}
