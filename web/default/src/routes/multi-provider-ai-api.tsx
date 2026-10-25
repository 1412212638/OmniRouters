import { createFileRoute } from '@tanstack/react-router'

import { MultiProviderAiApiPage } from '@/features/seo-pages'
import { multiProviderAiApiMeta } from '@/features/seo-pages/meta'

export const Route = createFileRoute('/multi-provider-ai-api')({
  head: () => ({
    meta: [
      { title: multiProviderAiApiMeta.title },
      { name: 'description', content: multiProviderAiApiMeta.description },
      { name: 'robots', content: 'index, follow' },
    ],
    links: [{ rel: 'canonical', href: multiProviderAiApiMeta.canonical }],
  }),
  component: MultiProviderAiApiPage,
})
