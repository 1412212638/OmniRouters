export type GeneratedImage = {
  key: string
  src: string
  mimeType?: string
  revisedPrompt?: string
  prompt: string
}

export type ImageGenerationMetadata = {
  size?: string
  aspectRatio?: string
  quality?: string
  count?: number
}

export type ImageGenerationEntry = {
  key: string
  prompt: string
  model: string
  group: string
  status: 'loading' | 'complete' | 'error'
  images: GeneratedImage[]
  metadata?: ImageGenerationMetadata
  error?: string
}

export type ImagePlaygroundState = {
  images: GeneratedImage[]
  entries: ImageGenerationEntry[]
  referenceImageKey: string | null
  prompt: string
  model: string
  group: string
  parameterValues: Record<string, Record<string, unknown>>
  parameterEnabled: Record<string, Record<string, boolean>>
}

const DATABASE_NAME = 'omnirouters-image-playground'
const DATABASE_VERSION = 1
const STORE_NAME = 'playground-state'
const FALLBACK_STORAGE_KEY = 'image-playground-state'
type StoredRecord = {
  key: string
  state: ImagePlaygroundState
}

function getStorageScope(): string {
  try {
    return window.localStorage.getItem('uid')?.trim() || 'anonymous'
  } catch {
    return 'anonymous'
  }
}

function getFallbackKey(): string {
  return `${FALLBACK_STORAGE_KEY}:${getStorageScope()}`
}

function openDatabase(): Promise<IDBDatabase | null> {
  if (typeof indexedDB === 'undefined') return Promise.resolve(null)

  return new Promise((resolve) => {
    const request = indexedDB.open(DATABASE_NAME, DATABASE_VERSION)
    request.addEventListener('upgradeneeded', () => {
      if (!request.result.objectStoreNames.contains(STORE_NAME)) {
        request.result.createObjectStore(STORE_NAME, { keyPath: 'key' })
      }
    })
    request.addEventListener('success', () => resolve(request.result))
    request.addEventListener('error', () => resolve(null))
    request.addEventListener('blocked', () => resolve(null))
  })
}

function parseState(value: unknown): ImagePlaygroundState | null {
  if (!value || typeof value !== 'object') return null
  const candidate = value as Partial<ImagePlaygroundState> & {
    count?: string
    size?: string
    quality?: string
    entries?: unknown[]
  }
  if (!Array.isArray(candidate.images)) return null

  const images = candidate.images.filter((image): image is GeneratedImage =>
    Boolean(
      image &&
      typeof image === 'object' &&
      typeof image.key === 'string' &&
      typeof image.src === 'string' &&
      typeof image.prompt === 'string'
    )
  )

  const entries =
    Array.isArray(candidate.entries) && candidate.entries.length > 0
      ? candidate.entries.flatMap((entry): ImageGenerationEntry[] => {
          const candidateEntry = entry as Partial<ImageGenerationEntry>
          if (
            !entry ||
            typeof entry !== 'object' ||
            typeof candidateEntry.key !== 'string' ||
            typeof candidateEntry.prompt !== 'string' ||
            typeof candidateEntry.model !== 'string' ||
            typeof candidateEntry.group !== 'string' ||
            !Array.isArray(candidateEntry.images)
          ) {
            return []
          }
          const status =
            candidateEntry.status === 'complete' ||
            candidateEntry.status === 'error'
              ? candidateEntry.status
              : 'error'
          const entryImages = candidateEntry.images.filter(
            (image): image is GeneratedImage =>
              Boolean(
                image &&
                typeof image === 'object' &&
                typeof image.key === 'string' &&
                typeof image.src === 'string' &&
                typeof image.prompt === 'string'
              )
          )
          const metadata =
            candidateEntry.metadata &&
            typeof candidateEntry.metadata === 'object'
              ? candidateEntry.metadata
              : undefined
          return [
            {
              key: candidateEntry.key,
              prompt: candidateEntry.prompt,
              model: candidateEntry.model,
              group: candidateEntry.group,
              status,
              images: entryImages,
              metadata,
              error:
                status === 'error'
                  ? candidateEntry.error || 'Generation interrupted'
                  : undefined,
            },
          ]
        })
      : images.map((image) => ({
          key: `legacy-${image.key}`,
          prompt: image.prompt,
          model: typeof candidate.model === 'string' ? candidate.model : '',
          group: typeof candidate.group === 'string' ? candidate.group : '',
          status: 'complete' as const,
          images: [image],
        }))

  const parameterValues =
    candidate.parameterValues && typeof candidate.parameterValues === 'object'
      ? candidate.parameterValues
      : {}
  const parameterEnabled =
    candidate.parameterEnabled && typeof candidate.parameterEnabled === 'object'
      ? candidate.parameterEnabled
      : {}
  if (!candidate.parameterValues) {
    const legacyValues: Record<string, unknown> = {}
    if (typeof candidate.count === 'string') {
      legacyValues.n = Number(candidate.count) || 1
    }
    if (typeof candidate.size === 'string') legacyValues.size = candidate.size
    if (typeof candidate.quality === 'string') {
      legacyValues.quality = candidate.quality
    }
    const legacyKey = `${candidate.group || ''}::${candidate.model || ''}::generation`
    if (Object.keys(legacyValues).length > 0) {
      ;(parameterValues as Record<string, Record<string, unknown>>)[legacyKey] =
        legacyValues
    }
  }

  return {
    images,
    entries,
    referenceImageKey:
      typeof candidate.referenceImageKey === 'string'
        ? candidate.referenceImageKey
        : null,
    prompt: typeof candidate.prompt === 'string' ? candidate.prompt : '',
    model: typeof candidate.model === 'string' ? candidate.model : '',
    group: typeof candidate.group === 'string' ? candidate.group : '',
    parameterValues: parameterValues as Record<string, Record<string, unknown>>,
    parameterEnabled: parameterEnabled as Record<
      string,
      Record<string, boolean>
    >,
  }
}

