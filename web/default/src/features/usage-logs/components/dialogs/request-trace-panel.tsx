import { EyeOff } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import {
  CodeBlock,
  CodeBlockCopyButton,
} from '@/components/ai-elements/code-block'
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '@/components/ui/empty'

import type { UsageLog } from '../../data/schema'
import type { LogOtherData } from '../../types'

type CallIOLog = LogOtherData['call_io_log']

function TraceJSON(props: { data: Record<string, unknown>; title: string }) {
  return (
    <CodeBlock
      code={JSON.stringify(props.data, null, 2)}
      language='json'
      title={props.title}
      className='my-0'
      enableCollapse={false}
      maxExpandedLines={20}
      showLineNumbers
    >
      <CodeBlockCopyButton />
    </CodeBlock>
  )
}

function CallJSONSection(props: {
  direction: 'input' | 'output'
  callIO: CallIOLog
}) {
  const { t } = useTranslation()
  const isInput = props.direction === 'input'
  const title = isInput ? t('Input JSON') : t('Output JSON')
  const payload = props.callIO?.enabled
    ? props.callIO[props.direction]
    : undefined
  const isDisabled = props.callIO?.enabled === false

  let message = t('No input or output was recorded for this request.')
  if (isDisabled) {
    message = t('Input and output logging was disabled for this request.')
  } else if (props.callIO?.enabled) {
    message = isInput
      ? t('No input summary is available for this request.')
      : t('No output summary is available for this request.')
  }

  return (
    <AccordionItem
      value={props.direction}
      className='bg-card overflow-hidden rounded-lg border'
    >
      <AccordionTrigger className='items-center px-3 py-3 focus-visible:ring-inset sm:px-4'>
        {title}
      </AccordionTrigger>
      <AccordionContent className='border-t px-3 pt-3 pb-3 sm:px-4'>
        {payload ? (
          <TraceJSON title={title} data={payload} />
        ) : (
          <Empty className='bg-muted/55 min-h-36'>
            <EmptyHeader className='max-w-md'>
              <EmptyMedia
                variant='icon'
                className='bg-background text-muted-foreground rounded-full'
              >
                <EyeOff aria-hidden='true' />
              </EmptyMedia>
              <EmptyTitle>{message}</EmptyTitle>
              {isDisabled && (
                <EmptyDescription className='text-xs'>
                  {t(
                    'Enable Input and Output Logging in your profile to see JSON for future requests.'
                  )}
                </EmptyDescription>
              )}
            </EmptyHeader>
          </Empty>
        )}
      </AccordionContent>
    </AccordionItem>
  )
}

export function RequestTracePanel(props: { log: UsageLog; callIO: CallIOLog }) {
  const { t } = useTranslation()
  const statusCode = props.callIO?.output?.status_code
  const metadata = {
    request_id: props.log.request_id || undefined,
    model: props.log.model_name || undefined,
    created_at: new Date(props.log.created_at * 1000).toJSON() ?? undefined,
    token_name: props.log.token_name || undefined,
    stream: props.log.is_stream,
    status_code: typeof statusCode === 'number' ? statusCode : undefined,
    response_time_seconds:
      props.log.use_time > 0 ? props.log.use_time : undefined,
  }

  return (
    <div className='flex min-w-0 flex-col gap-3 sm:gap-4'>
      <Accordion
        multiple
        defaultValue={['input', 'output']}
        className='gap-3 sm:gap-4'
      >
        <CallJSONSection direction='input' callIO={props.callIO} />
        <CallJSONSection direction='output' callIO={props.callIO} />
      </Accordion>
      <Card size='sm' className='rounded-lg'>
        <CardHeader>
          <CardTitle>{t('Request Metadata')}</CardTitle>
        </CardHeader>
        <CardContent>
          <TraceJSON title={t('Request Metadata')} data={metadata} />
        </CardContent>
      </Card>
    </div>
  )
}
