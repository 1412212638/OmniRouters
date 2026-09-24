import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useCallback, useMemo } from 'react'
import { useTranslation } from 'react-i18next'

import { SectionPageLayout } from '@/components/layout'
import {
  getEmailSettingsOptions,
  sendEmailSettingsMarketingEmail,
  updateEmailSettingOption,
} from '@/features/system-settings/api'
import { getOptionValue } from '@/features/system-settings/hooks/use-system-options'
import {
  EmailSettingsSection,
  defaultEmailSettingsFormValues,
  type EmailFormValues,
} from '@/features/system-settings/integrations/email-settings-section'
import type { UpdateOptionRequest } from '@/features/system-settings/types'

type EmailSettingsOptionValues = Omit<
  EmailFormValues,
  'EmailDomainWhitelist'
> & {
  EmailDomainWhitelist: string
}

const defaultEmailSettingsOptionValues: EmailSettingsOptionValues = {
  ...defaultEmailSettingsFormValues,
  EmailDomainWhitelist: '',
}

const parseDomainWhitelist = (value: string) =>
  value
    .split(',')
    .map((domain) => domain.trim())
    .filter(Boolean)

export function EmailSettings() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { data, isLoading } = useQuery({
    queryKey: ['email-settings-options'],
    queryFn: getEmailSettingsOptions,
    staleTime: 5 * 60 * 1000,
  })

  const settings = useMemo(
    () => getOptionValue(data?.data, defaultEmailSettingsOptionValues),
    [data?.data]
  )

  const defaultValues = useMemo<EmailFormValues>(
    () => ({
      ...settings,
      EmailDomainWhitelist: parseDomainWhitelist(settings.EmailDomainWhitelist),
    }),
    [settings]
  )

  const handleUpdateOption = useCallback(
    async (request: UpdateOptionRequest) => {
      return updateEmailSettingOption(request)
    },
    []
  )

  const handleSaved = useCallback(() => {
    void queryClient.invalidateQueries({
      queryKey: ['email-settings-options'],
    })
    void queryClient.invalidateQueries({ queryKey: ['system-options'] })
  }, [queryClient])

  if (isLoading) {
    return (
      <div className='text-muted-foreground flex h-full w-full flex-1 items-center justify-center'>
        {t('Loading email settings...')}
      </div>
    )
  }

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('Email Settings')}</SectionPageLayout.Title>
      <SectionPageLayout.Description>
        {t(
          'Manage SMTP delivery, templates, domain whitelist, and marketing emails.'
        )}
      </SectionPageLayout.Description>
      <SectionPageLayout.Content>
        <EmailSettingsSection
          defaultValues={defaultValues}
          updateOptionRequest={handleUpdateOption}
          sendMarketingEmailRequest={sendEmailSettingsMarketingEmail}
          onSaved={handleSaved}
        />
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
