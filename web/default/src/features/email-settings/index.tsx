/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { useCallback, useMemo } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
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
