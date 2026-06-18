/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import http from 'node:http'

const port = Number(process.env.MOCK_API_PORT || 3000)

const vendors = [
  {
    id: 1,
    name: 'OpenAI',
    icon: 'openai',
    description: 'Flagship reasoning, multimodal, and realtime models.',
  },
  {
    id: 2,
    name: 'Anthropic',
    icon: 'anthropic',
    description: 'Claude models for long-context writing and analysis.',
  },
  {
    id: 3,
    name: 'Google',
    icon: 'google',
    description: 'Gemini models for multimodal and high-context workloads.',
  },
  {
    id: 4,
    name: 'DeepSeek',
    icon: 'deepseek',
    description: 'Reasoning and coding focused open models.',
  },
  {
    id: 5,
    name: 'ByteDance',
    icon: 'bytedance',
    description: 'Video and creative generation models.',
  },
]

const groupRatio = {
  default: 1,
  premium: 0.85,
  research: 0.6,
}

const usableGroup = {
  default: { desc: 'Default', ratio: 1 },
  premium: { desc: 'Premium', ratio: 0.85 },
  research: { desc: 'Research', ratio: 0.6 },
}

const supportedEndpoint = {
  openai: '/v1/chat/completions',
  'openai-response': '/v1/responses',
  anthropic: '/v1/messages',
  gemini: '/v1beta/models',
  embeddings: '/v1/embeddings',
  'image-generation': '/v1/images/generations',
  'openai-video': '/v1/videos',
}

const models = [
  {
    id: 1,
    model_name: 'gpt-4.1',
    description: 'General purpose flagship model with strong coding and tool use.',
    vendor_id: 1,
    is_new: 1,
    quota_type: 0,
    model_ratio: 2,
    completion_ratio: 8,
    cache_ratio: 0.5,
    enable_groups: ['default', 'premium'],
    tags: 'chat,tools,vision,reasoning',
    supported_endpoint_types: ['openai', 'openai-response'],
  },
  {
    id: 2,
    model_name: 'gpt-4.1-mini',
    description: 'Fast and economical model for everyday product traffic.',
    vendor_id: 1,
    discount_enabled: 1,
    discount_percent: 20,
    discount_label: 'Mock promo',
    quota_type: 0,
    model_ratio: 0.4,
    completion_ratio: 1.6,
    cache_ratio: 0.1,
    enable_groups: ['default', 'premium', 'research'],
    tags: 'chat,fast,tools',
    supported_endpoint_types: ['openai', 'openai-response'],
  },
  {
    id: 3,
    model_name: 'o3',
    description: 'Reasoning model for complex analysis and multi-step tasks.',
    vendor_id: 1,
    quota_type: 0,
    model_ratio: 10,
    completion_ratio: 40,
    enable_groups: ['premium', 'research'],
    tags: 'reasoning,coding,math',
    supported_endpoint_types: ['openai-response'],
    billing_mode: 'tiered_expr',
    billing_expr:
      'tier(p, 0, 100000, 10) + tier(c, 0, 100000, 40) * if(reasoning_tokens > 0, 1.2, 1)',
    pricing_version: 'mock-2026-06',
  },
  {
    id: 4,
    model_name: 'claude-4-sonnet',
    description: 'Balanced Claude model for coding, agents, and long documents.',
    vendor_id: 2,
    quota_type: 0,
    model_ratio: 3,
    completion_ratio: 15,
    cache_ratio: 0.3,
    create_cache_ratio: 3.75,
    enable_groups: ['default', 'premium'],
    tags: 'chat,coding,long-context,tools',
    supported_endpoint_types: ['anthropic'],
  },
  {
    id: 5,
    model_name: 'claude-4-opus',
    description: 'High-capability Claude model for deep reasoning and difficult code.',
    vendor_id: 2,
    quota_type: 0,
    model_ratio: 15,
    completion_ratio: 75,
    cache_ratio: 1.5,
    create_cache_ratio: 18.75,
    enable_groups: ['premium'],
    tags: 'reasoning,coding,long-context',
    supported_endpoint_types: ['anthropic'],
  },
  {
    id: 6,
    model_name: 'gemini-2.5-pro',
    description: 'Large-context multimodal model for research and analysis.',
    vendor_id: 3,
    is_new: 1,
    quota_type: 0,
    model_ratio: 1.25,
    completion_ratio: 10,
    cache_ratio: 0.31,
    enable_groups: ['default', 'premium', 'research'],
    tags: 'chat,vision,long-context,reasoning',
    supported_endpoint_types: ['gemini'],
  },
  {
    id: 7,
    model_name: 'gemini-2.5-flash',
    description: 'Low-latency Gemini model for high-volume application traffic.',
    vendor_id: 3,
    quota_type: 0,
    model_ratio: 0.3,
    completion_ratio: 2.5,
    cache_ratio: 0.075,
    enable_groups: ['default', 'premium', 'research'],
    tags: 'chat,fast,vision',
    supported_endpoint_types: ['gemini'],
  },
  {
    id: 8,
    model_name: 'deepseek-r1',
    description: 'Reasoning model tuned for math, coding, and structured answers.',
    vendor_id: 4,
    quota_type: 0,
    model_ratio: 0.55,
    completion_ratio: 2.19,
    cache_ratio: 0.14,
    enable_groups: ['default', 'research'],
    tags: 'reasoning,coding,math',
    supported_endpoint_types: ['openai'],
  },
  {
    id: 9,
    model_name: 'text-embedding-3-large',
    description: 'Embedding model for semantic search and retrieval pipelines.',
    vendor_id: 1,
    quota_type: 0,
    model_ratio: 0.13,
    completion_ratio: 0,
    enable_groups: ['default', 'premium', 'research'],
    tags: 'embedding,retrieval',
    supported_endpoint_types: ['embeddings'],
  },
  {
    id: 10,
    model_name: 'sora-2',
    description: 'Video generation model with per-request resolution pricing.',
    vendor_id: 1,
    is_new: 1,
    quota_type: 1,
    model_ratio: 0,
    completion_ratio: 0,
    model_price: 0.12,
    enable_groups: ['premium'],
    tags: 'video,creative',
    supported_endpoint_types: ['openai-video'],
    sora_per_request_pricing: {
      enabled: true,
      resolution_tiers: [
        { value: '720p', multiplier: 1 },
        { value: '1080p', multiplier: 1.8 },
      ],
      audio_generation_surcharge: 0.03,
    },
  },
  {
    id: 11,
    model_name: 'seedream-4.0',
    description: 'Image generation model for marketing and product visuals.',
    vendor_id: 5,
    quota_type: 1,
    model_ratio: 0,
    completion_ratio: 0,
    model_price: 0.018,
    image_ratio: 1,
    enable_groups: ['default', 'premium'],
    tags: 'image,creative',
    supported_endpoint_types: ['image-generation'],
  },
]

