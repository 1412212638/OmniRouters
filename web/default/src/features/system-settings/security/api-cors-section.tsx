import { useEffect, useMemo, useRef } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'

import { SettingsForm } from '../components/settings-form-layout'
import { SettingsPageFormActions } from '../components/settings-page-context'
import { SettingsSection } from '../components/settings-section'
import { useUpdateOption } from '../hooks/use-update-option'

type CORSMode = 'disabled' | 'allowlist' | 'all'

type Props = {
  defaultValues: {
    'cors_setting.mode': CORSMode
    'cors_setting.allowed_origins': string[]
  }
}

const parseOrigins = (value: string) =>
  value
    .split('\n')
    .map((origin) => origin.trim())
    .filter(Boolean)

export function APICORSSection({ defaultValues }: Props) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()
  const baseline = useRef(defaultValues)
  const formDefaults = useMemo(
    () => ({
      mode: defaultValues['cors_setting.mode'],
      allowedOrigins: defaultValues['cors_setting.allowed_origins'].join('\n'),
    }),
    [defaultValues]
  )
  const form = useForm({ defaultValues: formDefaults })

  useEffect(() => {
    baseline.current = defaultValues
    form.reset(formDefaults)
  }, [defaultValues, formDefaults, form])

  const mode = form.watch('mode') as CORSMode

  const onSubmit = async (values: typeof formDefaults) => {
    const origins = parseOrigins(values.allowedOrigins)
    const updates: Array<{ key: string; value: string | string[] }> = []
    if (values.mode !== baseline.current['cors_setting.mode']) {
      updates.push({ key: 'cors_setting.mode', value: values.mode })
    }
    if (
      JSON.stringify(origins) !==
      JSON.stringify(baseline.current['cors_setting.allowed_origins'])
    ) {
      updates.push({ key: 'cors_setting.allowed_origins', value: origins })
    }
    if (updates.length === 0) {
      toast.info(t('No changes to save'))
      return
    }
    for (const update of updates) {
      await updateOption.mutateAsync({
        ...update,
        value:
          typeof update.value === 'string'
            ? update.value
            : JSON.stringify(update.value),
      })
    }
    baseline.current = {
      'cors_setting.mode': values.mode,
      'cors_setting.allowed_origins': origins,
    }
  }

  return (
    <SettingsSection
      title={t('API Cross-Origin Access')}
      description={t(
        'Configure browser cross-origin access for public API endpoints under /v1. Admin dashboard APIs keep their existing origin policy.'
      )}
    >
      <Form {...form}>
        <SettingsForm onSubmit={form.handleSubmit(onSubmit)}>
          <SettingsPageFormActions
            onSave={form.handleSubmit(onSubmit)}
            isSaving={updateOption.isPending}
          />
          <FormField
            control={form.control}
            name='mode'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('API CORS policy')}</FormLabel>
                <Select
                  items={[
                    { value: 'disabled', label: t('Disabled') },
                    { value: 'allowlist', label: t('Allow listed origins') },
                    { value: 'all', label: t('Allow all origins') },
                  ]}
                  value={field.value}
                  onValueChange={field.onChange}
                >
                  <FormControl>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent alignItemWithTrigger={false}>
                    <SelectGroup>
                      <SelectItem value='disabled'>{t('Disabled')}</SelectItem>
                      <SelectItem value='allowlist'>
                        {t('Allow listed origins')}
                      </SelectItem>
                      <SelectItem value='all'>
                        {t('Allow all origins')}
                      </SelectItem>
                    </SelectGroup>
                  </SelectContent>
                </Select>
                <FormDescription>
                  {mode === 'all'
                    ? t(
                        'Any website can read public API responses. Cookie credentials are not allowed; API keys are still required.'
                      )
                    : mode === 'disabled'
                      ? t('Browser cross-origin access to /v1 is disabled.')
                      : t(
                          'Only the listed website origins can read public API responses.'
                        )}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
          {mode === 'allowlist' && (
            <FormField
              control={form.control}
              name='allowedOrigins'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Allowed website origins')}</FormLabel>
                  <FormControl>
                    <Textarea
                      rows={6}
                      placeholder={
                        'https://app.example.com\nhttps://console.example.com'
                      }
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    {t(
                      'Enter one full origin per line, including http:// or https://. Do not include a path, query, or fragment.'
                    )}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          )}
          <p className='text-muted-foreground text-xs'>
            {t(
              'This policy only affects /v1 APIs. Admin and dashboard endpoints under /api keep the existing allowlist and cookie behavior.'
            )}
          </p>
        </SettingsForm>
      </Form>
    </SettingsSection>
  )
}
