import type { BillingUsageFieldSchema } from '../types'

// Usage and saved expressions stay per character. Only editor and storefront
// prices use the provider's 10,000-character denomination.
export function isTaskCharacterField(
  field: string,
  definition: BillingUsageFieldSchema | undefined
): boolean {
  return (
    field === 'characters' &&
    definition?.type === 'number' &&
    definition.unit === 'count'
  )
}

export function taskCharacterPriceScale(
  field: string,
  definition: BillingUsageFieldSchema | undefined
): number {
  return isTaskCharacterField(field, definition) ? 10_000 : 1
}

export const TASK_CHARACTER_PRICE_UNIT = {
  en: '10K characters',
  zh: '万字符',
  'zh-TW': '萬字元',
  fr: '10 000 caractères',
  ru: '10 000 символов',
  ja: '1万文字',
  vi: '10.000 ký tự',
}
