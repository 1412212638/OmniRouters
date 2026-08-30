import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Puzzle } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { SectionPageLayout } from '@/components/layout'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { getTaskPlugins, updateTaskPlugin } from './api'
import type { TaskPlugin } from './types'

const text = (value: Record<string, string> | undefined, language: string) =>
  value?.[language] ?? value?.en ?? Object.values(value ?? {})[0] ?? ''

export function TaskPlugins() {
  const { t, i18n } = useTranslation()
  const queryClient = useQueryClient()
  const query = useQuery({ queryKey: ['task-plugins'], queryFn: getTaskPlugins })
  const mutation = useMutation({
    mutationFn: ({ plugin, enabled }: { plugin: TaskPlugin; enabled: boolean }) => updateTaskPlugin(plugin, enabled),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['task-plugins'] }),
  })
  const data = query.data?.data

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('Task Plugins')}</SectionPageLayout.Title>
      <SectionPageLayout.Description>{t('Manage built-in task plugins and their availability.')}</SectionPageLayout.Description>
      <SectionPageLayout.Content>
        <div className='space-y-3'>
          <div className='bg-muted/40 flex items-center gap-3 rounded-md border px-4 py-3 text-sm'>
            <Puzzle className='text-muted-foreground size-4' />
            <span className='flex-1'>{t('Plugin runtime')}</span>
            <Badge variant={data?.runtime_enabled ? 'default' : 'secondary'}>{data?.runtime_enabled ? t('Enabled') : t('Disabled')}</Badge>
            <span className='text-muted-foreground text-xs'>{t('Controlled by TASK_PLUGIN_ENABLED')}</span>
          </div>
          {query.isLoading && <div className='text-muted-foreground py-8 text-center'>{t('Loading...')}</div>}
          {data?.plugins.map((plugin) => (
            <div key={plugin.key} className='flex flex-wrap items-center gap-4 rounded-md border px-4 py-4'>
              <div className='min-w-0 flex-1'>
                <div className='flex flex-wrap items-center gap-2'><span className='font-medium'>{plugin.name || plugin.key}</span><Badge variant='outline'>{plugin.version}</Badge></div>
                <div className='text-muted-foreground mt-1 text-sm'>{text(plugin.description, i18n.language)}</div>
                <div className='text-muted-foreground mt-2 text-xs'>{plugin.key} · {t('Author')}: {plugin.author?.name || '-'}</div>
              </div>
              <Switch checked={plugin.enabled} disabled={!data.runtime_enabled || mutation.isPending} aria-label={`${t('Enable')} ${plugin.name}`} onCheckedChange={(enabled) => mutation.mutate({ plugin, enabled })} />
            </div>
          ))}
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
