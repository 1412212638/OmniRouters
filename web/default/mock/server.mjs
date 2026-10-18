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
  {
    id: 6,
    name: 'Alibaba Cloud',
    icon: 'alibabacloud',
    description: 'Regional inference and cloud-hosted foundation models.',
  },
  {
    id: 7,
    name: 'Tencent Cloud',
    icon: 'tencentcloud',
    description: 'Hosted language and multimodal inference.',
  },
  {
    id: 8,
    name: 'SiliconFlow',
    icon: 'siliconcloud',
    description: 'Open model inference across text, image, and embedding tasks.',
  },
  {
    id: 9,
    name: 'OpenRouter',
    icon: 'openrouter',
    description: 'Multi-provider model routing and inference.',
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
  openai: { path: '/v1/chat/completions', method: 'POST' },
  'openai-response': { path: '/v1/responses', method: 'POST' },
  anthropic: { path: '/v1/messages', method: 'POST' },
  gemini: { path: '/v1beta/models/{model}:generateContent', method: 'POST' },
  embeddings: { path: '/v1/embeddings', method: 'POST' },
  'image-generation': { path: '/v1/images/generations', method: 'POST' },
  'openai-video': { path: '/v1/videos', method: 'POST' },
}

const availableVendorsByModel = {
  'gpt-4.1': [1, 9],
  'gpt-4.1-mini': [1, 9],
  o3: [1, 9],
  'claude-4-sonnet': [2, 9],
  'claude-4-opus': [2, 9],
  'gemini-2.5-pro': [3, 6, 7, 9],
  'gemini-2.5-flash': [3, 6, 7, 9],
  'deepseek-r1': [4, 6, 7, 8, 9],
  'deepseek-v4-flash': [4, 6, 7, 8, 9],
  'text-embedding-3-large': [1, 8, 9],
  'sora-2': [1, 9],
  'seedream-4.0': [5, 6, 9],
  'type-safe-demo': [6, 7, 9],
  'jev-1.14': [6, 7, 9],
}

const modelMetadata = {
  'gpt-4.1': {
    context_length: 1047576,
    max_output_tokens: 32768,
    knowledge_cutoff: '2024-06',
    release_date: '2025-04',
    parameter_count: 'Proprietary',
    capabilities: ['function_calling', 'streaming', 'vision', 'json_mode', 'structured_output', 'tools', 'web_search', 'code_interpreter', 'caching'],
  },
  'gpt-4.1-mini': {
    context_length: 1047576,
    max_output_tokens: 32768,
    knowledge_cutoff: '2024-06',
    release_date: '2025-04',
    parameter_count: 'Proprietary',
    capabilities: ['function_calling', 'streaming', 'vision', 'json_mode', 'structured_output', 'tools', 'caching'],
  },
  o3: {
    context_length: 200000,
    max_output_tokens: 100000,
    knowledge_cutoff: '2024-06',
    release_date: '2025-04',
    parameter_count: 'Proprietary',
    capabilities: ['reasoning', 'function_calling', 'streaming', 'vision', 'structured_output', 'tools', 'web_search', 'code_interpreter'],
  },
  'claude-4-sonnet': {
    context_length: 200000,
    max_output_tokens: 64000,
    knowledge_cutoff: '2025-03',
    release_date: '2025-05',
    parameter_count: 'Proprietary',
    capabilities: ['reasoning', 'function_calling', 'streaming', 'vision', 'tools', 'caching'],
  },
  'claude-4-opus': {
    context_length: 200000,
    max_output_tokens: 32000,
    knowledge_cutoff: '2025-03',
    release_date: '2025-05',
    parameter_count: 'Proprietary',
    capabilities: ['reasoning', 'function_calling', 'streaming', 'vision', 'tools', 'caching'],
  },
  'gemini-2.5-pro': {
    context_length: 1048576,
    max_output_tokens: 65536,
    knowledge_cutoff: '2025-01',
    release_date: '2025-03',
    parameter_count: 'Proprietary',
    capabilities: ['reasoning', 'function_calling', 'streaming', 'vision', 'json_mode', 'structured_output', 'tools', 'caching'],
  },
  'gemini-2.5-flash': {
    context_length: 1048576,
    max_output_tokens: 65536,
    knowledge_cutoff: '2025-01',
    release_date: '2025-04',
    parameter_count: 'Proprietary',
    capabilities: ['reasoning', 'function_calling', 'streaming', 'vision', 'json_mode', 'structured_output', 'tools', 'caching'],
  },
  'deepseek-r1': {
    context_length: 128000,
    max_output_tokens: 65536,
    knowledge_cutoff: '2024-07',
    release_date: '2025-01',
    parameter_count: '671B total / 37B active',
    capabilities: ['reasoning', 'streaming', 'json_mode', 'caching'],
  },
  'deepseek-v4-flash': {
    context_length: 131072,
    max_output_tokens: 32768,
    knowledge_cutoff: '2025-06',
    release_date: '2026-06',
    parameter_count: 'Open weights',
    capabilities: ['reasoning', 'streaming', 'json_mode', 'caching'],
  },
  'text-embedding-3-large': {
    context_length: 8191,
    max_output_tokens: 3072,
    release_date: '2024-01',
    parameter_count: 'Proprietary',
    capabilities: ['embeddings'],
  },
  'sora-2': {
    context_length: 100000,
    max_output_tokens: 10000,
    release_date: '2025-09',
    parameter_count: 'Proprietary',
    capabilities: ['streaming', 'vision'],
  },
  'seedream-4.0': {
    context_length: 12000,
    max_output_tokens: 1,
    release_date: '2025-09',
    parameter_count: 'Proprietary',
    capabilities: ['vision'],
  },
  'type-safe-demo': {
    context_length: 32768,
    max_output_tokens: 8192,
    knowledge_cutoff: '2025-12',
    release_date: '2026-06',
    parameter_count: 'Plugin-backed',
    capabilities: ['function_calling', 'streaming', 'structured_output'],
  },
  'jev-1.14': {
    context_length: 131072,
    max_output_tokens: 32768,
    knowledge_cutoff: '2025-12',
    release_date: '2026-09',
    parameter_count: 'Plugin-backed',
    capabilities: ['function_calling', 'streaming', 'structured_output'],
  },
}

