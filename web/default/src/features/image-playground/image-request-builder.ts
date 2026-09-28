import type {
  ImageModelCapabilities,
  ImageParameterCapability,
  ImageRequestPayload,
  ImageRequestProfile,
} from './api'

export type ImageParameterValues = Record<string, unknown>
export type ImageParameterEnabled = Record<string, boolean>

export type BuildImageRequestOptions = {
  model: string
  group?: string
  prompt: string
  capabilities: ImageModelCapabilities | null
  parameterValues: ImageParameterValues
  parameterEnabled: ImageParameterEnabled
  image?: string
  profile?: ImageRequestProfile
}

export const DEFAULT_IMAGE_REQUEST_PROFILE: ImageRequestProfile = {
  mode: 'auto',
  endpoint: 'generations',
  referenceField: 'image',
  customParameters: '',
}

const protectedRequestKeys = new Set([
  'model',
  'group',
  'prompt',
  'n',
  'image',
  'image_url',
  'input',
])

function parseCustomParameters(raw: string | undefined) {
  if (!raw?.trim()) return {}
  try {
    const value = JSON.parse(raw) as unknown
    if (!value || typeof value !== 'object' || Array.isArray(value)) return {}
    return Object.fromEntries(
      Object.entries(value).filter(([key]) => !protectedRequestKeys.has(key))
    )
  } catch {
    return {}
  }
}

export function validateImageCustomParameters(raw: string | undefined): string | null {
  if (!raw?.trim()) return null
  try {
    const value = JSON.parse(raw) as unknown
    if (!value || typeof value !== 'object' || Array.isArray(value)) {
      return 'Custom JSON parameters must be an object'
    }
    const findProtectedKey = (input: unknown): string | undefined => {
      if (!input || typeof input !== 'object') return undefined
      for (const [key, nested] of Object.entries(input)) {
        if (protectedRequestKeys.has(key)) return key
        const child = findProtectedKey(nested)
        if (child) return child
      }
      return undefined
    }
    const protectedKey = findProtectedKey(value)
    return protectedKey
      ? `${protectedKey} is controlled by Studio`
      : null
  } catch {
    return 'Custom JSON parameters must be valid JSON'
  }
}

export function resolveImageRequestEndpoint(
  model: string,
  hasReference: boolean,
  profile: ImageRequestProfile
): 'generations' | 'edits' | 'gemini' {
  const lowerModel = model.toLowerCase()
  if (lowerModel.startsWith('gemini-') || lowerModel.startsWith('nano-banana')) {
    return 'gemini'
  }
  if (profile.mode === 'custom') return profile.endpoint
  if (profile.mode === 'edits') return 'edits'
  if (profile.mode === 'generations') return 'generations'
  return hasReference && /^(gpt-image-|grok-imagine-image)/i.test(model)
    ? 'edits'
    : 'generations'
}

function setNestedValue(
  target: Record<string, unknown>,
  path: string,
  value: unknown
) {
  const segments = path.split('.').filter(Boolean)
  if (segments.length === 0) return
  let current = target
  segments.forEach((segment, index) => {
    if (index === segments.length - 1) {
      current[segment] = value
      return
    }
    const next = current[segment]
    if (!next || typeof next !== 'object' || Array.isArray(next)) {
      current[segment] = {}
    }
    current = current[segment] as Record<string, unknown>
  })
}

function valueForParameter(
  parameter: ImageParameterCapability,
  values: ImageParameterValues
) {
  const value = values[parameter.key]
  if (
    value !== undefined &&
    value !== null &&
    value !== '' &&
    (!parameter.options ||
      typeof value !== 'string' ||
      parameter.options.includes(value))
  ) {
    return value
  }
  return parameter.default
}

function boundedValue(parameter: ImageParameterCapability, value: unknown) {
  if (parameter.type !== 'integer' && parameter.type !== 'number') return value
  const number = Number(value)
  if (!Number.isFinite(number)) return parameter.default
  const bounded = Math.min(
    parameter.max ?? Number.MAX_SAFE_INTEGER,
    Math.max(parameter.min ?? Number.MIN_SAFE_INTEGER, number)
  )
  return parameter.type === 'integer' ? Math.round(bounded) : bounded
}

