import { useNavigate } from '@tanstack/react-router'
import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'

import { SectionPageLayout } from '@/components/layout'

import { ApiKeysDialogs } from './components/api-keys-dialogs'
import { ApiKeysPrimaryButtons } from './components/api-keys-primary-buttons'
import { ApiKeysProvider, useApiKeys } from './components/api-keys-provider'
import { ApiKeysTable } from './components/api-keys-table'

function OpenCreateDialogFromSearch({ enabled }: { enabled?: boolean }) {
  const { setOpen } = useApiKeys()
  const navigate = useNavigate()

  useEffect(() => {
    if (!enabled) return

    setOpen('create')
    void navigate({
      to: '/keys',
      search: (prev) => ({ ...prev, create: undefined }),
      replace: true,
    })
  }, [enabled, navigate, setOpen])

  return null
}

export function ApiKeys({ openCreate = false }: { openCreate?: boolean }) {
  const { t } = useTranslation()
  return (
    <ApiKeysProvider>
      <OpenCreateDialogFromSearch enabled={openCreate} />
      <SectionPageLayout fixedContent>
        <SectionPageLayout.Title>{t('API Keys')}</SectionPageLayout.Title>
        <SectionPageLayout.Actions>
          <ApiKeysPrimaryButtons />
        </SectionPageLayout.Actions>
        <SectionPageLayout.Content>
          <ApiKeysTable />
        </SectionPageLayout.Content>
      </SectionPageLayout>

      <ApiKeysDialogs />
    </ApiKeysProvider>
  )
}
