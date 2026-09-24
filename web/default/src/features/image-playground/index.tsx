import {
  Download,
  ImageIcon,
  LoaderCircle,
  Sparkles,
  Trash2,
  WandSparkles,
} from 'lucide-react'
import { nanoid } from 'nanoid'
import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { ModelGroupSelector } from '@/components/model-group-selector'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { cn } from '@/lib/utils'

import {
  generateImages,
  getImageGroups,
  getImageModels,
  type ImageGroupOption,
  type ImageModelOption,
} from './api'

type GeneratedImage = {
  key: string
  src: string
  mimeType?: string
  revisedPrompt?: string
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
    setIsLoadingOptions(true)
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

  const handleGenerate = async () => {
    const trimmedPrompt = prompt.trim()
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
      const response = await generateImages({
        model: selectedModel,
        group: selectedGroup || undefined,
        prompt: trimmedPrompt,
        n: Math.min(maxImageCount, Math.max(1, Number(count) || 1)),
        size,
        quality,
      })
      const generated = (response.data ?? [])
        .map((item) => {
          const image = imageSrc(item)
          if (!image) return null
          const generatedImage: GeneratedImage = {
            key: nanoid(),
            src: image.src,
          }
          if (image.mimeType) generatedImage.mimeType = image.mimeType
          if (item.revised_prompt) {
            generatedImage.revisedPrompt = item.revised_prompt
          }
          return generatedImage
        })
        .filter((item): item is GeneratedImage => item !== null)
      if (!generated.length) throw new Error(t('No images were returned'))
      setImages(generated)
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

  return (
    <div className='bg-background relative flex min-h-0 flex-1 flex-col overflow-auto'>
      <div className='mx-auto flex w-full max-w-6xl flex-1 flex-col gap-8 px-5 py-8 md:px-8 lg:py-12'>
        <header className='border-border/60 flex flex-col gap-3 border-b pb-7 md:flex-row md:items-end md:justify-between'>
          <div className='space-y-2'>
            <div className='text-primary flex items-center gap-2 text-xs font-semibold uppercase'>
              <Sparkles className='size-4' />
              {t('Creative workspace')}
            </div>
            <h1 className='text-3xl font-semibold md:text-4xl'>
              {t('Image Playground')}
            </h1>
          </div>
        </header>

        <div className='grid min-h-0 flex-1 gap-6 lg:grid-cols-[minmax(290px,360px)_1fr]'>
          <section className='bg-background/80 border-border/70 h-fit space-y-5 rounded-lg border p-5 shadow-[0_18px_55px_-30px_rgba(15,23,42,0.4)] backdrop-blur'>
            <div className='flex items-center justify-between'>
              <div>
                <h2 className='font-semibold'>{t('Create an image')}</h2>
              </div>
              <WandSparkles className='text-primary size-5' />
            </div>

            <div className='space-y-2'>
              <Label htmlFor='image-prompt'>{t('Prompt')}</Label>
              <Textarea
                id='image-prompt'
                value={prompt}
                onChange={(event) => setPrompt(event.target.value)}
                placeholder={t('Enter prompt')}
                className='bg-background/70 min-h-32 resize-y leading-6'
                disabled={isGenerating}
              />
            </div>

            <div className='space-y-2'>
              <Label>{t('Model')}</Label>
              <ModelGroupSelector
                selectedModel={selectedModel}
                models={models}
                onModelChange={handleModelChange}
                selectedGroup={selectedGroup}
                groups={groups}
                onGroupChange={setGroup}
                disabled={isGenerating || isLoadingOptions}
              />
              {models.length === 0 && !isLoadingOptions && (
                <p className='text-muted-foreground text-xs'>
                  {t('No image models available')}
                </p>
              )}
            </div>

            <div className='grid grid-cols-2 gap-3'>
              <div className='space-y-2'>
                <Label htmlFor='image-count'>{t('Images')}</Label>
                <Input
                  id='image-count'
                  type='number'
                  min={1}
                  max={maxImageCount}
                  value={count}
                  onChange={(event) =>
                    setCount(
                      String(
                        Math.min(
                          maxImageCount,
                          Math.max(1, Number(event.target.value) || 1)
                        )
                      )
                    )
                  }
                  disabled={isGenerating || maxImageCount === 1}
                />
              </div>
              <div className='space-y-2'>
                <Label>{t('Size')}</Label>
                <Select
                  value={size}
                  onValueChange={(value) => value && setSize(value)}
                >
                  <SelectTrigger className='w-full'>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {[
                      '1024x1024',
                      '1536x1024',
                      '1024x1536',
                      '1792x1024',
                      '1024x1792',
                    ].map((value) => (
                      <SelectItem key={value} value={value}>
                        {value}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className='space-y-2'>
              <Label>{t('Quality')}</Label>
              <Select
                value={quality}
                onValueChange={(value) => value && setQuality(value)}
              >
                <SelectTrigger className='w-full'>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {['auto', 'low', 'medium', 'high'].map((value) => (
                    <SelectItem key={value} value={value}>
                      {value}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <Button
              className='h-10 w-full gap-2'
              onClick={handleGenerate}
              disabled={isGenerating || isLoadingOptions || !selectedModel}
            >
              {isGenerating ? (
                <LoaderCircle className='size-4 animate-spin' />
              ) : (
                <WandSparkles className='size-4' />
              )}
              {isGenerating ? t('Generating...') : t('Generate image')}
            </Button>
          </section>

          <section className='flex min-h-[420px] min-w-0 flex-col gap-4'>
            <div className='flex items-center justify-between'>
              <div>
                <h2 className='font-semibold'>{t('Generated images')}</h2>
                <p className='text-muted-foreground mt-1 text-xs'>
                  {images.length
                    ? t('{{count}} images', { count: images.length })
                    : null}
                </p>
              </div>
              {images.length > 0 && (
                <Button
                  variant='ghost'
                  size='sm'
                  className='gap-2'
                  onClick={() => setImages([])}
                >
                  <Trash2 className='size-3.5' />
                  {t('Clear')}
                </Button>
              )}
            </div>
            {images.length === 0 ? (
              <div className='bg-background/45 text-muted-foreground border-border/80 flex min-h-[390px] flex-1 flex-col items-center justify-center rounded-lg border border-dashed p-8 text-center'>
                <div className='bg-primary/10 text-primary mb-4 flex size-14 items-center justify-center rounded-lg'>
                  <ImageIcon className='size-7' />
                </div>
                <p className='font-medium'>{t('No images yet')}</p>
              </div>
            ) : (
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
                      alt={prompt}
                      className='bg-muted aspect-square w-full object-contain'
                    />
                    <figcaption className='border-border/70 flex items-center justify-between gap-3 border-t px-4 py-3'>
                      <span className='text-muted-foreground line-clamp-2 text-xs'>
                        {image.revisedPrompt || prompt}
                      </span>
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
                    </figcaption>
                  </figure>
                ))}
              </div>
            )}
          </section>
        </div>
      </div>
    </div>
  )
}
