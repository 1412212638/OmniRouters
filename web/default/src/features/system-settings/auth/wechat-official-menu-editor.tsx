import { Plus, Trash2 } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'

type MenuButton = {
  id?: string
  name: string
  type?: 'click' | 'view'
  key?: string
  url?: string
  reply?: string
  sub_button?: MenuButton[]
}

type MenuConfig = { button: MenuButton[] }

type Props = {
  value: string
  onChange: (value: string) => void
  isDirty: boolean
  isSyncing: boolean
  onSync: () => void
}

const emptyConfig = (): MenuConfig => ({ button: [] })

let menuId = 0

function nextMenuId() {
  menuId += 1
  return `menu-${menuId}`
}

function parseMenu(value: string): MenuConfig {
  try {
    const parsed = JSON.parse(value) as Partial<MenuConfig>
    if (!Array.isArray(parsed.button)) return emptyConfig()
    return {
      button: parsed.button.map((button) => {
        const hasChildren =
          Array.isArray(button?.sub_button) && button.sub_button.length > 0
        const buttonType = button?.type === 'view' ? 'view' : 'click'
        const buttonKey = typeof button?.key === 'string' ? button.key : ''
        const buttonUrl = typeof button?.url === 'string' ? button.url : ''
        const buttonReply =
          typeof button?.reply === 'string' ? button.reply : ''
        return {
          id: nextMenuId(),
          name: typeof button?.name === 'string' ? button.name : '',
          type: hasChildren ? undefined : buttonType,
          key: hasChildren ? '' : buttonKey,
          url: hasChildren ? '' : buttonUrl,
          reply: hasChildren ? '' : buttonReply,
          sub_button: Array.isArray(button?.sub_button)
            ? button.sub_button.map((child) => ({
                id: nextMenuId(),
                name: typeof child?.name === 'string' ? child.name : '',
                type: child?.type === 'view' ? 'view' : 'click',
                key: typeof child?.key === 'string' ? child.key : '',
                url: typeof child?.url === 'string' ? child.url : '',
                reply: typeof child?.reply === 'string' ? child.reply : '',
              }))
            : [],
        }
      }),
    }
  } catch {
    return emptyConfig()
  }
}

function serializeMenu(config: MenuConfig) {
  return JSON.stringify({
    button: config.button.map(({ id: _id, sub_button, ...button }) => ({
      ...button,
      ...(sub_button?.length
        ? {
            sub_button: sub_button.map(({ id: _childId, ...child }) => child),
          }
        : {}),
    })),
  })
}

function makeButton(): MenuButton {
  return { id: nextMenuId(), name: '', type: 'click', key: '', reply: '' }
}

function makeChildButton(): MenuButton {
  return { id: nextMenuId(), name: '', type: 'click', key: '', reply: '' }
}

