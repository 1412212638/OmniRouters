import {
  Download,
  ImageIcon,
  LoaderCircle,
  PaperclipIcon,
  SendIcon,
  Trash2,
  Trash2Icon,
  X,
} from 'lucide-react'
import { nanoid } from 'nanoid'
import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import {
  Conversation,
  ConversationContent,
} from '@/components/ai-elements/conversation'
import {
  PromptInput,
  PromptInputActionAddAttachments,
  PromptInputAttachments,
  PromptInputAttachment,
  PromptInputButton,
  PromptInputFooter,
  PromptInputHeader,
  PromptInputTextarea,
  PromptInputTools,
  usePromptInputAttachments,
} from '@/components/ai-elements/prompt-input'
import { ModelGroupSelector } from '@/components/model-group-selector'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'

import {
  generateImages,
  editGeminiImage,
  editImage,
  getImageGroups,
  getImageModels,
  type ImageGroupOption,
  type ImageModelOption,
} from './api'
import { ImageParameterSettings } from './image-parameter-settings'

type GeneratedImage = {
  key: string
  src: string
  mimeType?: string
  revisedPrompt?: string
  prompt: string
}

type ImageSource = {
  src: string
  mimeType?: string
}

const FALLBACK_GROUP: ImageGroupOption = {
  label: 'default',
  value: 'default',
}

function imageSrc(item: {
  b64_json?: string
  url?: string
  mime_type?: string
}): ImageSource | null {
  if (item.url) return { src: item.url, mimeType: item.mime_type }
  if (item.b64_json) {
    const mimeType = item.mime_type || 'image/png'
    return {
      src: `data:${mimeType};base64,${item.b64_json}`,
      mimeType,
    }
  }
  return null
}

function downloadImage(src: string, index: number, mimeType?: string) {
  const anchor = document.createElement('a')
  anchor.href = src
  const extension = mimeType?.split('/')[1]?.split(';')[0] || 'png'
  anchor.download = `omnirouters-image-${index + 1}.${extension}`
  anchor.target = '_blank'
  anchor.rel = 'noreferrer'
  anchor.click()
}

function requestErrorMessage(error: unknown, fallback: string): string {
  if (error && typeof error === 'object') {
    const requestError = error as {
      message?: string
      response?: {
        data?: {
          error?: { message?: string }
          message?: string
        }
      }
    }
    return (
      requestError.response?.data?.error?.message ||
      requestError.response?.data?.message ||
      requestError.message ||
      fallback
    )
  }
  return fallback
}

function ImagePromptInputHeader({
  referenceImage,
  onClearReference,
}: {
  referenceImage: GeneratedImage | null
  onClearReference: () => void
}) {
  const { t } = useTranslation()
  const { files } = usePromptInputAttachments()

  if (!referenceImage && files.length === 0) return null

  return (
    <PromptInputHeader className='border-border/60 bg-muted/20 flex-wrap px-3 py-2'>
      {referenceImage && (
        <div className='flex items-center gap-2'>
          <img
            src={referenceImage.src}
            alt={t('Reference image')}
            className='size-10 rounded-md border object-cover'
          />
          <span className='text-muted-foreground text-xs'>
            {t('Editing image')}
          </span>
          <PromptInputButton
            aria-label={t('Remove reference image')}
            className='text-muted-foreground hover:text-foreground'
            onClick={onClearReference}
          >
            <X className='size-4' />
          </PromptInputButton>
        </div>
      )}
      <PromptInputAttachments>
        {(attachment) => <PromptInputAttachment data={attachment} />}
      </PromptInputAttachments>
    </PromptInputHeader>
  )
}

