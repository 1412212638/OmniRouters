import { Check, Copy, KeyRound, LoaderCircle, Pencil, Plus } from 'lucide-react'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { DateTimePicker } from '@/components/datetime-picker'
import { MultiSelect } from '@/components/multi-select'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import {
  ApiKeyGroupCombobox,
  type ApiKeyGroupOption,
} from '@/features/keys/components/api-key-group-combobox'
import { AutoGroupOrderEditor } from '@/features/keys/components/auto-group-order-editor'
import {
  getApiKeyFormSchema,
  transformFormDataToPayload,
  transformApiKeyToFormDefaults,
  type ApiKeyFormValues,
} from '@/features/keys/lib/api-key-form'
import type { ApiKey } from '@/features/keys/types'
import { getCurrencyDisplay, getCurrencyLabel } from '@/lib/currency'

import {
  createUserApiKey,
  getUserApiKeyEditorConfig,
  getUserApiKeys,
  updateUserApiKey,
} from '../../api'
import type { User } from '../../types'

type UserApiKey = {
  id: number
  name: string
  key: string
  status: number
  used_quota: number
  remain_quota: number
  unlimited_quota: boolean
  expired_time: number
  key_type: string
  group: string
  auto_groups: string[] | null
  cross_group_retry: boolean
  auto_route_strategy: string
  model_limits_enabled: boolean
  model_limits: string
  allow_ips: string | null
  allow_user_quota_query: boolean
}

type UserApiKeysDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  user: User
}

