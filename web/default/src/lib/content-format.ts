export function isHttpUrl(value: string): boolean {
  try {
    const url = new URL(value)
    return url.protocol === 'http:' || url.protocol === 'https:'
  } catch {
    return false
  }
}

export function isLikelyHtml(value: string): boolean {
  return /<!doctype html|<html[\s>]|<head[\s>]|<body[\s>]|<style[\s>]|<script[\s>]|<\/?[a-z][\s\S]*>/i.test(
    value
  )
}

/**
 * A complete document needs a document context for its head, scripts, and
 * global CSS. HTML fragments continue to use the sanitized inline renderer.
 */
export function isCompleteHtmlDocument(value: string): boolean {
  const trimmed = value.trim()

  return (
    /^<!doctype\s+html\b/i.test(trimmed) ||
    (/<html[\s>]/i.test(trimmed) && /<head[\s>]/i.test(trimmed))
  )
}
