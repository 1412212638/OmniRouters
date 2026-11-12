import { useRouterState } from '@tanstack/react-router'
import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'

import {
  multiProviderAiApiMeta,
  openAICompatibleApiMeta,
} from '@/features/seo-pages/meta'
import { DEFAULT_SYSTEM_NAME } from '@/lib/constants'
import { getPageTitleKey } from '@/lib/page-titles'
import { useSystemConfigStore } from '@/stores/system-config-store'

const PUBLIC_PAGE_META = new Map([
  [openAICompatibleApiMeta.canonical, openAICompatibleApiMeta],
  [multiProviderAiApiMeta.canonical, multiProviderAiApiMeta],
])

/** One title owner, reactive to navigation, language, and site configuration. */
export function DocumentTitle() {
  const pathname = useRouterState({
    select: (state) => state.location.pathname,
  })
  const systemName = useSystemConfigStore((state) => state.config.systemName)
  const { t } = useTranslation()

  useEffect(() => {
    const path = pathname.replace(/\/+$/, '') || '/'
    const name = systemName?.trim() || DEFAULT_SYSTEM_NAME
    const publicMeta = PUBLIC_PAGE_META.get(path)
    const pageKey = getPageTitleKey(path)
    let title = name
    if (publicMeta) {
      title = `${publicMeta.title} | ${name}`
    } else if (path !== '/') {
      title = `${t(pageKey ?? 'Oops! Page Not Found!')} | ${name}`
    }

    document.title = title
    document.querySelector('meta[name="title"]')?.setAttribute('content', title)

    // Preserve the existing branded SEO descriptions without leaving them on
    // unrelated pages after client-side navigation.
    const description = document.querySelector('meta[name="description"]')
    const previousDescription = description?.getAttribute('content')
    if (description && publicMeta) {
      description.setAttribute(
        'content',
        publicMeta.description.replaceAll('this gateway', name)
      )
    }
    return () => {
      if (description && publicMeta && previousDescription !== null) {
        description.setAttribute('content', previousDescription ?? '')
      }
    }
  }, [pathname, systemName, t])

  return null
}