const models = [
  {
    id: 1,
    model_name: 'gpt-4.1',
    description:
      'General purpose flagship model with strong coding and tool use.',
    vendor_id: 1,
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
    description:
      'Balanced Claude model for coding, agents, and long documents.',
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
    description:
      'High-capability Claude model for deep reasoning and difficult code.',
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
    description:
      'Low-latency Gemini model for high-volume application traffic.',
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
    description:
      'Reasoning model tuned for math, coding, and structured answers.',
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
    id: 12,
    model_name: 'deepseek-v4-flash',
    description:
      'Reasoning model with scheduled peak pricing for production traffic.',
    vendor_id: 4,
    quota_type: 0,
    model_ratio: 0.15,
    completion_ratio: 0.6,
    cache_ratio: 0.003,
    enable_groups: ['default', 'premium', 'research'],
    tags: 'reasoning,coding,peak-hours',
    supported_endpoint_types: ['openai'],
    billing_mode: 'tiered_expr',
    billing_expr:
      '(tier("base", p * 0.15 + c * 0.6 + cr * 0.003)) * (weekday("Asia/Shanghai") >= 1 && weekday("Asia/Shanghai") < 6 && hour("Asia/Shanghai") >= 9 && hour("Asia/Shanghai") < 12 ? 2 : 1) * (weekday("Asia/Shanghai") >= 1 && weekday("Asia/Shanghai") < 6 && hour("Asia/Shanghai") >= 14 && hour("Asia/Shanghai") < 18 ? 2 : 1)',
    pricing_version: 'mock-2026-06',
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
    billing_mode: 'tiered_expr',
    billing_expr:
      'param("resolution") == "1K" ? tier("1K", 68000) : param("resolution") == "2K" ? tier("2K", 75000) : param("resolution") == "4K" ? tier("4K", 80000) : tier("default", 80000)',
    pricing_version: 'mock-2026-06',
  },
  {
    id: 13,
    model_name: 'type-safe-demo',
    description:
      'Task billing demo with TypeSafe and Quality Check plugin pricing.',
    vendor_id: 6,
    quota_type: 1,
    model_ratio: 0,
    completion_ratio: 0,
    model_price: 0.02,
    enable_groups: ['default', 'premium'],
    tags: 'task,plugin,structured-output',
    supported_endpoint_types: ['openai'],
    billing_plugin_variants: [
      {
        plugin_key: 'type-safe',
        plugin_name: 'TypeSafe',
        billing_mode: 'tiered_expr',
        billing_expr:
          'tier("default", u("input_tokens") * 0.00001 + u("output_tokens") * 0.00002)',
        billing_usage_schema: {
          input_tokens: {
            type: 'number',
            unit: 'token',
            description: 'Input tokens',
          },
          output_tokens: {
            type: 'number',
            unit: 'token',
            description: 'Output tokens',
          },
        },
        billing_usage_examples: [
          {
            label: 'Standard validation',
            facts: { input_tokens: 1000, output_tokens: 500 },
          },
        ],
      },
      {
        plugin_key: 'quality-check',
        plugin_name: 'Quality Check',
        billing_mode: 'tiered_expr',
        billing_expr:
          'tier("default", u("input_tokens") * 0.000006 + u("output_tokens") * 0.000012)',
        billing_usage_schema: {
          input_tokens: {
            type: 'number',
            unit: 'token',
            description: 'Input tokens',
          },
          output_tokens: {
            type: 'number',
            unit: 'token',
            description: 'Output tokens',
          },
        },
        billing_usage_examples: [
          {
            label: 'Quality scan',
            facts: { input_tokens: 1000, output_tokens: 300 },
          },
        ],
      },
    ],
  },
  {
    id: 14,
    model_name: 'jev-1.14',
    description:
      'JEV 1.14 model with TypeSafe structured-output validation pricing.',
    vendor_id: 6,
    quota_type: 1,
    model_ratio: 0,
    completion_ratio: 0,
    model_price: 0.02,
    enable_groups: ['default', 'premium'],
    tags: 'task,plugin,structured-output,jev',
    supported_endpoint_types: ['openai'],
    billing_plugin_variants: [
      {
        plugin_key: 'type-safe',
        plugin_name: 'TypeSafe',
        billing_mode: 'tiered_expr',
        billing_expr:
          'tier("default", u("input_tokens") * 0.00001 + u("output_tokens") * 0.00002)',
        billing_usage_schema: {
          input_tokens: {
            type: 'number',
            unit: 'token',
            description: 'Input tokens',
          },
          output_tokens: {
            type: 'number',
            unit: 'token',
            description: 'Output tokens',
          },
        },
        billing_usage_examples: [
          {
            label: 'JEV validation',
            facts: { input_tokens: 1200, output_tokens: 450 },
          },
        ],
      },
    ],
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
  }
})

