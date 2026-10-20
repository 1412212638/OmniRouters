import { useNavigate, useParams, useSearch } from '@tanstack/react-router'
import {
  ArrowLeft,
  ArrowRight,
  AudioLines,
  Code2,
  FileText,
  HeartPulse,
  Info,
  Image as ImageIcon,
  Type,
  Video,
} from 'lucide-react'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { CopyButton } from '@/components/copy-button'
import {
  StaticDataTable,
  type StaticDataTableColumn,
} from '@/components/data-table'
import { sideDrawerContentClassName } from '@/components/drawer-layout'
import { GroupBadge } from '@/components/group-badge'
import { PublicLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { getLobeIcon } from '@/lib/lobe-icon'

import { DEFAULT_TOKEN_UNIT, QUOTA_TYPE_VALUES } from '../constants'
import { usePricingData } from '../hooks/use-pricing-data'
import {
  getDynamicPriceEntries,
  getDynamicPricingSummary,
  getDynamicPricingTiers,
  getTaskPriceEntries,
  isDynamicPricingModel,
} from '../lib/dynamic-price'
import { getAvailableGroups, isTokenBasedModel } from '../lib/model-helpers'
import {
  formatFixedPrice,
  formatGroupPrice,
  getSoraPricingDisplay,
  isSoraPerRequestPricingModel,
  stripTrailingZeros,
} from '../lib/price'
import { formatToolPrice, getPricingToolPrices } from '../lib/tool-price'
import type {
  PriceType,
  PricingModel,
  PricingToolPrice,
  PricingUsableGroup,
  TokenUnit,
} from '../types'
import { getDynamicPricingRuleOptions } from './dynamic-pricing-breakdown'
import { ModelDetailsApi } from './model-details-api'
import { ModelDetailsPerformance } from './model-details-performance'
import { TaskPricingSection } from './task-pricing-section'

// ----------------------------------------------------------------------------
// Local UI helpers
// ----------------------------------------------------------------------------

const MODALITY_LABEL_KEYS: Record<string, string> = {
  decisions: 'Decisions',
  text: 'Text',
  image: 'Image',
  audio: 'Audio',
  video: 'Video',
  file: 'File',
}

const MODALITY_ICON_MAP: Record<
  string,
  React.ComponentType<{ className?: string }>
> = {
  decisions: Type,
  text: Type,
  image: ImageIcon,
  audio: AudioLines,
  video: Video,
  file: FileText,
  embedding: FileText,
  rerank: FileText,
  speech: AudioLines,
  transcription: AudioLines,
}

const TOKEN_FORMAT = new Intl.NumberFormat(undefined, {
  maximumFractionDigits: 1,
})

function formatCatalogTokenCount(tokens: number): string {
  if (!Number.isFinite(tokens) || tokens <= 0) return ''
  if (tokens >= 1_000_000) {
    return `${TOKEN_FORMAT.format(tokens / 1_000_000)}M`
  }
  if (tokens >= 1_000) {
    return `${TOKEN_FORMAT.format(tokens / 1_000)}K`
  }
  return TOKEN_FORMAT.format(tokens)
}

function normalizeCatalogItems(items?: readonly string[]): string[] {
  if (!items) return []
  return items.filter((item) => item.trim().length > 0)
}

function ModalityIcons(props: { items: string[] }) {
  const { t } = useTranslation()
  if (props.items.length === 0) return null

  return (
    <span className='inline-flex min-w-0 flex-wrap items-center gap-1.5'>
      {props.items.map((item) => {
        const Icon = MODALITY_ICON_MAP[item] ?? FileText
        const label = t(MODALITY_LABEL_KEYS[item] ?? item)
        return (
          <span
            key={item}
            title={label}
            aria-label={label}
            className='bg-muted text-foreground inline-flex size-8 items-center justify-center rounded-md'
          >
            <Icon className='size-4' aria-hidden='true' />
          </span>
        )
      })}
    </span>
  )
}

function ModelHeaderMeta(props: { model: PricingModel }) {
  const { t } = useTranslation()
  const model = props.model
  const inputModalities = normalizeCatalogItems(model.input_modalities)
  const outputModalities = normalizeCatalogItems(model.output_modalities)
  const contextLength =
    formatCatalogTokenCount(model.context_length ?? 0) ||
    model.context_length_display ||
    ''
  const maxOutput =
    formatCatalogTokenCount(model.max_output_tokens ?? 0) ||
    model.max_output_tokens_display ||
    ''

  return (
    <div className='mt-3 grid grid-cols-1 gap-2 @md/details:grid-cols-2'>
      <div className='bg-card min-w-0 rounded-lg border px-3 py-2.5'>
        <div className='text-muted-foreground truncate text-[10px] font-medium tracking-wider uppercase'>
          {t('Modality')}
        </div>
        <div className='text-foreground mt-1 flex min-w-0 items-center gap-2'>
          {inputModalities.length > 0 ? (
            <ModalityIcons items={inputModalities} />
          ) : (
            '—'
          )}
          <ArrowRight className='text-muted-foreground/60 size-3.5 shrink-0' />
          {outputModalities.length > 0 ? (
            <ModalityIcons items={outputModalities} />
          ) : (
            '—'
          )}
        </div>
      </div>
      <div className='bg-card min-w-0 rounded-lg border px-3 py-2.5'>
        <div className='text-foreground mt-1 grid min-h-8 grid-cols-2 items-center gap-3 text-sm font-semibold tabular-nums'>
          <div className='min-w-0'>
            <div className='text-muted-foreground text-[10px] font-medium'>
              {t('Context')}
            </div>
            <div className='mt-2 truncate'>{contextLength || '—'}</div>
          </div>
          <div className='min-w-0'>
            <div className='text-muted-foreground text-[10px] font-medium'>
              {t('Max output')}
            </div>
            <div className='mt-2 truncate'>{maxOutput || '—'}</div>
          </div>
        </div>
      </div>
    </div>
  )
}

// ----------------------------------------------------------------------------
// Model header (always visible above the detail sections)
// ----------------------------------------------------------------------------

function ModelHeader(props: { model: PricingModel }) {
  const { t } = useTranslation()
  const model = props.model
  const modelIconKey = model.icon || model.vendor_icon
  const modelIcon = modelIconKey ? getLobeIcon(modelIconKey, 20) : null
  const description = model.description || model.vendor_description || null
  const isSpecialExpression =
    getDynamicPricingSummary(model, { tokenUnit: DEFAULT_TOKEN_UNIT })
      ?.isSpecialExpression ?? false
  const isSoraPricing = isSoraPerRequestPricingModel(model)
  const billingLabel =
    model.quota_type === QUOTA_TYPE_VALUES.TOKEN || isSoraPricing
      ? t('Token-based')
      : t('Per Request')
  const showDynamicPricingBadge =
    isSoraPricing ||
    (model.billing_mode === 'tiered_expr' && Boolean(model.billing_expr))

  return (
    <header className='pb-4'>
      <div className='flex flex-wrap items-center gap-2.5'>
        {modelIcon}
        <h1 className='font-mono text-xl font-bold tracking-tight sm:text-2xl'>
          {model.model_name}
        </h1>
        <CopyButton
          value={model.model_name || ''}
          className='size-6'
          iconClassName='size-3'
          tooltip={t('Copy model name')}
          successTooltip={t('Copied!')}
          aria-label={t('Copy model name')}
        />
      </div>
      <div className='mt-1 flex flex-wrap items-center gap-1.5 text-xs'>
        {model.vendor_name && (
          <span className='text-muted-foreground'>{model.vendor_name}</span>
        )}
        <span className='text-muted-foreground/30'>·</span>
        <span className='text-muted-foreground/70'>{billingLabel}</span>
        {showDynamicPricingBadge && (
          <>
            <span className='text-muted-foreground/30'>·</span>
            <span className='rounded bg-amber-100 px-1.5 py-0.5 text-[10px] font-medium text-amber-700 dark:bg-amber-500/20 dark:text-amber-300'>
              {isSpecialExpression
                ? t('Special billing expression')
                : t('Dynamic Pricing')}
            </span>
          </>
        )}
      </div>
      {description && (
        <p className='text-muted-foreground mt-2 text-sm leading-relaxed'>
          {description}
        </p>
      )}
      <ModelHeaderMeta model={model} />
    </header>
  )
}

// ----------------------------------------------------------------------------
// Auto group chain (used inside group pricing section)
// ----------------------------------------------------------------------------

function AutoGroupChain(props: { model: PricingModel; autoGroups: string[] }) {
  const { t } = useTranslation()
  const modelEnableGroups = Array.isArray(props.model.enable_groups)
    ? props.model.enable_groups
    : []
  const autoChain = props.autoGroups.filter((g) =>
    modelEnableGroups.includes(g)
  )

  if (autoChain.length === 0) return null

  return (
    <div className='text-muted-foreground mb-3 flex flex-wrap items-center gap-1 text-xs'>
      <span className='font-medium'>{t('Auto Group Chain')}</span>
      <span className='text-muted-foreground/40'>→</span>
      {autoChain.map((g, idx) => (
        <span key={g} className='flex items-center gap-1'>
          <GroupBadge group={g} size='sm' />
          {idx < autoChain.length - 1 && (
            <span className='text-muted-foreground/40'>→</span>
          )}
        </span>
      ))}
    </div>
  )
}

type DynamicPriceOptions = Parameters<typeof getDynamicPriceEntries>[1]
type DynamicPricingTier = ReturnType<typeof getDynamicPricingTiers>[number]
type DynamicFormattedPricesByTier = Map<DynamicPricingTier, Map<string, string>>

function getSimpleTaskPricingModel(model: PricingModel): PricingModel | null {
  const variants = model.billing_plugin_variants ?? []
  const mergedSchema = Object.assign(
    {},
    ...variants.map((variant) => variant.billing_usage_schema),
    model.billing_usage_schema
  )
  const candidates = [
    // Prefer the model-level expression, but allow its usage schema to come
    // from plugin variants when the API splits expression metadata by source.
    { billingExpr: model.billing_expr, schema: mergedSchema },
    ...variants.map((variant) => ({
      billingExpr: variant.billing_expr,
      schema: Object.assign({}, mergedSchema, variant.billing_usage_schema),
    })),
  ]

  for (const candidate of candidates) {
    const schema = candidate.schema
    const billingExpr = candidate.billingExpr?.trim()
    if (!schema || !billingExpr) continue

    const fields = Object.keys(schema).filter((field) =>
      ['input_tokens', 'output_tokens'].includes(field)
    )
    if (
      fields.length === 0 ||
      !fields.every(
        (field) =>
          schema[field]?.type === 'number' && schema[field]?.unit === 'token'
      )
    ) {
      continue
    }

    const simpleModel = {
      ...model,
      billing_plugin_variants: undefined,
      billing_mode: 'tiered_expr' as const,
      billing_expr: billingExpr,
      billing_usage_schema: schema,
    }
    if (getTaskPriceEntries(simpleModel, { tokenUnit: 'M' }).length > 0) {
      return simpleModel
    }
  }

  return null
}

function getSimpleTaskPricingFields(model: PricingModel): string[] {
  const entries = getTaskPriceEntries(model, {
    tokenUnit: 'M',
  })
  return ['input_tokens', 'output_tokens'].filter((field) =>
    entries.some((entry) => entry.field === field)
  )
}

function getTaskPricingSignature(model: PricingModel): string {
  const schema = model.billing_usage_schema ?? {}
  const fields = Object.keys(schema)
    .sort()
    .map((field) => [field, schema[field]])
  return JSON.stringify([model.billing_mode, model.billing_expr, fields])
}

function getTaskPricingSections(model: PricingModel, taskPricingTitle: string) {
  const rawSections = [
    ...(model.billing_usage_schema
      ? [{ key: 'model', model, title: taskPricingTitle }]
      : []),
    ...(model.billing_plugin_variants ?? []).map((variant) => ({
      key: variant.plugin_key,
      model: {
        ...model,
        billing_plugin_variants: undefined,
        billing_mode: variant.billing_mode,
        billing_expr: variant.billing_expr,
        billing_usage_schema: variant.billing_usage_schema,
        billing_usage_examples: variant.billing_usage_examples,
      },
      title: `${variant.plugin_name} ${taskPricingTitle}`,
    })),
  ]
  const uniqueSections = rawSections.filter(
    (section, index) =>
      rawSections.findIndex(
        (candidate) =>
          getTaskPricingSignature(candidate.model) ===
          getTaskPricingSignature(section.model)
      ) === index
  )
  return uniqueSections.filter(
    (section) =>
      getTaskPriceEntries(section.model, { tokenUnit: 'M' }).length > 0
  )
}

function getDynamicPriceFields(
  tiers: DynamicPricingTier[],
  options: DynamicPriceOptions
) {
  return Array.from(
    new Map(
      tiers
        .flatMap((tier) => getDynamicPriceEntries(tier, options))
        .map((entry) => [entry.field, entry])
    ).values()
  )
}

function getDynamicFormattedPricesByTier(
  tiers: DynamicPricingTier[],
  options: DynamicPriceOptions
): DynamicFormattedPricesByTier {
  return new Map(
    tiers.map((tier) => [
      tier,
      new Map(
        getDynamicPriceEntries(tier, options).map((entry) => [
          entry.field,
          entry.formatted,
        ])
      ),
    ])
  )
}

function GroupPricingHeader(props: {
  group: string
  multiplier: number
  expiry?: number
}) {
  const { t } = useTranslation()

  return (
    <div className='bg-muted/20 flex items-center justify-between gap-3 border-b px-3 py-2'>
      <GroupBadge group={props.group} size='sm' />
      <span className='text-muted-foreground flex items-center gap-3 font-mono text-xs'>
        {props.multiplier}x
        {props.expiry && (
          <span className='font-sans'>
            {t('Expires at')} {new Date(props.expiry * 1000).toLocaleString()}
          </span>
        )}
      </span>
    </div>
  )
}

type PricingConditionOption = ReturnType<
  typeof getDynamicPricingRuleOptions
>[number]

function getSoraResolutionOptions(
  model: PricingModel
): PricingConditionOption[] {
  return (model.sora_per_request_pricing?.resolution_tiers ?? [])
    .map((tier) => String(tier.value || '').trim())
    .filter(
      (value, index, values) =>
        value.length > 0 && values.indexOf(value) === index
    )
    .map((value) => ({ value, label: value, multiplier: 1 }))
}

function getResolutionParamValue(conditionText?: string): string | null {
  if (!conditionText) return null
  const match = conditionText.match(
    /param\(\s*["']resolution["']\s*\)\s*==\s*(["']((?:\\.|[^"'])*)["'])/
  )
  if (!match) return null
  try {
    return JSON.parse(match[1]) as string
  } catch {
    return match[2]
  }
}

function getImageResolutionOptions(
  model: PricingModel,
  tiers: DynamicPricingTier[]
): PricingConditionOption[] {
  if (
    !model.supported_endpoint_types?.includes('image-generation') ||
    tiers.length < 2 ||
    tiers.some(
      (tier) =>
        !Number.isFinite(Number(tier.fixedPrice)) || Number(tier.fixedPrice) < 0
    )
  ) {
    return []
  }

  const values = tiers
    .map((tier) => getResolutionParamValue(tier.conditionText))
    .filter((value): value is string => Boolean(value))
  if (values.length < 2) return []

  return values
    .filter((value, index, all) => all.indexOf(value) === index)
    .map((value) => ({ value, label: value, multiplier: 1 }))
}

function GroupTaskPricing(props: {
  model: PricingModel
  groupRatio: number
  priceRate: number
  usdExchangeRate: number
}) {
  const { t } = useTranslation()
  const sections = getTaskPricingSections(props.model, t('Task pricing'))

  if (sections.length === 0) return null

  return (
    <div className='border-t px-3 py-3'>
      <div className='space-y-5'>
        {sections.map((section) => (
          <TaskPricingSection
            key={section.key}
            model={section.model}
            priceRate={props.priceRate}
            usdExchangeRate={props.usdExchangeRate}
            groupRatio={props.groupRatio}
            title={section.title}
          />
        ))}
      </div>
    </div>
  )
}

function PricingByGroupHeader(props: {
  conditionLabel?: string
  conditionOptions?: PricingConditionOption[]
  selectedCondition?: string
  onConditionChange?: (value: string) => void
}) {
  const { t } = useTranslation()
  const showRuleSelector =
    Boolean(props.conditionOptions && props.conditionOptions.length > 1) &&
    Boolean(props.selectedCondition && props.onConditionChange)

  return (
    <div className='mb-3 flex min-w-0 flex-wrap items-center justify-between gap-2'>
      <h2 className='text-muted-foreground min-w-0 text-xs font-semibold tracking-wider uppercase'>
        {t('Pricing by Group')}
      </h2>
      {showRuleSelector && (
        <div className='flex min-w-0 items-center gap-2'>
          <span className='text-muted-foreground shrink-0 text-xs font-medium'>
            {props.conditionLabel ?? t('Billing period')}
          </span>
          <Select
            items={props.conditionOptions!.map((option) => ({
              value: option.value,
              label: option.label,
            }))}
            value={props.selectedCondition}
            onValueChange={(value) => {
              if (typeof value === 'string') props.onConditionChange?.(value)
            }}
          >
            <SelectTrigger
              size='sm'
              className='max-w-[min(100%,36rem)] min-w-0 sm:w-auto'
              aria-label={props.conditionLabel ?? t('Billing period')}
            >
              <SelectValue />
            </SelectTrigger>
            <SelectContent align='end' alignItemWithTrigger={false}>
              {props.conditionOptions!.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      )}
    </div>
  )
}

function getToolPricingColumns<TData>(props: {
  tools: PricingToolPrice[]
  thClass: string
  translate: (key: string) => string
  getGroupMultiplier: (row: TData) => number
  showRechargePrice: boolean
  priceRate: number
  usdExchangeRate: number
}): StaticDataTableColumn<TData>[] {
  return props.tools.map((tool) => ({
    id: `tool-${tool.name}`,
    header: props.translate(tool.label),
    className: `${props.thClass} text-right`,
    cellClassName: 'py-2.5 text-right font-mono',
    cell: (row: TData) => {
      const ratio = props.getGroupMultiplier(row)
      return (
        <span className={ratio < 1 ? 'text-success' : undefined}>
          {formatToolPrice(tool.price_per_1k, ratio, {
            showRechargePrice: props.showRechargePrice,
            priceRate: props.priceRate,
            usdExchangeRate: props.usdExchangeRate,
          })}
          <span className='text-muted-foreground/60 ml-1 text-xs font-normal'>
            /{props.translate('1K calls')}
          </span>
        </span>
      )
    },
  }))
}

function DynamicGroupPricingCard(props: {
  model: PricingModel
  group: string
  groupRatio: number
  ruleInfo: { multiplier: number; expiry?: number }
  dynamicTiers: DynamicPricingTier[]
  priceFields: ReturnType<typeof getDynamicPriceFields>
  tokenUnit: TokenUnit
  showRechargePrice: boolean
  priceRate: number
  usdExchangeRate: number
  selectedRuleMultiplier: number
}) {
  const { t } = useTranslation()
  const toolPricingColumns = getToolPricingColumns({
    tools: getPricingToolPrices(props.model),
    thClass:
      'text-muted-foreground py-2 text-[10px] font-medium tracking-wider uppercase',
    translate: t,
    getGroupMultiplier: () => props.groupRatio,
    showRechargePrice: props.showRechargePrice,
    priceRate: props.priceRate,
    usdExchangeRate: props.usdExchangeRate,
  })
  const formattedPricesByTier = useMemo(
    () =>
      getDynamicFormattedPricesByTier(props.dynamicTiers, {
        tokenUnit: props.tokenUnit,
        showRechargePrice: props.showRechargePrice,
        priceRate: props.priceRate,
        usdExchangeRate: props.usdExchangeRate,
        groupRatioMultiplier: props.groupRatio * props.selectedRuleMultiplier,
      }),
    [
      props.dynamicTiers,
      props.tokenUnit,
      props.showRechargePrice,
      props.priceRate,
      props.usdExchangeRate,
      props.groupRatio,
      props.selectedRuleMultiplier,
    ]
  )

  return (
    <div className='overflow-hidden rounded-lg border'>
      <GroupPricingHeader
        group={props.group}
        multiplier={props.ruleInfo.multiplier}
        expiry={props.ruleInfo.expiry}
      />
      <StaticDataTable
        className='rounded-none border-0'
        tableClassName='text-sm'
        headerRowClassName='hover:bg-transparent'
        data={props.dynamicTiers}
        getRowKey={(tier, tierIndex) =>
          `${props.group}-${tier.label || tierIndex}`
        }
        columns={[
          {
            id: 'tier',
            header: t('Tier'),
            className:
              'text-muted-foreground py-2 text-[10px] font-medium tracking-wider uppercase',
            cellClassName: 'text-muted-foreground py-2.5',
            cell: (tier) => tier.label || t('Default'),
          },
          ...props.priceFields.map((fieldEntry) => ({
            id: fieldEntry.field,
            header: t(fieldEntry.shortLabel),
            className:
              'text-muted-foreground py-2 text-right text-[10px] font-medium tracking-wider uppercase',
            cellClassName: 'py-2.5 text-right font-mono',
            cell: (tier: (typeof props.dynamicTiers)[number]) => (
              <span
                className={
                  props.groupRatio * props.selectedRuleMultiplier < 1
                    ? 'text-success'
                    : undefined
                }
              >
                {formattedPricesByTier.get(tier)?.get(fieldEntry.field) ?? '-'}
              </span>
            ),
          })),
          ...toolPricingColumns,
        ]}
      />
      {props.model.billing_plugin_variants?.length ? (
        <GroupTaskPricing
          model={{
            ...props.model,
            billing_usage_schema: undefined,
          }}
          groupRatio={props.groupRatio}
          priceRate={props.priceRate}
          usdExchangeRate={props.usdExchangeRate}
        />
      ) : null}
    </div>
  )
}

// ----------------------------------------------------------------------------
// Group pricing table
// ----------------------------------------------------------------------------

function GroupPricingSection(props: {
  model: PricingModel
  groupRatio: Record<string, number>
  usableGroup: PricingUsableGroup
  autoGroups: string[]
  priceRate: number
  usdExchangeRate: number
  tokenUnit: TokenUnit
  showRechargePrice?: boolean
  groupModelRatioExpiry?: Record<string, Record<string, number>>
}) {
  const { t, i18n } = useTranslation()
  const [selectedCondition, setSelectedCondition] = useState('')
  const soraResolutionOptions = useMemo(
    () => getSoraResolutionOptions(props.model),
    [props.model]
  )
  const dynamicRuleOptions = useMemo(
    () =>
      getDynamicPricingRuleOptions(props.model.billing_expr, t, i18n.language),
    [props.model.billing_expr, t, i18n.language]
  )
  const isTokenBased = isTokenBasedModel(props.model)
  const isSoraPricing = isSoraPerRequestPricingModel(props.model)
  const dynamicPricingTiers = useMemo(
    () => getDynamicPricingTiers(props.model),
    [props.model]
  )
  const imageResolutionOptions = useMemo(
    () => getImageResolutionOptions(props.model, dynamicPricingTiers),
    [props.model, dynamicPricingTiers]
  )
  const isImageResolutionPricing =
    !isSoraPricing && imageResolutionOptions.length > 1
  const conditionOptions = isSoraPricing
    ? soraResolutionOptions
    : isImageResolutionPricing
      ? imageResolutionOptions
      : dynamicRuleOptions
  const activeCondition =
    conditionOptions.find((option) => option.value === selectedCondition)
      ?.value ??
    conditionOptions[0]?.value ??
    ''
  const selectedRuleMultiplier =
    dynamicRuleOptions.find((option) => option.value === activeCondition)
      ?.multiplier ?? 1
  const showRechargePrice = props.showRechargePrice ?? false
  const effectiveGroupRatio = props.model.group_ratio ?? props.groupRatio
  const simpleTaskPricingModel = useMemo(
    () => getSimpleTaskPricingModel(props.model),
    [props.model]
  )
  const getRuleInfo = (group: string) => {
    const base = props.groupRatio[group] ?? 1
    const effective = effectiveGroupRatio[group] ?? 1
    const modelMultiplier = props.model.model_group_ratio?.[group]
    const expiry =
      props.groupModelRatioExpiry?.[group]?.[props.model.model_name]
    const active = !expiry || expiry <= 0 || expiry > Date.now() / 1000
    return {
      multiplier: effective,
      baseMultiplier: base,
      modelMultiplier,
      expiry: active && expiry && expiry > 0 ? expiry : undefined,
    }
  }

  const availableGroups = useMemo(
    () => getAvailableGroups(props.model, props.usableGroup || {}),
    [props.model, props.usableGroup]
  )

  const tokenUnitLabel = props.tokenUnit === 'K' ? '1K' : '1M'

  const extraPriceTypes = useMemo(() => {
    const types: { label: string; type: PriceType }[] = []
    if (props.model.cache_ratio != null)
      types.push({ label: t('Cache'), type: 'cache' })
    if (props.model.create_cache_ratio != null)
      types.push({ label: t('Cache Write'), type: 'create_cache' })
    if (props.model.image_ratio != null)
      types.push({ label: t('Image'), type: 'image' })
    if (props.model.audio_ratio != null)
      types.push({ label: t('Audio In'), type: 'audio_input' })
    if (
      props.model.audio_ratio != null &&
      props.model.audio_completion_ratio != null
    )
      types.push({ label: t('Audio Out'), type: 'audio_output' })
    return types
  }, [props.model, t])

  if (availableGroups.length === 0) {
    return (
      <section>
        <PricingByGroupHeader />
        <AutoGroupChain model={props.model} autoGroups={props.autoGroups} />
        <p className='text-muted-foreground text-sm'>
          {t(
            'This model is not available in any group, or no group pricing information is configured.'
          )}
        </p>
      </section>
    )
  }

  const thClass =
    'text-muted-foreground py-2 text-[10px] font-medium tracking-wider uppercase'
  const toolPricingColumns = getToolPricingColumns({
    tools: getPricingToolPrices(props.model),
    thClass,
    translate: t,
    getGroupMultiplier: (group) => getRuleInfo(group).multiplier,
    showRechargePrice,
    priceRate: props.priceRate,
    usdExchangeRate: props.usdExchangeRate,
  })

  if (isSoraPricing) {
    const hasAudioGenerationSurcharge = availableGroups.some((group) =>
      Boolean(
        getSoraPricingDisplay(props.model, {
          showRechargePrice,
          priceRate: props.priceRate,
          usdExchangeRate: props.usdExchangeRate,
          selectedGroup: group,
        })?.audioGenerationSurcharge
      )
    )

    return (
      <section>
        <PricingByGroupHeader
          conditionLabel={t('Resolution')}
          conditionOptions={conditionOptions}
          selectedCondition={activeCondition}
          onConditionChange={setSelectedCondition}
        />
        <AutoGroupChain model={props.model} autoGroups={props.autoGroups} />
        <StaticDataTable
          className='-mx-4 rounded-none border-0 sm:mx-0'
          tableClassName='text-sm'
          headerRowClassName='hover:bg-transparent'
          data={availableGroups}
          getRowKey={(group) => group}
          columns={[
            {
              id: 'group',
              header: t('Group'),
              className: thClass,
              cellClassName: 'py-2.5',
              cell: (group) => <GroupBadge group={group} size='sm' />,
            },
            {
              id: 'ratio',
              header: t('Ratio'),
              className: thClass,
              cellClassName: 'text-muted-foreground py-2.5 font-mono',
              cell: (group) => `${getRuleInfo(group).multiplier}x`,
            },
            {
              id: 'expiresAt',
              header: t('Discount expiration time'),
              className: thClass,
              cellClassName: 'text-muted-foreground py-2.5 text-xs',
              cell: (group) => {
                const expiry = getRuleInfo(group).expiry
                return expiry ? new Date(expiry * 1000).toLocaleString() : '-'
              },
            },
            {
              id: 'resolution',
              header: t('Video output'),
              className: `${thClass} text-right`,
              cellClassName: 'py-2.5 text-right font-mono',
              cell: (group) => {
                const summary = getSoraPricingDisplay(props.model, {
                  showRechargePrice,
                  priceRate: props.priceRate,
                  usdExchangeRate: props.usdExchangeRate,
                  selectedGroup: group,
                })
                const selectedTier = summary?.resolutionTiers.find(
                  (tier) => tier.key === activeCondition
                )
                return (
                  <span
                    className={
                      getRuleInfo(group).multiplier < 1
                        ? 'text-success'
                        : undefined
                    }
                  >
                    {selectedTier
                      ? `${stripTrailingZeros(selectedTier.price)}/s`
                      : '-'}
                  </span>
                )
              },
            },
            ...(hasAudioGenerationSurcharge
              ? [
                  {
                    id: 'audioGenerationSurcharge',
                    header: t('Audio generation surcharge'),
                    className: `${thClass} text-right`,
                    cellClassName: 'py-2.5 text-right font-mono',
                    cell: (group: string) => {
                      const summary = getSoraPricingDisplay(props.model, {
                        showRechargePrice,
                        priceRate: props.priceRate,
                        usdExchangeRate: props.usdExchangeRate,
                        selectedGroup: group,
                      })
                      return (
                        <span
                          className={
                            getRuleInfo(group).multiplier < 1
                              ? 'text-success'
                              : undefined
                          }
                        >
                          {summary?.audioGenerationSurcharge
                            ? `${stripTrailingZeros(summary.audioGenerationSurcharge)}/${t('request')}`
                            : '-'}
                        </span>
                      )
                    },
                  },
                ]
              : []),
            ...toolPricingColumns,
          ]}
        />
      </section>
    )
  }

  if (isImageResolutionPricing) {
    const selectedTier =
      dynamicPricingTiers.find(
        (tier) =>
          getResolutionParamValue(tier.conditionText) === activeCondition
      ) ?? dynamicPricingTiers[dynamicPricingTiers.length - 1]

    return (
      <section>
        <PricingByGroupHeader
          conditionLabel={t('Resolution')}
          conditionOptions={conditionOptions}
          selectedCondition={activeCondition}
          onConditionChange={setSelectedCondition}
        />
        <AutoGroupChain model={props.model} autoGroups={props.autoGroups} />
        <StaticDataTable
          className='-mx-4 rounded-none border-0 sm:mx-0'
          tableClassName='text-sm'
          headerRowClassName='hover:bg-transparent'
          data={availableGroups}
          getRowKey={(group) => group}
          columns={[
            {
              id: 'group',
              header: t('Group'),
              className: thClass,
              cellClassName: 'py-2.5',
              cell: (group) => <GroupBadge group={group} size='sm' />,
            },
            {
              id: 'ratio',
              header: t('Ratio'),
              className: thClass,
              cellClassName: 'text-muted-foreground py-2.5 font-mono',
              cell: (group) => `${getRuleInfo(group).multiplier}x`,
            },
            {
              id: 'expiresAt',
              header: t('Discount expiration time'),
              className: thClass,
              cellClassName: 'text-muted-foreground py-2.5 text-xs',
              cell: (group) => {
                const expiry = getRuleInfo(group).expiry
                return expiry ? new Date(expiry * 1000).toLocaleString() : '-'
              },
            },
            {
              id: 'imageOutput',
              header: t('Image output'),
              className: `${thClass} text-right`,
              cellClassName: 'py-2.5 text-right font-mono',
              cell: (group: string) => {
                const entries = selectedTier
                  ? getDynamicPriceEntries(selectedTier, {
                      tokenUnit: props.tokenUnit,
                      showRechargePrice,
                      priceRate: props.priceRate,
                      usdExchangeRate: props.usdExchangeRate,
                      groupRatioMultiplier: effectiveGroupRatio[group] ?? 1,
                    })
                  : []
                const entry = entries.find(
                  (item) => item.field === 'fixedPrice'
                )
                return (
                  <span
                    className={
                      getRuleInfo(group).multiplier < 1
                        ? 'text-success'
                        : undefined
                    }
                  >
                    {entry ? `${entry.formatted}/${t('Image unit')}` : '-'}
                  </span>
                )
              },
            },
            ...toolPricingColumns,
          ]}
        />
      </section>
    )
  }

  if (isDynamicPricingModel(props.model) && !simpleTaskPricingModel) {
    const dynamicTiers = dynamicPricingTiers

    if (dynamicTiers.length === 0) {
      if (
        props.model.billing_usage_schema ||
        (props.model.billing_plugin_variants?.length ?? 0) > 0
      ) {
        return (
          <section>
            <PricingByGroupHeader />
            <AutoGroupChain model={props.model} autoGroups={props.autoGroups} />
            <div className='space-y-3'>
              {availableGroups.map((group) => {
                const ruleInfo = getRuleInfo(group)
                return (
                  <div
                    key={group}
                    className='overflow-hidden rounded-lg border'
                  >
                    <GroupPricingHeader
                      group={group}
                      multiplier={ruleInfo.multiplier}
                      expiry={ruleInfo.expiry}
                    />
                    <GroupTaskPricing
                      model={props.model}
                      groupRatio={effectiveGroupRatio[group] ?? 1}
                      priceRate={props.priceRate}
                      usdExchangeRate={props.usdExchangeRate}
                    />
                  </div>
                )
              })}
            </div>
          </section>
        )
      }

      return (
        <section>
          <PricingByGroupHeader />
          <AutoGroupChain model={props.model} autoGroups={props.autoGroups} />
          <div className='space-y-3'>
            {availableGroups.map((group) => {
              const ruleInfo = getRuleInfo(group)
              return (
                <div key={group} className='overflow-hidden rounded-lg border'>
                  <GroupPricingHeader
                    group={group}
                    multiplier={ruleInfo.multiplier}
                    expiry={ruleInfo.expiry}
                  />
                  <div className='border-amber-200/70 bg-amber-50/70 p-3 dark:border-amber-500/20 dark:bg-amber-500/10'>
                    <div className='text-sm font-medium text-amber-800 dark:text-amber-200'>
                      {t('Special billing expression')}
                    </div>
                    <p className='text-muted-foreground mt-1 text-xs'>
                      {t(
                        'Group prices cannot be expanded because this expression is not a standard tiered pricing expression.'
                      )}
                    </p>
                    <div className='mt-3'>
                      <div className='text-muted-foreground mb-1 text-[10px] font-medium tracking-wider uppercase'>
                        {t('Raw expression')}
                      </div>
                      <code className='text-muted-foreground bg-background/80 block max-h-28 overflow-auto rounded-md border px-2 py-1.5 font-mono text-xs break-all'>
                        {props.model.billing_expr}
                      </code>
                    </div>
                  </div>
                </div>
              )
            })}
          </div>
        </section>
      )
    }

    const priceFields = getDynamicPriceFields(dynamicTiers, {
      tokenUnit: props.tokenUnit,
      showRechargePrice,
      priceRate: props.priceRate,
      usdExchangeRate: props.usdExchangeRate,
      groupRatioMultiplier: 1,
    })
    return (
      <section>
        <PricingByGroupHeader
          conditionOptions={dynamicRuleOptions}
          selectedCondition={activeCondition}
          onConditionChange={setSelectedCondition}
        />
        <AutoGroupChain model={props.model} autoGroups={props.autoGroups} />
        <div className='space-y-3'>
          {availableGroups.map((group) => {
            const ruleInfo = getRuleInfo(group)
            return (
              <DynamicGroupPricingCard
                key={group}
                model={props.model}
                group={group}
                groupRatio={effectiveGroupRatio[group] ?? 1}
                ruleInfo={ruleInfo}
                dynamicTiers={dynamicTiers}
                priceFields={priceFields}
                tokenUnit={props.tokenUnit}
                showRechargePrice={showRechargePrice}
                priceRate={props.priceRate}
                usdExchangeRate={props.usdExchangeRate}
                selectedRuleMultiplier={selectedRuleMultiplier}
              />
            )
          })}
          <p className='text-muted-foreground/40 mt-1.5 text-[10px]'>
            {t('Prices shown per')} {tokenUnitLabel} tokens
          </p>
        </div>
      </section>
    )
  }

  const renderGroupPrice = (group: string, type: PriceType) =>
    formatGroupPrice(
      props.model,
      group,
      type,
      props.tokenUnit,
      showRechargePrice,
      props.priceRate,
      props.usdExchangeRate,
      effectiveGroupRatio
    )
  const renderFixedGroupPrice = (group: string) =>
    formatFixedPrice(
      props.model,
      group,
      showRechargePrice,
      props.priceRate,
      props.usdExchangeRate,
      effectiveGroupRatio
    )

  const renderSimpleTaskPrice = (group: string, field: string) => {
    if (!simpleTaskPricingModel) return '-'
    const entries = getTaskPriceEntries(simpleTaskPricingModel, {
      tokenUnit: props.tokenUnit,
      showRechargePrice,
      priceRate: props.priceRate,
      usdExchangeRate: props.usdExchangeRate,
      groupRatioMultiplier: effectiveGroupRatio[group] ?? 1,
    })
    const entry = entries.find((item) => item.field === field)
    return entry?.formatted ?? '-'
  }
  const simpleTaskPricingFields = simpleTaskPricingModel
    ? getSimpleTaskPricingFields(simpleTaskPricingModel)
    : []

  return (
    <section>
      <PricingByGroupHeader />
      <AutoGroupChain model={props.model} autoGroups={props.autoGroups} />
      <StaticDataTable
        className='-mx-4 rounded-none border-0 sm:mx-0'
        tableClassName='text-sm'
        headerRowClassName='hover:bg-transparent'
        data={availableGroups}
        getRowKey={(group) => group}
        columns={[
          {
            id: 'group',
            header: t('Group'),
            className: thClass,
            cellClassName: 'py-2.5',
            cell: (group) => <GroupBadge group={group} size='sm' />,
          },
          {
            id: 'ratio',
            header: t('Ratio'),
            className: thClass,
            cellClassName: 'text-muted-foreground py-2.5 font-mono',
            cell: (group) => {
              const info = getRuleInfo(group)
              return `${info.multiplier}x`
            },
          },
          {
            id: 'expiresAt',
            header: t('Discount expiration time'),
            className: thClass,
            cellClassName: 'text-muted-foreground py-2.5 text-xs',
            cell: (group) => {
              const expiry = getRuleInfo(group).expiry
              return expiry ? new Date(expiry * 1000).toLocaleString() : '-'
            },
          },
          ...(simpleTaskPricingModel
            ? [
                ...simpleTaskPricingFields.map((field) => ({
                  id: field,
                  header: t(field === 'input_tokens' ? 'Input' : 'Output'),
                  className: `${thClass} text-right`,
                  cellClassName: 'py-2.5 text-right font-mono',
                  cell: (group: string) => (
                    <span
                      className={
                        getRuleInfo(group).multiplier < 1
                          ? 'text-success'
                          : undefined
                      }
                    >
                      {renderSimpleTaskPrice(group, field)}
                    </span>
                  ),
                })),
                ...toolPricingColumns,
              ]
            : isTokenBased
              ? [
                  {
                    id: 'input',
                    header: t('Input'),
                    className: `${thClass} text-right`,
                    cellClassName: 'py-2.5 text-right font-mono',
                    cell: (group: string) => (
                      <span
                        className={
                          getRuleInfo(group).multiplier < 1
                            ? 'text-success'
                            : undefined
                        }
                      >
                        {renderGroupPrice(group, 'input')}
                      </span>
                    ),
                  },
                  {
                    id: 'output',
                    header: t('Output'),
                    className: `${thClass} text-right`,
                    cellClassName: 'py-2.5 text-right font-mono',
                    cell: (group: string) => (
                      <span
                        className={
                          getRuleInfo(group).multiplier < 1
                            ? 'text-success'
                            : undefined
                        }
                      >
                        {renderGroupPrice(group, 'output')}
                      </span>
                    ),
                  },
                  ...extraPriceTypes.map((ep) => ({
                    id: ep.type,
                    header: ep.label,
                    className: `${thClass} text-right`,
                    cellClassName: 'py-2.5 text-right font-mono',
                    cell: (group: string) => (
                      <span
                        className={
                          getRuleInfo(group).multiplier < 1
                            ? 'text-success'
                            : undefined
                        }
                      >
                        {renderGroupPrice(group, ep.type)}
                      </span>
                    ),
                  })),
                  ...toolPricingColumns,
                ]
              : [
                  {
                    id: 'price',
                    header: isSoraPricing ? t('Price') : t('Price'),
                    className: `${thClass} text-right`,
                    cellClassName: 'py-2.5 text-right font-mono',
                    cell: (group: string) => (
                      <span
                        className={
                          getRuleInfo(group).multiplier < 1
                            ? 'text-success'
                            : undefined
                        }
                      >
                        {renderFixedGroupPrice(group)}
                        {isSoraPricing && (
                          <span className='text-muted-foreground/40 ml-1 text-xs font-normal'>
                            /s
                          </span>
                        )}
                      </span>
                    ),
                  },
                  ...toolPricingColumns,
                ]),
        ]}
      />
      {!simpleTaskPricingModel &&
        getTaskPricingSections(props.model, t('Task pricing')).length > 0 && (
          <div className='mt-3 space-y-3'>
            {availableGroups.map((group) => {
              const ruleInfo = getRuleInfo(group)
              return (
                <div
                  key={`tasks-${group}`}
                  className='overflow-hidden rounded-lg border'
                >
                  <GroupPricingHeader
                    group={group}
                    multiplier={ruleInfo.multiplier}
                    expiry={ruleInfo.expiry}
                  />
                  <GroupTaskPricing
                    model={props.model}
                    groupRatio={effectiveGroupRatio[group] ?? 1}
                    priceRate={props.priceRate}
                    usdExchangeRate={props.usdExchangeRate}
                  />
                </div>
              )
            })}
          </div>
        )}
      <div className='-mx-4 sm:mx-0'>
        {(isTokenBased || simpleTaskPricingModel) && (
          <p className='text-muted-foreground/40 mt-1.5 px-4 text-[10px] sm:px-0'>
            {t('Prices shown per')} {tokenUnitLabel} tokens
          </p>
        )}
      </div>
    </section>
  )
}

const TAB_VALUES = ['overview', 'performance', 'api'] as const
type TabValue = (typeof TAB_VALUES)[number]

const TAB_META: Record<
  TabValue,
  { icon: React.ComponentType<{ className?: string }>; labelKey: string }
> = {
  overview: { icon: Info, labelKey: 'Overview' },
  performance: { icon: HeartPulse, labelKey: 'Performance' },
  api: { icon: Code2, labelKey: 'API' },
}

export interface ModelDetailsContentProps {
  model: PricingModel
  groupRatio: Record<string, number>
  usableGroup: PricingUsableGroup
  endpointMap: Record<string, { path?: string; method?: string }>
  autoGroups: string[]
  priceRate: number
  usdExchangeRate: number
  tokenUnit: TokenUnit
  showRechargePrice?: boolean
  groupModelRatioExpiry?: Record<string, Record<string, number>>
}

export function ModelDetailsContent(props: ModelDetailsContentProps) {
  const { t } = useTranslation()
  const showRechargePrice = props.showRechargePrice ?? false

  return (
    <div className='@container/details space-y-4'>
      <ModelHeader model={props.model} />

      <Tabs defaultValue='overview' className='gap-4'>
        <TabsList className='bg-muted/60 grid w-full grid-cols-3 gap-1 rounded-lg p-1 group-data-horizontal/tabs:h-auto'>
          {TAB_VALUES.map((value) => {
            const Icon = TAB_META[value].icon
            return (
              <TabsTrigger
                key={value}
                value={value}
                className='h-8 min-w-0 gap-1.5 rounded-md px-3 text-xs sm:text-sm'
              >
                <Icon className='size-3.5' />
                <span className='truncate'>{t(TAB_META[value].labelKey)}</span>
              </TabsTrigger>
            )
          })}
        </TabsList>

        <TabsContent value='overview' className='space-y-6 outline-none'>
          <section className='bg-card/60 space-y-5 rounded-xl border p-4 shadow-sm'>
            <GroupPricingSection
              model={props.model}
              groupRatio={props.groupRatio}
              usableGroup={props.usableGroup}
              autoGroups={props.autoGroups}
              priceRate={props.priceRate}
              usdExchangeRate={props.usdExchangeRate}
              tokenUnit={props.tokenUnit}
              showRechargePrice={showRechargePrice}
              groupModelRatioExpiry={props.groupModelRatioExpiry}
            />
          </section>
        </TabsContent>

        <TabsContent value='performance' className='outline-none'>
          <ModelDetailsPerformance model={props.model} />
        </TabsContent>

        <TabsContent value='api' className='outline-none'>
          <ModelDetailsApi
            model={props.model}
            endpointMap={props.endpointMap}
          />
        </TabsContent>
      </Tabs>
    </div>
  )
}

// ----------------------------------------------------------------------------
// Drawer & page wrappers
// ----------------------------------------------------------------------------

export interface ModelDetailsDrawerProps extends ModelDetailsContentProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function ModelDetailsDrawer(props: ModelDetailsDrawerProps) {
  const { t } = useTranslation()
  const { open, onOpenChange, ...contentProps } = props

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent
        side='right'
        className={sideDrawerContentClassName(
          'sm:max-w-2xl lg:max-w-3xl xl:max-w-4xl 2xl:max-w-5xl'
        )}
      >
        <SheetHeader className='sr-only'>
          <SheetTitle>{props.model.model_name}</SheetTitle>
          <SheetDescription>{t('Model details')}</SheetDescription>
        </SheetHeader>
        <div className='flex-1 overflow-y-auto px-4 pt-11 pb-5 sm:px-6 sm:pt-12 sm:pb-6'>
          <ModelDetailsContent {...contentProps} />
        </div>
      </SheetContent>
    </Sheet>
  )
}

export function ModelDetails() {
  const { t } = useTranslation()
  const { modelId } = useParams({ from: '/pricing/$modelId/' })
  const search = useSearch({ from: '/pricing/$modelId/' })
  const navigate = useNavigate()

  const {
    models,
    groupRatio,
    usableGroup,
    endpointMap,
    autoGroups,
    isLoading,
    priceRate,
    usdExchangeRate,
  } = usePricingData()

  const tokenUnit: TokenUnit =
    search.tokenUnit === 'K' ? 'K' : DEFAULT_TOKEN_UNIT

  const model = useMemo(() => {
    if (!models || !modelId) return null
    return models.find((m) => m.model_name === modelId) || null
  }, [models, modelId])

  const handleBack = () => {
    navigate({ to: '/pricing', search })
  }

  if (isLoading) {
    return (
      <PublicLayout>
        <div className='mx-auto max-w-5xl px-4 sm:px-6'>
          <Skeleton className='mb-4 h-5 w-16' />
          <div className='space-y-2'>
            <Skeleton className='h-7 w-64' />
            <Skeleton className='h-4 w-40' />
            <Skeleton className='h-4 w-full max-w-md' />
          </div>
          <div className='mt-6 grid grid-cols-2 gap-2 sm:grid-cols-4'>
            {Array.from({ length: 4 }).map((_, i) => (
              <Skeleton key={i} className='h-16 w-full' />
            ))}
          </div>
          <div className='mt-6 space-y-3'>
            {Array.from({ length: 4 }).map((_, i) => (
              <Skeleton key={i} className='h-24 w-full' />
            ))}
          </div>
        </div>
      </PublicLayout>
    )
  }

  if (!model) {
    return (
      <PublicLayout>
        <div className='mx-auto max-w-2xl px-4 text-center sm:px-6'>
          <h2 className='mb-1 text-base font-semibold'>
            {t('Model not found')}
          </h2>
          <p className='text-muted-foreground mb-4 text-sm'>
            {t("The model you're looking for doesn't exist.")}
          </p>
          <Button onClick={handleBack} variant='outline' size='sm'>
            {t('Back to Models')}
          </Button>
        </div>
      </PublicLayout>
    )
  }

  return (
    <PublicLayout>
      <div className='mx-auto max-w-5xl px-4 sm:px-6'>
        <Button
          variant='ghost'
          size='sm'
          onClick={handleBack}
          className='text-muted-foreground hover:text-foreground mb-4 h-auto gap-1 px-0 py-1 text-xs'
        >
          <ArrowLeft className='size-3.5' />
          {t('Back')}
        </Button>

        <ModelDetailsContent
          model={model}
          groupRatio={groupRatio || {}}
          usableGroup={usableGroup || {}}
          autoGroups={autoGroups || []}
          priceRate={priceRate ?? 1}
          usdExchangeRate={usdExchangeRate ?? 1}
          tokenUnit={tokenUnit}
          showRechargePrice={search.rechargePrice ?? false}
          endpointMap={
            (endpointMap as Record<
              string,
              { path?: string; method?: string }
            >) || {}
          }
        />
      </div>
    </PublicLayout>
  )
}
