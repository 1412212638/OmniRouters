export type LocalizedText = string | Record<string, string> | undefined | null

export function resolveLocalizedText(
  value: LocalizedText,
  language?: string
): string {
  if (!value) return ''
  if (typeof value === 'string') return value
  return value[language ?? ''] ?? value.en ?? Object.values(value)[0] ?? ''
}
