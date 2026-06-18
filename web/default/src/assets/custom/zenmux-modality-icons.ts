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
import { type ComponentType, type SVGProps } from 'react'
import {
  ZenMuxAudioIcon,
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
} satisfies Record<string, ZenMuxIconComponent>

export function getZenMuxModalityIcon(
  modality: string | undefined
): ZenMuxIconComponent | undefined {
  if (!modality) return undefined
  return ZENMUX_MODALITY_ICONS[modality as keyof typeof ZENMUX_MODALITY_ICONS]
}
