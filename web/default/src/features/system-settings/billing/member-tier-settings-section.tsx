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
import { useEffect, useMemo, useState, type ElementType } from 'react'
import {
  ArrowDown,
  ArrowUp,
  Info,
  Plus,
  RefreshCw,
  RotateCcw,
  Save,
  Trash2,
} from 'lucide-react'
import { useMutation } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  NativeSelect,
  NativeSelectOption,
} from '@/components/ui/native-select'
import { Switch } from '@/components/ui/switch'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { formatQuota } from '@/lib/format'
import { cn } from '@/lib/utils'
import { recalculateMemberTiers } from '../api'
import { SettingsPageActionsPortal } from '../components/settings-page-context'
import { SettingsSection } from '../components/settings-section'
import { useUpdateOption } from '../hooks/use-update-option'
import { safeJsonParse } from '../utils/json-parser'

type MemberTierRule = {
  enabled: boolean
  group: string
  display_name: string
  description?: string
  min_topup_quota: number
  min_used_quota: number
}

type MemberTierSettingsSectionProps = {
  defaultValue: string
  groupRatio: string
  topupGroupRatio: string
}

type IconButtonProps = {
  label: string
  disabled?: boolean
  onClick: () => void
  icon: ElementType
  destructive?: boolean
}

const normalizeRule = (rule: Partial<MemberTierRule>): MemberTierRule => {
  const group = String(rule.group ?? '').trim()
  const displayName = String(rule.display_name ?? '').trim()
  return {
    enabled: rule.enabled !== false,
    group,
    display_name: displayName || group,
    description: String(rule.description ?? '').trim(),
    min_topup_quota: Math.max(0, Number(rule.min_topup_quota) || 0),
    min_used_quota: Math.max(0, Number(rule.min_used_quota) || 0),
  }
}

const parseMemberTierRules = (value: string): MemberTierRule[] => {
  const parsed = safeJsonParse<unknown[]>(value, {
    fallback: [],
    silent: true,
    context: 'member tier rules',
  })
  if (!Array.isArray(parsed)) return []

  return parsed
    .filter((item): item is Partial<MemberTierRule> => {
      return item != null && typeof item === 'object'
    })
    .map(normalizeRule)
}

const serializeRules = (rules: MemberTierRule[]) =>
  JSON.stringify(rules.map(normalizeRule), null, 2)

function IconButton({
  label,
  disabled,
  onClick,
  icon: Icon,
  destructive,
}: IconButtonProps) {
  return (
    <Tooltip>
      <TooltipTrigger
        render={
          <Button
            type='button'
            variant={destructive ? 'destructive' : 'ghost'}
            size='icon-sm'
            disabled={disabled}
            onClick={onClick}
          >
            <Icon />
            <span className='sr-only'>{label}</span>
          </Button>
        }
      />
      <TooltipContent>
        <p>{label}</p>
      </TooltipContent>
    </Tooltip>
  )
}

function parseRatioMap(value: string): Record<string, unknown> {
  const parsed = safeJsonParse<Record<string, unknown>>(value, {
    fallback: {},
    silent: true,
    context: 'group ratios',
  })
  if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
    return {}
  }
  return parsed
}

function formatRatio(value: unknown, fallback = '1') {
  const numeric =
    typeof value === 'number' ? value : Number.parseFloat(String(value ?? ''))
  if (!Number.isFinite(numeric)) return fallback
  return String(numeric)
}