export function UserApiKeysDialog({
  open,
  onOpenChange,
  user,
}: UserApiKeysDialogProps) {
  const { t } = useTranslation()
  const [keys, setKeys] = useState<UserApiKey[]>([])
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [name, setName] = useState('')
  const [editingId, setEditingId] = useState<number | null>(null)
  const [editingKey, setEditingKey] = useState<UserApiKey | null>(null)
  const [formValues, setFormValues] = useState<ApiKeyFormValues | null>(null)
  const [editorConfig, setEditorConfig] =
    useState<Awaited<ReturnType<typeof getUserApiKeyEditorConfig>>['data']>()
  const [newKey, setNewKey] = useState('')

  const refresh = useCallback(async () => {
    setLoading(true)
    try {
      const result = await getUserApiKeys(user.id)
      if (result.success) setKeys(result.data?.items ?? [])
      else toast.error(result.message || t('Failed to load API keys'))
    } catch {
      toast.error(t('Failed to load API keys'))
    } finally {
      setLoading(false)
    }
  }, [t, user.id])

  useEffect(() => {
    if (!open) {
      setName('')
      setEditingId(null)
      setEditingKey(null)
      setFormValues(null)
      setEditorConfig(undefined)
      setNewKey('')
      return
    }
    void refresh()
    void getUserApiKeyEditorConfig(user.id)
      .then((result) => {
        if (result.success) setEditorConfig(result.data)
        else toast.error(result.message || t('Failed to load API key settings'))
      })
      .catch(() => toast.error(t('Failed to load API key settings')))
  }, [open, refresh, user.id, t])

  const groupOptions = useMemo<ApiKeyGroupOption[]>(
    () =>
      Object.entries(editorConfig?.groups ?? {}).map(([value, info]) => ({
        value,
        label: value,
        desc: info.desc || value,
        ratio: info.ratio,
      })),
    [editorConfig]
  )
  const globalAutoGroupOptions = useMemo(
    () =>
      groupOptions.filter((option) =>
        editorConfig?.auto_groups.includes(option.value)
      ),
    [editorConfig, groupOptions]
  )
  const { meta: currencyMeta } = getCurrencyDisplay()
  const currencyLabel = getCurrencyLabel()

  const save = async () => {
    const trimmedName = name.trim()
    if (!trimmedName) {
      toast.error(t('API key name is required'))
      return
    }
    setSaving(true)
    try {
      if (editingId === null) {
        const result = await createUserApiKey(user.id, trimmedName)
        if (!result.success || !result.data?.key) {
          toast.error(result.message || t('Failed to create API key'))
          return
        }
        setNewKey(result.data.key)
        toast.success(t('API key created'))
      } else {
        if (!editingKey || !formValues) return
        const parsedForm = getApiKeyFormSchema(
          t,
          editorConfig?.max_auto_groups
        ).safeParse({
          ...formValues,
          name: trimmedName,
        })
        if (!parsedForm.success) {
          toast.error(
            parsedForm.error.issues[0]?.message ||
              t('Please check the API key settings')
          )
          return
        }
        const payload = transformFormDataToPayload(parsedForm.data)
        const result = await updateUserApiKey(user.id, editingId, payload)
        if (!result.success) {
          toast.error(result.message || t('Failed to update API key'))
          return
        }
        toast.success(t('API key updated'))
        setEditingId(null)
        setEditingKey(null)
        setFormValues(null)
      }
      setName('')
      await refresh()
    } catch {
      toast.error(t('Unexpected error'))
    } finally {
      setSaving(false)
    }
  }

  const edit = (item: UserApiKey) => {
    setEditingId(item.id)
    setEditingKey(item)
    setName(item.name)
    if (editorConfig) {
      setFormValues(
        transformApiKeyToFormDefaults(
          item as ApiKey,
          editorConfig.auto_groups,
          editorConfig.max_auto_groups
        )
      )
    }
    setNewKey('')
  }

  const toggleStatus = async (item: UserApiKey) => {
    try {
      const result = await updateUserApiKey(user.id, item.id, {
        name: item.name,
        status: item.status === 1 ? 2 : 1,
      })
      if (!result.success) {
        toast.error(result.message || t('Failed to update API key'))
        return
      }
      await refresh()
    } catch {
      toast.error(t('Unexpected error'))
    }
  }

  const copyKey = async () => {
    try {
      await navigator.clipboard.writeText(newKey)
      toast.success(t('Copied to clipboard'))
    } catch {
      toast.error(t('Could not copy API key'))
    }
  }

  let saveIcon =
    editingId === null ? (
      <Plus data-icon='inline-start' />
    ) : (
      <Check data-icon='inline-start' />
    )
  if (saving) {
    saveIcon = (
      <LoaderCircle className='animate-spin' data-icon='inline-start' />
    )
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='max-h-[85vh] overflow-y-auto sm:max-w-2xl'>
        <DialogHeader>
          <DialogTitle>{t('Manage API keys')}</DialogTitle>
          <DialogDescription>
            {t('Manage API keys for {{username}}', { username: user.username })}
          </DialogDescription>
        </DialogHeader>

        {newKey && (
          <div className='flex flex-col gap-2 rounded-md border p-3'>
            <Label htmlFor='new-user-api-key'>
              {t('New API key. Copy it now.')}
            </Label>
            <div className='flex gap-2'>
              <Input id='new-user-api-key' value={newKey} readOnly />
              <Button
                type='button'
                variant='outline'
                onClick={() => void copyKey()}
                aria-label={t('Copy')}
              >
                <Copy data-icon='inline-start' />
                {t('Copy')}
              </Button>
            </div>
          </div>
        )}

        <div className='flex flex-col gap-2'>
          <Label htmlFor='user-api-key-name'>{t('API key name')}</Label>
          <div className='flex gap-2'>
            <Input
              id='user-api-key-name'
              value={name}
              onChange={(event) => setName(event.target.value)}
              maxLength={50}
              placeholder={t('Enter a name')}
            />
            <Button
              type='button'
              onClick={() => void save()}
              disabled={saving || loading}
            >
              {saveIcon}
              {editingId === null ? t('Create') : t('Save')}
            </Button>
            {editingId !== null && (
              <Button
                type='button'
                variant='outline'
                onClick={() => {
                  setEditingId(null)
                  setEditingKey(null)
                  setFormValues(null)
                  setName('')
                }}
              >
                {t('Cancel')}
              </Button>
            )}
          </div>
        </div>

        {editingId !== null && formValues && (
          <div className='grid gap-4 rounded-md border p-4 sm:grid-cols-2'>
            <div className='flex flex-col gap-2 sm:col-span-2'>
              <Label>{t('Group')}</Label>
              <ApiKeyGroupCombobox
                options={groupOptions}
                value={formValues.group}
                onValueChange={(group) =>
                  setFormValues(
                    (current) =>
                      current && {
                        ...current,
                        group,
                        auto_route_strategy:
                          group === 'auto' ? current.auto_route_strategy : '',
                        auto_groups_mode:
                          group === 'auto'
                            ? current.auto_groups_mode
                            : 'inherit',
                        auto_groups:
                          group === 'auto' ? current.auto_groups : [],
                        cross_group_retry: group === 'auto',
                      }
                  )
                }
                placeholder={t('Select a group')}
              />
            </div>

            {formValues.group === 'auto' && (
              <div className='flex flex-col gap-2 sm:col-span-2'>
                <Label>{t('Auto group order')}</Label>
                <AutoGroupOrderEditor
                  value={formValues.auto_groups}
                  mode={formValues.auto_groups_mode}
                  options={groupOptions}
                  globalOptions={globalAutoGroupOptions}
                  maxCount={editorConfig?.max_auto_groups ?? 5}
                  onChange={(value) =>
                    setFormValues(
                      (current) =>
                        current && {
                          ...current,
                          auto_groups: value.groups,
                          auto_groups_mode: value.mode,
                        }
                    )
                  }
                />
                <div className='flex items-center justify-between gap-3 rounded-md border p-3'>
                  <Label>{t('Cross-group retry')}</Label>
                  <Switch
                    checked={!!formValues.cross_group_retry}
                    onCheckedChange={(checked) =>
                      setFormValues(
                        (current) =>
                          current && { ...current, cross_group_retry: checked }
                      )
                    }
                  />
                </div>
                <div className='flex items-center justify-between gap-3 rounded-md border p-3'>
                  <Label>{t('Routing Strategy')}</Label>
                  <Switch
                    checked={formValues.auto_route_strategy === 'price'}
                    onCheckedChange={(checked) =>
                      setFormValues(
                        (current) =>
                          current && {
                            ...current,
                            auto_route_strategy: checked ? 'price' : '',
                          }
                      )
                    }
                  />
                </div>
              </div>
            )}

            <div className='flex flex-col gap-2 sm:col-span-2'>
              <Label>{t('Expiration Time')}</Label>
              <div className='flex gap-2'>
                <DateTimePicker
                  value={formValues.expired_time}
                  onChange={(expired_time) =>
                    setFormValues(
                      (current) => current && { ...current, expired_time }
                    )
                  }
                  placeholder={t('Never expires')}
                  className='min-w-0 flex-1'
                />
                <Button
                  type='button'
                  variant='outline'
                  onClick={() =>
                    setFormValues(
                      (current) =>
                        current && { ...current, expired_time: undefined }
                    )
                  }
                >
                  {t('Never')}
                </Button>
              </div>
            </div>

            <div className='flex flex-col gap-2 sm:col-span-2'>
              <div className='flex items-center justify-between gap-3 rounded-md border p-3'>
                <Label>{t('Unlimited quota')}</Label>
                <Switch
                  checked={formValues.unlimited_quota}
                  onCheckedChange={(unlimited_quota) =>
                    setFormValues(
                      (current) => current && { ...current, unlimited_quota }
                    )
                  }
                />
              </div>
              {!formValues.unlimited_quota && (
                <>
                  <Label>
                    {t('Quota ({{currency}})', { currency: currencyLabel })}
                  </Label>
                  <Input
                    type='number'
                    min={0}
                    step={currencyMeta.kind === 'tokens' ? 1 : 0.01}
                    value={formValues.remain_quota_dollars ?? ''}
                    onChange={(event) =>
                      setFormValues(
                        (current) =>
                          current && {
                            ...current,
                            remain_quota_dollars:
                              event.target.value === ''
                                ? undefined
                                : Number(event.target.value),
                          }
                      )
                    }
                    placeholder={t('Enter quota in {{currency}}', {
                      currency: currencyLabel,
                    })}
                  />
                </>
              )}
            </div>

            <div className='flex flex-col gap-2 sm:col-span-2'>
              <Label>{t('Model limits')}</Label>
              <MultiSelect
                options={(editorConfig?.models ?? []).map((model) => ({
                  label: model,
                  value: model,
                }))}
                selected={formValues.model_limits}
                onChange={(model_limits) =>
                  setFormValues(
                    (current) => current && { ...current, model_limits }
                  )
                }
                placeholder={t('Select models (empty for allow all)')}
              />
            </div>

            <div className='flex flex-col gap-2 sm:col-span-2'>
              <Label>{t('IP Whitelist')}</Label>
              <Textarea
                value={formValues.allow_ips}
                onChange={(event) =>
                  setFormValues(
                    (current) =>
                      current && { ...current, allow_ips: event.target.value }
                  )
                }
                placeholder={t('Enter IP addresses, one per line')}
              />
            </div>

            <div className='flex items-center justify-between gap-3 rounded-md border p-3 sm:col-span-2'>
              <Label>{t('Allow querying owner quota')}</Label>
              <Switch
                checked={formValues.allow_user_quota_query}
                onCheckedChange={(allow_user_quota_query) =>
                  setFormValues(
                    (current) =>
                      current && { ...current, allow_user_quota_query }
                  )
                }
              />
            </div>
          </div>
        )}

        <div className='flex flex-col gap-2'>
          <div className='text-sm font-medium'>{t('API keys')}</div>
          {loading && (
            <div className='text-muted-foreground flex items-center justify-center py-8'>
              <LoaderCircle className='animate-spin' />
            </div>
          )}
          {!loading && keys.length === 0 && (
            <div className='text-muted-foreground rounded-md border border-dashed p-6 text-center text-sm'>
              {t('No API keys found')}
            </div>
          )}
          {!loading && keys.length > 0 && (
            <div className='flex max-h-72 flex-col gap-2 overflow-y-auto'>
              {keys.map((item) => (
                <div
                  key={item.id}
                  className='flex items-center gap-3 rounded-md border p-3'
                >
                  <KeyRound className='text-muted-foreground shrink-0' />
                  <div className='min-w-0 flex-1'>
                    <div className='truncate text-sm font-medium'>
                      {item.name}
                    </div>
                    <div className='text-muted-foreground truncate font-mono text-xs'>
                      {item.key}
                    </div>
                    <div className='text-muted-foreground text-xs'>
                      {item.key_type === 'general' || !item.key_type
                        ? t(item.status === 1 ? 'Enabled' : 'Disabled')
                        : t(
                            item.key_type === 'campaign'
                              ? 'Campaign key'
                              : 'Subscription key'
                          )}
                    </div>
                  </div>
                  {(item.key_type === 'general' || !item.key_type) && (
                    <>
                      <Button
                        type='button'
                        size='sm'
                        variant='outline'
                        onClick={() => edit(item)}
                        disabled={!editorConfig}
                        aria-label={t('Edit')}
                      >
                        <Pencil data-icon='inline-start' />
                        {t('Edit')}
                      </Button>
                      <Button
                        type='button'
                        size='sm'
                        variant='ghost'
                        onClick={() => void toggleStatus(item)}
                      >
                        {t(item.status === 1 ? 'Disable' : 'Enable')}
                      </Button>
                    </>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
