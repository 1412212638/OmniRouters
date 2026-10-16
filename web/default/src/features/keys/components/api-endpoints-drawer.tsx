import { ExternalLink, Gauge, Route, Server, Waypoints } from 'lucide-react'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { CopyButton } from '@/components/copy-button'
import {
  sideDrawerContentClassName,
  sideDrawerHeaderClassName,
} from '@/components/drawer-layout'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { useApiInfo } from '@/features/dashboard/hooks/use-status-data'
import {
  getDefaultPingStatus,
  testUrlLatency,
} from '@/features/dashboard/lib/api-info'
import type { ApiInfoItem, PingStatus } from '@/features/dashboard/types'
import { useStatus } from '@/hooks/use-status'
import { getBgColorClass } from '@/lib/colors'
import { cn } from '@/lib/utils'

type Endpoint = {
  id: string
  method: 'GET' | 'POST'
  path: string
  label: string
  description?: string
}

type Protocol = {
  id: string
  label: string
  description: string
  endpoints: Endpoint[]
}

const PROTOCOLS: Protocol[] = [
  {
    id: 'openai',
    label: 'OpenAI Compatible',
    description: 'OpenAI-compatible APIs',
    endpoints: [
      { id: 'models', method: 'GET', path: '/v1/models', label: 'List Models' },
      {
        id: 'chat-completions',
        method: 'POST',
        path: '/v1/chat/completions',
        label: 'Create a Chat Completion',
      },
      {
        id: 'completions',
        method: 'POST',
        path: '/v1/completions',
        label: 'Create a Completion',
      },
      {
        id: 'responses',
        method: 'POST',
        path: '/v1/responses',
        label: 'Create a Model Response',
      },
      {
        id: 'embeddings',
        method: 'POST',
        path: '/v1/embeddings',
        label: 'Create an Embedding',
      },
      {
        id: 'images-generations',
        method: 'POST',
        path: '/v1/images/generations',
        label: 'Create an Image',
      },
      {
        id: 'images-edits',
        method: 'POST',
        path: '/v1/images/edits',
        label: 'Edit an Image',
      },
      {
        id: 'audio-speech',
        method: 'POST',
        path: '/v1/audio/speech',
        label: 'Create Speech',
      },
      {
        id: 'audio-transcriptions',
        method: 'POST',
        path: '/v1/audio/transcriptions',
        label: 'Create Transcription',
      },
      {
        id: 'audio-translations',
        method: 'POST',
        path: '/v1/audio/translations',
        label: 'Create Translation',
      },
      {
        id: 'rerank',
        method: 'POST',
        path: '/v1/rerank',
        label: 'Create Rerank',
      },
      {
        id: 'moderations',
        method: 'POST',
        path: '/v1/moderations',
        label: 'Create Moderation',
      },
    ],
  },
  {
    id: 'anthropic',
    label: 'Anthropic',
    description: 'Anthropic Messages',
    endpoints: [
      {
        id: 'messages',
        method: 'POST',
        path: '/v1/messages',
        label: 'Create a Message',
      },
      { id: 'models', method: 'GET', path: '/v1/models', label: 'List Models' },
    ],
  },
  {
    id: 'gemini',
    label: 'Google Gemini',
    description: 'Gemini-compatible APIs',
    endpoints: [
      {
        id: 'models',
        method: 'GET',
        path: '/v1beta/models',
        label: 'List Models',
      },
      {
        id: 'generate-content',
        method: 'POST',
        path: '/v1beta/models/{model}:generateContent',
        label: 'Generate Content',
      },
      {
        id: 'stream-generate-content',
        method: 'POST',
        path: '/v1beta/models/{model}:streamGenerateContent',
        label: 'Stream Generate Content',
      },
    ],
  },
  {
    id: 'decision',
    label: 'Decision Model',
    description: 'TypeSafe task plugin',
    endpoints: [
      {
        id: 'systemone',
        method: 'POST',
        path: '/v1/systemone',
        label: 'Decision Model',
        description: 'Requires the TypeSafe plugin',
      },
    ],
  },
]

type ApiEndpointsDrawerProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
}

function normalizeBaseUrl(url: string): string {
  return url.trim().replace(/\/+$/, '')
}

function buildEndpointUrl(baseUrl: string, path: string): string {
  const base = normalizeBaseUrl(baseUrl)
  for (const prefix of ['/v1beta', '/v1']) {
    if (base.endsWith(prefix) && path.startsWith(`${prefix}/`)) {
      return `${base.slice(0, -prefix.length)}${path}`
    }
  }
  return `${base}${path}`
}

