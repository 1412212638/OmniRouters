import { Plus, Search } from 'lucide-react'
import { useState, useMemo } from 'react'
import { useTranslation } from 'react-i18next'

import { StaticDataTable } from '@/components/data-table/static/static-data-table'
import { StaticRowActions } from '@/components/data-table/static/static-row-actions'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

import { safeJsonParseWithValidation } from '../utils/json-parser'
import { isObjectRecord } from '../utils/json-validators'
import { RateLimitDialog, type RateLimitEntryData } from './rate-limit-dialog'

type RateLimitVisualEditorProps = {
  value: string
  onChange: (value: string) => void
}

type RateLimitEntry = RateLimitEntryData

function parseRateLimitValue(
  groupName: string,
  limits: unknown
): RateLimitEntry | null {
  if (limits && typeof limits === 'object' && !Array.isArray(limits)) {
    const value = limits as Record<string, unknown>
    return {
      groupName,
      rpm: typeof value.rpm === 'number' ? value.rpm : 0,
      tpm: typeof value.tpm === 'number' ? value.tpm : 0,
      rpd: typeof value.rpd === 'number' ? value.rpd : 0,
    }
  }
  if (
    Array.isArray(limits) &&
    limits.length === 2 &&
    limits.every((item) => typeof item === 'number')
  ) {
    return {
      groupName,
      rpm: limits[0] as number,
      tpm: 0,
      rpd: limits[1] as number,
    }
  }
  return null
}

export function RateLimitVisualEditor({
  value,
  onChange,
}: RateLimitVisualEditorProps) {
  const { t } = useTranslation()
  const [searchText, setSearchText] = useState('')
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editData, setEditData] = useState<RateLimitEntry | null>(null)

  const rateLimits = useMemo(() => {
    if (!value || value.trim() === '') return []

    const parsed = safeJsonParseWithValidation<Record<string, unknown>>(value, {
      fallback: {},
      validator: isObjectRecord,
      validatorMessage: 'Rate limits must be a JSON object',
      context: 'rate limits',
    })

    return Object.entries(parsed)
      .map(([groupName, limits]) => parseRateLimitValue(groupName, limits))
      .filter((item): item is RateLimitEntry => item !== null)
  }, [value])

  const filteredRateLimits = useMemo(() => {
    if (!searchText) return rateLimits
    const lowerSearch = searchText.toLowerCase()
    return rateLimits.filter((limit) =>
      limit.groupName.toLowerCase().includes(lowerSearch)
    )
  }, [rateLimits, searchText])

  const handleSave = (data: RateLimitEntryData) => {
    const parsed = safeJsonParseWithValidation<Record<string, unknown>>(value, {
      fallback: {},
      validator: isObjectRecord,
      silent: true,
    })

    if (editData && editData.groupName !== data.groupName) {
      delete parsed[editData.groupName]
    }

    parsed[data.groupName] = {
      rpm: data.rpm,
      tpm: data.tpm,
      rpd: data.rpd,
    }

    onChange(JSON.stringify(parsed, null, 2))
  }

  const handleDelete = (groupName: string) => {
    const parsed = safeJsonParseWithValidation<Record<string, unknown>>(value, {
      fallback: {},
      validator: isObjectRecord,
      silent: true,
    })

    delete parsed[groupName]

    onChange(JSON.stringify(parsed, null, 2))
  }

  const handleEdit = (limit: RateLimitEntry) => {
    setEditData(limit)
    setDialogOpen(true)
  }

  const handleAdd = () => {
    setEditData(null)
    setDialogOpen(true)
  }

  return (
    <div className='space-y-4'>
      <div className='flex items-center gap-4'>
        <div className='relative flex-1'>
          <Search className='text-muted-foreground absolute top-2.5 left-2.5 h-4 w-4' />
          <Input
            placeholder={t('Search group names...')}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            className='pl-9'
          />
        </div>
        <Button onClick={handleAdd}>
          <Plus className='mr-2 h-4 w-4' />
          {t('Add group')}
        </Button>
      </div>

      <StaticDataTable
        data={filteredRateLimits}
        getRowKey={(limit) => limit.groupName}
        emptyContent={
          searchText
            ? t('No groups match your search')
            : t(
                'No group-based rate limits configured. Click "Add group" to get started.'
              )
        }
        columns={[
          {
            id: 'group',
            header: t('Group Name'),
            cellClassName: 'font-medium',
            cell: (limit) => limit.groupName,
          },
          {
            id: 'rpm',
            header: 'RPM',
            className: 'text-right',
            cellClassName: 'text-right',
            cell: (limit) => (
              <span className='font-mono'>
                {limit.rpm === 0 ? t('Unlimited') : limit.rpm.toLocaleString()}
              </span>
            ),
          },
          {
            id: 'tpm',
            header: 'TPM',
            className: 'text-right',
            cellClassName: 'text-right',
            cell: (limit) => (
              <span className='font-mono'>
                {limit.tpm === 0 ? t('Unlimited') : limit.tpm.toLocaleString()}
              </span>
            ),
          },
          {
            id: 'rpd',
            header: 'RPD',
            className: 'text-right',
            cellClassName: 'text-right',
            cell: (limit) => (
              <span className='font-mono'>
                {limit.rpd === 0 ? t('Unlimited') : limit.rpd.toLocaleString()}
              </span>
            ),
          },
          {
            id: 'actions',
            header: t('Actions'),
            className: 'text-right',
            cellClassName: 'text-right',
            cell: (limit) => (
              <StaticRowActions
                editLabel={t('Edit')}
                deleteLabel={t('Delete')}
                menuLabel={t('Open menu')}
                onEdit={() => handleEdit(limit)}
                onDelete={() => handleDelete(limit.groupName)}
              />
            ),
          },
        ]}
      />

      <RateLimitDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        onSave={handleSave}
        editData={editData}
      />
    </div>
  )
}
