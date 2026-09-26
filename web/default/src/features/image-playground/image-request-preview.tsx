import { Copy, Eye, Terminal } from 'lucide-react'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import {
  CodeBlock,
  CodeBlockCopyButton,
} from '@/components/ai-elements/code-block'
import {
  PromptInputButton,
  usePromptInputAttachments,
} from '@/components/ai-elements/prompt-input'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'

import type { ImageRequestPayload } from './api'
import { imageRequestCurl, imageRequestJson } from './image-request-builder'

type ImageRequestPreviewProps = {
  buildRequest: (image?: string) => ImageRequestPayload
  isEdit: boolean
  isGeminiEdit: boolean
  disabled?: boolean
}

export function ImageRequestPreview({
  buildRequest,
  isEdit,
  isGeminiEdit,
  disabled,
}: ImageRequestPreviewProps) {
  const { t } = useTranslation()
  const { files } = usePromptInputAttachments()
  const [tab, setTab] = useState<'json' | 'curl'>('json')
  const attachmentImage = files.at(0)?.url
  const request = buildRequest(attachmentImage)
  const editRequest = isEdit || Boolean(attachmentImage)
  const endpoint = editRequest ? '/pg/images/edits' : '/pg/images/generations'
  const code = useMemo(() => {
    if (tab === 'curl') {
      return imageRequestCurl(request, endpoint, editRequest && !isGeminiEdit)
    }
    return imageRequestJson(request)
  }, [endpoint, editRequest, isGeminiEdit, request, tab])

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(code)
      toast.success(t('Copied'))
    } catch {
      toast.error(t('Failed to copy'))
    }
  }

  return (
    <Dialog>
      <Tooltip>
        <TooltipTrigger
          render={
            <DialogTrigger
              render={
                <PromptInputButton
                  aria-label={t('Request preview')}
                  className='text-muted-foreground hover:bg-muted/70 hover:text-foreground font-medium'
                  disabled={disabled}
                  variant='ghost'
                />
              }
            >
              <Eye size={16} />
            </DialogTrigger>
          }
        />
        <TooltipContent>
          <p>{t('Request preview')}</p>
        </TooltipContent>
      </Tooltip>
      <DialogContent className='max-h-[90dvh] overflow-y-auto sm:max-w-3xl'>
        <DialogHeader>
          <DialogTitle>{t('Request preview')}</DialogTitle>
          <DialogDescription>
            {t('Preview of the request sent to the image playground endpoint.')}
          </DialogDescription>
        </DialogHeader>
        <div className='flex flex-wrap items-center justify-between gap-3'>
          <div className='text-muted-foreground flex items-center gap-2 text-sm'>
            {tab === 'curl' ? (
              <Terminal className='size-4' />
            ) : (
              <span className='font-mono text-xs'>POST</span>
            )}
            <code className='break-all'>{endpoint}</code>
          </div>
          <Tabs
            value={tab}
            onValueChange={(value) => setTab(value as typeof tab)}
          >
            <TabsList className='h-8'>
              <TabsTrigger className='h-7 px-3 text-xs' value='json'>
                JSON
              </TabsTrigger>
              <TabsTrigger className='h-7 px-3 text-xs' value='curl'>
                cURL
              </TabsTrigger>
            </TabsList>
          </Tabs>
        </div>
        <CodeBlock
          code={code}
          language={tab === 'json' ? 'json' : 'bash'}
          maxExpandedLines={28}
          title={tab.toUpperCase()}
        >
          <CodeBlockCopyButton />
        </CodeBlock>
        <div className='flex justify-end'>
          <Button
            className='gap-2'
            onClick={copy}
            size='sm'
            type='button'
            variant='outline'
          >
            <Copy className='size-4' />
            {t('Copy')}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