models.push(...mockVariants)

const baseCreatedTime = Math.floor(
  new Date('2026-06-16T09:00:00+08:00').getTime() / 1000
)

for (const [index, model] of models.entries()) {
  const endpoints = model.supported_endpoint_types || []
  const tags = String(model.tags || '').split(/[,;|\s]+/)
  const baseModelName = model.model_name.replace(/-mock-\d+$/, '')
  const metadata = modelMetadata[baseModelName] || {}
  const listedVendors = availableVendorsByModel[baseModelName] || [model.vendor_id]

  model.created_time = baseCreatedTime - index * 60 * 60
  model.context_length = metadata.context_length ?? 128000
  model.context_length_display = model.context_length.toLocaleString()
  model.max_output_tokens = metadata.max_output_tokens ?? 16384
  model.max_output_tokens_display = model.max_output_tokens.toLocaleString()
  model.knowledge_cutoff = metadata.knowledge_cutoff ?? '2025-01'
  model.release_date = metadata.release_date ?? '2025-06'
  model.parameter_count = metadata.parameter_count ?? 'Open weights'
  model.capabilities = metadata.capabilities ?? ['streaming', 'function_calling']
  model.available_vendors = listedVendors
    .filter((vendorId) => vendorId && vendors.some((vendor) => vendor.id === vendorId))
    .map((vendorId) => vendors.find((vendor) => vendor.id === vendorId))
  model.rate_limits = {
    default: { rpm: 60, tpm: 120000, rpd: 10000 },
    premium: { rpm: 120, tpm: 300000, rpd: 30000 },
    research: { rpm: 30, tpm: 60000, rpd: 3000 },
  }
  model.model_group_ratio = {
    default: 1,
    premium:
      model.model_name.includes('mini') || model.model_name.includes('flash')
        ? 0.85
        : 1.1,
    research: 0.7,
  }

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
    model.input_modalities = tags.includes('vision')
      ? ['text', 'image']
      : ['text']
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
    group_model_ratio: Object.fromEntries(
      Object.entries(usableGroup).map(([group]) => [
        group,
        Object.fromEntries(
          models.map((model) => [
            model.model_name,
            model.model_group_ratio?.[group] ?? 1,
          ])
        ),
      ])
    ),
    group_model_ratio_expiry: {},
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
