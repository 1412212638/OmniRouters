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

export function imageRequestJson(request: ImageRequestPayload) {
  const copy = { ...request }
  if (typeof copy.image === 'string') copy.image = '<reference image omitted>'
  return JSON.stringify(copy, null, 2)
}

export function imageRequestCurl(
  request: ImageRequestPayload,
  endpoint: string,
  isEdit: boolean
) {
  const body = imageRequestJson(request)
  if (isEdit) {
    const fields = Object.entries(request)
      .filter(([key]) => key !== 'image')
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
