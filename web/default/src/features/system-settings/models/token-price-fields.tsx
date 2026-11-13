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
import { useId, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field'
import { Separator } from '@/components/ui/separator'
import { Switch } from '@/components/ui/switch'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import { BILLING_EXTRA_VARS } from '@/features/pricing/lib/billing-expr'
import type { VisualTier } from '@/features/pricing/lib/tier-expr'

import { DraftNumberInput } from './pricing-number-input'

const SCENARIOS = [
  {
    key: 'general',
    fields: ['cr'],
    label: 'pricingEditor.general',
    help: 'pricingEditor.generalHelp',
  },
  {
    key: 'claude',
    fields: ['cr', 'cc', 'cc1h'],
    label: 'pricingEditor.claude',
    help: 'pricingEditor.claudeHelp',
  },
  {
    key: 'gemini',
    fields: ['cr', 'img', 'img_o', 'ai', 'ao'],
    label: 'pricingEditor.gemini',
    help: 'pricingEditor.geminiHelp',
  },
  {
    key: 'openai-image',
    fields: ['cr', 'img', 'img_cr'],
    label: 'pricingEditor.openaiImage',
    help: 'pricingEditor.openaiImageHelp',
  },
  {
    key: 'custom',
    fields: BILLING_EXTRA_VARS.map((variable) => variable.key),
    label: 'pricingEditor.custom',
    help: 'pricingEditor.customHelp',
  },
] as const

type ScenarioKey = (typeof SCENARIOS)[number]['key']

function isScenarioKey(value: unknown): value is ScenarioKey {
  return SCENARIOS.some((scenario) => scenario.key === value)
}

type OptionalPriceProps = {
  variable: (typeof BILLING_EXTRA_VARS)[number]
  value: number | undefined
  onChange: (value: number | undefined) => void
}

function OptionalPrice(props: OptionalPriceProps) {
  const { t } = useTranslation()
  const id = useId()
  const enabled = props.value != null
  const label = t(`pricingEditor.${props.variable.key}Label`)

  return (
    <Field className='min-w-0 rounded-lg border p-3'>
      <div className='flex items-start justify-between gap-2'>
        <FieldLabel htmlFor={id}>{label}</FieldLabel>
        <Switch
          id={id}
          checked={enabled}
          onCheckedChange={(checked) => props.onChange(checked ? 0 : undefined)}
          aria-label={t('pricingEditor.enableCategory', { category: label })}
        />
      </div>
      <FieldDescription>
        {t(`pricingEditor.${props.variable.key}Help`)}
      </FieldDescription>
      {enabled ? (
        <>
          <DraftNumberInput
            min={0}
            step={0.000001}
            value={props.value ?? 0}
            onValueChange={props.onChange}
            aria-label={label}
          />
          <div className='flex flex-wrap items-center gap-2'>
            <Badge variant='secondary'>{t('pricingEditor.tokenUnit')}</Badge>
            {props.value === 0 && (
              <Badge variant='outline'>{t('pricingEditor.categoryFree')}</Badge>
            )}
          </div>
        </>
      ) : (
        <FieldDescription>{t('pricingEditor.baseFallback')}</FieldDescription>
      )}
    </Field>
  )
}

type TokenPriceFieldsProps = {
  tier: VisualTier
  onChange: (tier: VisualTier) => void
}

export function TokenPriceFields(props: TokenPriceFieldsProps) {
  const { t } = useTranslation()
  const id = useId()
  const [scenarioKey, setScenarioKey] = useState<ScenarioKey>('general')
  const [advancedOpen, setAdvancedOpen] = useState(false)
  const scenario =
    SCENARIOS.find((entry) => entry.key === scenarioKey) ?? SCENARIOS[0]
  const recommended = new Set<string>(scenario.fields)
  // Already configured categories stay visible in every scenario, including
  // zero-price categories. Scenario selection must never change billing.
  const primaryFields = BILLING_EXTRA_VARS.filter(
    (variable) =>
      recommended.has(variable.key) ||
      (variable.tierField != null && props.tier[variable.tierField] != null)
  )
  const advancedFields = BILLING_EXTRA_VARS.filter(
    (variable) => !primaryFields.includes(variable)
  )
  const changePrice = (field: string, value: number | undefined) => {
    const next = { ...props.tier }
    if (value == null) delete next[field]
    else next[field] = value
    props.onChange(next)
  }
  const renderCategory = (variable: (typeof BILLING_EXTRA_VARS)[number]) => {
    const field = variable.tierField
    if (field == null) return null
    return (
      <OptionalPrice
        key={variable.key}
        variable={variable}
        value={props.tier[field] as number | undefined}
        onChange={(value) => changePrice(field, value)}
      />
    )
  }

  return (
    <FieldGroup>
      <Field>
        <FieldLabel id={`${id}-scenario`}>
          {t('pricingEditor.scenario')}
        </FieldLabel>
        <ToggleGroup
          value={[scenarioKey]}
          onValueChange={(values) => {
            if (isScenarioKey(values[0])) setScenarioKey(values[0])
          }}
          variant='outline'
          size='sm'
          spacing={1}
          className='max-w-full flex-wrap'
          aria-labelledby={`${id}-scenario`}
        >
          {SCENARIOS.map((entry) => (
            <ToggleGroupItem key={entry.key} value={entry.key}>
              {t(entry.label)}
            </ToggleGroupItem>
          ))}
        </ToggleGroup>
        <FieldDescription>{t(scenario.help)}</FieldDescription>
        <FieldDescription>{t('pricingEditor.displayOnly')}</FieldDescription>
      </Field>

      <Alert>
        <AlertDescription>{t('pricingEditor.unitsHelp')}</AlertDescription>
      </Alert>

      <FieldGroup className='grid grid-cols-1 gap-3 sm:grid-cols-2'>
        {(['input_unit_cost', 'output_unit_cost'] as const).map((field) => {
          const label =
            field === 'input_unit_cost'
              ? t('pricingEditor.inputLabel')
              : t('pricingEditor.outputLabel')
          return (
            <Field key={field}>
              <FieldLabel htmlFor={`${id}-${field}`}>{label}</FieldLabel>
              <DraftNumberInput
                id={`${id}-${field}`}
                min={0}
                step={0.000001}
                value={props.tier[field]}
                onValueChange={(value) => changePrice(field, value)}
              />
              <FieldDescription>
                {t('pricingEditor.tokenUnit')}
              </FieldDescription>
            </Field>
          )
        })}
      </FieldGroup>

      <Separator />
      <Field>
        <FieldLabel>{t('pricingEditor.separatePricing')}</FieldLabel>
        <FieldDescription>{t('pricingEditor.separateHelp')}</FieldDescription>
      </Field>
      <FieldGroup className='grid grid-cols-1 items-start gap-3 sm:grid-cols-2'>
        {primaryFields.map(renderCategory)}
      </FieldGroup>
      {advancedFields.length > 0 && (
        <div className='flex flex-col gap-3'>
          <Button
            type='button'
            variant='ghost'
            size='sm'
            className='self-start'
            aria-expanded={advancedOpen}
            aria-controls={`${id}-advanced`}
            onClick={() => setAdvancedOpen((open) => !open)}
          >
            {advancedOpen
              ? t('pricingEditor.hideAdvanced')
              : t('pricingEditor.showAdvanced')}
          </Button>
          {advancedOpen && (
            <FieldGroup
              id={`${id}-advanced`}
              className='grid grid-cols-1 items-start gap-3 sm:grid-cols-2'
            >
              {advancedFields.map(renderCategory)}
            </FieldGroup>
          )}
        </div>
      )}
    </FieldGroup>
  )
}