export function MemberTierSettingsSection({
  defaultValue,
  groupRatio,
  topupGroupRatio,
}: MemberTierSettingsSectionProps) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()
  const defaultRules = useMemo(
    () => parseMemberTierRules(defaultValue),
    [defaultValue]
  )
  const [rules, setRules] = useState<MemberTierRule[]>(defaultRules)

  useEffect(() => {
    setRules(defaultRules)
  }, [defaultRules])

  const groupRatioMap = useMemo(() => parseRatioMap(groupRatio), [groupRatio])
  const topupRatioMap = useMemo(
    () => parseRatioMap(topupGroupRatio),
    [topupGroupRatio]
  )

  const groupOptions = useMemo(() => {
    const keys = new Set<string>([
      ...Object.keys(groupRatioMap),
      ...Object.keys(topupRatioMap),
      ...rules.map((rule) => rule.group).filter(Boolean),
    ])
    if (keys.size === 0) keys.add('default')
    return Array.from(keys).sort((a, b) => {
      if (a === 'default') return -1
      if (b === 'default') return 1
      return a.localeCompare(b)
    })
  }, [groupRatioMap, rules, topupRatioMap])

  const isDirty = serializeRules(rules) !== serializeRules(defaultRules)

  const recalculateMutation = useMutation({
    mutationFn: recalculateMemberTiers,
    onSuccess: (res) => {
      if (!res.success || !res.data) {
        toast.error(res.message || t('Recalculation failed'))
        return
      }
      toast.success(
        t(
          'Recalculation finished: {{scanned}} scanned, {{upgraded}} upgraded, {{skipped}} skipped by active subscription.',
          {
            scanned: res.data.scanned,
            upgraded: res.data.upgraded,
            skipped: res.data.skipped_by_active_subscription,
          }
        )
      )
    },
    onError: (error: Error) => {
      toast.error(error.message || t('Recalculation failed'))
    },
  })

  const updateRule = (index: number, patch: Partial<MemberTierRule>) => {
    setRules((current) =>
      current.map((rule, idx) =>
        idx === index ? normalizeRule({ ...rule, ...patch }) : rule
      )
    )
  }

  const addRule = () => {
    const usedGroups = new Set(rules.map((rule) => rule.group))
    const nextGroup = groupOptions.find((group) => !usedGroups.has(group)) ?? ''
    setRules((current) => [
      ...current,
      normalizeRule({
        enabled: true,
        group: nextGroup,
        display_name: `Tier ${current.length + 1}`,
        min_topup_quota: 0,
        min_used_quota: 0,
      }),
    ])
  }

  const removeRule = (index: number) => {
    setRules((current) => current.filter((_, idx) => idx !== index))
  }

  const moveRule = (index: number, direction: -1 | 1) => {
    setRules((current) => {
      const nextIndex = index + direction
      if (nextIndex < 0 || nextIndex >= current.length) return current
      const next = [...current]
      const [item] = next.splice(index, 1)
      next.splice(nextIndex, 0, item)
      return next
    })
  }

  const validateRules = () => {
    const seen = new Set<string>()
    for (const [index, rule] of rules.entries()) {
      if (!rule.group) {
        toast.error(t('Tier {{index}} must select a group.', { index: index + 1 }))
        return false
      }
      if (seen.has(rule.group)) {
        toast.error(t('Tier group cannot be duplicated: {{group}}', { group: rule.group }))
        return false
      }
      seen.add(rule.group)
      if (rule.min_topup_quota < 0 || rule.min_used_quota < 0) {
        toast.error(t('Tier thresholds must be non-negative.'))
        return false
      }
    }
    return true
  }

  const saveRules = async () => {
    if (!validateRules()) return
    await updateOption.mutateAsync({
      key: 'MemberTierRules',
      value: serializeRules(rules),
    })
  }

  const resetRules = () => {
    setRules(defaultRules)
  }

  const renderRatioSummary = (group: string) => (
    <div className='text-muted-foreground grid gap-1 text-xs'>
      <span>
        {t('Model ratio')}: {formatRatio(groupRatioMap[group])}
      </span>
      <span>
        {t('Top-up ratio')}: {formatRatio(topupRatioMap[group])}
      </span>
    </div>
  )

  const renderConditionSummary = (rule: MemberTierRule) => (
    <div className='text-muted-foreground grid gap-1 text-xs'>
      <span>
        {t('Minimum top-up')}: {formatQuota(rule.min_topup_quota)}
      </span>
      <span>
        {t('Minimum consumption')}: {formatQuota(rule.min_used_quota)}
      </span>
    </div>
  )

  return (
    <TooltipProvider>
      <SettingsSection title={t('Member Tiers')}>
        <SettingsPageActionsPortal>
          <Button
            type='button'
            variant='outline'
            size='sm'
            disabled={recalculateMutation.isPending}
            onClick={() => recalculateMutation.mutate()}
          >
            <RefreshCw
              data-icon='inline-start'
              className={cn(recalculateMutation.isPending && 'animate-spin')}
            />
            <span>
              {recalculateMutation.isPending
                ? t('Recalculating...')
                : t('Recalculate')}
            </span>
          </Button>
          <Button
            type='button'
            variant='outline'
            size='sm'
            disabled={!isDirty || updateOption.isPending}
            onClick={resetRules}
          >
            <RotateCcw data-icon='inline-start' />
            <span>{t('Reset')}</span>
          </Button>
          <Button
            type='button'
            size='sm'
            disabled={!isDirty || updateOption.isPending}
            onClick={saveRules}
          >
            <Save data-icon='inline-start' />
            <span>{updateOption.isPending ? t('Saving...') : t('Save Rules')}</span>
          </Button>
        </SettingsPageActionsPortal>

        <Alert>
          <Info className='h-4 w-4' />
          <AlertDescription>
            {t(
              'Upgrade checks run after any successful top-up and automatically at local midnight every day.'
            )}
          </AlertDescription>
        </Alert>

        <div className='flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between'>
          <div className='text-muted-foreground text-sm'>
            {t('Rules are evaluated from top to bottom; later matched tiers win.')}
          </div>
          <Button type='button' variant='outline' onClick={addRule}>
            <Plus data-icon='inline-start' />
            <span>{t('Add Tier')}</span>
          </Button>
        </div>

        {rules.length === 0 ? (
          <div className='text-muted-foreground rounded-lg border border-dashed p-8 text-center text-sm'>
            {t('No member tiers configured. Add a tier to get started.')}
          </div>
        ) : (
          <>
            <div className='hidden rounded-lg border md:block'>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className='w-16'>{t('Enabled')}</TableHead>
                    <TableHead className='min-w-48'>{t('Tier')}</TableHead>
                    <TableHead className='min-w-36'>{t('Group')}</TableHead>
                    <TableHead className='min-w-40'>
                      {t('Minimum top-up')}
                    </TableHead>
                    <TableHead className='min-w-40'>
                      {t('Minimum consumption')}
                    </TableHead>
                    <TableHead className='min-w-36'>{t('Benefits')}</TableHead>
                    <TableHead className='w-36 text-right'>
                      {t('Actions')}
                    </TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {rules.map((rule, index) => (
                    <TableRow key={`${rule.group}-${index}`}>
                      <TableCell>
                        <Switch
                          checked={rule.enabled}
                          onCheckedChange={(enabled) =>
                            updateRule(index, { enabled })
                          }
                        />
                      </TableCell>
                      <TableCell className='whitespace-normal'>
                        <div className='grid gap-2'>
                          <Input
                            value={rule.display_name}
                            onChange={(event) =>
                              updateRule(index, {
                                display_name: event.target.value,
                              })
                            }
                            placeholder={t('Tier name')}
                          />
                          <Textarea
                            value={rule.description ?? ''}
                            onChange={(event) =>
                              updateRule(index, {
                                description: event.target.value,
                              })
                            }
                            placeholder={t('Benefit description')}
                            className='min-h-12 resize-y'
                          />
                        </div>
                      </TableCell>
                      <TableCell>
                        <NativeSelect
                          className='w-full'
                          value={rule.group}
                          onChange={(event) =>
                            updateRule(index, { group: event.target.value })
                          }
                        >
                          <NativeSelectOption value=''>
                            {t('Select group')}
                          </NativeSelectOption>
                          {groupOptions.map((group) => (
                            <NativeSelectOption key={group} value={group}>
                              {group}
                            </NativeSelectOption>
                          ))}
                        </NativeSelect>
                      </TableCell>
                      <TableCell>
                        <Input
                          type='number'
                          min={0}
                          step={1}
                          value={rule.min_topup_quota}
                          onChange={(event) =>
                            updateRule(index, {
                              min_topup_quota: Number(event.target.value),
                            })
                          }
                        />
                      </TableCell>
                      <TableCell>
                        <Input
                          type='number'
                          min={0}
                          step={1}
                          value={rule.min_used_quota}
                          onChange={(event) =>
                            updateRule(index, {
                              min_used_quota: Number(event.target.value),
                            })
                          }
                        />
                      </TableCell>
                      <TableCell className='whitespace-normal'>
                        {renderRatioSummary(rule.group)}
                      </TableCell>
                      <TableCell className='text-right'>
                        <div className='flex justify-end gap-1'>
                          <IconButton
                            label={t('Move up')}
                            icon={ArrowUp}
                            disabled={index === 0}
                            onClick={() => moveRule(index, -1)}
                          />
                          <IconButton
                            label={t('Move down')}
                            icon={ArrowDown}
                            disabled={index === rules.length - 1}
                            onClick={() => moveRule(index, 1)}
                          />
                          <IconButton
                            label={t('Delete')}
                            icon={Trash2}
                            destructive
                            onClick={() => removeRule(index)}
                          />
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>

            <div className='grid gap-3 md:hidden'>
              {rules.map((rule, index) => (
                <div key={`${rule.group}-${index}`} className='rounded-lg border p-3'>
                  <div className='mb-3 flex items-start justify-between gap-3'>
                    <div className='min-w-0'>
                      <div className='text-sm font-medium'>
                        {rule.display_name || rule.group || t('Untitled')}
                      </div>
                      <div className='text-muted-foreground text-xs'>
                        {renderConditionSummary(rule)}
                      </div>
                    </div>
                    <Switch
                      checked={rule.enabled}
                      onCheckedChange={(enabled) =>
                        updateRule(index, { enabled })
                      }
                    />
                  </div>
                  <div className='grid gap-3'>
                    <Input
                      value={rule.display_name}
                      onChange={(event) =>
                        updateRule(index, { display_name: event.target.value })
                      }
                      placeholder={t('Tier name')}
                    />
                    <NativeSelect
                      className='w-full'
                      value={rule.group}
                      onChange={(event) =>
                        updateRule(index, { group: event.target.value })
                      }
                    >
                      <NativeSelectOption value=''>
                        {t('Select group')}
                      </NativeSelectOption>
                      {groupOptions.map((group) => (
                        <NativeSelectOption key={group} value={group}>
                          {group}
                        </NativeSelectOption>
                      ))}
                    </NativeSelect>
                    <div className='grid gap-2 sm:grid-cols-2'>
                      <Input
                        type='number'
                        min={0}
                        step={1}
                        value={rule.min_topup_quota}
                        onChange={(event) =>
                          updateRule(index, {
                            min_topup_quota: Number(event.target.value),
                          })
                        }
                        aria-label={t('Minimum top-up')}
                      />
                      <Input
                        type='number'
                        min={0}
                        step={1}
                        value={rule.min_used_quota}
                        onChange={(event) =>
                          updateRule(index, {
                            min_used_quota: Number(event.target.value),
                          })
                        }
                        aria-label={t('Minimum consumption')}
                      />
                    </div>
                    <Textarea
                      value={rule.description ?? ''}
                      onChange={(event) =>
                        updateRule(index, { description: event.target.value })
                      }
                      placeholder={t('Benefit description')}
                    />
                    <div className='flex items-center justify-between gap-3'>
                      {renderRatioSummary(rule.group)}
                      <div className='flex gap-1'>
                        <IconButton
                          label={t('Move up')}
                          icon={ArrowUp}
                          disabled={index === 0}
                          onClick={() => moveRule(index, -1)}
                        />
                        <IconButton
                          label={t('Move down')}
                          icon={ArrowDown}
                          disabled={index === rules.length - 1}
                          onClick={() => moveRule(index, 1)}
                        />
                        <IconButton
                          label={t('Delete')}
                          icon={Trash2}
                          destructive
                          onClick={() => removeRule(index)}
                        />
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </>
        )}
      </SettingsSection>
    </TooltipProvider>
  )
}
