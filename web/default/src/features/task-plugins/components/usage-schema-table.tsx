import { useTranslation } from 'react-i18next'

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import type { BillingUsageSchema } from '@/features/pricing/types'
import { resolveLocalizedText } from '@/lib/localized-text'

type UsageSchemaTableProps = {
  schema: BillingUsageSchema
}

function getUsageTypeLabelKey(
  type: BillingUsageSchema[string]['type']
): string {
  if (type === 'number') return 'Number'
  if (type === 'boolean') return 'Boolean'
  return 'Enum'
}

function formatUsageUnit(
  unit: BillingUsageSchema[string]['unit'],
  t: (key: string) => string
): string {
  if (unit === 'second') return t('Second')
  if (unit === 'count') return t('Count')
  if (unit === 'token') return t('token (unit)')
  if (unit === 'credit') return t('credit')
  return '—'
}

export function UsageSchemaTable(props: UsageSchemaTableProps) {
  const { t, i18n } = useTranslation()
  const entries = Object.entries(props.schema).sort(([left], [right]) =>
    left.localeCompare(right)
  )

  return (
    <div className='overflow-x-auto rounded-md border'>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t('Name')}</TableHead>
            <TableHead>{t('Type')}</TableHead>
            <TableHead>{t('Unit')}</TableHead>
            <TableHead>{t('Enum values')}</TableHead>
            <TableHead>{t('Description')}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {entries.map(([name, definition]) => (
            <TableRow key={name}>
              <TableCell className='font-mono'>{name}</TableCell>
              <TableCell>{t(getUsageTypeLabelKey(definition.type))}</TableCell>
              <TableCell>{formatUsageUnit(definition.unit, t)}</TableCell>
              <TableCell className='font-mono'>
                {definition.enum?.join(', ') || '—'}
              </TableCell>
              <TableCell className='min-w-48 whitespace-normal'>
                {resolveLocalizedText(definition.description, i18n.language) ||
                  '—'}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