export function buildImageRequest({
  model,
  group,
  prompt,
  capabilities,
  parameterValues,
  parameterEnabled,
  image,
  profile,
}: BuildImageRequestOptions): ImageRequestPayload {
  const request: ImageRequestPayload = { model, prompt }
  if (group) request.group = group
  const geminiModel = /^(gemini-|nano-banana)/i.test(model)
  if (image && (geminiModel || profile?.referenceField !== 'none')) {
    if (geminiModel) request.image = image
    else if (profile?.referenceField === 'image_url') request.image_url = image
    else if (profile?.referenceField === 'input.image') {
      request.input = { image }
    } else request.image = image
  }

  const outputFormatParameter = capabilities?.parameters.find(
    (parameter) => parameter.key === 'output_format'
  )
  const outputFormat =
    outputFormatParameter &&
    parameterEnabled[outputFormatParameter.key] !== false
      ? valueForParameter(outputFormatParameter, parameterValues)
      : undefined

  capabilities?.parameters.forEach((parameter) => {
    if (parameterEnabled[parameter.key] === false) return
    const value = valueForParameter(parameter, parameterValues)
    if (value === undefined || value === null || value === '') return
    if (
      parameter.key === 'output_compression' &&
      request.model.toLowerCase().startsWith('gpt-image-2.5') &&
      outputFormat !== 'jpeg'
    ) {
      return
    }
    setNestedValue(
      request,
      parameter.request_key,
      boundedValue(parameter, value)
    )
  })

  Object.assign(request, parseCustomParameters(profile?.customParameters))
  request.model = model
  request.prompt = prompt
  if (group) request.group = group
  else delete request.group
  if (image && (geminiModel || profile?.referenceField !== 'none')) {
    if (geminiModel) request.image = image
    else if (profile.referenceField === 'image_url') request.image_url = image
    else if (profile.referenceField === 'input.image') {
      request.input = { image }
    } else request.image = image
  }
  return request
}

function geminiImageRequest(request: ImageRequestPayload) {
  const parts: Array<Record<string, unknown>> = [{ text: request.prompt }]
  if (typeof request.image === 'string' && request.image.length > 0) {
    const mimeType = request.image.match(/^data:([^;,]+)/)?.[1] ?? 'image/*'
    parts.unshift({
      inlineData: {
        mimeType,
        data: '<reference image omitted>',
      },
    })
  }

  const generationConfig: Record<string, unknown> = {
    responseModalities: ['TEXT', 'IMAGE'],
  }
  const imageConfig: Record<string, string> = {}
  if (typeof request.size === 'string' && request.size.includes(':')) {
    imageConfig.aspectRatio = request.size
  }
  if (request.quality === '2K' || request.quality === '4K') {
    imageConfig.imageSize = request.quality
  }
  if (Object.keys(imageConfig).length > 0) {
    generationConfig.imageConfig = imageConfig
  }

  const result: Record<string, unknown> = {
    contents: [{ role: 'user', parts }],
    generationConfig,
  }
  for (const [key, value] of Object.entries(request)) {
    if (
      !['model', 'group', 'prompt', 'image', 'image_url', 'input', 'size', 'quality', 'n'].includes(key)
    ) {
      result[key] = value
    }
  }
  return result
}

export function imageRequestJson(
  request: ImageRequestPayload,
  isGemini = false,
  referenceField?: string
) {
  if (isGemini) {
    return JSON.stringify(geminiImageRequest(request), null, 2)
  }
  const copy = { ...request }
  delete copy.group
  if (typeof copy.image === 'string') copy.image = '<reference image omitted>'
  if (typeof copy.image_url === 'string') copy.image_url = '<reference image omitted>'
  if (referenceField === 'input.image' && copy.input && typeof copy.input === 'object') {
    copy.input = { ...(copy.input as Record<string, unknown>), image: '<reference image omitted>' }
  }
  return JSON.stringify(copy, null, 2)
}

export function imageRequestEndpoint(
  model: string,
  isEdit: boolean,
  isGemini: boolean,
  profile?: ImageRequestProfile
) {
  if (isGemini) {
    return `/v1beta/models/${encodeURIComponent(model)}:generateContent`
  }
  if (profile?.mode === 'custom') {
    if (profile.endpoint === 'gemini') {
      return `/v1beta/models/${encodeURIComponent(model)}:generateContent`
    }
    return profile.endpoint === 'edits'
      ? '/v1/images/edits'
      : '/v1/images/generations'
  }
  return isEdit ? '/v1/images/edits' : '/v1/images/generations'
}

export function imageRequestCurl(
  request: ImageRequestPayload,
  endpoint: string,
  isEdit: boolean,
  isGemini = false,
  referenceField?: string
) {
  const body = imageRequestJson(request, isGemini, referenceField)
  if (isEdit && !isGemini) {
    const fields = Object.entries(request)
      .filter(([key]) => key !== 'image' && key !== 'group')
      .map(([key, value]) => {
        const rendered =
          typeof value === 'object' ? JSON.stringify(value) : String(value)
        return `  -F ${shellQuote(`${key}=${rendered}`)}`
      })
    fields.push('  -F "image=@reference.png"')
    return [
      `curl -X POST ${shellQuote(endpoint)} \\\n  -H 'Authorization: Bearer <your-token>'`,
      ...fields,
    ].join(' \\\n')
  }
  return `curl -X POST ${shellQuote(endpoint)} \\\n  -H 'Content-Type: application/json' \\\n  -H 'Authorization: Bearer <your-token>' \\\n  -d '${body.replaceAll("'", "'\\''")}'`
}

function shellQuote(value: string) {
  return `'${value.replaceAll("'", "'\\''")}'`
}
