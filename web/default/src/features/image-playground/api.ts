import { api } from '@/lib/api'

export const IMAGE_PLAYGROUND_ENDPOINTS = {
  GENERATIONS: '/pg/images/generations',
  USER_MODELS: '/api/user/models',
  USER_GROUPS: '/api/user/self/groups',
} as const

export type ImageGenerationRequest = {
  model: string
  group?: string
  prompt: string
  n: number
  size?: string
  quality?: string
}

export type ImageGenerationResponse = {
  data?: Array<{
    b64_json?: string
    url?: string
    revised_prompt?: string
    mime_type?: string
  }>
  created?: number
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
