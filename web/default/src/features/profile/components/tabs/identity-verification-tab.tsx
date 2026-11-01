import { useQuery } from '@tanstack/react-query'
import { CheckCircle2, ShieldCheck } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

import { getIdentityVerification, submitEnterpriseVerification, submitPersonalVerification } from '../../api'

export function IdentityVerificationTab() {
  const { t } = useTranslation()
  const query = useQuery({ queryKey: ['identity-verification'], queryFn: getIdentityVerification })
  const data = query.data?.data
  const [method, setMethod] = useState<'personal' | 'enterprise'>('personal')
  const [personal, setPersonal] = useState({ name: '', idCard: '' })
  const [enterprise, setEnterprise] = useState({ company: '', registration: '', name: '', idCard: '' })
  const [submitting, setSubmitting] = useState(false)

  if (query.isLoading) return <p className='text-muted-foreground text-sm'>{t('Loading...')}</p>
  if (!data?.enabled) return <p className='text-muted-foreground text-sm'>{t('Real-name verification is not enabled')}</p>
  if (data.verification) {
    return (
      <div className='flex items-start gap-3 rounded-xl border p-4'>
        <CheckCircle2 className='text-emerald-600 mt-0.5 h-5 w-5' />
        <div className='space-y-1 text-sm'>
          <p className='font-medium'>{t('Verified')}</p>
          <p className='text-muted-foreground'>{data.verification.type === 'personal' ? t('Personal verification') : t('Enterprise verification')}</p>
          <p className='text-muted-foreground'>{data.verification.display_name || data.verification.company_name_masked}</p>
        </div>
      </div>
    )
  }
  const methods = data.available_methods
  if (!methods.length) return <p className='text-muted-foreground text-sm'>{t('No verification method is currently available')}</p>

  const submit = async () => {
    setSubmitting(true)
    try {
      const result = method === 'personal'
        ? await submitPersonalVerification(personal.name, personal.idCard)
        : await submitEnterpriseVerification({ company_name: enterprise.company, registration_number: enterprise.registration, name: enterprise.name, id_card: enterprise.idCard })
      if (result.success) {
        toast.success(t('Verification completed'))
        await query.refetch()
      } else toast.error(result.message || t('Verification failed'))
    } catch (error) {
      toast.error(error instanceof Error ? error.message : t('Verification service is temporarily unavailable'))
    } finally { setSubmitting(false) }
  }

  return (
    <div className='space-y-4'>
      <div className='flex items-center gap-2'><ShieldCheck className='h-4 w-4' /><p className='text-muted-foreground text-sm'>{t('Choose one verification method. Successful verification cannot be changed by the user.')}</p></div>
      {methods.length > 1 && <div className='flex gap-2'>{methods.map((value) => <Button key={value} type='button' size='sm' variant={method === value ? 'default' : 'outline'} onClick={() => setMethod(value)}>{value === 'personal' ? t('Personal verification') : t('Enterprise verification')}</Button>)}</div>}
      {method === 'personal' && methods.includes('personal') && <div className='space-y-3'>
        <Input placeholder={t('Full name')} value={personal.name} onChange={(e) => setPersonal({ ...personal, name: e.target.value })} />
        <Input placeholder={t('Identity card number')} value={personal.idCard} onChange={(e) => setPersonal({ ...personal, idCard: e.target.value })} />
      </div>}
      {method === 'enterprise' && methods.includes('enterprise') && <div className='space-y-3'>
        <Input placeholder={t('Company name')} value={enterprise.company} onChange={(e) => setEnterprise({ ...enterprise, company: e.target.value })} />
        <Input placeholder={t('Unified social credit code or registration number')} value={enterprise.registration} onChange={(e) => setEnterprise({ ...enterprise, registration: e.target.value })} />
        <Input placeholder={t('Legal representative or key person name')} value={enterprise.name} onChange={(e) => setEnterprise({ ...enterprise, name: e.target.value })} />
        <Input placeholder={t('Identity card number')} value={enterprise.idCard} onChange={(e) => setEnterprise({ ...enterprise, idCard: e.target.value })} />
      </div>}
      <Button type='button' onClick={() => void submit()} disabled={submitting}>{submitting ? t('Submitting...') : t('Submit verification')}</Button>
    </div>
  )
}
