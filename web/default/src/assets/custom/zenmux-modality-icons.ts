import { type ComponentType, type SVGProps } from 'react'

import {
  ZenMuxAudioIcon,
  ZenMuxDecisionsIcon,
  ZenMuxEmbeddingIcon,
  ZenMuxFileIcon,
  ZenMuxImageIcon,
  ZenMuxRerankIcon,
  ZenMuxSpeechIcon,
  ZenMuxTextIcon,
  ZenMuxTranscriptionIcon,
  ZenMuxVideoIcon,
} from './zenmux-icons'

export type ZenMuxIconComponent = ComponentType<SVGProps<SVGSVGElement>>

export const ZENMUX_MODALITY_ICONS = {
  text: ZenMuxTextIcon,
  image: ZenMuxImageIcon,
  file: ZenMuxFileIcon,
  audio: ZenMuxAudioIcon,
  video: ZenMuxVideoIcon,
  embedding: ZenMuxEmbeddingIcon,
  rerank: ZenMuxRerankIcon,
  speech: ZenMuxSpeechIcon,
  transcription: ZenMuxTranscriptionIcon,
  decisions: ZenMuxDecisionsIcon,
} satisfies Record<string, ZenMuxIconComponent>

export function getZenMuxModalityIcon(
  modality: string | undefined
): ZenMuxIconComponent | undefined {
  if (!modality) return undefined
  return ZENMUX_MODALITY_ICONS[modality as keyof typeof ZENMUX_MODALITY_ICONS]
}
