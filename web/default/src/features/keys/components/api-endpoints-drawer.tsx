import { ExternalLink, Gauge, Route, Server } from 'lucide-react'
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
      { id: 'models', method: 'GET', path: '/v1/models', label: 'List Models' },
    ],
  },
  {
    id: 'web-tools',
    label: 'Web Tools',
    description: 'Web search and page fetch',
    endpoints: [
      {
        id: 'tools-search',
        method: 'POST',
        path: '/v1/tools/search',
        label: 'Search the Web',
      },
      {
        id: 'tools-fetch',
        method: 'POST',
        path: '/v1/tools/fetch',
        label: 'Fetch a Web Page',
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
      {
        id: 'models',
        method: 'GET',
        path: '/v1beta/models',
        label: 'List Models',
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
  const [selectedEndpointIds, setSelectedEndpointIds] = useState<
    Record<string, string>
  >(() =>
    Object.fromEntries(
      PROTOCOLS.map((protocol) => [protocol.id, protocol.endpoints[0].id])
    )
  )
  const [activeEndpointKey, setActiveEndpointKey] = useState(
    `${PROTOCOLS[0].id}:${PROTOCOLS[0].endpoints[0].id}`
  )
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

  const [selectedProtocolId] = activeEndpointKey.split(':')
  const selectedProtocol =
    PROTOCOLS.find((protocol) => protocol.id === selectedProtocolId) ??
    PROTOCOLS[0]
  const selectedNode = nodes.find((node) => node.url === nodeUrl) ?? nodes[0]

  function getProtocolEndpoint(protocol: Protocol): Endpoint {
    const endpointId = selectedEndpointIds[protocol.id]
    return (
      protocol.endpoints.find((endpoint) => endpoint.id === endpointId) ??
      protocol.endpoints[0]
    )
  }

  const selectedEndpoint = getProtocolEndpoint(selectedProtocol)

  function getProtocolPrefix(protocol: Protocol): string {
    return protocol.id === 'gemini' ? '/v1beta' : '/v1'
  }

  function getProtocolBaseUrl(protocol: Protocol): string {
    if (!selectedNode) return getProtocolPrefix(protocol)
    const root = normalizeBaseUrl(selectedNode.url).replace(
      /\/(?:v1beta|v1)$/i,
      ''
    )
    return `${root}${getProtocolPrefix(protocol)}`
  }

  function getDisplayPath(protocol: Protocol, path: string): string {
    const prefix = getProtocolPrefix(protocol)
    return path.startsWith(prefix) ? path.slice(prefix.length) || '/' : path
  }

  const selectedBaseUrl = getProtocolBaseUrl(selectedProtocol)
  const selectedUrl = `${selectedBaseUrl}${getDisplayPath(selectedProtocol, selectedEndpoint.path)}`

  function selectEndpoint(protocol: Protocol, endpoint: Endpoint) {
    setSelectedEndpointIds((current) => ({
      ...current,
      [protocol.id]: endpoint.id,
    }))
    setActiveEndpointKey(`${protocol.id}:${endpoint.id}`)
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
            {t('API Access Points')}
          </SheetTitle>
          <SheetDescription className='flex flex-wrap items-center gap-2'>
            {t('Supported API protocols and endpoints')}
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
          </SheetDescription>
        </SheetHeader>

        <ScrollArea className='min-h-0 flex-1'>
          <div className='space-y-5 px-4 pb-6 sm:px-6'>
            {nodes.length > 1 && (
              <div className='bg-muted/20 flex items-center gap-3 rounded-xl border px-4 py-3'>
                <Server className='text-muted-foreground size-4 shrink-0' />
                <div className='min-w-0 flex-1'>
                  <div className='text-xs font-medium'>{t('Server Node')}</div>
                  <div className='text-muted-foreground truncate text-xs'>
                    {selectedNode?.description || selectedNode?.url}
                  </div>
                </div>
                <Select
                  value={selectedNode?.url ?? ''}
                  onValueChange={(value) => {
                    setNodeUrl(value ?? '')
                    setPingStatus(getDefaultPingStatus())
                  }}
                >
                  <SelectTrigger
                    className='w-44 min-w-0 shrink-0'
                    title={selectedNode?.url}
                  >
                    <SelectValue placeholder={t('Select Server Node')}>
                      <span className='min-w-0 truncate'>
                        {selectedNode?.route || selectedNode?.url}
                      </span>
                    </SelectValue>
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
                          <span className='min-w-0'>
                            <span className='block truncate'>{node.route}</span>
                            <span className='text-muted-foreground block truncate text-xs'>
                              {node.url}
                            </span>
                          </span>
                        </span>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}

            {PROTOCOLS.map((protocol) => {
              const protocolBaseUrl = getProtocolBaseUrl(protocol)
              const protocolEndpoint = getProtocolEndpoint(protocol)
              const protocolApiUrl = `${protocolBaseUrl}${getDisplayPath(protocol, protocolEndpoint.path)}`
              return (
                <section
                  key={protocol.id}
                  className='bg-card rounded-xl border p-4 shadow-xs'
                >
                  <div className='flex items-center gap-3'>
                    <div className='bg-muted flex size-10 shrink-0 items-center justify-center rounded-full'>
                      <Route className='text-muted-foreground size-5' />
                    </div>
                    <div className='min-w-0'>
                      <h3 className='font-medium'>{t(protocol.label)}</h3>
                      <p className='text-muted-foreground text-xs'>
                        {t(protocol.description)}
                      </p>
                    </div>
                  </div>

                  <div className='my-4 border-t border-dashed' />

                  <div className='space-y-2'>
                    <div className='text-muted-foreground text-xs'>
                      {t('API Address')}
                    </div>
                    <div className='bg-muted/30 flex items-center gap-2 rounded-lg border px-3 py-2'>
                      <code className='min-w-0 flex-1 text-xs break-all'>
                        {protocolApiUrl}
                      </code>
                      <CopyButton
                        value={protocolApiUrl}
                        size='sm'
                        className='size-7 shrink-0 p-0'
                        iconClassName='size-3.5'
                        tooltip={t('Copy endpoint URL')}
                        aria-label={t('Copy endpoint URL')}
                      />
                    </div>
                  </div>

                  <div className='mt-5 space-y-2'>
                    <div className='text-muted-foreground text-xs'>
                      {t('Endpoints')}
                    </div>
                    <div className='grid gap-1 sm:grid-cols-2'>
                      {protocol.endpoints.map((endpoint) => {
                        const isSelected = endpoint.id === protocolEndpoint.id
                        return (
                          <button
                            key={endpoint.id}
                            type='button'
                            className={cn(
                              'flex min-w-0 items-start gap-3 rounded-lg p-2 text-left transition-colors hover:bg-muted/60',
                              isSelected &&
                                'bg-primary/8 ring-primary/30 ring-1'
                            )}
                            onClick={() => selectEndpoint(protocol, endpoint)}
                          >
                            <span className='bg-muted flex size-9 shrink-0 items-center justify-center rounded-lg'>
                              <Route className='text-muted-foreground size-4' />
                            </span>
                            <span className='min-w-0 pt-0.5'>
                              <span className='flex items-center gap-2'>
                                <span className='truncate text-xs font-medium'>
                                  {t(endpoint.label)}
                                </span>
                                <Badge
                                  variant='outline'
                                  className='shrink-0 px-1.5 font-mono text-[9px]'
                                >
                                  {endpoint.method}
                                </Badge>
                              </span>
                              <code className='text-muted-foreground mt-0.5 block truncate text-[11px]'>
                                {getDisplayPath(protocol, endpoint.path)}
                              </code>
                            </span>
                          </button>
                        )
                      })}
                    </div>
                  </div>
                </section>
              )
            })}

            <div className='rounded-xl border border-dashed p-4 text-sm'>
              <div className='flex flex-wrap items-center justify-between gap-3'>
                <div className='font-medium'>{t('Selected Endpoint')}</div>
                <div className='flex flex-wrap gap-2'>
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
                  <Button
                    variant='link'
                    size='sm'
                    className='h-auto px-0'
                    render={
                      <a
                        href={docsLink}
                        target={
                          docsLink.startsWith('http') ? '_blank' : undefined
                        }
                        rel='noreferrer'
                      />
                    }
                  >
                    {t('View Docs')}
                  </Button>
                </div>
              </div>
              <p className='text-muted-foreground mt-1 text-xs'>
                {t(
                  'Use the selected endpoint with your API key in the Authorization header.'
                )}
              </p>
              {pingStatus.latency !== null && !pingStatus.testing && (
                <Badge variant='secondary' className='mt-3 font-mono'>
                  {pingStatus.latency}
                  {t('ms')}
                </Badge>
              )}
              {pingStatus.error && (
                <Badge variant='outline' className='mt-3'>
                  {t('N/A')}
                </Badge>
              )}
            </div>
          </div>
        </ScrollArea>
      </SheetContent>
    </Sheet>
  )
}
