import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'

import {
  SettingsForm,
  SettingsSwitchContent,
  SettingsSwitchRow,
} from '../components/settings-form-layout'
import { SettingsPageFormActions } from '../components/settings-page-context'
import { SettingsSection } from '../components/settings-section'
import { useUpdateOption } from '../hooks/use-update-option'

export type RealNameSettingsValues = {
  'real_name_verification.enabled': boolean
  'real_name_verification.enforcement_enabled': boolean
  'real_name_verification.personal_enabled': boolean
  'real_name_verification.personal_app_code': string
  'real_name_verification.personal_app_code_configured': boolean
  'real_name_verification.enterprise_enabled': boolean
  'real_name_verification.enterprise_app_code': string
  'real_name_verification.enterprise_app_code_configured': boolean
}

type Props = { defaultValues: RealNameSettingsValues }

export function RealNameSection({ defaultValues }: Props) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()
  const [values, setValues] = useState(defaultValues)

  useEffect(() => setValues(defaultValues), [defaultValues])

  const set = <K extends keyof RealNameSettingsValues>(
    key: K,
    value: RealNameSettingsValues[K]
  ) => setValues((current) => ({ ...current, [key]: value }))

  const save = async () => {
    try {
      const credentialUpdates: Array<{ key: string; value: string }> = []
      if (values['real_name_verification.personal_app_code'].trim()) {
        credentialUpdates.push({
          key: 'real_name_verification.personal_app_code',
          value: values['real_name_verification.personal_app_code'].trim(),
        })
      }
      if (values['real_name_verification.enterprise_app_code'].trim()) {
        credentialUpdates.push({
          key: 'real_name_verification.enterprise_app_code',
          value: values['real_name_verification.enterprise_app_code'].trim(),
        })
      }
      for (const update of credentialUpdates) await updateOption.mutateAsync(update)

      const switches: Array<[keyof RealNameSettingsValues, boolean]> = [
        ['real_name_verification.enabled', values['real_name_verification.enabled']],
        ['real_name_verification.personal_enabled', values['real_name_verification.personal_enabled']],
        ['real_name_verification.enterprise_enabled', values['real_name_verification.enterprise_enabled']],
        ['real_name_verification.enforcement_enabled', values['real_name_verification.enforcement_enabled']],
      ]
      for (const [key, value] of switches) {
        if (value !== defaultValues[key]) await updateOption.mutateAsync({ key, value })
      }
      toast.success(t('Setting updated successfully'))
    } catch (error) {
      toast.error(error instanceof Error ? error.message : t('Failed to update setting'))
    }
  }

  return (
    <SettingsSection title={t('Real-name Verification')}>
      <SettingsForm onSubmit={(event) => { event.preventDefault(); void save() }}>
        <SettingsPageFormActions onSave={save} isSaving={updateOption.isPending} />

        <SettingsSwitchRow>
          <SettingsSwitchContent>
            <span className='text-sm font-medium'>{t('Enable Real-name Verification')}</span>
            <span className='text-muted-foreground text-xs'>{t('Allow users to verify their personal or enterprise identity')}</span>
          </SettingsSwitchContent>
          <Switch checked={values['real_name_verification.enabled']} onCheckedChange={(value) => set('real_name_verification.enabled', value)} />
        </SettingsSwitchRow>

        <SettingsSwitchRow>
          <SettingsSwitchContent>
            <span className='text-sm font-medium'>{t('Require Verification for Platform Services')}</span>
            <span className='text-muted-foreground text-xs'>{t('Unverified users cannot use model, API, subscription, payment, or other business services')}</span>
          </SettingsSwitchContent>
          <Switch checked={values['real_name_verification.enforcement_enabled']} onCheckedChange={(value) => set('real_name_verification.enforcement_enabled', value)} />
        </SettingsSwitchRow>

        <div className='bg-muted/20 min-w-0 space-y-4 rounded-xl border p-4 lg:col-span-2'>
          <div>
            <h3 className='text-sm font-medium'>{t('Personal Verification')}</h3>
            <p className='text-muted-foreground text-xs'>{t('Verify a name and identity card number through the configured provider')}</p>
          </div>
          <SettingsSwitchRow>
            <SettingsSwitchContent>
              <span className='text-sm font-medium'>{t('Enable Personal Verification')}</span>
              <span className='text-muted-foreground text-xs'>{t('Available only after an AppCode is configured')}</span>
            </SettingsSwitchContent>
            <Switch checked={values['real_name_verification.personal_enabled']} onCheckedChange={(value) => set('real_name_verification.personal_enabled', value)} />
          </SettingsSwitchRow>
          <div className='space-y-2'>
            <label className='text-sm font-medium'>{t('Personal Verification AppCode')}</label>
            <Input type='password' value={values['real_name_verification.personal_app_code']} onChange={(event) => set('real_name_verification.personal_app_code', event.target.value)} placeholder={values['real_name_verification.personal_app_code_configured'] ? t('Configured; leave blank to keep it') : t('Enter AppCode')} />
            <p className='text-muted-foreground text-xs'>{t('The AppCode is stored on the server and is never returned to the browser')}</p>
          </div>
        </div>

        <div className='bg-muted/20 min-w-0 space-y-4 rounded-xl border p-4 lg:col-span-2'>
          <div>
            <h3 className='text-sm font-medium'>{t('Enterprise Verification')}</h3>
            <p className='text-muted-foreground text-xs'>{t('Verify company, registration number, representative name, and identity card number')}</p>
          </div>
          <SettingsSwitchRow>
            <SettingsSwitchContent>
              <span className='text-sm font-medium'>{t('Enable Enterprise Verification')}</span>
              <span className='text-muted-foreground text-xs'>{t('Available only after an AppCode is configured')}</span>
            </SettingsSwitchContent>
            <Switch checked={values['real_name_verification.enterprise_enabled']} onCheckedChange={(value) => set('real_name_verification.enterprise_enabled', value)} />
          </SettingsSwitchRow>
          <div className='space-y-2'>
            <label className='text-sm font-medium'>{t('Enterprise Verification AppCode')}</label>
            <Input type='password' value={values['real_name_verification.enterprise_app_code']} onChange={(event) => set('real_name_verification.enterprise_app_code', event.target.value)} placeholder={values['real_name_verification.enterprise_app_code_configured'] ? t('Configured; leave blank to keep it') : t('Enter AppCode')} />
            <p className='text-muted-foreground text-xs'>{t('The AppCode is stored on the server and is never returned to the browser')}</p>
          </div>
        </div>
      </SettingsForm>
    </SettingsSection>
  )
}
