import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import { AlertTriangle, RefreshCw } from 'lucide-react'
import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { ROLE } from '@/lib/roles'
import { useAuthStore } from '@/stores/auth-store'

import { getTaskPricingWarnings } from '../api'

const REASONS: Record<string, string> = {
  missing_price: 'No task price configured',
  missing_expression: 'Task billing expression is missing',
  invalid_expression: 'Task billing expression is invalid',
  incompatible_expression: 'Task billing expression does not match this plugin',
}

export function TaskPricingWarnings({
  showPricingLink = true,
}: {
  showPricingLink?: boolean
}) {
  const { t } = useTranslation()
  const isRoot = useAuthStore(
    (state) => state.auth.user?.role === ROLE.SUPER_ADMIN
  )
  const queryClient = useQueryClient()
  const query = useQuery({
    queryKey: ['task-pricing-warnings'],
    queryFn: getTaskPricingWarnings,
    enabled: isRoot,
    retry: false,
    refetchOnMount: 'always',
    refetchOnWindowFocus: 'always',
  })

  // Recheck after the existing editors read back persisted pricing, channels,
  // plugin versions or the global switch. No billing mutation is introduced.
  useEffect(() => {
    if (!isRoot) return
    return queryClient.getQueryCache().subscribe((event) => {
      if (event.type !== 'updated' || event.action.type !== 'success') return
      if (
        [
          'system-options',
          'model-pricing-snapshot',
          'task-plugins',
          'task-plugin-enabled',
          'channels',
        ].includes(String(event.query.queryKey[0]))
      ) {
        void queryClient.invalidateQueries({
          queryKey: ['task-pricing-warnings'],
        })
      }
    })
  }, [isRoot, queryClient])

  if (!isRoot) return null
  if (query.isPending) return <Skeleton className='h-16 w-full shrink-0' />
  const data = query.data
  const hasWarnings = Boolean(data?.warnings.length)
  return (
    <Alert className='shrink-0'>
      <AlertTriangle />
      <AlertTitle>
        <div className='flex flex-wrap items-center justify-between gap-2'>
          <span>{t('Task pricing check')}</span>
          <div className='flex items-center gap-2'>
            {showPricingLink && hasWarnings && (
              <Button
                variant='outline'
                size='sm'
                render={
                  <Link
                    to='/system-settings/billing/$section'
                    params={{ section: 'model-pricing' }}
                  />
                }
              >
                {t('Configure prices')}
              </Button>
            )}
            <Button
              variant='ghost'
              size='sm'
              disabled={query.isFetching}
              onClick={() => void query.refetch()}
            >
              <RefreshCw data-icon='inline-start' />
              {t('Refresh')}
            </Button>
          </div>
        </div>
      </AlertTitle>
      <AlertDescription className='min-w-0'>
        {query.isError && (
          <span>{t('Failed to check task pricing. Please retry.')}</span>
        )}
        {!query.isError && !data?.enabled && (
          <span>
            {t('Task plugins are disabled; pricing checks are paused.')}
          </span>
        )}
        {!query.isError && data?.enabled && (
          <div className='flex min-w-0 flex-col gap-2'>
            <span>
              {t(
                'Checks enabled channels and model mappings. Zero prices are valid. Request-dependent billing is still validated when calling the model.'
              )}
            </span>
            {hasWarnings ? (
              <div className='max-h-52 overflow-auto'>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('Model')}</TableHead>
                      <TableHead>{t('Plugin')}</TableHead>
                      <TableHead>{t('Channel IDs')}</TableHead>
                      <TableHead>{t('Reason')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {data.warnings.map((warning) => (
                      <TableRow
                        key={`${warning.model_name}:${warning.plugin_key}:${warning.reason}`}
                      >
                        <TableCell className='max-w-64 break-all whitespace-normal'>
                          {warning.model_name}
                        </TableCell>
                        <TableCell>
                          {warning.plugin_name || warning.plugin_key}
                        </TableCell>
                        <TableCell className='max-w-48 break-words whitespace-normal'>
                          {warning.channel_ids.join(', ')}
                        </TableCell>
                        <TableCell>
                          {t(
                            REASONS[warning.reason] ??
                              'Task pricing needs review'
                          )}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            ) : (
              <span>
                {t(
                  'No task pricing configuration issues found in the checked channels.'
                )}
              </span>
            )}
            {Boolean(data?.skipped_channels.length) && (
              <span>
                {t(
                  'Some channels could not be fully checked because mappings or plugin ownership are ambiguous. Review channel IDs: {{ids}}',
                  { ids: data?.skipped_channels.join(', ') }
                )}
              </span>
            )}
          </div>
        )}
      </AlertDescription>
    </Alert>
  )
}
