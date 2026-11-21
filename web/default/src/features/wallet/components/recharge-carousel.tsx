import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { cn } from '@/lib/utils'

import type { RechargeCarouselItem } from '../types'

interface RechargeCarouselProps {
  items?: RechargeCarouselItem[]
}

export function RechargeCarousel({ items = [] }: RechargeCarouselProps) {
  const { t } = useTranslation()
  const visibleItems = useMemo(
    () => items.filter((item) => item.enabled !== false && item.image),
    [items]
  )
  const [activeIndex, setActiveIndex] = useState(0)

  useEffect(() => {
    setActiveIndex((current) =>
      visibleItems.length === 0 ? 0 : Math.min(current, visibleItems.length - 1)
    )
  }, [visibleItems.length])

  useEffect(() => {
    if (visibleItems.length < 2) return
    const timer = window.setInterval(() => {
      setActiveIndex((current) => (current + 1) % visibleItems.length)
    }, 6000)
    return () => window.clearInterval(timer)
  }, [visibleItems.length])

  if (visibleItems.length === 0) return null

  const item = visibleItems[activeIndex]
  const image = (
    <img
      src={item.image}
      alt={item.title || t('Recharge promotion')}
      className='block h-auto max-h-56 w-full object-cover sm:max-h-72'
      loading='lazy'
    />
  )

  return (
    <div className='border-border/70 bg-muted/20 overflow-hidden rounded-xl border'>
      {item.link ? (
        <a
          href={item.link}
          target={item.link.startsWith('http') ? '_blank' : undefined}
          rel={item.link.startsWith('http') ? 'noopener noreferrer' : undefined}
          aria-label={item.title || t('Recharge promotion')}
        >
          {image}
        </a>
      ) : (
        image
      )}
      {(item.title || item.description || visibleItems.length > 1) && (
        <div className='flex items-end justify-between gap-4 px-4 py-3'>
          <div className='min-w-0'>
            {item.title && <p className='font-medium'>{item.title}</p>}
            {item.description && (
              <p className='text-muted-foreground mt-0.5 text-sm'>
                {item.description}
              </p>
            )}
          </div>
          {visibleItems.length > 1 && (
            <div
              className='flex shrink-0 items-center gap-1.5'
              aria-label={t('Recharge promotion slides')}
            >
              {visibleItems.map((slide, index) => (
                <button
                  key={slide.id || `${slide.image}-${index}`}
                  type='button'
                  aria-label={t('Go to slide {{number}}', {
                    number: index + 1,
                  })}
                  aria-current={index === activeIndex ? 'true' : undefined}
                  className={cn(
                    'bg-muted-foreground/30 h-1.5 w-1.5 rounded-full transition-all',
                    index === activeIndex && 'bg-primary w-4'
                  )}
                  onClick={() => setActiveIndex(index)}
                />
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
