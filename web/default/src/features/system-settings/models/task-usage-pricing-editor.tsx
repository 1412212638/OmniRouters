import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Field, FieldDescription, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import {
  createDefaultTaskMatrixConfig,
  generateTaskExprFromConfig,
  getTaskEnumFields,
  getTaskNumberFields,
  taskMatrixRowLabel,
  taskMatrixToTiers,
  tryParseTaskMatrixConfig,
  type TaskMatrixConfig,
} from '@/features/pricing/lib/task-expr'
import {
  taskPriceLabel,
  taskUsageUnitLabel,
} from '@/features/pricing/lib/task-price-display'
import type { BillingUsageSchema } from '@/features/pricing/types'

function normalizeConfig(config: TaskMatrixConfig, schema: BillingUsageSchema) {
  return {
    rows:
      config.rows.length > 0
        ? config.rows
        : createDefaultTaskMatrixConfig(schema).rows,
  }
}

type TaskUsagePricingEditorProps = {
  schema?: BillingUsageSchema
  expression: string
  onExpressionChange: (expression: string) => void
}

export function TaskUsagePricingEditor({
  schema,
  expression,
  onExpressionChange,
}: TaskUsagePricingEditorProps) {
  const { t, i18n } = useTranslation()
  const [config, setConfig] = useState<TaskMatrixConfig | null>(null)
  const numberFields = useMemo(() => getTaskNumberFields(schema), [schema])
  const enumFields = useMemo(() => getTaskEnumFields(schema), [schema])

  useEffect(() => {
    if (!schema || numberFields.length === 0) {
      setConfig(null)
      return
    }
    const parsed = tryParseTaskMatrixConfig(expression, schema)
    setConfig(
      normalizeConfig(parsed ?? createDefaultTaskMatrixConfig(schema), schema)
    )
  }, [expression, numberFields.length, schema])

  if (!schema || numberFields.length === 0 || !config) return null

  const updateConfig = (next: TaskMatrixConfig) => {
    setConfig(next)
    const tiers = taskMatrixToTiers(next, schema)
    const nextExpression = generateTaskExprFromConfig({ tiers }, schema)
    if (nextExpression) onExpressionChange(nextExpression)
  }

  const updatePrice = (rowIndex: number, field: string, value: string) => {
    const numeric = Number(value)
    const safeValue = Number.isFinite(numeric) && numeric >= 0 ? numeric : 0
    updateConfig({
      rows: config.rows.map((row, index) =>
        index === rowIndex
          ? { ...row, unitPrices: { ...row.unitPrices, [field]: safeValue } }
          : row
      ),
    })
  }

  const updateConstant = (rowIndex: number, value: string) => {
    const numeric = Number(value)
    const safeValue = Number.isFinite(numeric) && numeric >= 0 ? numeric : 0
    updateConfig({
      rows: config.rows.map((row, index) =>
        index === rowIndex ? { ...row, constant: safeValue } : row
      ),
    })
  }

  return (
    <Field className='rounded-lg border p-4'>
      <FieldLabel>{t('Task pricing')}</FieldLabel>
      <FieldDescription>
        {t(
          'Configure task prices by usage fields such as duration, count, or resolution. The generated expression is used for billing.'
        )}
      </FieldDescription>
      <Alert className='mt-3'>
        <AlertDescription className='text-xs leading-5'>
          {t(
            'Prices are charged per usage unit. Token usage fields are priced per 1M tokens.'
          )}
        </AlertDescription>
      </Alert>
      <div className='mt-4 overflow-x-auto rounded-md border'>
        <table className='w-full min-w-[620px] text-sm'>
          <thead className='bg-muted/40'>
            <tr className='border-b'>
              <th className='px-3 py-2 text-left font-medium'>
                {t('Task condition')}
              </th>
              <th className='px-3 py-2 text-right font-medium'>
                {t('Fixed price')}
              </th>
              {numberFields.map(([field, definition]) => (
                <th key={field} className='px-3 py-2 text-right font-medium'>
                  {taskPriceLabel(definition.description, field, i18n.language)}
                  <span className='text-muted-foreground ml-1 text-xs font-normal'>
                    /
                    {taskUsageUnitLabel(
                      definition,
                      i18n.language,
                      definition.unit || t('unit')
                    )}
                  </span>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {config.rows.map((row, index) => (
              <tr
                key={`${taskMatrixRowLabel(row.combination)}-${index}`}
                className='border-b last:border-b-0'
              >
                <td className='px-3 py-2 font-medium'>
                  {taskMatrixRowLabel(row.combination)}
                </td>
                <td className='px-3 py-2'>
                  <Input
                    inputMode='decimal'
                    className='h-8 min-w-24 text-right font-mono'
                    value={String(row.constant || 0)}
                    onChange={(event) =>
                      updateConstant(index, event.target.value)
                    }
                    aria-label={`${t('Fixed price')} ${taskMatrixRowLabel(row.combination)}`}
                  />
                </td>
                {numberFields.map(([field]) => (
                  <td key={field} className='px-3 py-2'>
                    <Input
                      inputMode='decimal'
                      className='h-8 min-w-24 text-right font-mono'
                      value={String(row.unitPrices[field] || 0)}
                      onChange={(event) =>
                        updatePrice(index, field, event.target.value)
                      }
                      aria-label={`${field} ${taskMatrixRowLabel(row.combination)}`}
                    />
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {enumFields.length === 0 && (
        <p className='text-muted-foreground mt-2 text-xs'>
          {t(
            'This model has no enum conditions, so one base pricing row is shown.'
          )}
        </p>
      )}
      <div className='mt-3 flex items-center justify-between gap-3'>
        <code className='bg-muted min-w-0 flex-1 truncate rounded px-2 py-1 text-xs'>
          {expression ||
            t('Expression will be generated after entering prices')}
        </code>
        <Button
          type='button'
          variant='ghost'
          size='sm'
          onClick={() => {
            if (!schema) return
            const reset = createDefaultTaskMatrixConfig(schema)
            updateConfig(reset)
          }}
        >
          {t('Reset')}
        </Button>
      </div>
    </Field>
  )
}
