import { useEffect, useMemo, useState } from 'react'
import * as z from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Plus, Save } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { StaticDataTable } from '@/components/data-table/static/static-data-table'
import { StaticRowActions } from '@/components/data-table/static/static-row-actions'
import { Dialog } from '@/components/dialog'
import { SettingsSection } from '../components/settings-section'
import { useUpdateOption } from '../hooks/use-update-option'

type CarouselItem = { id: number; title: string; description: string; image: string; enabled: boolean; sort: number }
const schema = z.object({ title: z.string().min(1).max(200), description: z.string().min(1).max(5000), image: z.string().min(1).max(5000), enabled: z.boolean(), sort: z.coerce.number().int().min(0).max(9999) })
type Values = z.infer<typeof schema>
type FormInput = z.input<typeof schema>

export function ModelSquareCarouselSection({ data }: { data: string }) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()
  const [items, setItems] = useState<CarouselItem[]>([])
  const [editing, setEditing] = useState<CarouselItem | null>(null)
  const [open, setOpen] = useState(false)
  const [dirty, setDirty] = useState(false)
  const form = useForm<FormInput, unknown, Values>({ resolver: zodResolver(schema), defaultValues: { title: '', description: '', image: '', enabled: true, sort: 0 } })

  useEffect(() => { try { const parsed = JSON.parse(data || '[]'); setItems(Array.isArray(parsed) ? parsed.map((item, index) => ({ ...item, id: item.id || index + 1, enabled: item.enabled !== false, sort: Number(item.sort) || 0 })) : []) } catch { setItems([]) } }, [data])
  const sorted = useMemo(() => [...items].sort((a, b) => a.sort - b.sort || a.id - b.id), [items])
  const add = () => { setEditing(null); form.reset({ title: '', description: '', image: '', enabled: true, sort: items.length }); setOpen(true) }
  const edit = (item: CarouselItem) => { setEditing(item); form.reset(item); setOpen(true) }
  const submit = (values: Values) => { setItems((current) => editing ? current.map((item) => item.id === editing.id ? { ...item, ...values } : item) : [...current, { id: Math.max(0, ...current.map((item) => item.id)) + 1, ...values }]); setDirty(true); setOpen(false) }
  const remove = (item: CarouselItem) => { setItems((current) => current.filter((entry) => entry.id !== item.id)); setDirty(true) }
  const save = async () => { try { await updateOption.mutateAsync({ key: 'console_setting.model_square_carousel', value: JSON.stringify(items) }); setDirty(false); toast.success(t('Model plaza carousel saved successfully')) } catch { toast.error(t('Failed to save model plaza carousel')) } }

  return <SettingsSection title={t('Model Plaza Carousel')} description={t('Manage the fixed promotional carousel shown first in the model plaza.') }>
    <div className='flex flex-wrap gap-2'><Button size='sm' onClick={add}><Plus className='mr-2 size-4' />{t('Add Carousel Item')}</Button><Button size='sm' variant='secondary' disabled={!dirty || updateOption.isPending} onClick={save}><Save className='mr-2 size-4' />{t('Save Settings')}</Button></div>
    <StaticDataTable data={sorted} getRowKey={(item) => item.id} emptyContent={t('No carousel items yet.')} columns={[{ id: 'title', header: t('Title'), cell: (item) => item.title }, { id: 'image', header: t('Image'), cellClassName: 'max-w-xs truncate', cell: (item) => item.image }, { id: 'sort', header: t('Sort'), cell: (item) => item.sort }, { id: 'enabled', header: t('Enabled'), cell: (item) => item.enabled ? t('Yes') : t('No') }, { id: 'actions', header: t('Actions'), cell: (item) => <StaticRowActions editLabel={t('Edit')} deleteLabel={t('Delete')} menuLabel={t('Open menu')} onEdit={() => edit(item)} onDelete={() => remove(item)} /> }]} />
    <Dialog open={open} onOpenChange={setOpen} title={editing ? t('Edit Carousel Item') : t('Add Carousel Item')} contentClassName='max-w-2xl' footer={<><Button variant='outline' onClick={() => setOpen(false)}>{t('Cancel')}</Button><Button type='submit' form='model-square-carousel-form'>{editing ? t('Update') : t('Add')}</Button></>}>
    <Form {...form}><form id='model-square-carousel-form' onSubmit={form.handleSubmit(submit)} className='space-y-4'>{(['title', 'description', 'image'] as const).map((name) => <FormField key={name} control={form.control} name={name} render={({ field }) => <FormItem><FormLabel>{t(name === 'title' ? 'Title' : name === 'description' ? 'Description' : 'Image URL')}</FormLabel><FormControl>{name === 'description' ? <Textarea rows={4} {...field} /> : <Input placeholder={name === 'image' ? 'https://...' : undefined} {...field} />}</FormControl><FormMessage /></FormItem>} />)}<FormField control={form.control} name='sort' render={({ field }) => <FormItem><FormLabel>{t('Sort')}</FormLabel><FormControl><Input type='number' value={field.value as number} onChange={(event) => field.onChange(event.target.value)} onBlur={field.onBlur} name={field.name} ref={field.ref} /></FormControl><FormDescription>{t('Lower values are displayed first.')}</FormDescription><FormMessage /></FormItem>} /><FormField control={form.control} name='enabled' render={({ field }) => <FormItem className='flex items-center gap-2'><FormControl><Checkbox checked={field.value} onCheckedChange={field.onChange} /></FormControl><FormLabel>{t('Enabled')}</FormLabel></FormItem>} /></form></Form>
    </Dialog>
  </SettingsSection>
}