const mockVariants = Array.from({ length: 24 }, (_, index) => {
  const source = models[index % models.length]
  const sequence = String(index + 1).padStart(2, '0')
  return {
    ...source,
    id: 100 + index,
    model_name: `${source.model_name}-mock-${sequence}`,
    description: `${source.description} Mock catalog variant ${sequence}.`,
    is_new: index < 4 ? 1 : 0,
  }
})

models.push(...mockVariants)

const baseCreatedTime = Math.floor(
  new Date('2026-06-16T09:00:00+08:00').getTime() / 1000
)

for (const [index, model] of models.entries()) {
  const endpoints = model.supported_endpoint_types || []
  const tags = String(model.tags || '').split(/[,;|\s]+/)

  model.created_time = baseCreatedTime - index * 60 * 60

  if (endpoints.includes('embeddings')) {
    model.input_modalities = ['text']
    model.output_modalities = ['embedding']
  } else if (endpoints.includes('image-generation')) {
    model.input_modalities = ['text', 'image']
    model.output_modalities = ['image']
  } else if (endpoints.includes('openai-video')) {
    model.input_modalities = ['text', 'image']
    model.output_modalities = ['video']
  } else {
    model.input_modalities = tags.includes('vision') ? ['text', 'image'] : ['text']
    model.output_modalities = ['text']
  }
}