export function ImagePlayground() {
  const { t } = useTranslation()
  const [prompt, setPrompt] = useState('')
  const [model, setModel] = useState('')
  const [group, setGroup] = useState('')
  const [count, setCount] = useState('1')
  const [size, setSize] = useState('1024x1024')
  const [quality, setQuality] = useState('auto')
  const [models, setModels] = useState<ImageModelOption[]>([])
  const [groups, setGroups] = useState<ImageGroupOption[]>([FALLBACK_GROUP])
  const [images, setImages] = useState<GeneratedImage[]>([])
  const [referenceImageKey, setReferenceImageKey] = useState<string | null>(
    null
  )
  const [isLoadingOptions, setIsLoadingOptions] = useState(true)
  const [isGenerating, setIsGenerating] = useState(false)

  const selectedModel = useMemo(
    () => model || models[0]?.value || '',
    [model, models]
  )
  const selectedGroup = useMemo(
    () => group || groups[0]?.value || '',
    [group, groups]
  )
  const referenceImage = useMemo(
    () => images.find((image) => image.key === referenceImageKey) ?? null,
    [images, referenceImageKey]
  )
  const maxImageCount =
    selectedModel.toLowerCase().startsWith('gemini-') ||
    selectedModel.toLowerCase().startsWith('nano-banana')
      ? 1
      : 4

  useEffect(() => {
    let cancelled = false
    void getImageGroups()
      .then((loadedGroups) => {
        if (cancelled) return
        if (loadedGroups.length) {
          setGroups(loadedGroups)
          setGroup((current) => current || loadedGroups[0].value)
        }
      })
      .catch(() => {
        toast.error(t('Failed to load image groups'))
      })
      .finally(() => {
        if (!cancelled) setIsLoadingOptions(false)
      })
    return () => {
      cancelled = true
    }
  }, [t])

  useEffect(() => {
    if (!selectedGroup) return
    let cancelled = false
    void getImageModels(selectedGroup)
      .then((loadedModels) => {
        if (cancelled) return
        setModels(loadedModels)
        setModel((current) =>
          loadedModels.some((item) => item.value === current)
            ? current
            : (loadedModels[0]?.value ?? '')
        )
      })
      .catch(() => {
        if (!cancelled) {
          setModels([])
          setModel('')
          toast.error(t('Failed to load image models'))
        }
      })
      .finally(() => {
        if (!cancelled) setIsLoadingOptions(false)
      })
    return () => {
      cancelled = true
    }
  }, [selectedGroup, t])

  const handleGenerate = async (
    promptValue = prompt,
    attachedImage?: string
  ) => {
    const trimmedPrompt = promptValue.trim()
    if (!trimmedPrompt) {
      toast.error(t('Prompt is required'))
      return
    }
    if (!selectedModel) {
      toast.error(t('Select a model'))
      return
    }
    setIsGenerating(true)
    try {
      const request = {
        model: selectedModel,
        group: selectedGroup || undefined,
        prompt: trimmedPrompt,
        n: Math.min(maxImageCount, Math.max(1, Number(count) || 1)),
        size,
        quality,
      }
      const isGeminiImageModel =
        selectedModel.toLowerCase().startsWith('gemini-') ||
        selectedModel.toLowerCase().startsWith('nano-banana')
      let response
      const referenceSource = attachedImage || referenceImage?.src
      if (referenceSource) {
        response = isGeminiImageModel
          ? await editGeminiImage({ ...request, image: referenceSource })
          : await editImage({ ...request, image: referenceSource })
      } else {
        response = await generateImages(request)
      }
      const generated = (response.data ?? [])
        .map((item) => {
          const image = imageSrc(item)
          if (!image) return null
          const generatedImage: GeneratedImage = {
            key: nanoid(),
            src: image.src,
            prompt: trimmedPrompt,
          }
          if (image.mimeType) generatedImage.mimeType = image.mimeType
          if (item.revised_prompt) {
            generatedImage.revisedPrompt = item.revised_prompt
          }
          return generatedImage
        })
        .filter((item): item is GeneratedImage => item !== null)
      if (!generated.length) throw new Error(t('No images were returned'))
      setImages((current) => [...current, ...generated])
      setReferenceImageKey(generated.at(-1)?.key ?? null)
    } catch (error) {
      toast.error(requestErrorMessage(error, t('Request failed')))
    } finally {
      setIsGenerating(false)
    }
  }

  const handleModelChange = (value: string) => {
    setModel(value)
    const normalized = value.toLowerCase()
    if (
      normalized.startsWith('gemini-') ||
      normalized.startsWith('nano-banana')
    ) {
      setCount('1')
    }
  }

  const handleGroupChange = (value: string) => {
    setIsLoadingOptions(true)
    setGroup(value)
  }

  let submitLabel = referenceImage ? t('Edit image') : t('Generate image')
  if (isGenerating) submitLabel = t('Generating...')

  return (
    <div className='relative flex size-full min-h-0 flex-col overflow-hidden'>
      <div className='flex min-h-0 flex-1 flex-col overflow-hidden'>
        <Conversation>
          <ConversationContent className='p-0'>
            <div className='mx-auto w-full max-w-4xl px-4 py-4'>
              {images.length === 0 ? (
                <div className='flex min-h-[min(520px,calc(100svh-18rem))] items-center justify-center px-1 py-8 md:py-12'>
                  <div className='grid w-full max-w-2xl gap-5 text-center'>
                    <div className='bg-muted/50 text-muted-foreground mx-auto flex size-11 items-center justify-center rounded-xl border'>
                      <ImageIcon className='size-5' aria-hidden='true' />
                    </div>
                    <div className='grid gap-2'>
                      <h2 className='text-xl font-semibold tracking-tight text-balance md:text-2xl'>
                        {t('Create an image')}
                      </h2>
                      <p className='text-muted-foreground text-sm leading-6'>
                        {t('No images yet')}
                      </p>
                    </div>
                  </div>
                </div>
              ) : (
                <div className='grid gap-4'>
                  <div className='flex items-center justify-between'>
                    <div>
                      <h2 className='font-semibold'>{t('Generated images')}</h2>
                      <p className='text-muted-foreground mt-1 text-xs'>
                        {t('{{count}} images', { count: images.length })}
                      </p>
                    </div>
                    <Button
                      variant='ghost'
                      size='sm'
                      className='gap-2'
                      onClick={() => {
                        setImages([])
                        setReferenceImageKey(null)
                      }}
                    >
                      <Trash2 className='size-3.5' />
                      {t('Clear')}
                    </Button>
                  </div>
                  <div
                    className={cn(
                      'grid gap-4',
                      images.length === 1 ? 'grid-cols-1' : 'sm:grid-cols-2'
                    )}
                  >
                    {images.map((image, index) => (
                      <figure
                        key={image.key}
                        className='group bg-background/75 border-border/70 relative overflow-hidden rounded-lg border shadow-[0_18px_55px_-30px_rgba(15,23,42,0.5)]'
                      >
                        <img
                          src={image.src}
                          alt={image.revisedPrompt || image.prompt}
                          className='bg-muted aspect-square w-full object-contain'
                        />
                        <figcaption className='border-border/70 flex items-center justify-between gap-3 border-t px-4 py-3'>
                          <span className='text-muted-foreground line-clamp-2 text-xs'>
                            {image.revisedPrompt || image.prompt}
                          </span>
                          <div className='flex shrink-0 items-center gap-1'>
                            <Button
                              size='sm'
                              variant={
                                referenceImageKey === image.key
                                  ? 'secondary'
                                  : 'ghost'
                              }
                              onClick={() => setReferenceImageKey(image.key)}
                            >
                              {referenceImageKey === image.key
                                ? t('Editing')
                                : t('Edit')}
                            </Button>
                            <Button
                              size='icon-sm'
                              variant='secondary'
                              aria-label={t('Download')}
                              onClick={() =>
                                downloadImage(image.src, index, image.mimeType)
                              }
                            >
                              <Download className='size-4' />
                            </Button>
                          </div>
                        </figcaption>
                      </figure>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </ConversationContent>
        </Conversation>
      </div>

      <div className='mx-auto w-full max-w-4xl'>
        <PromptInput
          accept='image/*'
          className='relative px-1 md:pb-4'
          groupClassName='bg-background/95 dark:bg-background/80 border-border/70 shadow-[0_18px_60px_-32px_rgba(0,0,0,0.65)] ring-1 ring-foreground/5 rounded-xl overflow-hidden transition-all duration-200 focus-within:border-primary/45 focus-within:ring-primary/15 focus-within:shadow-[0_22px_70px_-34px_rgba(0,0,0,0.75)]'
          multiple={false}
          onSubmit={({ text, files }) =>
            handleGenerate(text || '', files?.at(0)?.url)
          }
        >
          <ImagePromptInputHeader
            referenceImage={referenceImage}
            onClearReference={() => setReferenceImageKey(null)}
          />
          <PromptInputTextarea
            autoComplete='off'
            autoCorrect='off'
            autoCapitalize='off'
            spellCheck={false}
            className='min-h-20 px-5 pt-4 pb-3 leading-7 md:min-h-24 md:text-base'
            disabled={isGenerating}
            onChange={(event) => setPrompt(event.target.value)}
            placeholder={t('Enter prompt')}
          />
          <PromptInputFooter className='border-border/60 bg-muted/20 dark:bg-muted/10 border-t px-3 py-2.5 backdrop-blur'>
            <div className='flex w-full flex-col gap-2.5 md:flex-row md:items-center md:justify-between'>
              <div className='flex items-center justify-between gap-2 md:justify-start'>
                <PromptInputTools className='bg-background/70 border-border/60 rounded-lg border p-1 shadow-xs'>
                  <ImageParameterSettings
                    count={count}
                    maxCount={maxImageCount}
                    size={size}
                    quality={quality}
                    disabled={isGenerating}
                    onCountChange={setCount}
                    onSizeChange={setSize}
                    onQualityChange={setQuality}
                  />

                  <Tooltip>
                    <DropdownMenu>
                      <TooltipTrigger
                        render={
                          <DropdownMenuTrigger
                            render={
                              <PromptInputButton
                                aria-label={t('Attach')}
                                className='text-muted-foreground hover:bg-muted/70 hover:text-foreground font-medium'
                                disabled={isGenerating}
                                variant='ghost'
                              />
                            }
                          >
                            <PaperclipIcon size={16} />
                          </DropdownMenuTrigger>
                        }
                      />
                      <TooltipContent>
                        <p>{t('Attach')}</p>
                      </TooltipContent>
                      <DropdownMenuContent align='start'>
                        <PromptInputActionAddAttachments
                          label={t('Add photos or files')}
                        />
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </Tooltip>

                  <Tooltip>
                    <TooltipTrigger
                      render={
                        <PromptInputButton
                          aria-label={t('Clear chat history')}
                          className='text-muted-foreground hover:bg-destructive/10 hover:text-destructive font-medium'
                          disabled={isGenerating || images.length === 0}
                          onClick={() => {
                            setImages([])
                            setReferenceImageKey(null)
                          }}
                          variant='ghost'
                        >
                          <Trash2Icon size={16} />
                        </PromptInputButton>
                      }
                    />
                    <TooltipContent>
                      <p>{t('Clear chat history')}</p>
                    </TooltipContent>
                  </Tooltip>
                </PromptInputTools>
                <div className='flex items-center gap-1.5 md:hidden'>
                  <PromptInputButton
                    className='bg-primary text-primary-foreground hover:bg-primary/90 disabled:bg-muted disabled:text-muted-foreground h-8 px-3 font-medium shadow-sm'
                    disabled={
                      isGenerating || isLoadingOptions || !selectedModel
                    }
                    type='submit'
                    variant='default'
                  >
                    {isGenerating ? (
                      <LoaderCircle className='size-4 animate-spin' />
                    ) : (
                      <SendIcon className='size-4' />
                    )}
                    <span className='sr-only'>{submitLabel}</span>
                  </PromptInputButton>
                </div>
              </div>

              <div className='hidden min-w-0 items-center gap-2 md:flex'>
                <ModelGroupSelector
                  selectedModel={selectedModel}
                  models={models}
                  onModelChange={handleModelChange}
                  selectedGroup={selectedGroup}
                  groups={groups}
                  onGroupChange={handleGroupChange}
                  disabled={isGenerating || isLoadingOptions}
                />
                <PromptInputButton
                  className='bg-primary text-primary-foreground hover:bg-primary/90 disabled:bg-muted disabled:text-muted-foreground h-8 px-3 font-medium shadow-sm'
                  disabled={isGenerating || isLoadingOptions || !selectedModel}
                  type='submit'
                  variant='default'
                >
                  {isGenerating ? (
                    <LoaderCircle className='size-4 animate-spin' />
                  ) : (
                    <SendIcon className='size-4' />
                  )}
                  <span className='hidden sm:inline'>{submitLabel}</span>
                </PromptInputButton>
              </div>

              <div className='flex min-w-0 items-center md:hidden'>
                <ModelGroupSelector
                  className='w-full max-w-none'
                  selectedModel={selectedModel}
                  models={models}
                  onModelChange={handleModelChange}
                  selectedGroup={selectedGroup}
                  groups={groups}
                  onGroupChange={handleGroupChange}
                  disabled={isGenerating || isLoadingOptions}
                />
              </div>
            </div>
          </PromptInputFooter>
        </PromptInput>
      </div>
    </div>
  )
}
