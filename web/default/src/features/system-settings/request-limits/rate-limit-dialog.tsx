import { zodResolver } from '@hookform/resolvers/zod'
import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import * as z from 'zod'

import { Dialog } from '@/components/dialog'
import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'

const rateLimitDialogSchema = z.object({
  groupName: z.string().min(1, 'Group name is required'),
  rpm: z
    .number()
    .min(0, 'Must be ≥ 0')
    .max(2147483647, 'Must be ≤ 2,147,483,647'),
  tpm: z
    .number()
    .min(0, 'Must be ≥ 0')
    .max(2147483647, 'Must be ≤ 2,147,483,647'),
  rpd: z
    .number()
    .min(0, 'Must be ≥ 0')
    .max(2147483647, 'Must be ≤ 2,147,483,647'),
})

type RateLimitDialogFormValues = z.infer<typeof rateLimitDialogSchema>

const RATE_LIMIT_FORM_ID = 'rate-limit-form'

export type RateLimitEntryData = {
  groupName: string
  rpm: number
  tpm: number
  rpd: number
}

type RateLimitDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSave: (data: RateLimitEntryData) => void
  editData?: RateLimitEntryData | null
}

export function RateLimitDialog({
  open,
  onOpenChange,
  onSave,
  editData,
}: RateLimitDialogProps) {
  const { t } = useTranslation()
  const isEditMode = !!editData

  const form = useForm<RateLimitDialogFormValues>({
    resolver: zodResolver(rateLimitDialogSchema),
    defaultValues: {
      groupName: '',
      rpm: 0,
      tpm: 0,
      rpd: 0,
    },
  })

  useEffect(() => {
    if (editData) {
      form.reset(editData)
    } else {
      form.reset({
        groupName: '',
        rpm: 0,
        tpm: 0,
        rpd: 0,
      })
    }
  }, [editData, form, open])

  const handleSubmit = (values: RateLimitDialogFormValues) => {
    onSave(values)
    form.reset()
    onOpenChange(false)
  }

  return (
    <Dialog
      open={open}
      onOpenChange={onOpenChange}
      title={
        isEditMode ? t('Edit group rate limit') : t('Add group rate limit')
      }
      description={t(
        'Configure rate limiting rules for a specific user group.'
      )}
      contentClassName='sm:max-w-[500px]'
      contentHeight='auto'
      bodyClassName='space-y-4'
      footer={
        <>
          <Button
            type='button'
            variant='outline'
            onClick={() => onOpenChange(false)}
          >
            {t('Cancel')}
          </Button>
          <Button type='submit' form={RATE_LIMIT_FORM_ID}>
            {isEditMode ? t('Update') : t('Add')}
          </Button>
        </>
      }
    >
      <Form {...form}>
        <form
          id={RATE_LIMIT_FORM_ID}
          onSubmit={form.handleSubmit(handleSubmit)}
          className='space-y-4'
        >
          <FormField
            control={form.control}
            name='groupName'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Group Name')}</FormLabel>
                <FormControl>
                  <Input
                    placeholder={t('e.g., default, vip, premium')}
                    {...field}
                    disabled={isEditMode}
                  />
                </FormControl>
                <FormDescription>
                  {isEditMode
                    ? t('Group name cannot be changed when editing.')
                    : t('Unique identifier for this group.')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name='rpm'
            render={({ field }) => (
              <FormItem>
                <FormLabel>RPM</FormLabel>
                <FormControl>
                  <div className='flex items-center gap-2'>
                    <Input
                      type='number'
                      min={0}
                      max={2147483647}
                      step={1}
                      {...field}
                      onChange={(e) =>
                        field.onChange(parseInt(e.target.value) || 0)
                      }
                    />
                    <span className='text-muted-foreground text-sm'>
                      {t('requests/minute')}
                    </span>
                  </div>
                </FormControl>
                <FormDescription>
                  {t('Requests per minute. 0 = unlimited.')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name='tpm'
            render={({ field }) => (
              <FormItem>
                <FormLabel>TPM</FormLabel>
                <FormControl>
                  <div className='flex items-center gap-2'>
                    <Input
                      type='number'
                      min={0}
                      max={2147483647}
                      step={1}
                      {...field}
                      onChange={(e) =>
                        field.onChange(parseInt(e.target.value) || 0)
                      }
                    />
                    <span className='text-muted-foreground text-sm'>
                      {t('tokens/minute')}
                    </span>
                  </div>
                </FormControl>
                <FormDescription>
                  {t('Tokens per minute. 0 = unlimited.')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name='rpd'
            render={({ field }) => (
              <FormItem>
                <FormLabel>RPD</FormLabel>
                <FormControl>
                  <div className='flex items-center gap-2'>
                    <Input
                      type='number'
                      min={0}
                      max={2147483647}
                      step={1}
                      {...field}
                      onChange={(e) =>
                        field.onChange(parseInt(e.target.value) || 0)
                      }
                    />
                    <span className='text-muted-foreground text-sm'>
                      {t('requests/day')}
                    </span>
                  </div>
                </FormControl>
                <FormDescription>
                  {t('Requests per day. 0 = unlimited.')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
        </form>
      </Form>
    </Dialog>
  )
}