function statusResponse() {
  return {
    success: true,
    data: {
      system_name: 'OmniRouters Mock',
      model_square_theme: 'catalog',
      logo: '/logo.png',
      footer_html: '',
      demo_site_enabled: true,
      display_token_stat_enabled: true,
      display_in_currency: true,
      quota_display_type: 'currency',
      quota_per_unit: 500000,
      price: 7,
      usd_exchange_rate: 7.25,
      custom_currency_symbol: '¥',
      custom_currency_exchange_rate: 7.25,
      HeaderNavModules: JSON.stringify({
        home: true,
        console: true,
        pricing: { enabled: true, requireAuth: false },
        playground: { enabled: true, requireAuth: false },
        rankings: { enabled: true, requireAuth: false },
        docs: true,
        about: true,
      }),
    },
  }
}

function pricingResponse() {
  return {
    success: true,
    data: models,
    vendors,
    group_ratio: groupRatio,
    usable_group: usableGroup,
    supported_endpoint: supportedEndpoint,
    auto_groups: ['default'],
  }
}

function perfSummaryResponse() {
  return {
    success: true,
    data: {
      models: models.map((model, index) => ({
        model_name: model.model_name,
        avg_latency_ms: 850 + index * 85,
        success_rate: Math.max(0.94, 0.995 - index * 0.004),
        avg_tps: Math.max(18, 75 - index * 4),
        request_count: 12000 + index * 3100,
      })),
    },
  }
}

function perfMetricsResponse(url) {
  const modelName = url.searchParams.get('model') || 'gpt-4.1'
  const now = Date.now()
  const series = Array.from({ length: 24 }, (_, index) => ({
    ts: now - (23 - index) * 60 * 60 * 1000,
    avg_ttft_ms: 180 + index * 4,
    avg_latency_ms: 900 + index * 20,
    success_rate: 0.985 - (index % 5) * 0.003,
    avg_tps: 62 - (index % 7),
  }))

  return {
    success: true,
    data: {
      model_name: modelName,
      series_schema: 'hourly',
      groups: Object.keys(usableGroup).map((group, index) => ({
        group,
        avg_ttft_ms: 190 + index * 25,
        avg_latency_ms: 960 + index * 120,
        success_rate: 0.99 - index * 0.012,
        avg_tps: 64 - index * 8,
        series,
      })),
    },
  }
}

function json(res, status, payload) {
  const body = JSON.stringify(payload, null, 2)
  res.writeHead(status, {
    'content-type': 'application/json; charset=utf-8',
    'cache-control': 'no-store',
    'access-control-allow-origin': 'http://localhost:8080',
    'access-control-allow-credentials': 'true',
  })
  res.end(body)
}

function notFound(res, pathname) {
  json(res, 404, {
    success: false,
    message: `No mock route for ${pathname}`,
  })
}

const server = http.createServer((req, res) => {
  const url = new URL(req.url || '/', `http://localhost:${port}`)

  if (req.method === 'OPTIONS') {
    res.writeHead(204, {
      'access-control-allow-origin': 'http://localhost:8080',
      'access-control-allow-credentials': 'true',
      'access-control-allow-methods': 'GET,POST,OPTIONS',
      'access-control-allow-headers': 'content-type,new-api-user',
    })
    res.end()
    return
  }

  if (url.pathname === '/') {
    res.writeHead(200, { 'content-type': 'text/plain; charset=utf-8' })
    res.end(
      [
        'OmniRouters mock API is running.',
        'Start the frontend separately, then open /pricing on the frontend port.',
      ].join('\n')
    )
    return
  }

  if (url.pathname === '/api/status') return json(res, 200, statusResponse())
  if (url.pathname === '/api/setup') {
    return json(res, 200, { success: true, data: { status: true } })
  }
  if (url.pathname === '/api/home_page_content') {
    return json(res, 200, { success: true, data: '' })
  }
  if (url.pathname === '/api/pricing') return json(res, 200, pricingResponse())
  if (url.pathname === '/api/notice') {
    return json(res, 200, { success: true, data: '' })
  }
  if (url.pathname === '/api/perf-metrics/summary') {
    return json(res, 200, perfSummaryResponse())
  }
  if (url.pathname === '/api/perf-metrics') {
    return json(res, 200, perfMetricsResponse(url))
  }

  notFound(res, url.pathname)
})

server.listen(port, () => {
  console.log(`Mock API listening on http://localhost:${port}`)
})
