import type {
  ImageModelCapabilities,
  ImageParameterCapability,
  ImageRequestPayload,
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
  if (value !== undefined && value !== null && value !== '') return value
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
}: BuildImageRequestOptions): ImageRequestPayload {
  const request: ImageRequestPayload = { model, prompt }
  if (group) request.group = group
  if (image) request.image = image

  capabilities?.parameters.forEach((parameter) => {
    if (parameterEnabled[parameter.key] === false) return
    const value = valueForParameter(parameter, parameterValues)
    if (value === undefined || value === null || value === '') return
    setNestedValue(
      request,
      parameter.request_key,
      boundedValue(parameter, value)
    )
  })
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

  return {
    contents: [{ role: 'user', parts }],
    generationConfig,
  }
}

export function imageRequestJson(
  request: ImageRequestPayload,
  isGemini = false
) {
  if (isGemini) {
    return JSON.stringify(geminiImageRequest(request), null, 2)
  }
  const copy = { ...request }
  delete copy.group
  if (typeof copy.image === 'string') copy.image = '<reference image omitted>'
  return JSON.stringify(copy, null, 2)
}

export function imageRequestEndpoint(
  model: string,
  isEdit: boolean,
  isGemini: boolean
) {
  if (isGemini) {
    return `/v1beta/models/${encodeURIComponent(model)}:generateContent`
  }
  return isEdit ? '/v1/images/edits' : '/v1/images/generations'
}

export function imageRequestCurl(
  request: ImageRequestPayload,
  endpoint: string,
  isEdit: boolean,
  isGemini = false
) {
  const body = imageRequestJson(request, isGemini)
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
