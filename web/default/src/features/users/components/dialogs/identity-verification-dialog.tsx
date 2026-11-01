import { useQuery } from '@tanstack/react-query'
import { ShieldCheck } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from '@/components/ui/sheet'

import { getUserIdentityVerification, resetUserIdentityVerification, viewUserIdentityVerification } from '../../api'

type Props = { userId: number; username: string; open: boolean; onOpenChange: (open: boolean) => void; onReset: () => void }

export function IdentityVerificationDialog({ userId, username, open, onOpenChange, onReset }: Props) {
  const { t } = useTranslation()
  const [full, setFull] = useState(false)
  const [reason, setReason] = useState('')
  const summary = useQuery({ queryKey: ['admin-identity-verification', userId], queryFn: () => getUserIdentityVerification(userId), enabled: open })
  const details = useQuery({ queryKey: ['admin-identity-verification-full', userId], queryFn: () => viewUserIdentityVerification(userId), enabled: open && full })
  const reset = async () => {
    const result = await resetUserIdentityVerification(userId, reason)
    if (result.success) { toast.success(t('Verification reset')); setReason(''); onOpenChange(false); onReset() }
    else toast.error(result.message || t('Failed to reset verification'))
  }
  const record = summary.data?.data
  const detail = details.data?.data
  return <Sheet open={open} onOpenChange={onOpenChange}>
    <SheetContent className='sm:max-w-[520px]'>
      <SheetHeader><SheetTitle>{t('Real-name Verification')}</SheetTitle><SheetDescription>{username}</SheetDescription></SheetHeader>
      <div className='space-y-4 p-4'>
        {!record ? <p className='text-muted-foreground text-sm'>{t('Not verified')}</p> : <>
          <div className='flex items-center gap-2 text-sm'><ShieldCheck className='h-4 w-4 text-emerald-600' />{record.type === 'personal' ? t('Personal verification') : t('Enterprise verification')}</div>
          <p className='text-sm'>{record.display_name || record.company_name_masked}</p>
          <p className='text-muted-foreground text-xs'>{t('ID card suffix')}: {record.id_card_suffix || '-'}</p>
          {full && detail && <div className='space-y-2 rounded-lg border p-3 text-sm'><p>{detail.name}</p><p>{detail.id_card}</p><p>{detail.company_name}</p><p>{detail.registration_number}</p></div>}
          <Button type='button' variant='outline' onClick={() => setFull(true)}>{t('View full information')}</Button>
          <Separator />
          <div className='space-y-2'><Label>{t('Reset reason')}</Label><Input value={reason} onChange={(e) => setReason(e.target.value)} placeholder={t('Explain why this verification must be reset')} /></div>
          <SheetFooter><Button type='button' variant='destructive' disabled={reason.trim().length < 5} onClick={() => void reset()}>{t('Reset verification')}</Button></SheetFooter>
        </>}
      </div>
    </SheetContent>
  </Sheet>
}
