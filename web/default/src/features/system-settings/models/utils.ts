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
import type {
  SoraPerRequestPricing,
  SoraResolutionTier,
} from '@/features/pricing/types'

export function formatJsonForTextarea(value: string) {
  if (!value || !value.trim()) {
    return ''
  }

  try {
    const parsed = JSON.parse(value)
    return JSON.stringify(parsed, null, 2)
  } catch {
    return value
  }
}

export function normalizeJsonString(value: string) {
  const trimmed = value.trim()
  if (!trimmed) {
    return ''
  }

  try {
    const parsed = JSON.parse(trimmed)
    return JSON.stringify(parsed)
  } catch {
    return trimmed
  }
}

export type SoraResolutionTierDraft = {
  value: string
  multiplier: string
}

export function cloneSoraResolutionTiers(
  tiers?: Array<
    | Partial<SoraResolutionTier>
    | Partial<SoraResolutionTierDraft>
    | null
    | undefined
  > | null
): SoraResolutionTierDraft[] {
  if (!Array.isArray(tiers)) return []

  return tiers.map((tier) => ({
    value: String(tier?.value ?? ''),
    multiplier: String(tier?.multiplier ?? ''),
  }))
}

export function normalizeSoraResolutionTiers(
  tiers: SoraResolutionTierDraft[]
): SoraResolutionTierDraft[] {
  return cloneSoraResolutionTiers(tiers)
    .map((tier) => ({
      value: tier.value.trim(),
      multiplier: tier.multiplier.trim(),
    }))
    .filter((tier) => tier.value || tier.multiplier)
}

export function serializeSoraPerRequestPricing(
  enabled: boolean,
  tiers: SoraResolutionTierDraft[],
  audioGenerationSurcharge?: string
): SoraPerRequestPricing | null {
  const normalized = normalizeSoraResolutionTiers(tiers)
  const parsedAudioGenerationSurcharge = parseOptionalSoraSurcharge(
    audioGenerationSurcharge,
    'Sora audio generation surcharge must be greater than or equal to 0'
  )
  if (
    !enabled &&
    normalized.length === 0 &&
    parsedAudioGenerationSurcharge === undefined
  ) {
    return null
  }

  const tierNames = normalized.map((tier) => tier.value)
  if (new Set(tierNames).size !== tierNames.length) {
    throw new Error('Sora resolution tiers must be unique')
  }

  const parsedTiers = normalized.map((tier) => {
    const multiplier = Number(tier.multiplier)
    if (!Number.isFinite(multiplier) || multiplier <= 0) {
      throw new Error('Sora tier multiplier must be greater than 0')
    }
    return {
      value: tier.value,
      multiplier,
    }
  })

  return {
    enabled,
    resolution_tiers: parsedTiers,
    ...(parsedAudioGenerationSurcharge !== undefined
      ? { audio_generation_surcharge: parsedAudioGenerationSurcharge }
      : {}),
  }
}

function parseOptionalSoraSurcharge(
  value: string | undefined,
  errorMessage: string
): number | undefined {
  const trimmed = String(value ?? '').trim()
  if (!trimmed) return undefined

  const surcharge = Number(trimmed)
  if (!Number.isFinite(surcharge) || surcharge < 0) {
    throw new Error(errorMessage)
  }
  return surcharge
}

type JsonValidationOptions = {
  allowEmpty?: boolean
  predicate?: (value: unknown) => boolean
  predicateMessage?: string
}

export type JsonValidationError = {
  type: 'required' | 'structure' | 'syntax'
  line?: number
  column?: number
  position?: number
  missingCommaLine?: number
}

function extractErrorPosition(
  error: unknown,
  jsonString: string
): { line?: number; column?: number; position?: number } {
  if (!(error instanceof Error)) return {}

  const message = error.message

  // Format 1: "Unexpected token } in JSON at position 15"
  const positionMatch = message.match(/at position (\d+)/i)
  if (positionMatch) {
    const position = parseInt(positionMatch[1], 10)
    const lines = jsonString.substring(0, position).split('\n')
    return {
      line: lines.length,
      column: lines[lines.length - 1].length + 1,
      position,
    }
  }

  // Format 2: "JSON.parse: ... at line 2 column 3"
  const lineColMatch = message.match(/at line (\d+) column (\d+)/i)
  if (lineColMatch) {
    return {
      line: parseInt(lineColMatch[1], 10),
      column: parseInt(lineColMatch[2], 10),
    }
  }

  return {}
}

function buildSyntaxError(
  error: unknown,
  jsonString: string
): JsonValidationError {
  if (!(error instanceof Error)) {
    return {
      type: 'syntax',
    } satisfies JsonValidationError
  }

  const position = extractErrorPosition(error, jsonString)
  const message = error.message

  // Check if it's a "missing comma" type error
  const isMissingCommaError =
    message.includes("Expected ','") ||
    message.includes('Expected property name') ||
    message.includes('Unexpected string')

  const missingCommaLine =
    isMissingCommaError && position.line && position.line > 1
      ? position.line - 1
      : undefined

  return {
    type: 'syntax',
    ...position,
    missingCommaLine,
  } satisfies JsonValidationError
}

function formatErrorMessage(error: unknown, jsonString: string): string {
  if (!(error instanceof Error)) return 'Invalid JSON'

  const position = extractErrorPosition(error, jsonString)
  const message = error.message
  const syntaxError = buildSyntaxError(error, jsonString)

  if (position.line && position.column) {
    let hint = ''
    if (syntaxError.missingCommaLine) {
      hint = ` (check line ${syntaxError.missingCommaLine} for missing comma)`
    }
    return `Error at line ${position.line}, column ${position.column}: ${message}${hint}`
  }

  if (position.position !== undefined) {
    return `Error at position ${position.position}: ${message}`
  }

  return message
}

export function validateJsonString(
  value: string,
  options: JsonValidationOptions = {}
) {
  const { allowEmpty = true, predicate, predicateMessage } = options
  const trimmed = value.trim()

  if (!trimmed) {
    return {
      valid: allowEmpty,
      message: allowEmpty ? undefined : 'Value is required',
      error: allowEmpty
        ? undefined
        : ({
            type: 'required',
          } satisfies JsonValidationError),
    }
  }

  try {
    const parsed = JSON.parse(trimmed)
    if (predicate && !predicate(parsed)) {
      return {
        valid: false,
        message: predicateMessage || 'JSON structure is invalid',
        error: {
          type: 'structure',
        } satisfies JsonValidationError,
      }
    }

    return { valid: true, error: undefined }
  } catch (error: unknown) {
    return {
      valid: false,
      message: formatErrorMessage(error, trimmed),
      error: buildSyntaxError(error, trimmed),
    }
  }
}
