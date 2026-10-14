import { Building2, FileText, Pencil, Trash2, UserRound } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'

import { deleteInvoiceProfile, updateInvoiceProfile } from '../api'
import type { InvoiceProfile } from '../types'

interface InvoiceProfileCardProps {
  profile?: InvoiceProfile | null
  onSaved: (profile: InvoiceProfile | null) => void
}

const emptyProfile: InvoiceProfile = { type: 'personal', name: '', email: '' }

export function InvoiceProfileCard({
  profile,
  onSaved,
}: InvoiceProfileCardProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [clearOpen, setClearOpen] = useState(false)
  const [saving, setSaving] = useState(false)
  const [clearing, setClearing] = useState(false)
  const [form, setForm] = useState<InvoiceProfile>(profile || emptyProfile)

  useEffect(() => {
    if (open) setForm(profile ? { ...profile } : { ...emptyProfile })
  }, [open, profile])

  const setField = <K extends keyof InvoiceProfile>(
    key: K,
    value: InvoiceProfile[K]
  ) => {
    setForm((current) => ({ ...current, [key]: value }))
  }

  const save = async () => {
    setSaving(true)
    try {
      const response = await updateInvoiceProfile(form)
      if (!response.success && response.message !== 'success') {
        toast.error(t('Failed to save invoice information'))
        return
      }
      const saved = response.data || form
      onSaved(saved)
      setOpen(false)
      toast.success(t('Invoice information saved'))
    } catch {
      toast.error(t('Failed to save invoice information'))
    } finally {
      setSaving(false)
    }
  }

  const clear = async () => {
    setClearing(true)
    try {
      const response = await deleteInvoiceProfile()
      if (!response.success && response.message !== 'success') {
        toast.error(t('Failed to clear invoice information'))
        return
      }
      onSaved(null)
      setClearOpen(false)
      toast.success(t('Invoice information cleared'))
    } catch {
      toast.error(t('Failed to clear invoice information'))
    } finally {
      setClearing(false)
    }
  }

  const displayName =
    profile?.type === 'company' ? profile.company_name : profile?.name
  const address = [profile?.city, profile?.country].filter(Boolean).join(', ')

  return (
    <>
      <Card data-card-hover='false' className='gap-0 rounded-lg py-0'>
        <CardHeader className='flex flex-row items-start justify-between gap-3 border-b p-4 sm:p-5'>
          <div className='flex items-start gap-3'>
            <div className='bg-muted flex size-9 shrink-0 items-center justify-center rounded-md'>
              <FileText className='text-muted-foreground size-4' />
            </div>
            <div className='space-y-1'>
              <CardTitle>{t('Invoice Information')}</CardTitle>
              <CardDescription>
                {t(
                  'Saved to your account and used for Stripe and Waffo Pancake invoices'
                )}
              </CardDescription>
            </div>
          </div>
          <div className='flex shrink-0 gap-2'>
            {profile && (
              <Button
                variant='ghost'
                size='icon-sm'
                title={t('Clear invoice information')}
                onClick={() => setClearOpen(true)}
              >
                <Trash2 className='size-4' />
              </Button>
            )}
            <Button
              variant='outline'
              size='sm'
              className='gap-2'
              onClick={() => setOpen(true)}
            >
              <Pencil className='size-4' />
              {profile ? t('Edit') : t('Add information')}
            </Button>
          </div>
        </CardHeader>
        <CardContent className='p-4 sm:p-5'>
          {profile ? (
            <div className='flex flex-wrap items-center gap-x-5 gap-y-2 text-sm'>
              <span className='text-foreground inline-flex items-center gap-2 font-medium'>
                {profile.type === 'company' ? (
                  <Building2 className='size-4' />
                ) : (
                  <UserRound className='size-4' />
                )}
                {profile.type === 'company' ? t('Company') : t('Personal')}:{' '}
                {displayName}
              </span>
              <span className='text-muted-foreground'>{profile.email}</span>
              {address && (
                <span className='text-muted-foreground'>{address}</span>
              )}
              {profile.tax_id && (
                <span className='text-muted-foreground'>
                  {t('Tax ID')}: {profile.tax_id}
                </span>
              )}
            </div>
          ) : (
            <p className='text-muted-foreground text-sm'>
              {t('No invoice information saved')}
            </p>
          )}
        </CardContent>
      </Card>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className='max-h-[calc(100dvh-2rem)] overflow-y-auto sm:max-w-2xl'>
          <DialogHeader>
            <DialogTitle>{t('Invoice Information')}</DialogTitle>
            <DialogDescription>
              {t(
                'This information is optional and can be updated at any time.'
              )}
            </DialogDescription>
          </DialogHeader>
          <div className='space-y-5'>
            <div
              className='grid grid-cols-2 gap-2'
              role='group'
              aria-label={t('Invoice type')}
            >
              <Button
                type='button'
                variant={form.type === 'personal' ? 'default' : 'outline'}
                className='gap-2'
                onClick={() => setField('type', 'personal')}
              >
                <UserRound className='size-4' /> {t('Personal')}
              </Button>
              <Button
                type='button'
                variant={form.type === 'company' ? 'default' : 'outline'}
                className='gap-2'
                onClick={() => setField('type', 'company')}
              >
                <Building2 className='size-4' /> {t('Company')}
              </Button>
            </div>

            <div className='grid gap-4 sm:grid-cols-2'>
              {form.type === 'personal' ? (
                <Field
                  label={t('Full name')}
                  value={form.name || ''}
                  onChange={(v) => setField('name', v)}
                  required
                />
              ) : (
                <>
                  <Field
                    label={t('Legal company name')}
                    value={form.company_name || ''}
                    onChange={(v) => setField('company_name', v)}
                    required
                  />
                  <Field
                    label={t('Tax ID')}
                    value={form.tax_id || ''}
                    onChange={(v) => setField('tax_id', v)}
                  />
                  <Field
                    label={t('Contact name')}
                    value={form.contact_name || ''}
                    onChange={(v) => setField('contact_name', v)}
                  />
                </>
              )}
              <Field
                label={t('Invoice email')}
                value={form.email}
                onChange={(v) => setField('email', v)}
                type='email'
                required
              />
              <Field
                label={t('Country/Region (2-letter code)')}
                value={form.country || ''}
                onChange={(v) =>
                  setField('country', v.toUpperCase().slice(0, 2))
                }
                placeholder='US'
              />
              <Field
                label={t('State/Province')}
                value={form.state || ''}
                onChange={(v) => setField('state', v)}
              />
              <Field
                label={t('City')}
                value={form.city || ''}
                onChange={(v) => setField('city', v)}
              />
              <Field
                label={t('Postal code')}
                value={form.postcode || ''}
                onChange={(v) => setField('postcode', v)}
              />
              <div className='space-y-2 sm:col-span-2'>
                <Label htmlFor='invoice-address'>{t('Street address')}</Label>
                <Textarea
                  id='invoice-address'
                  value={form.address || ''}
                  onChange={(e) => setField('address', e.target.value)}
                  rows={2}
                />
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button
              variant='outline'
              onClick={() => setOpen(false)}
              disabled={saving}
            >
              {t('Cancel')}
            </Button>
            <Button
              onClick={save}
              disabled={
                saving ||
                (form.type === 'personal'
                  ? !form.name?.trim()
                  : !form.company_name?.trim()) ||
                !form.email.trim()
              }
            >
              {saving ? t('Saving...') : t('Save information')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog open={clearOpen} onOpenChange={setClearOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {t('Clear invoice information?')}
            </AlertDialogTitle>
            <AlertDialogDescription>
              {t(
                'This removes the saved information from your account. Existing invoices are not changed.'
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={clearing}>
              {t('Cancel')}
            </AlertDialogCancel>
            <Button variant='destructive' onClick={clear} disabled={clearing}>
              {clearing ? t('Clearing...') : t('Clear')}
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}

function Field({
  label,
  value,
  onChange,
  type = 'text',
  placeholder,
  required = false,
}: {
  label: string
  value: string
  onChange: (value: string) => void
  type?: string
  placeholder?: string
  required?: boolean
}) {
  const id = `invoice-${label.toLowerCase().replaceAll(/[^a-z0-9]+/g, '-')}`
  return (
    <div className='space-y-2'>
      <Label htmlFor={id}>
        {label}
        {required ? ' *' : ''}
      </Label>
      <Input
        id={id}
        type={type}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        required={required}
      />
    </div>
  )
}
