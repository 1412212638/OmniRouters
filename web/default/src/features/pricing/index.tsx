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
import {
  type ReactNode,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react'
import { useQuery } from '@tanstack/react-query'
import { ArrowRight01Icon, SearchIcon } from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { ChevronDown } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import {
  ZenMuxCopyIcon,
  ZenMuxModelsIcon,
  ZenMuxRefreshCircleIcon,
  ZenMuxSearchIcon,
  ZenMuxSortIcon,
} from '@/assets/custom/zenmux-icons'
import {
  ZENMUX_MODALITY_ICONS,
  type ZenMuxIconComponent,
} from '@/assets/custom/zenmux-modality-icons'
import { getLobeIcon } from '@/lib/lobe-icon'
import { cn } from '@/lib/utils'
import { useCopyToClipboard } from '@/hooks/use-copy-to-clipboard'
import { useStatus } from '@/hooks/use-status'
import { Button } from '@/components/ui/button'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import { Skeleton } from '@/components/ui/skeleton'
import { PublicLayout } from '@/components/layout'
import { PageTransition } from '@/components/page-transition'
import { getPerfMetricsSummary } from '@/features/performance-metrics/api'
import type { PerfModelSummary } from '@/features/performance-metrics/types'
import { ClassicPricing } from './classic-pricing'
import { ModelDetailsDrawer } from './components'
import {
  DEFAULT_PRICING_PAGE_SIZE,
  DEFAULT_TOKEN_UNIT,
  EXCLUDED_GROUPS,
  FILTER_ALL,
} from './constants'
import { usePricingData } from './hooks/use-pricing-data'
import {
  getDynamicDisplayGroupRatio,
  getDynamicPricingSummary,
} from './lib/dynamic-price'
import { isTokenBasedModel } from './lib/model-helpers'
import {
  formatPrice,
  formatRequestPrice,
  getSoraPricingDisplay,
  stripTrailingZeros,
} from './lib/price'
import type {
  Modality,
  PricingModel,
  PricingUsableGroup,
  PricingVendor,
  TokenUnit,
} from './types'

type FilterOption = {
  value: string
  label: string
  count: number
  icon?: ReactNode
  translateLabel?: boolean
}

type SidebarSectionId = 'input' | 'output' | 'vendor' | 'group'

type ModalityOption = {
  value: Modality
  label: string
  icon: ZenMuxIconComponent
}

const MODALITY_OPTIONS: ModalityOption[] = [
  { value: 'text', label: 'Text', icon: ZENMUX_MODALITY_ICONS.text },
  { value: 'image', label: 'Image', icon: ZENMUX_MODALITY_ICONS.image },
  { value: 'file', label: 'File', icon: ZENMUX_MODALITY_ICONS.file },
  { value: 'audio', label: 'Audio', icon: ZENMUX_MODALITY_ICONS.audio },
  { value: 'video', label: 'Video', icon: ZENMUX_MODALITY_ICONS.video },
  {
    value: 'embedding',
    label: 'Embedding',
    icon: ZENMUX_MODALITY_ICONS.embedding,
  },
  { value: 'rerank', label: 'Rerank', icon: ZENMUX_MODALITY_ICONS.rerank },
  { value: 'speech', label: 'Speech', icon: ZENMUX_MODALITY_ICONS.speech },
  {
    value: 'transcription',
    label: 'Transcription',
    icon: ZENMUX_MODALITY_ICONS.transcription,
  },
]

const MODALITY_VALUES = MODALITY_OPTIONS.map((option) => option.value)
const TOP_MODALITY_OPTIONS = MODALITY_OPTIONS.filter(
  (option) => option.value !== 'file'
)
const LOAD_BATCH_SIZE = DEFAULT_PRICING_PAGE_SIZE
const FILTER_PREVIEW_COUNT = 8
const TOKEN_UNIT: TokenUnit = DEFAULT_TOKEN_UNIT
const VENDOR_ICON_ALIASES: Record<string, string> = {
  anthropic: 'Anthropic',
  bytedance: 'ByteDance',
  claude: 'Claude',
  deepseek: 'DeepSeek',
  google: 'Google',
  googlecloud: 'GoogleCloud',
  openai: 'OpenAI',
  openrouter: 'OpenRouter',
}

function normalizeModalities(values: Modality[] | undefined): Modality[] {
  if (!Array.isArray(values)) return []
  return MODALITY_VALUES.filter((value) => values.includes(value))
}

function hasModality(
  model: PricingModel,
  field: 'input_modalities' | 'output_modalities',
  value: string
): boolean {
  if (value === FILTER_ALL) return true
  return normalizeModalities(model[field]).includes(value as Modality)
}

function formatCreatedTime(timestamp: number | undefined): string {
  if (!timestamp || !Number.isFinite(timestamp)) return '-'
  const ms = timestamp < 1_000_000_000_000 ? timestamp * 1000 : timestamp
  const date = new Date(ms)
  if (Number.isNaN(date.getTime())) return '-'

  const pad = (value: number) => String(value).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(
    date.getDate()
  )} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function normalizeSuccessRate(rate: number | undefined): number | null {
  if (rate === undefined || rate === null || !Number.isFinite(rate)) return null
  const normalized = rate <= 1 ? rate * 100 : rate
  return Math.max(0, Math.min(100, normalized))
}

function formatSuccessRate(rate: number | undefined): string {
  const normalized = normalizeSuccessRate(rate)
  return normalized === null ? '-' : `${normalized.toFixed(2)}%`
}

function getSearchText(model: PricingModel): string {
  return [
    model.model_name,
    model.description,
    model.vendor_name,
    model.tags,
    ...(model.supported_endpoint_types || []),
    ...normalizeModalities(model.input_modalities),
    ...normalizeModalities(model.output_modalities),
  ]
    .filter(Boolean)
    .join(' ')
    .toLowerCase()
}

function sortByCreatedTime(a: PricingModel, b: PricingModel): number {
  const createdDiff = (b.created_time || 0) - (a.created_time || 0)
  if (createdDiff !== 0) return createdDiff
  return (a.model_name || '').localeCompare(b.model_name || '')
}

function getCatalogModelTitle(model: PricingModel): string {
  const modelName = model.model_name || ''
  const vendorName = model.vendor_name?.trim()
  if (!vendorName) return modelName

  const normalizedVendor = vendorName.toLowerCase()
  const normalizedModelName = modelName.toLowerCase()
  if (normalizedModelName.startsWith(`${normalizedVendor}/`)) {
    return modelName
  }

  return `${normalizedVendor}/${modelName}`
}

function buildModalityOptions(
  models: PricingModel[],
  field: 'input_modalities' | 'output_modalities'
): FilterOption[] {
  return MODALITY_OPTIONS.map((option) => {
    const count = models.filter((model) =>
      normalizeModalities(model[field]).includes(option.value)
    ).length
    const Icon = option.icon

    return {
      value: option.value,
      label: option.label,
      count,
      translateLabel: true,
      icon: (
        <span className='inline-flex size-5 shrink-0 items-center justify-center'>
          <Icon className='size-3.5' />
        </span>
      ),
    }
  }).filter((option) => option.count > 0)
}

function buildVendorOptions(
  models: PricingModel[],
  vendors: PricingVendor[]
): FilterOption[] {
  const counts = new Map<number, number>()
  for (const model of models) {
    if (!model.vendor_id) continue
    counts.set(model.vendor_id, (counts.get(model.vendor_id) || 0) + 1)
  }

  return vendors
    .filter((vendor) => counts.has(vendor.id))
    .map((vendor) => ({
      value: String(vendor.id),
      label: vendor.name,
      count: counts.get(vendor.id) || 0,
      icon: renderVendorFilterIcon(vendor.name, vendor.icon),
    }))
    .sort((a, b) => b.count - a.count || a.label.localeCompare(b.label))
}

function renderVendorFilterIcon(name: string, icon?: string): ReactNode {
  const iconName = normalizeVendorIconName(icon || name)

  return (
    <span className='bg-background inline-flex size-5 shrink-0 items-center justify-center rounded-full dark:bg-white/5'>
      {iconName ? (
        getLobeIcon(iconName, 16)
      ) : (
        <span className='text-muted-foreground text-[10px] font-semibold'>
          {name.charAt(0).toUpperCase()}
        </span>
      )}
    </span>
  )
}

function normalizeVendorIconName(icon: string | undefined): string | undefined {
  const trimmed = icon?.trim()
  if (!trimmed) return undefined
  if (
    /^(https?:)?\/\//.test(trimmed) ||
    trimmed.startsWith('/') ||
    trimmed.startsWith('data:image/') ||
    trimmed.startsWith('blob:')
  ) {
    return trimmed
  }

  const [base, ...rest] = trimmed.split('.')
  const aliasKey = base.toLowerCase().replace(/[\s_-]+/g, '')
  const normalizedBase = VENDOR_ICON_ALIASES[aliasKey] || base
  return [normalizedBase, ...rest].join('.')
}

function buildGroupOptions(
  models: PricingModel[],
  usableGroup: PricingUsableGroup
): FilterOption[] {
  return Object.entries(usableGroup)
    .filter(([value]) => !EXCLUDED_GROUPS.includes(value))
    .map(([value]) => ({
      value,
      label: value,
      count: models.filter((model) => isModelAvailableInGroup(model, value))
        .length,
    }))
    .filter((option) => option.count > 0)
    .sort((a, b) => b.count - a.count || a.label.localeCompare(b.label))
}

function isModelAvailableInGroup(model: PricingModel, group: string): boolean {
  const groups = model.enable_groups || []
  return groups.includes(group) || groups.includes('all')
}

function FilterButton(props: {
  active: boolean
  label: string
  count?: number
  icon?: ReactNode
  onClick: () => void
  className?: string
}) {
  return (
    <button
      type='button'
      onClick={props.onClick}
      className={cn(
        'flex h-9 w-full items-center justify-between gap-3 rounded-md border px-3.5 text-left text-sm font-normal shadow-none transition-colors',
        props.active
          ? 'border-border bg-white text-neutral-950 shadow-[0_2px_4px_rgba(0,0,0,0.05)] hover:bg-white dark:border-white/15 dark:bg-neutral-800 dark:text-neutral-50 dark:shadow-none dark:hover:bg-neutral-800'
          : 'border-border bg-neutral-100 text-neutral-500 hover:bg-neutral-200 hover:text-neutral-700 dark:border-white/10 dark:bg-neutral-900 dark:text-neutral-400 dark:hover:bg-neutral-800 dark:hover:text-neutral-100',
        props.className
      )}
    >
      <span className='flex min-w-0 items-center gap-2'>
        {props.icon}
        <span className='min-w-0 truncate'>{props.label}</span>
      </span>
      {props.count !== undefined && (
        <span
          className={cn(
            'shrink-0 font-mono text-[13px] tabular-nums',
            props.active
              ? 'text-neutral-500 dark:text-neutral-300'
              : 'text-neutral-400 dark:text-neutral-500'
          )}
        >
          {props.count}
        </span>
      )}
    </button>
  )
}

function FilterSection(props: {
  title: string
  options: FilterOption[]
  value: string
  onChange: (value: string) => void
  allLabel: string
  allCount: number
  open: boolean
  onOpenChange: (open: boolean) => void
  previewCount?: number
}) {
  const [visibleOptionCount, setVisibleOptionCount] = useState<number | null>(
    null
  )
  const { t } = useTranslation()
  const hasSelection = props.value !== FILTER_ALL
  const resetLabel = `${t('Reset')} ${props.title}`
  const toggleLabel = `${props.open ? t('Collapse') : t('Expand')} ${props.title}`
  const previewCount = props.previewCount ?? props.options.length
  const visibleCount = visibleOptionCount ?? previewCount
  const visibleOptions = props.options.slice(0, visibleCount)
  const selectedOption = hasSelection
    ? props.options.find((option) => option.value === props.value)
    : undefined
  const selectedOptionIsPreviewed =
    selectedOption &&
    visibleOptions.some((option) => option.value === selectedOption.value)
  const options = selectedOptionIsPreviewed
    ? visibleOptions
    : selectedOption
      ? [...visibleOptions, selectedOption]
      : visibleOptions
  const hiddenOptionCount = Math.max(props.options.length - visibleCount, 0)
  const nextOptionCount = Math.min(previewCount, hiddenOptionCount)
  const canShowMore = hiddenOptionCount > 0
  const canShowLess = visibleOptionCount !== null && visibleCount > previewCount

  return (
    <Collapsible
      open={props.open}
      onOpenChange={props.onOpenChange}
      className='flex flex-col gap-2'
    >
      <div className='flex h-6 items-center justify-between gap-2'>
        <CollapsibleTrigger
          render={
            <button
              type='button'
              title={toggleLabel}
              aria-label={toggleLabel}
              className='group text-foreground hover:text-foreground/80 focus-visible:ring-ring/40 flex min-w-0 flex-1 items-center justify-between gap-2 rounded-md text-left text-sm font-semibold transition-colors outline-none focus-visible:ring-2'
            />
          }
        >
          <span className='min-w-0 truncate'>{props.title}</span>
          <ChevronDown
            className={cn(
              'text-muted-foreground size-4 shrink-0 transition-transform',
              props.open && 'rotate-180'
            )}
            aria-hidden='true'
          />
        </CollapsibleTrigger>
        {hasSelection && (
          <Button
            type='button'
            variant='ghost'
            size='icon-xs'
            title={resetLabel}
            aria-label={resetLabel}
            className='text-muted-foreground hover:text-foreground size-6 rounded-md'
            onClick={() => props.onChange(FILTER_ALL)}
          >
            <ZenMuxRefreshCircleIcon className='size-4' />
          </Button>
        )}
      </div>
      <CollapsibleContent className={cn(!props.open && 'hidden')}>
        <div className='flex flex-col gap-1'>
          <FilterButton
            active={props.value === FILTER_ALL}
            label={props.allLabel}
            count={props.allCount}
            onClick={() => props.onChange(FILTER_ALL)}
          />
          <div className='flex flex-col gap-1'>
            {options.map((option) => (
              <FilterButton
                key={option.value}
                active={props.value === option.value}
                label={option.translateLabel ? t(option.label) : option.label}
                count={option.count}
                icon={option.icon}
                onClick={() => props.onChange(option.value)}
              />
            ))}
          </div>
          {(canShowMore || canShowLess) && (
            <FilterButton
              active={false}
              label={
                canShowMore
                  ? t('Show {{count}} more', { count: nextOptionCount })
                  : t('Show less')
              }
              onClick={() => {
                if (canShowMore) {
                  setVisibleOptionCount(visibleCount + previewCount)
                  return
                }
                setVisibleOptionCount(null)
              }}
            />
          )}
        </div>
      </CollapsibleContent>
    </Collapsible>
  )
}

function TopModalityTabs(props: {
  value: string
  counts: Record<string, number>
  total: number
  onChange: (value: string) => void
}) {
  const { t } = useTranslation()

  return (
    <div className='flex min-w-0 flex-wrap items-center gap-2'>
      <Button
        type='button'
        variant='ghost'
        size='sm'
        onClick={() => props.onChange(FILTER_ALL)}
        className={cn(
          "h-9 gap-2 rounded-md border px-3.5 text-sm font-normal shadow-none [&_svg:not([class*='size-'])]:size-4",
          props.value === FILTER_ALL
            ? 'border-border bg-white text-neutral-950 shadow-[0_2px_4px_rgba(0,0,0,0.05)] hover:bg-white dark:border-white/15 dark:bg-neutral-800 dark:text-neutral-50 dark:shadow-none dark:hover:bg-neutral-800'
            : 'border-border bg-neutral-100 text-neutral-500 hover:bg-neutral-200 hover:text-neutral-700 dark:border-white/10 dark:bg-neutral-900 dark:text-neutral-400 dark:hover:bg-neutral-800 dark:hover:text-neutral-100'
        )}
      >
        <ZenMuxModelsIcon data-icon='inline-start' />
        <span>All</span>
        <span className='text-muted-foreground font-mono text-[13px] tabular-nums dark:text-neutral-300'>
          {props.total}
        </span>
      </Button>
      {TOP_MODALITY_OPTIONS.map((option) => {
        const Icon = option.icon
        const active = props.value === option.value
        const count = props.counts[option.value] || 0

        return (
          <Button
            key={option.value}
            type='button'
            variant='ghost'
            size='sm'
            onClick={() => props.onChange(option.value)}
            className={cn(
              "h-9 gap-2 rounded-md border px-3.5 text-sm font-normal shadow-none [&_svg:not([class*='size-'])]:size-4",
              active
                ? 'border-border bg-white text-neutral-950 shadow-[0_2px_4px_rgba(0,0,0,0.05)] hover:bg-white dark:border-white/15 dark:bg-neutral-800 dark:text-neutral-50 dark:shadow-none dark:hover:bg-neutral-800'
                : 'border-border bg-neutral-100 text-neutral-500 hover:bg-neutral-200 hover:text-neutral-700 dark:border-white/10 dark:bg-neutral-900 dark:text-neutral-400 dark:hover:bg-neutral-800 dark:hover:text-neutral-100'
            )}
          >
            <Icon data-icon='inline-start' />
            {t(option.label)}
            {count > 0 && (
              <span className='text-muted-foreground font-mono text-[13px] tabular-nums dark:text-neutral-300'>
                {count}
              </span>
            )}
          </Button>
        )
      })}
    </div>
  )
}

function SuccessRateStrip(props: { perf?: PerfModelSummary }) {
  const normalized = normalizeSuccessRate(props.perf?.success_rate)
  const segments = 36
  const filled =
    normalized === null ? 0 : Math.round((normalized / 100) * segments)
  const tone =
    normalized === null
      ? 'bg-muted'
      : normalized >= 99
        ? 'bg-emerald-500'
        : normalized >= 95
          ? 'bg-amber-500'
          : 'bg-destructive'

  return (
    <div className='min-w-0'>
      <div className='grid grid-cols-[repeat(36,minmax(0,1fr))] gap-0.5'>
        {Array.from({ length: segments }).map((_, index) => (
          <span
            key={index}
            className={cn(
              'h-2 rounded-[2px]',
              index < filled ? tone : 'bg-muted'
            )}
          />
        ))}
      </div>
    </div>
  )
}

function TinyModalities(props: { modalities: Modality[] | undefined }) {
  const { t } = useTranslation()
  const modalities = normalizeModalities(props.modalities)

  return (
    <span className='inline-flex min-w-0 flex-wrap items-center gap-1'>
      {modalities.length === 0 ? (
        <span className='text-muted-foreground/60 text-xs'>-</span>
      ) : (
        modalities.map((value) => {
          const option = MODALITY_OPTIONS.find((item) => item.value === value)
          const Icon = option?.icon
          return (
            <span
              key={value}
              title={t(option?.label || value)}
              className='text-muted-foreground bg-background inline-flex size-4 items-center justify-center rounded-[3px] border dark:border-white/10 dark:bg-neutral-900 dark:text-neutral-400'
            >
              {Icon && <Icon className='size-3' />}
            </span>
          )
        })
      )}
    </span>
  )
}

function InfoLine(props: {
  label: string
  children: ReactNode
  valueClassName?: string
}) {
  return (
    <div className='grid min-h-5 grid-cols-[max-content_minmax(0,1fr)] items-center gap-1 text-sm'>
      <span className='text-muted-foreground shrink-0 whitespace-nowrap'>
        {props.label}
      </span>
      <span className={cn('text-foreground min-w-0', props.valueClassName)}>
        {props.children}
      </span>
    </div>
  )
}

type CatalogPriceSummary =
  | {
      kind: 'paired'
      inputValue: string
      outputValue: string
      inputUnit: string
      outputUnit: string
    }
  | {
      kind: 'single'
      label: string
      value: string
      unit?: string
    }

function usePriceRows(props: {
  model: PricingModel
  tokenUnit: TokenUnit
  priceRate: number
  usdExchangeRate: number
  selectedGroup?: string
}): CatalogPriceSummary {
  const { t } = useTranslation()
  const isTokenBased = isTokenBasedModel(props.model)
  const tokenUnitLabel = props.tokenUnit === 'K' ? '1K tokens' : '1M tokens'
  const dynamicSummary =
    props.model.billing_mode === 'tiered_expr' && props.model.billing_expr
      ? getDynamicPricingSummary(props.model, {
          tokenUnit: props.tokenUnit,
          priceRate: props.priceRate,
          usdExchangeRate: props.usdExchangeRate,
          groupRatioMultiplier: getDynamicDisplayGroupRatio(
            props.model,
            props.selectedGroup
          ),
        })
      : null
  const soraSummary = getSoraPricingDisplay(props.model, {
    priceRate: props.priceRate,
    usdExchangeRate: props.usdExchangeRate,
    selectedGroup: props.selectedGroup,
  })

  if (
    dynamicSummary &&
    (dynamicSummary.isSpecialExpression || dynamicSummary.entries.length === 0)
  ) {
    return {
      kind: 'single',
      label: t('Billing'),
      value: t('Dynamic Pricing'),
    }
  }

  if (dynamicSummary && dynamicSummary.entries.length > 0) {
    const inputEntry =
      dynamicSummary.primaryEntries.find(
        (entry) => entry.field === 'inputPrice'
      ) ||
      dynamicSummary.primaryEntries[0] ||
      dynamicSummary.secondaryEntries[0]
    const outputEntry =
      dynamicSummary.primaryEntries.find(
        (entry) => entry.field === 'outputPrice'
      ) ||
      dynamicSummary.primaryEntries[1] ||
      dynamicSummary.secondaryEntries[1]

    return {
      kind: 'paired',
      inputValue: inputEntry?.formatted || t('Dynamic Pricing'),
      outputValue: outputEntry?.formatted || '-',
      inputUnit:
        inputEntry?.displayUnit === 'request' ? t('request') : tokenUnitLabel,
      outputUnit:
        outputEntry?.displayUnit === 'request' ? t('request') : tokenUnitLabel,
    }
  }

  if (soraSummary) {
    return {
      kind: 'paired',
      inputValue: stripTrailingZeros(soraSummary.basePrice),
      outputValue: '-',
      inputUnit: t('request'),
      outputUnit: '',
    }
  }

  if (!isTokenBased) {
    return {
      kind: 'paired',
      inputValue: formatRequestPrice(
        props.model,
        false,
        props.priceRate,
        props.usdExchangeRate,
        props.selectedGroup
      ),
      outputValue: '-',
      inputUnit: t('request'),
      outputUnit: '',
    }
  }

  return {
    kind: 'paired',
    inputValue: formatPrice(
      props.model,
      'input',
      props.tokenUnit,
      false,
      props.priceRate,
      props.usdExchangeRate,
      props.selectedGroup
    ),
    outputValue: formatPrice(
      props.model,
      'output',
      props.tokenUnit,
      false,
      props.priceRate,
      props.usdExchangeRate,
      props.selectedGroup
    ),
    inputUnit: tokenUnitLabel,
    outputUnit: tokenUnitLabel,
  }
}

function CatalogModelCard(props: {
  model: PricingModel
  perf?: PerfModelSummary
  priceRate: number
  usdExchangeRate: number
  selectedGroup?: string
  onOpen: () => void
}) {
  const { t } = useTranslation()
  const { copyToClipboard } = useCopyToClipboard()
  const modelIconKey = props.model.icon || props.model.vendor_icon
  const modelIcon = modelIconKey ? getLobeIcon(modelIconKey, 30) : null
  const initial = props.model.model_name?.charAt(0).toUpperCase() || '?'
  const modelTitle = getCatalogModelTitle(props.model)
  const priceRows = usePriceRows({
    model: props.model,
    tokenUnit: TOKEN_UNIT,
    priceRate: props.priceRate,
    usdExchangeRate: props.usdExchangeRate,
    selectedGroup: props.selectedGroup,
  })

  return (
    <article className='group bg-background hover:bg-muted/20 flex min-h-[284px] flex-col border-b p-4 transition-colors lg:border-r 2xl:p-5 dark:border-white/10 dark:hover:bg-white/[0.03]'>
      <div className='flex items-start justify-between gap-3'>
        <div className='flex min-w-0 items-start gap-3'>
          <div className='bg-muted/60 flex size-10 shrink-0 items-center justify-center rounded-lg dark:bg-white/5'>
            {modelIcon || (
              <span className='text-muted-foreground text-sm font-bold'>
                {initial}
              </span>
            )}
          </div>
          <div className='min-w-0'>
            <div className='flex min-w-0 flex-wrap items-center gap-1.5'>
              <button
                type='button'
                onClick={props.onOpen}
                className='focus-visible:ring-ring/50 max-w-full min-w-0 cursor-pointer rounded-sm text-left focus-visible:ring-2 focus-visible:outline-none'
              >
                <h2 className='text-foreground min-w-0 truncate font-mono text-base leading-5 font-semibold underline-offset-2 hover:underline'>
                  {modelTitle}
                </h2>
              </button>
            </div>
            <div className='text-muted-foreground mt-0.5 flex min-w-0 items-center gap-1.5 text-xs'>
              <span className='truncate font-mono'>
                {props.model.model_name}
              </span>
              <Button
                type='button'
                variant='ghost'
                size='icon-xs'
                title={t('Copy')}
                className='text-muted-foreground size-5'
                onClick={() => copyToClipboard(props.model.model_name || '')}
              >
                <ZenMuxCopyIcon />
              </Button>
            </div>
          </div>
        </div>

        <div className='flex shrink-0 items-center gap-1'>
          <Button
            type='button'
            variant='ghost'
            size='icon'
            title={t('Details')}
            aria-label={t('Details')}
            className='text-muted-foreground hover:text-foreground'
            onClick={props.onOpen}
          >
            <HugeiconsIcon icon={ArrowRight01Icon} strokeWidth={2} />
          </Button>
        </div>
      </div>

      <p className='text-foreground/90 mt-8 line-clamp-2 min-h-[2.5rem] text-sm leading-5'>
        {props.model.description || t('No description available.')}
      </p>

      <div className='mt-4 grid gap-x-3 gap-y-2.5 sm:grid-cols-2 2xl:gap-x-5'>
        <div className='flex min-w-0 flex-col gap-2'>
          <InfoLine label={`${t('Input Types')}:`}>
            <TinyModalities modalities={props.model.input_modalities} />
          </InfoLine>
          {priceRows.kind === 'paired' && (
            <InfoLine label={`${t('Input')}:`} valueClassName='truncate'>
              <span className='font-mono font-medium'>
                {priceRows.inputValue}
              </span>
              {priceRows.inputUnit && (
                <span className='text-muted-foreground ml-1 text-xs'>
                  /{priceRows.inputUnit}
                </span>
              )}
            </InfoLine>
          )}
        </div>
        <div className='flex min-w-0 flex-col gap-2'>
          <InfoLine label={`${t('Output Types')}:`}>
            <TinyModalities modalities={props.model.output_modalities} />
          </InfoLine>
          {priceRows.kind === 'paired' && (
            <InfoLine label={`${t('Output')}:`} valueClassName='truncate'>
              <span className='font-mono font-medium'>
                {priceRows.outputValue}
              </span>
              {priceRows.outputUnit && (
                <span className='text-muted-foreground ml-1 text-xs'>
                  /{priceRows.outputUnit}
                </span>
              )}
            </InfoLine>
          )}
        </div>
        {priceRows.kind === 'single' && (
          <div className='min-w-0 sm:col-span-2'>
            <InfoLine label={`${priceRows.label}:`} valueClassName='truncate'>
              <span className='font-medium'>{priceRows.value}</span>
              {priceRows.unit && (
                <span className='text-muted-foreground ml-1 text-xs'>
                  /{priceRows.unit}
                </span>
              )}
            </InfoLine>
          </div>
        )}
      </div>

      <div className='mt-auto grid gap-4 pt-5 sm:grid-cols-[minmax(0,1fr)_minmax(158px,50%)]'>
        <div className='min-w-0 space-y-2'>
          <div className='text-muted-foreground text-sm'>
            {t('Available from {{count}} providers', { count: 1 })}
          </div>
          <div className='flex min-w-0 items-center gap-2'>
            <div className='bg-muted/60 flex size-5 shrink-0 items-center justify-center rounded-full dark:bg-white/5'>
              {props.model.vendor_icon ? (
                getLobeIcon(props.model.vendor_icon, 16)
              ) : (
                <span className='text-muted-foreground text-[10px] font-semibold'>
                  {(props.model.vendor_name || '?').charAt(0)}
                </span>
              )}
            </div>
          </div>
        </div>

        <div className='min-w-0 space-y-2'>
          <div className='flex items-center justify-between gap-2 text-xs'>
            <span className='text-muted-foreground truncate'>
              {formatCreatedTime(props.model.created_time)}
            </span>
            <span className='text-foreground shrink-0 font-mono tabular-nums'>
              {formatSuccessRate(props.perf?.success_rate)}
            </span>
          </div>
          <SuccessRateStrip perf={props.perf} />
        </div>
      </div>
    </article>
  )
}

function CatalogSkeleton() {
  return (
    <div className='bg-background grid min-h-[calc(100svh-4rem)] pt-16 lg:grid-cols-[248px_minmax(0,1fr)]'>
      <aside className='border-r p-4'>
        <div className='space-y-7'>
          {Array.from({ length: 5 }).map((_, index) => (
            <div key={index} className='space-y-2'>
              <Skeleton className='h-4 w-24' />
              <Skeleton className='h-8 w-full' />
              <Skeleton className='h-8 w-full' />
              <Skeleton className='h-8 w-4/5' />
            </div>
          ))}
        </div>
      </aside>
      <main className='min-w-0'>
        <div className='border-b px-5 py-7'>
          <Skeleton className='h-8 w-36' />
          <Skeleton className='mt-3 h-4 w-24' />
        </div>
        <div className='grid xl:grid-cols-3'>
          {Array.from({ length: 9 }).map((_, index) => (
            <div key={index} className='space-y-4 border-r border-b p-5'>
              <Skeleton className='h-10 w-3/4' />
              <Skeleton className='h-10 w-full' />
              <Skeleton className='h-16 w-full' />
              <Skeleton className='h-8 w-full' />
            </div>
          ))}
        </div>
      </main>
    </div>
  )
}

function getModelSquareTheme(
  status: Record<string, unknown> | null | undefined
): 'catalog' | 'classic' {
  return status?.model_square_theme === 'classic' ? 'classic' : 'catalog'
}

export function Pricing() {
  const { status } = useStatus()

  if (getModelSquareTheme(status) === 'classic') {
    return <ClassicPricing />
  }

  return <CatalogPricing />
}

function CatalogPricing() {
  const { t } = useTranslation()
  const [selectedModelName, setSelectedModelName] = useState<string | null>(
    null
  )
  const [searchInput, setSearchInput] = useState('')
  const [inputFilter, setInputFilter] = useState<string>(FILTER_ALL)
  const [outputFilter, setOutputFilter] = useState<string>(FILTER_ALL)
  const [vendorFilter, setVendorFilter] = useState<string>(FILTER_ALL)
  const [groupFilter, setGroupFilter] = useState<string>(FILTER_ALL)
  const [sidebarSectionOpen, setSidebarSectionOpen] = useState<
    Record<SidebarSectionId, boolean>
  >({
    input: true,
    output: true,
    vendor: true,
    group: true,
  })
  const [visibleState, setVisibleState] = useState({
    filterKey: '',
    count: LOAD_BATCH_SIZE,
  })
  const sentinelRef = useRef<HTMLDivElement | null>(null)

  const {
    models,
    vendors,
    groupRatio,
    usableGroup,
    endpointMap,
    autoGroups,
    isLoading,
    priceRate,
    usdExchangeRate,
  } = usePricingData()

  const perfQuery = useQuery({
    queryKey: ['perf-metrics-summary', 24],
    queryFn: () => getPerfMetricsSummary(24),
    staleTime: 60 * 1000,
    retry: false,
  })

  const perfMap = useMemo(() => {
    const map = new Map<string, PerfModelSummary>()
    for (const model of perfQuery.data?.data?.models ?? []) {
      map.set(model.model_name, model)
    }
    return map
  }, [perfQuery.data])

  const inputOptions = useMemo(
    () => buildModalityOptions(models, 'input_modalities'),
    [models]
  )
  const outputOptions = useMemo(
    () => buildModalityOptions(models, 'output_modalities'),
    [models]
  )
  const vendorOptions = useMemo(
    () => buildVendorOptions(models, vendors),
    [models, vendors]
  )
  const groupOptions = useMemo(
    () => buildGroupOptions(models, usableGroup),
    [models, usableGroup]
  )

  const inputCounts = useMemo(
    () =>
      Object.fromEntries(
        inputOptions.map((option) => [option.value, option.count])
      ),
    [inputOptions]
  )

  const searchQuery = searchInput.trim().toLowerCase()
  const filteredModels = useMemo(() => {
    return models
      .filter((model) => {
        if (searchQuery && !getSearchText(model).includes(searchQuery)) {
          return false
        }
        if (!hasModality(model, 'input_modalities', inputFilter)) {
          return false
        }
        if (!hasModality(model, 'output_modalities', outputFilter)) {
          return false
        }
        if (
          vendorFilter !== FILTER_ALL &&
          String(model.vendor_id || '') !== vendorFilter
        ) {
          return false
        }
        if (
          groupFilter !== FILTER_ALL &&
          !isModelAvailableInGroup(model, groupFilter)
        ) {
          return false
        }
        return true
      })
      .sort(sortByCreatedTime)
  }, [
    groupFilter,
    inputFilter,
    models,
    outputFilter,
    searchQuery,
    vendorFilter,
  ])

  const filterKey = [
    searchQuery,
    inputFilter,
    outputFilter,
    vendorFilter,
    groupFilter,
  ].join('|')

  const visibleCount =
    visibleState.filterKey === filterKey ? visibleState.count : LOAD_BATCH_SIZE

  useEffect(() => {
    const sentinel = sentinelRef.current
    if (!sentinel) return

    const observer = new IntersectionObserver(
      (entries) => {
        const entry = entries[0]
        if (!entry?.isIntersecting) return
        setVisibleState((current) => {
          const currentCount =
            current.filterKey === filterKey ? current.count : LOAD_BATCH_SIZE
          return {
            filterKey,
            count: Math.min(
              currentCount + LOAD_BATCH_SIZE,
              filteredModels.length
            ),
          }
        })
      },
      { rootMargin: '640px 0px' }
    )

    observer.observe(sentinel)
    return () => observer.disconnect()
  }, [filterKey, filteredModels.length])

  const visibleModels = filteredModels.slice(0, visibleCount)
  const selectedModel = useMemo(
    () =>
      selectedModelName
        ? models.find((model) => model.model_name === selectedModelName) || null
        : null,
    [models, selectedModelName]
  )

  const hasActiveFilters =
    Boolean(searchQuery) ||
    inputFilter !== FILTER_ALL ||
    outputFilter !== FILTER_ALL ||
    vendorFilter !== FILTER_ALL ||
    groupFilter !== FILTER_ALL

  const clearFilters = useCallback(() => {
    setSearchInput('')
    setInputFilter(FILTER_ALL)
    setOutputFilter(FILTER_ALL)
    setVendorFilter(FILTER_ALL)
    setGroupFilter(FILTER_ALL)
  }, [])

  const setSidebarSectionExpanded = useCallback(
    (section: SidebarSectionId, open: boolean) => {
      setSidebarSectionOpen((current) => ({
        ...current,
        [section]: open,
      }))
    },
    []
  )

  if (isLoading) {
    return (
      <PublicLayout showMainContainer={false}>
        <CatalogSkeleton />
      </PublicLayout>
    )
  }

  return (
    <PublicLayout showMainContainer={false}>
      <PageTransition className='bg-background text-foreground min-h-svh pt-16'>
        <div className='grid min-h-[calc(100svh-4rem)] lg:grid-cols-[248px_minmax(0,1fr)]'>
          <aside className='bg-background/95 border-b p-4 lg:sticky lg:top-16 lg:h-[calc(100svh-4rem)] lg:overflow-y-auto lg:border-r lg:border-b-0'>
            <div className='space-y-7'>
              <FilterSection
                title={t('Input Types')}
                options={inputOptions}
                value={inputFilter}
                onChange={setInputFilter}
                allLabel={t('All')}
                allCount={models.length}
                open={sidebarSectionOpen.input}
                onOpenChange={(open) =>
                  setSidebarSectionExpanded('input', open)
                }
              />
              <FilterSection
                title={t('Output Types')}
                options={outputOptions}
                value={outputFilter}
                onChange={setOutputFilter}
                allLabel={t('All')}
                allCount={models.length}
                open={sidebarSectionOpen.output}
                onOpenChange={(open) =>
                  setSidebarSectionExpanded('output', open)
                }
              />
              <FilterSection
                title={t('Vendor')}
                options={vendorOptions}
                value={vendorFilter}
                onChange={setVendorFilter}
                allLabel={t('All')}
                allCount={vendorOptions.length}
                open={sidebarSectionOpen.vendor}
                onOpenChange={(open) =>
                  setSidebarSectionExpanded('vendor', open)
                }
                previewCount={FILTER_PREVIEW_COUNT}
              />
              <FilterSection
                title={t('Group')}
                options={groupOptions}
                value={groupFilter}
                onChange={setGroupFilter}
                allLabel={t('All')}
                allCount={groupOptions.length}
                open={sidebarSectionOpen.group}
                onOpenChange={(open) =>
                  setSidebarSectionExpanded('group', open)
                }
                previewCount={FILTER_PREVIEW_COUNT}
              />
              {hasActiveFilters && (
                <Button
                  type='button'
                  variant='outline'
                  size='sm'
                  className='w-full'
                  onClick={clearFilters}
                >
                  {t('Clear all filters')}
                </Button>
              )}
            </div>
          </aside>

          <main className='min-w-0'>
            <header className='border-b px-4 py-5 sm:px-6 lg:px-8'>
              <div className='flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between'>
                <div className='min-w-0'>
                  <h1 className='text-2xl leading-8 font-bold'>
                    {t('Models')}
                  </h1>
                  <p className='text-muted-foreground mt-1 text-sm'>
                    {t('{{count}} models', { count: models.length })}
                  </p>
                </div>

                <div className='flex w-full items-center gap-2 sm:w-auto'>
                  <label className='border-border bg-background text-muted-foreground focus-within:ring-ring/20 flex h-8 w-[190px] max-w-full shrink-0 items-center rounded-[6px] border transition-colors focus-within:border-neutral-300 focus-within:ring-2 dark:border-white/10 dark:bg-neutral-950 dark:focus-within:border-white/20'>
                    <ZenMuxSearchIcon className='ml-3 size-3.5 shrink-0' />
                    <input
                      type='text'
                      aria-label={t('Search models')}
                      value={searchInput}
                      onChange={(event) => setSearchInput(event.target.value)}
                      className='text-foreground h-full min-w-0 flex-1 bg-transparent px-2 text-sm outline-none'
                    />
                  </label>
                  <Button
                    type='button'
                    variant='outline'
                    size='sm'
                    className='border-border bg-background hover:bg-muted h-8 gap-1.5 rounded-[6px] px-3 text-sm font-normal dark:border-white/10 dark:bg-neutral-950 dark:hover:bg-neutral-900'
                  >
                    <ZenMuxSortIcon className='text-muted-foreground size-4' />
                    {t('Latest')}
                  </Button>
                </div>
              </div>

              <div className='mt-5'>
                <TopModalityTabs
                  value={inputFilter}
                  counts={inputCounts}
                  total={models.length}
                  onChange={setInputFilter}
                />
              </div>
            </header>

            {filteredModels.length === 0 ? (
              <div className='flex min-h-[420px] flex-col items-center justify-center px-6 text-center'>
                <HugeiconsIcon
                  icon={SearchIcon}
                  strokeWidth={1.8}
                  className='text-muted-foreground/45 mb-3 size-10'
                />
                <h2 className='text-base font-semibold'>
                  {t('No models found')}
                </h2>
                <p className='text-muted-foreground mt-1 max-w-sm text-sm'>
                  {t('No models match your current filters.')}
                </p>
                {hasActiveFilters && (
                  <Button
                    type='button'
                    variant='outline'
                    size='sm'
                    className='mt-5'
                    onClick={clearFilters}
                  >
                    {t('Clear all filters')}
                  </Button>
                )}
              </div>
            ) : (
              <>
                <div className='grid min-w-0 xl:grid-cols-3'>
                  {visibleModels.map((model) => (
                    <CatalogModelCard
                      key={model.model_name}
                      model={model}
                      perf={perfMap.get(model.model_name || '')}
                      priceRate={priceRate}
                      usdExchangeRate={usdExchangeRate}
                      selectedGroup={groupFilter}
                      onOpen={() =>
                        setSelectedModelName(model.model_name || '')
                      }
                    />
                  ))}
                </div>

                <div
                  ref={sentinelRef}
                  className='text-muted-foreground flex h-20 items-center justify-center border-t text-sm'
                >
                  {visibleModels.length < filteredModels.length
                    ? t('Loading more...')
                    : t('{{count}} models shown', {
                        count: filteredModels.length,
                      })}
                </div>
              </>
            )}
          </main>
        </div>

        {selectedModel && (
          <ModelDetailsDrawer
            open={Boolean(selectedModel)}
            onOpenChange={(open) => {
              if (!open) setSelectedModelName(null)
            }}
            model={selectedModel}
            groupRatio={groupRatio || {}}
            usableGroup={usableGroup || {}}
            endpointMap={
              (endpointMap as Record<
                string,
                { path?: string; method?: string }
              >) || {}
            }
            autoGroups={autoGroups || []}
            priceRate={priceRate ?? 1}
            usdExchangeRate={usdExchangeRate ?? 1}
            tokenUnit={TOKEN_UNIT}
            showRechargePrice={false}
          />
        )}
      </PageTransition>
    </PublicLayout>
  )
}
