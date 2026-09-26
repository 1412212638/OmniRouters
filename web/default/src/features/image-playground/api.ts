import { api } from '@/lib/api'

export const IMAGE_PLAYGROUND_ENDPOINTS = {
  GENERATIONS: '/pg/images/generations',
  EDITS: '/pg/images/edits',
  USER_MODELS: '/api/user/models',
  USER_GROUPS: '/api/user/self/groups',
  MODEL_CAPABILITIES: '/api/user/image-model-capabilities',
} as const

export type ImageRequestPayload = {
  model: string
  group?: string
  prompt: string
  [key: string]: unknown
}

export type ImageGenerationRequest = ImageRequestPayload

export type ImageGenerationResponse = {
  data?: Array<{
    b64_json?: string
    url?: string
    revised_prompt?: string
    mime_type?: string
  }>
  created?: number
}

export type ImageEditRequest = ImageRequestPayload & {
  image: string
}

export type ImageParameterCapability = {
  key: string
  request_key: string
  type: 'integer' | 'number' | 'boolean' | 'enum' | 'string'
  default?: unknown
  min?: number
  max?: number
  step?: number
  options?: string[]
  enabled_by_default: boolean
  operations: string[]
}

export type ImageModelCapabilities = {
  model: string
  operation: 'generation' | 'edit'
  parameters: ImageParameterCapability[]
}

export type ImageModelOption = { label: string; value: string }
export type ImageGroupOption = {
  label: string
  value: string
  ratio?: number
  desc?: string
}

export async function generateImages(
  payload: ImageGenerationRequest,
  signal?: AbortSignal
): Promise<ImageGenerationResponse> {
  const response = await api.post(
    IMAGE_PLAYGROUND_ENDPOINTS.GENERATIONS,
    payload,
    {
      signal,
      skipErrorHandler: true,
    } as Record<string, unknown>
  )
  return response.data
}

export async function editImage(
  payload: ImageEditRequest,
  signal?: AbortSignal
): Promise<ImageGenerationResponse> {
  const formData = new FormData()
  formData.append('model', payload.model)
  formData.append('prompt', payload.prompt)
  Object.entries(payload).forEach(([key, value]) => {
    if (
      key === 'model' ||
      key === 'prompt' ||
      key === 'group' ||
      key === 'image'
    ) {
      return
    }
    if (value === undefined || value === null) {
      return
    }
    if (['n', 'size', 'quality', 'response_format', 'stream'].includes(key)) {
      formData.append(key, String(value))
      return
    }
    formData.append(key, JSON.stringify(value))
  })
  if (payload.group) formData.append('group', payload.group)

  const imageResponse = await fetch(payload.image, { signal })
  if (!imageResponse.ok) {
    throw new Error('Unable to read the reference image')
  }
  const imageBlob = await imageResponse.blob()
  const extension = imageBlob.type.split('/')[1] || 'png'
  formData.append('image', imageBlob, `reference.${extension}`)

  const response = await api.post(IMAGE_PLAYGROUND_ENDPOINTS.EDITS, formData, {
    signal,
    skipErrorHandler: true,
  } as Record<string, unknown>)
  return response.data
}

export async function editGeminiImage(
  payload: ImageEditRequest,
  signal?: AbortSignal
): Promise<ImageGenerationResponse> {
  const imageResponse = await fetch(payload.image, { signal })
  if (!imageResponse.ok) {
    throw new Error('Unable to read the reference image')
  }
  const imageBlob = await imageResponse.blob()
  const imageData = await new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.addEventListener('load', () => {
      if (typeof reader.result === 'string') resolve(reader.result)
      else reject(new Error('Unable to encode the reference image'))
    })
    reader.addEventListener('error', () =>
      reject(new Error('Unable to encode the reference image'))
    )
    reader.readAsDataURL(imageBlob)
  })
  const response = await api.post(
    IMAGE_PLAYGROUND_ENDPOINTS.EDITS,
    { ...payload, image: imageData },
    {
      signal,
      skipErrorHandler: true,
    } as Record<string, unknown>
  )
  return response.data
}

export async function getImageModels(
  group: string
): Promise<ImageModelOption[]> {
  const response = await api.get(IMAGE_PLAYGROUND_ENDPOINTS.USER_MODELS, {
    params: { endpoint_type: 'image-generation', group },
  })
  const data = response.data?.data
  return Array.isArray(data)
    ? data.map((model: string) => ({ label: model, value: model }))
    : []
}

export async function getImageGroups(): Promise<ImageGroupOption[]> {
  const response = await api.get(IMAGE_PLAYGROUND_ENDPOINTS.USER_GROUPS)
  const data = response.data?.data
  if (!data || typeof data !== 'object') return []
  return Object.entries(
    data as Record<string, { desc?: string; ratio?: number }>
  ).map(([value, info]) => ({
    label: value,
    value,
    desc: info?.desc,
    ratio: info?.ratio,
  }))
}

export async function getImageModelCapabilities(
  group: string,
  model: string,
  operation: 'generation' | 'edit'
): Promise<ImageModelCapabilities> {
  const response = await api.get(
    IMAGE_PLAYGROUND_ENDPOINTS.MODEL_CAPABILITIES,
    {
      params: { group, model, operation },
    }
  )
  const data = response.data?.data
  return {
    model,
    operation,
    parameters: Array.isArray(data?.parameters) ? data.parameters : [],
  }
}
