import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { SectionPageLayout } from '@/components/layout'
import { getAuditLogs } from './api'

export function AuditLogs() {
  const { t } = useTranslation()
  const { data, isLoading } = useQuery({
    queryKey: ['audit-logs'],
    queryFn: () => getAuditLogs(),
  })

  return (
    <SectionPageLayout fixedContent>
      <SectionPageLayout.Title>{t('Audit Logs')}</SectionPageLayout.Title>
      <SectionPageLayout.Content>
        <div className='overflow-auto rounded-md border'>
          <table className='w-full text-sm'>
            <thead className='bg-muted/50 text-left'>
              <tr>
                <th className='p-3'>{t('Time')}</th>
                <th className='p-3'>{t('User')}</th>
                <th className='p-3'>{t('Category')}</th>
                <th className='p-3'>{t('Action')}</th>
                <th className='p-3'>{t('Status')}</th>
                <th className='p-3'>{t('IP')}</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                <tr><td className='p-4' colSpan={6}>{t('Loading...')}</td></tr>
              ) : data?.items?.length ? data.items.map((item) => (
                <tr key={item.id} className='border-t'>
                  <td className='p-3'>{new Date(item.created_at * 1000).toLocaleString()}</td>
                  <td className='p-3'>{item.username || item.user_id}</td>
                  <td className='p-3'>{item.category}</td>
                  <td className='p-3'>{item.action}</td>
                  <td className='p-3'>{item.success ? t('Success') : t('Failed')}</td>
                  <td className='p-3'>{item.ip || '-'}</td>
                </tr>
              )) : (
                <tr><td className='p-4' colSpan={6}>{t('No records')}</td></tr>
              )}
            </tbody>
          </table>
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
