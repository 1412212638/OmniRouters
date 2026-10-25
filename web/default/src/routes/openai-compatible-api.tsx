import { createFileRoute } from '@tanstack/react-router'

import { OpenAICompatibleApiPage } from '@/features/seo-pages'
import { openAICompatibleApiMeta } from '@/features/seo-pages/meta'

export const Route = createFileRoute('/openai-compatible-api')({
  head: () => ({
    meta: [
      { title: openAICompatibleApiMeta.title },
      { name: 'description', content: openAICompatibleApiMeta.description },
      { name: 'robots', content: 'index, follow' },
    ],
    links: [{ rel: 'canonical', href: openAICompatibleApiMeta.canonical }],
  }),
  component: OpenAICompatibleApiPage,
})
