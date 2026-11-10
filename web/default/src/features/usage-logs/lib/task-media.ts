import type { TaskLog } from '../types'

export type PreviewMediaKind = 'image' | 'video' | 'audio'

export interface PreviewMedia {
  kind: PreviewMediaKind
  url: string
  title?: string
}

function mediaKindFromValue(
  key: string,
  value: string
): PreviewMediaKind | null {
  const normalizedKey = key.toLowerCase()
  const normalizedValue = value.split('?')[0].toLowerCase()
  if (
    normalizedKey.includes('video') ||
    normalizedValue.includes('video/') ||
    /\.(mp4|webm|mov|m4v|m3u8)$/.test(normalizedValue)
  ) {
    return 'video'
  }
  if (
    normalizedKey.includes('audio') ||
    normalizedValue.includes('audio/') ||
    /\.(mp3|wav|m4a|aac|ogg|flac)$/.test(normalizedValue)
  ) {
    return 'audio'
  }
  if (
    normalizedKey.includes('image') ||
    normalizedValue.includes('image/') ||
    /^(png|jpe?g|webp|gif|avif)$/.test(normalizedValue) ||
    /\.(png|jpe?g|webp|gif|avif)$/.test(normalizedValue)
  ) {
    return 'image'
  }
  return null
}

function parseTaskData(data: unknown): unknown {
  if (data && typeof data === 'object') return data
  if (typeof data !== 'string') return null
  try {
    return JSON.parse(data)
  } catch {
    return null
  }
}

export function extractTaskMedia(
  data: unknown,
  fallbackKind?: PreviewMediaKind
): PreviewMedia[] {
  const media: PreviewMedia[] = []
  const seen = new Set<string>()
  const add = (url: unknown, kind: PreviewMediaKind | null, title?: string) => {
    if (typeof url !== 'string' || !/^https?:\/\//i.test(url) || !kind) return
    const key = `${kind}:${url}`
    if (seen.has(key)) return
    seen.add(key)
    media.push(title ? { kind, url, title } : { kind, url })
  }

  const visit = (value: unknown, key = '', path = '') => {
    if (typeof value === 'string') {
      if (!/^https?:\/\//i.test(value)) return
      const isGenericUrl = key === 'url' || key === 'urls'
      const inferredKind =
        mediaKindFromValue(key, value) ||
        (isGenericUrl && /\.(output|content|result|urls)(\.|\[|$)/.test(path)
          ? fallbackKind
          : null)
      add(value, inferredKind ?? null)
      return
    }
    if (Array.isArray(value)) {
      value.forEach((item, index) =>
        visit(item, key, `${path}.${key}[${index}]`)
      )
      return
    }
    if (!value || typeof value !== 'object') return

    const record = value as Record<string, unknown>
    const mime = record.mime_type ?? record.content_type ?? record.output_format
    const metadataKind =
      typeof mime === 'string' ? mediaKindFromValue('mime', mime) : null
    let title: string | undefined
    if (typeof record.title === 'string') title = record.title
    else if (typeof record.name === 'string') title = record.name

    for (const [childKey, childValue] of Object.entries(record)) {
      const childPath = `${path}.${childKey}`.toLowerCase()
      if (/\.((input)|(reference))([._]|\[|$)/.test(childPath)) continue
      if (typeof childValue === 'string' && /^https?:\/\//i.test(childValue)) {
        const isGenericUrl = childKey === 'url' || childKey === 'urls'
        const inferredKind =
          mediaKindFromValue(childKey, childValue) ||
          metadataKind ||
          (isGenericUrl &&
          /\.(output|content|result|urls)(\.|\[|$)/.test(childPath)
            ? fallbackKind
            : null)
        add(childValue, inferredKind ?? null, title)
      } else {
        visit(childValue, childKey, childPath)
      }
    }
  }

  visit(parseTaskData(data))
  return media
}

export function getTaskPreviewMedia(
  log: Pick<
    TaskLog,
    'data' | 'status' | 'action' | 'fail_reason' | 'result_url' | 'task_id'
  >,
  successfulStatus: string,
  isVideoAction: boolean
): PreviewMedia[] {
  const extracted = extractTaskMedia(
    log.data,
    isVideoAction ? 'video' : undefined
  )
  if (log.status !== successfulStatus || !isVideoAction) return extracted

  const resultURL = log.result_url || log.fail_reason
  // Omni stores Base64 in its native result; dashboard lists intentionally
  // omit it. Offer preview so the cell can resolve its signed artifact URL.
  if ((resultURL || log.action === 'omni_video') && log.task_id) {
    return [
      {
        kind: 'video',
        url: `/v1/videos/${encodeURIComponent(log.task_id)}/content`,
      },
      ...extracted.filter((item) => item.kind !== 'video'),
    ]
  }
  return extracted
}

export function getDashboardTaskArtifactsURL(taskId: string): string {
  return `/api/task/${encodeURIComponent(taskId)}/artifacts`
}
