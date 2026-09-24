import { useTranslation } from 'react-i18next'

import { getChannelTypeLabel } from '@/features/channels/lib'

import type { MarketplaceIndexVersion, MarketplacePlugin } from '../types'

type MarketplaceCapabilitiesProps = {
  plugin: MarketplacePlugin
  version?: MarketplaceIndexVersion
}

/**
 * The sensitive declarations an administrator needs before installing: where the
 * plugin may send requests, which channel types it can bind, how it
 * authenticates, and whether the source pins an integrity hash. Reading these is
 * far more effective than expecting a full source review, so they are surfaced
 * above the source viewer rather than buried in it.
 */
export function MarketplaceCapabilities(props: MarketplaceCapabilitiesProps) {
  const { t } = useTranslation()
  const allowedHosts = props.version?.allowedHosts
  const channelTypes = props.plugin.channelTypes

  return (
    <div className='space-y-2 rounded-md border p-3 text-sm'>
      <p className='font-medium'>{t('Declared capabilities')}</p>
      <dl className='grid gap-2 sm:grid-cols-2'>
        <div>
          <dt className='text-muted-foreground text-xs'>
            {t('Allowed hosts')}
          </dt>
          <dd className='font-mono text-xs break-all'>
            {allowedHosts?.length ? allowedHosts.join(', ') : t('Not declared')}
          </dd>
        </div>
        <div>
          <dt className='text-muted-foreground text-xs'>
            {t('Channel types')}
          </dt>
          <dd className='text-xs'>
            {channelTypes?.length
              ? channelTypes
                  .map((type) => `${getChannelTypeLabel(type)} (#${type})`)
                  .join(', ')
              : t('Not declared')}
          </dd>
        </div>
        <div>
          <dt className='text-muted-foreground text-xs'>
            {t('Authentication')}
          </dt>
          <dd className='font-mono text-xs'>
            {props.version?.auth || t('Not declared')}
          </dd>
        </div>
        <div>
          <dt className='text-muted-foreground text-xs'>
            {t('Integrity hash')}
          </dt>
          <dd className='font-mono text-xs break-all'>
            {props.version?.sha256 ?? t('Not provided by this source')}
          </dd>
        </div>
      </dl>
      <p className='text-muted-foreground text-xs'>
        {t(
          'These values come from the source index and are shown for review only. The gateway admits the plugin based on the metadata compiled from its source.'
        )}
      </p>
    </div>
  )
}