function getStatusServerAddress(
  status: Record<string, unknown> | null
): string {
  const data = status?.data as Record<string, unknown> | undefined
  const candidate =
    status?.server_address ??
    status?.serverAddress ??
    data?.server_address ??
    data?.serverAddress
  if (typeof candidate === 'string' && candidate.trim()) {
    return normalizeBaseUrl(candidate)
  }
  if (typeof window !== 'undefined') {
    return window.location.origin
  }
  return ''
}

export function ApiEndpointsDrawer({
  open,
  onOpenChange,
}: ApiEndpointsDrawerProps) {
  const { t } = useTranslation()
  const { status } = useStatus()
  const { items: configuredNodes } = useApiInfo()
  const docsLink =
    typeof status?.docs_link === 'string' && status.docs_link.trim()
      ? status.docs_link
      : '/docs'
  const [protocolId, setProtocolId] = useState(PROTOCOLS[0].id)
  const [endpointId, setEndpointId] = useState(PROTOCOLS[0].endpoints[0].id)
  const [nodeUrl, setNodeUrl] = useState('')
  const [pingStatus, setPingStatus] = useState<PingStatus>(
    getDefaultPingStatus()
  )

  const nodes = useMemo<ApiInfoItem[]>(() => {
    if (configuredNodes.length > 0) return configuredNodes
    const fallback = getStatusServerAddress(
      status as Record<string, unknown> | null
    )
    return fallback
      ? [
          {
            url: fallback,
            route: t('Current domain'),
            description: t('Current server address'),
            color: 'blue',
          },
        ]
      : []
  }, [configuredNodes, status, t])

  const selectedProtocol =
    PROTOCOLS.find((protocol) => protocol.id === protocolId) ?? PROTOCOLS[0]
  const selectedEndpoint =
    selectedProtocol.endpoints.find((endpoint) => endpoint.id === endpointId) ??
    selectedProtocol.endpoints[0]
  const selectedNode = nodes.find((node) => node.url === nodeUrl) ?? nodes[0]
  const selectedUrl = selectedNode
    ? buildEndpointUrl(selectedNode.url, selectedEndpoint.path)
    : selectedEndpoint.path

  function handleProtocolChange(value: string | null) {
    if (!value) return
    const nextProtocol = PROTOCOLS.find((protocol) => protocol.id === value)
    setProtocolId(value)
    setEndpointId(nextProtocol?.endpoints[0]?.id ?? '')
    setPingStatus(getDefaultPingStatus())
  }

  async function handleLatencyCheck() {
    setPingStatus({ latency: null, testing: true, error: false })
    setPingStatus(await testUrlLatency(selectedUrl))
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent
        className={sideDrawerContentClassName('w-full sm:max-w-xl')}
      >
        <SheetHeader className={sideDrawerHeaderClassName()}>
          <SheetTitle className='flex items-center gap-2'>
            <Route className='text-muted-foreground size-4' />
            {t('API Endpoints')}
          </SheetTitle>
          <SheetDescription>
            {t('Choose a server node and endpoint to generate a request URL.')}
          </SheetDescription>
        </SheetHeader>

        <ScrollArea className='min-h-0 flex-1'>
          <div className='space-y-5 px-4 pb-6 sm:px-6'>
            <div className='bg-muted/20 rounded-xl border p-4'>
              <div className='mb-3 flex items-center gap-2 text-sm font-medium'>
                <Server className='text-muted-foreground size-4' />
                {t('Server Node')}
              </div>
              <Select
                value={selectedNode?.url ?? ''}
                onValueChange={(value) => setNodeUrl(value ?? '')}
              >
                <SelectTrigger className='w-full'>
                  <SelectValue placeholder={t('Select Server Node')} />
                </SelectTrigger>
                <SelectContent>
                  {nodes.map((node) => (
                    <SelectItem key={node.url} value={node.url}>
                      <span className='flex min-w-0 items-center gap-2'>
                        <span
                          className={cn(
                            'size-2 rounded-full',
                            getBgColorClass(node.color)
                          )}
                        />
                        <span className='truncate'>{node.route}</span>
                        <span className='text-muted-foreground truncate text-xs'>
                          {node.url}
                        </span>
                      </span>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className='grid gap-4 sm:grid-cols-2'>
              <div className='space-y-2'>
                <div className='text-sm font-medium'>{t('Protocol')}</div>
                <Select value={protocolId} onValueChange={handleProtocolChange}>
                  <SelectTrigger className='w-full'>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {PROTOCOLS.map((protocol) => (
                      <SelectItem key={protocol.id} value={protocol.id}>
                        {t(protocol.label)}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className='space-y-2'>
                <div className='text-sm font-medium'>{t('Endpoint')}</div>
                <Select
                  value={selectedEndpoint.id}
                  onValueChange={(value) => {
                    if (value) {
                      setEndpointId(value)
                      setPingStatus(getDefaultPingStatus())
                    }
                  }}
                >
                  <SelectTrigger className='w-full'>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {selectedProtocol.endpoints.map((endpoint) => (
                      <SelectItem key={endpoint.id} value={endpoint.id}>
                        <span className='flex items-center gap-2'>
                          <Badge
                            variant='outline'
                            className='font-mono text-[10px]'
                          >
                            {endpoint.method}
                          </Badge>
                          {t(endpoint.label)}
                        </span>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className='bg-card rounded-xl border p-4 shadow-xs'>
              <div className='mb-3 flex items-center justify-between gap-3'>
                <div className='flex min-w-0 items-center gap-2'>
                  <Waypoints className='text-muted-foreground size-4' />
                  <span className='truncate font-medium'>
                    {t(selectedProtocol.label)}
                  </span>
                </div>
                <Badge
                  variant='secondary'
                  className='shrink-0 font-mono text-[10px]'
                >
                  {selectedEndpoint.method}
                </Badge>
              </div>
              <div className='space-y-2'>
                <div className='text-muted-foreground text-xs'>
                  {t('Generated URL')}
                </div>
                <div className='bg-muted/30 flex items-center gap-2 rounded-lg border px-3 py-2'>
                  <code className='min-w-0 flex-1 text-xs break-all'>
                    {selectedUrl}
                  </code>
                  <CopyButton
                    value={selectedUrl}
                    size='sm'
                    className='size-7 p-0'
                    iconClassName='size-3.5'
                    tooltip={t('Copy endpoint URL')}
                    aria-label={t('Copy endpoint URL')}
                  />
                </div>
                <div className='text-muted-foreground flex items-center gap-2 text-xs'>
                  <span>{t(selectedEndpoint.label)}</span>
                  <span aria-hidden='true'>·</span>
                  <code>{selectedEndpoint.path}</code>
                </div>
                {selectedEndpoint.description && (
                  <p className='text-muted-foreground text-xs'>
                    {t(selectedEndpoint.description)}
                  </p>
                )}
              </div>
              <div className='mt-4 flex flex-wrap gap-2'>
                <CopyButton
                  value={selectedNode?.url ?? ''}
                  variant='outline'
                  size='sm'
                  tooltip={t('Copy base URL')}
                >
                  {t('Copy Base URL')}
                </CopyButton>
                <Button
                  variant='outline'
                  size='sm'
                  onClick={handleLatencyCheck}
                  disabled={!selectedUrl || pingStatus.testing}
                >
                  <Gauge
                    className={cn(
                      'size-4',
                      pingStatus.testing && 'animate-pulse'
                    )}
                  />
                  {pingStatus.testing ? t('Testing...') : t('Check Latency')}
                </Button>
                <Button
                  variant='outline'
                  size='sm'
                  render={
                    <a href={selectedUrl} target='_blank' rel='noreferrer' />
                  }
                >
                  <ExternalLink className='size-4' />
                  {t('Open Endpoint')}
                </Button>
                {pingStatus.latency !== null && !pingStatus.testing && (
                  <Badge variant='secondary' className='font-mono'>
                    {pingStatus.latency}
                    {t('ms')}
                  </Badge>
                )}
                {pingStatus.error && (
                  <Badge variant='outline'>{t('N/A')}</Badge>
                )}
              </div>
            </div>

            <div className='rounded-xl border border-dashed p-4 text-sm'>
              <div className='flex items-center justify-between gap-3'>
                <div className='font-medium'>{t('API documentation')}</div>
                <Button
                  variant='link'
                  size='sm'
                  className='h-auto px-0'
                  render={
                    <a
                      href={docsLink}
                      target={docsLink.startsWith('http') ? '_blank' : undefined}
                      rel='noreferrer'
                    />
                  }
                >
                  {t('View Docs')}
                </Button>
              </div>
              <p className='text-muted-foreground mt-1 text-xs'>
                {t(
                  'Use the selected endpoint with your API key in the Authorization header.'
                )}
              </p>
            </div>
          </div>
        </ScrollArea>
      </SheetContent>
    </Sheet>
  )
}
