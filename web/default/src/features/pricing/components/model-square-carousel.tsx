import { ChevronLeft, ChevronRight } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import type { ModelSquareCarouselItem } from '@/features/auth/types'
import { cn } from '@/lib/utils'

export function ModelSquareCarousel({
  items,
}: {
  items: ModelSquareCarouselItem[]
}) {
  const { t } = useTranslation()
  const activeItems = items
    .filter((item) => item.enabled !== false && item.image && item.title)
    .sort((a, b) => (a.sort ?? 0) - (b.sort ?? 0) || (a.id ?? 0) - (b.id ?? 0))
  const [index, setIndex] = useState(0)
  useEffect(() => {
    setIndex(0)
  }, [activeItems.length])
  useEffect(() => {
    if (activeItems.length < 2) return
    const timer = window.setInterval(
      () => setIndex((current) => (current + 1) % activeItems.length),
      6000
    )
    return () => window.clearInterval(timer)
  }, [activeItems.length])
  if (!activeItems.length) return null
  const item = activeItems[index % activeItems.length]
  return (
    <article className='group bg-muted/20 relative min-h-[260px] overflow-hidden border-r border-b sm:min-h-[300px] lg:min-h-[340px]'>
      <img
        src={item.image}
        alt=''
        className='absolute inset-0 size-full object-cover transition-opacity duration-500'
      />
      <div className='absolute inset-0 bg-gradient-to-t from-black/75 via-black/20 to-transparent' />
      <div className='absolute inset-x-0 bottom-0 p-5 text-white sm:p-6'>
        <h2 className='text-lg font-semibold'>{item.title}</h2>
        <p className='mt-1 line-clamp-3 text-sm text-white/85'>
          {item.description}
        </p>
      </div>
      {activeItems.length > 1 && (
        <>
          <Button
            type='button'
            variant='ghost'
            size='icon'
            aria-label={t('Previous')}
            className='absolute top-1/2 left-3 -translate-y-1/2 text-white opacity-0 transition-opacity group-hover:opacity-100 hover:bg-black/30 hover:text-white'
            onClick={() =>
              setIndex(
                (current) =>
                  (current - 1 + activeItems.length) % activeItems.length
              )
            }
          >
            <ChevronLeft className='size-5' />
          </Button>
          <Button
            type='button'
            variant='ghost'
            size='icon'
            aria-label={t('Next')}
            className='absolute top-1/2 right-3 -translate-y-1/2 text-white opacity-0 transition-opacity group-hover:opacity-100 hover:bg-black/30 hover:text-white'
            onClick={() =>
              setIndex((current) => (current + 1) % activeItems.length)
            }
          >
            <ChevronRight className='size-5' />
          </Button>
          <div className='absolute right-4 bottom-3 flex gap-1.5'>
            {activeItems.map((entry, dotIndex) => (
              <button
                key={entry.id ?? dotIndex}
                type='button'
                aria-label={t('Go to slide {{count}}', { count: dotIndex + 1 })}
                onClick={() => setIndex(dotIndex)}
                className={cn(
                  'size-1.5 rounded-full bg-white/50',
                  dotIndex === index && 'bg-white'
                )}
              />
            ))}
          </div>
        </>
      )}
    </article>
  )
}
