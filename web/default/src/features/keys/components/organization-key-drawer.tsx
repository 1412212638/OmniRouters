import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'

import { createApiKey, updateApiKey } from '../api'
import type { ApiKey, ApiKeyFormData } from '../types'
import { useApiKeys } from './api-keys-provider'

export function OrganizationKeyDrawer({
  open,
  onOpenChange,
  currentRow,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentRow?: ApiKey
}) {
  const { t } = useTranslation()
  const { triggerRefresh } = useApiKeys()
  const [name, setName] = useState('')
  const [saving, setSaving] = useState(false)
  useEffect(() => {
    if (open) setName(currentRow?.name ?? '')
  }, [open, currentRow?.id, currentRow?.name])

  const save = async () => {
    const value = name.trim()
    if (!value || saving) return
    setSaving(true)
    try {
      const payload = { name: value } as ApiKeyFormData
      const response = currentRow
        ? await updateApiKey({ ...payload, id: currentRow.id })
        : await createApiKey(payload)
      if (!response.success) {
        throw new Error(response.message || t('Request failed'))
      }
      toast.success(t(currentRow ? 'Key updated' : 'Organization key created'))
      triggerRefresh()
      onOpenChange(false)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : t('Request failed'))
    } finally {
      setSaving(false)
    }
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent>
        <SheetHeader>
          <SheetTitle>
            {t(currentRow ? 'Update API Key' : 'Create API Key')}
          </SheetTitle>
          <SheetDescription>
            {t('Keys created here charge the organization balance.')}
          </SheetDescription>
        </SheetHeader>
        <FieldGroup className='p-4'>
          <Field>
            <FieldLabel htmlFor='organization-key-name'>
              {t('Key name')}
            </FieldLabel>
            <Input
              id='organization-key-name'
              value={name}
              maxLength={50}
              onChange={(event) => setName(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === 'Enter') {
                  event.preventDefault()
                  void save()
                }
              }}
            />
            <FieldDescription>
              {t(
                'Members share the organization balance according to their role and monthly limit.'
              )}
            </FieldDescription>
          </Field>
        </FieldGroup>
        <SheetFooter>
          <SheetClose render={<Button variant='outline' />}>
            {t('Close')}
          </SheetClose>
          <Button onClick={() => void save()} disabled={!name.trim() || saving}>
            {saving ? t('Saving...') : t('Save changes')}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
