import { getMarketplaceDisplay } from '@/lib/marketplace-display'
import { StatusBadge } from '@/components/status-badge'
import type { PricingModel } from '../types'

interface ModelMarketplaceBadgesProps {
  model: PricingModel
  discountFallback: string
}

export function ModelMarketplaceBadges(props: ModelMarketplaceBadgesProps) {
  const display = getMarketplaceDisplay(props.model, props.discountFallback)

  if (!display.showNew && !display.discountLabel && !display.promotionNote) {
    return null
  }

  return (
    <>
      {display.showNew && (
        <StatusBadge label='NEW' variant='red' copyable={false} size='sm' />
      )}
      {display.discountLabel && (
        <StatusBadge
          label={display.discountLabel}
          variant='orange'
          copyable={false}
          size='sm'
        />
      )}
      {display.discountLabel && display.promotionNote && (
        <StatusBadge
          label={display.promotionNote}
          variant='red'
          copyable={false}
          size='sm'
        />
      )}
    </>
  )
}