export function WeChatOfficialMenuEditor({
  value,
  onChange,
  isDirty,
  isSyncing,
  onSync,
}: Props) {
  const { t } = useTranslation()
  const [config, setConfig] = useState<MenuConfig>(() => parseMenu(value))
  const lastValueRef = useRef(value)

  useEffect(() => {
    if (value === lastValueRef.current) return
    lastValueRef.current = value
    setConfig(parseMenu(value))
  }, [value])

  const update = (next: MenuConfig) => {
    const serialized = serializeMenu(next)
    lastValueRef.current = serialized
    setConfig(next)
    onChange(serialized)
  }

  const updateRoot = (index: number, changes: Partial<MenuButton>) => {
    const button = config.button[index]
    if (!button) return
    const next = [...config.button]
    next[index] = { ...button, ...changes }
    update({ button: next })
  }

  const updateChild = (
    rootIndex: number,
    childIndex: number,
    changes: Partial<MenuButton>
  ) => {
    const root = config.button[rootIndex]
    const child = root?.sub_button?.[childIndex]
    if (!root || !child) return
    const next = [...config.button]
    const children = [...(root.sub_button ?? [])]
    children[childIndex] = { ...child, ...changes }
    next[rootIndex] = { ...root, sub_button: children }
    update({ button: next })
  }

  const addRoot = () => {
    if (config.button.length >= 3) return
    update({ button: [...config.button, makeButton()] })
  }

  const removeRoot = (index: number) => {
    update({
      button: config.button.filter((_, buttonIndex) => buttonIndex !== index),
    })
  }

  const toggleChildren = (index: number) => {
    const button = config.button[index]
    if (!button) return
    updateRoot(index, {
      type: undefined,
      key: undefined,
      url: undefined,
      reply: undefined,
      sub_button: button.sub_button?.length ? undefined : [makeChildButton()],
    })
  }

  const addChild = (rootIndex: number) => {
    const root = config.button[rootIndex]
    if (!root || (root.sub_button?.length ?? 0) >= 5) return
    updateRoot(rootIndex, {
      sub_button: [...(root.sub_button ?? []), makeChildButton()],
    })
  }

  const removeChild = (rootIndex: number, childIndex: number) => {
    const root = config.button[rootIndex]
    if (!root) return
    const children = (root.sub_button ?? []).filter(
      (_, index) => index !== childIndex
    )
    updateRoot(rootIndex, {
      sub_button: children.length ? children : undefined,
    })
  }

  const renderActionFields = (
    button: MenuButton,
    updateButton: (changes: Partial<MenuButton>) => void
  ) => (
    <div className='grid gap-3 sm:grid-cols-2'>
      <label className='space-y-1.5 text-sm'>
        <span className='text-muted-foreground'>{t('Menu action')}</span>
        <select
          className='border-input bg-background h-8 w-full rounded-lg border px-2 text-sm'
          value={button.type ?? 'click'}
          onChange={(event) =>
            updateButton({
              type: event.target.value as 'click' | 'view',
              key: '',
              url: '',
              reply: '',
            })
          }
        >
          <option value='click'>{t('Click reply')}</option>
          <option value='view'>{t('Open webpage')}</option>
        </select>
      </label>
      {button.type === 'view' ? (
        <label className='space-y-1.5 text-sm sm:col-span-1'>
          <span className='text-muted-foreground'>{t('Webpage URL')}</span>
          <Input
            value={button.url ?? ''}
            onChange={(event) => updateButton({ url: event.target.value })}
            placeholder='https://'
          />
        </label>
      ) : (
        <label className='space-y-1.5 text-sm sm:col-span-1'>
          <span className='text-muted-foreground'>{t('Menu key')}</span>
          <Input
            value={button.key ?? ''}
            onChange={(event) => updateButton({ key: event.target.value })}
            placeholder={t('Unique click key')}
          />
        </label>
      )}
      {button.type !== 'view' && (
        <label className='space-y-1.5 text-sm sm:col-span-2'>
          <span className='text-muted-foreground'>{t('Reply content')}</span>
          <Textarea
            value={button.reply ?? ''}
            onChange={(event) => updateButton({ reply: event.target.value })}
            placeholder={t('Reply sent after the menu is clicked')}
            rows={2}
          />
        </label>
      )}
    </div>
  )

  return (
    <div className='bg-muted/20 space-y-4 rounded-xl border p-4'>
      <div className='flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between'>
        <div className='space-y-1'>
          <p className='text-sm font-medium'>{t('Custom WeChat menu')}</p>
          <p className='text-muted-foreground text-xs'>
            {t('Save the menu locally, then sync it to WeChat when ready')}
          </p>
        </div>
        <Button
          type='button'
          variant='outline'
          size='sm'
          onClick={onSync}
          disabled={isSyncing || isDirty}
        >
          {isSyncing ? t('Syncing...') : t('Sync menu to WeChat')}
        </Button>
      </div>

      <div className='space-y-3'>
        {config.button.map((button, rootIndex) => (
          <div
            key={button.id}
            className='bg-background space-y-3 rounded-lg border p-3'
          >
            <div className='flex items-center gap-2'>
              <Input
                className='min-w-0 flex-1'
                value={button.name}
                onChange={(event) =>
                  updateRoot(rootIndex, { name: event.target.value })
                }
                placeholder={t('Menu name')}
              />
              <Button
                type='button'
                variant='ghost'
                size='icon-sm'
                onClick={() => removeRoot(rootIndex)}
                aria-label={t('Remove menu')}
                title={t('Remove menu')}
              >
                <Trash2 />
              </Button>
            </div>

            {button.sub_button?.length ? (
              <div className='space-y-3 border-l-2 pl-3'>
                {button.sub_button.map((child, childIndex) => (
                  <div
                    key={child.id}
                    className='bg-muted/20 space-y-3 rounded-lg border p-3'
                  >
                    <div className='flex items-center gap-2'>
                      <Input
                        className='min-w-0 flex-1'
                        value={child.name}
                        onChange={(event) =>
                          updateChild(rootIndex, childIndex, {
                            name: event.target.value,
                          })
                        }
                        placeholder={t('Menu name')}
                      />
                      <Button
                        type='button'
                        variant='ghost'
                        size='icon-sm'
                        onClick={() => removeChild(rootIndex, childIndex)}
                        aria-label={t('Remove menu')}
                        title={t('Remove menu')}
                      >
                        <Trash2 />
                      </Button>
                    </div>
                    {renderActionFields(child, (changes) =>
                      updateChild(rootIndex, childIndex, changes)
                    )}
                  </div>
                ))}
                <Button
                  type='button'
                  variant='outline'
                  size='sm'
                  onClick={() => addChild(rootIndex)}
                  disabled={button.sub_button.length >= 5}
                >
                  <Plus />
                  {t('Add submenu')}
                </Button>
              </div>
            ) : (
              <>
                {renderActionFields(button, (changes) =>
                  updateRoot(rootIndex, changes)
                )}
                <Button
                  type='button'
                  variant='ghost'
                  size='sm'
                  onClick={() => toggleChildren(rootIndex)}
                >
                  <Plus />
                  {t('Add submenu')}
                </Button>
              </>
            )}
          </div>
        ))}
      </div>

      <Button
        type='button'
        variant='outline'
        size='sm'
        onClick={addRoot}
        disabled={config.button.length >= 3}
      >
        <Plus />
        {t('Add menu')}
      </Button>
    </div>
  )
}
