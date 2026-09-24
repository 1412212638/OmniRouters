import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import * as z from 'zod'

import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'

import {
  SettingsForm,
  SettingsSwitchContent,
  SettingsSwitchItem,
} from '../components/settings-form-layout'
import { SettingsPageFormActions } from '../components/settings-page-context'
import { SettingsSection } from '../components/settings-section'
import { useResetForm } from '../hooks/use-reset-form'
import { useUpdateOption } from '../hooks/use-update-option'
import { safeNumberFieldProps } from '../utils/numeric-field'

const behaviorSchema = z.object({
  DefaultCollapseSidebar: z.boolean(),
  DemoSiteEnabled: z.boolean(),
  SelfUseModeEnabled: z.boolean(),
  ClientGoneDrainTimeout: z.number().int().min(1).max(300),
})

type BehaviorFormValues = z.infer<typeof behaviorSchema>

type SystemBehaviorSectionProps = {
  defaultValues: BehaviorFormValues
}

export function SystemBehaviorSection({
  defaultValues,
}: SystemBehaviorSectionProps) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()

  const form = useForm({
    resolver: zodResolver(behaviorSchema),
    defaultValues,
  })

  useResetForm(form, defaultValues)

  const onSubmit = async (data: BehaviorFormValues) => {
    const updates = Object.entries(data).filter(
      ([key, value]) => value !== defaultValues[key as keyof BehaviorFormValues]
    )

    for (const [key, value] of updates) {
      await updateOption.mutateAsync({ key, value })
    }
  }

  return (
    <SettingsSection title={t('System Behavior')}>
      <Form {...form}>
        <SettingsForm onSubmit={form.handleSubmit(onSubmit)}>
          <SettingsPageFormActions
            onSave={form.handleSubmit(onSubmit)}
            isSaving={updateOption.isPending}
          />
          <FormField
            control={form.control}
            name='DefaultCollapseSidebar'
            render={({ field }) => (
              <SettingsSwitchItem>
                <SettingsSwitchContent>
                  <FormLabel>{t('Default Collapse Sidebar')}</FormLabel>
                  <FormDescription>
                    {t('Sidebar collapsed by default for new users')}
                  </FormDescription>
                </SettingsSwitchContent>
                <FormControl>
                  <Switch
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                </FormControl>
              </SettingsSwitchItem>
            )}
          />

          <FormField
            control={form.control}
            name='DemoSiteEnabled'
            render={({ field }) => (
              <SettingsSwitchItem>
                <SettingsSwitchContent>
                  <FormLabel>{t('Demo Site Mode')}</FormLabel>
                  <FormDescription>
                    {t('Enable demo mode with limited functionality')}
                  </FormDescription>
                </SettingsSwitchContent>
                <FormControl>
                  <Switch
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                </FormControl>
              </SettingsSwitchItem>
            )}
          />

          <FormField
            control={form.control}
            name='SelfUseModeEnabled'
            render={({ field }) => (
              <SettingsSwitchItem>
                <SettingsSwitchContent>
                  <FormLabel>{t('Self-Use Mode')}</FormLabel>
                  <FormDescription>
                    {t('Optimize system for self-hosted single-user usage')}
                  </FormDescription>
                </SettingsSwitchContent>
                <FormControl>
                  <Switch
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                </FormControl>
              </SettingsSwitchItem>
            )}
          />

          <FormField
            control={form.control}
            name='ClientGoneDrainTimeout'
            render={({ field }) => (
              <FormItem>
                <FormLabel>
                  {t('Client disconnect drain timeout (seconds)')}
                </FormLabel>
                <FormControl>
                  <Input
                    type='number'
                    min={1}
                    max={300}
                    step={1}
                    {...safeNumberFieldProps(field)}
                  />
                </FormControl>
                <FormDescription>
                  {t(
                    'How long to keep reading the upstream after a streaming client disconnects, so final usage can be received (1-300 seconds)'
                  )}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
        </SettingsForm>
      </Form>
    </SettingsSection>
  )
}