function readFallback(): ImagePlaygroundState | null {
  try {
    const raw = window.localStorage.getItem(getFallbackKey())
    return raw ? parseState(JSON.parse(raw)) : null
  } catch {
    return null
  }
}

function writeFallback(state: ImagePlaygroundState): void {
  try {
    window.localStorage.setItem(getFallbackKey(), JSON.stringify(state))
  } catch {
    // IndexedDB is the primary store and has substantially more capacity.
  }
}

export async function loadImagePlaygroundState(): Promise<ImagePlaygroundState | null> {
  const database = await openDatabase()
  if (!database) return readFallback()

  return new Promise((resolve) => {
    const transaction = database.transaction(STORE_NAME, 'readonly')
    const request = transaction.objectStore(STORE_NAME).get(getStorageScope())
    request.addEventListener('success', () => {
      const record = request.result as StoredRecord | undefined
      resolve(parseState(record?.state) ?? readFallback())
    })
    request.addEventListener('error', () => resolve(readFallback()))
  })
}

export async function saveImagePlaygroundState(
  state: ImagePlaygroundState
): Promise<void> {
  const database = await openDatabase()
  if (!database) {
    writeFallback(state)
    return
  }

  await new Promise<void>((resolve) => {
    const transaction = database.transaction(STORE_NAME, 'readwrite')
    transaction.objectStore(STORE_NAME).put({
      key: getStorageScope(),
      state,
    } satisfies StoredRecord)
    transaction.addEventListener('complete', () => resolve())
    transaction.addEventListener('error', () => {
      writeFallback(state)
      resolve()
    })
    transaction.addEventListener('abort', () => {
      writeFallback(state)
      resolve()
    })
  })
}

export async function clearImagePlaygroundState(): Promise<void> {
  const database = await openDatabase()
  if (database) {
    await new Promise<void>((resolve) => {
      const transaction = database.transaction(STORE_NAME, 'readwrite')
      transaction.objectStore(STORE_NAME).delete(getStorageScope())
      transaction.addEventListener('complete', () => resolve())
      transaction.addEventListener('error', () => resolve())
      transaction.addEventListener('abort', () => resolve())
    })
  }

  try {
    window.localStorage.removeItem(getFallbackKey())
  } catch {
    // Ignore storage cleanup errors.
  }
}
