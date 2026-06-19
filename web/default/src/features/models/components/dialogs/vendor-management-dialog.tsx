/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { useMemo, useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Building2, Loader2, Pencil, Plus, Search, Trash2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { getLobeIcon } from '@/lib/lobe-icon'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { Dialog } from '@/components/dialog'
import { getVendors } from '../../api'
import { handleDeleteVendor, vendorsQueryKeys } from '../../lib'
import type { Vendor } from '../../types'
import { VendorMutateDialog } from './vendor-mutate-dialog'

type VendorManagementDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
}

function VendorIcon(props: { vendor: Vendor }) {
  const iconKey = props.vendor.icon?.trim()
  const icon = iconKey ? getLobeIcon(iconKey, 20) : null

  return (
    <span className='bg-muted flex size-8 shrink-0 items-center justify-center rounded-md'>
      {icon || (
        <span className='text-muted-foreground text-xs font-semibold'>
          {props.vendor.name.charAt(0).toUpperCase()}
        </span>
      )}
    </span>
  )
}

export function VendorManagementDialog({
  open,
  onOpenChange,
}: VendorManagementDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [keyword, setKeyword] = useState('')
  const [mutateOpen, setMutateOpen] = useState(false)
  const [editingVendor, setEditingVendor] = useState<Vendor | null>(null)
  const [deletingVendor, setDeletingVendor] = useState<Vendor | null>(null)

  const vendorsQuery = useQuery({
    queryKey: vendorsQueryKeys.list({ page_size: 1000 }),
    queryFn: () => getVendors({ page_size: 1000 }),
    enabled: open,
  })

  const filteredVendors = useMemo(() => {
    const vendors = vendorsQuery.data?.data?.items || []
    const query = keyword.trim().toLowerCase()
    if (!query) return vendors
    return vendors.filter((vendor) =>
      [vendor.name, vendor.description, vendor.icon]
        .filter(Boolean)
        .join(' ')
        .toLowerCase()
        .includes(query)
    )
  }, [keyword, vendorsQuery.data?.data?.items])

  const handleCreate = () => {
    setEditingVendor(null)
    setMutateOpen(true)
  }

  const handleEdit = (vendor: Vendor) => {
    setEditingVendor(vendor)
    setMutateOpen(true)
  }

  const handleMutateOpenChange = (nextOpen: boolean) => {
    setMutateOpen(nextOpen)
    if (!nextOpen) {
      setEditingVendor(null)
    }
  }

  return (
    <>
      <Dialog
        open={open}
        onOpenChange={onOpenChange}
        title={t('Manage Vendors')}
        description={t('Create, edit, or delete model vendors.')}
        contentClassName='sm:max-w-4xl'
        contentHeight='min(640px,calc(100vh-14rem))'
        footer={
          <>
            <Button
              type='button'
              variant='outline'
              onClick={() => onOpenChange(false)}
            >
              {t('Close')}
            </Button>
            <Button type='button' onClick={handleCreate}>
              <Plus data-icon='inline-start' />
              {t('Create Vendor')}
            </Button>
          </>
        }
      >
        <div className='flex min-h-0 flex-col gap-3'>
          <div className='relative'>
            <Search className='text-muted-foreground pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2' />
            <Input
              value={keyword}
              onChange={(event) => setKeyword(event.target.value)}
              placeholder={t('Search vendors...')}
              className='pl-9'
            />
          </div>

          <div className='min-h-0 rounded-md border'>
            {vendorsQuery.isLoading ? (
              <div className='text-muted-foreground flex h-40 items-center justify-center gap-2 text-sm'>
                <Loader2 className='size-4 animate-spin' />
                {t('Loading...')}
              </div>
            ) : filteredVendors.length === 0 ? (
              <div className='flex h-40 flex-col items-center justify-center gap-2 text-center'>
                <Building2 className='text-muted-foreground size-8' />
                <div className='text-sm font-medium'>
                  {keyword ? t('No vendors found') : t('No vendors yet')}
                </div>
                <div className='text-muted-foreground text-xs'>
                  {keyword
                    ? t('Try another vendor keyword.')
                    : t('Create a vendor before assigning it to models.')}
                </div>
              </div>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t('Vendor')}</TableHead>
                    <TableHead>{t('Description')}</TableHead>
                    <TableHead className='w-28 text-right'>
                      {t('Actions')}
                    </TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredVendors.map((vendor) => (
                    <TableRow key={vendor.id}>
                      <TableCell>
                        <div className='flex min-w-0 items-center gap-2'>
                          <VendorIcon vendor={vendor} />
                          <div className='min-w-0'>
                            <div className='truncate font-medium'>
                              {vendor.name}
                            </div>
                            {vendor.icon ? (
                              <div className='text-muted-foreground truncate text-xs'>
                                {vendor.icon}
                              </div>
                            ) : null}
                          </div>
                        </div>
                      </TableCell>
                      <TableCell className='text-muted-foreground max-w-[320px] truncate'>
                        {vendor.description || '-'}
                      </TableCell>
                      <TableCell>
                        <div className='flex justify-end gap-1'>
                          <Button
                            type='button'
                            variant='ghost'
                            size='icon-sm'
                            title={t('Edit')}
                            aria-label={t('Edit')}
                            onClick={() => handleEdit(vendor)}
                          >
                            <Pencil />
                          </Button>
                          <Button
                            type='button'
                            variant='ghost'
                            size='icon-sm'
                            title={t('Delete')}
                            aria-label={t('Delete')}
                            onClick={() => setDeletingVendor(vendor)}
                          >
                            <Trash2 />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </div>
        </div>
      </Dialog>

      <VendorMutateDialog
        open={mutateOpen}
        onOpenChange={handleMutateOpenChange}
        currentVendor={editingVendor}
      />

      <ConfirmDialog
        open={Boolean(deletingVendor)}
        onOpenChange={(nextOpen) => !nextOpen && setDeletingVendor(null)}
        title={t('Delete Vendor')}
        desc={t('Are you sure you want to delete vendor "{{name}}"?', {
          name: deletingVendor?.name,
        })}
        confirmText={t('Delete')}
        destructive
        handleConfirm={() => {
          if (!deletingVendor) return
          handleDeleteVendor(deletingVendor.id, queryClient, () => {
            setDeletingVendor(null)
          })
        }}
      />
    </>
  )
}
