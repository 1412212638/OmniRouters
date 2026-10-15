import axios from 'axios'
import { QrCode } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Dialog } from '@/components/dialog'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import {
  bindWeChatOfficial,
  createWeChatOfficialBindQr,
  pollWeChatOfficialBind,
} from '@/features/auth/api'

interface WeChatBindDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess: () => void
  officialEnabled?: boolean
}

export function WeChatBindDialog({
  open,
  onOpenChange,
  onSuccess,
  officialEnabled = false,
}: WeChatBindDialogProps) {
  const { t } = useTranslation()
  const [imageUrl, setImageUrl] = useState('')
  const [statusMessage, setStatusMessage] = useState('')
  const [loading, setLoading] = useState(false)
  const onSuccessRef = useRef(onSuccess)

  useEffect(() => {
    onSuccessRef.current = onSuccess
  }, [onSuccess])

  useEffect(() => {
    if (!open || !officialEnabled) return
    let cancelled = false
    let timer: ReturnType<typeof setTimeout> | undefined

    const load = async () => {
      setLoading(true)
      setStatusMessage('')
      try {
        const response = await createWeChatOfficialBindQr()
        if (cancelled) return
        if (!response.success || !response.data) {
          throw new Error(
            response.message || t('Failed to create WeChat QR code')
          )
        }
        setImageUrl(response.data.image_url)
        const code = response.data.code

        const poll = async () => {
          if (cancelled) return
          try {
            const response = await pollWeChatOfficialBind(code)
            if (cancelled) return
            if (!response.success) throw new Error(response.message)
            if (response.data?.status === 'consumed') {
              throw new Error(t('This QR code has already been used'))
            }
            if (response.data?.status === 'ready') {
              const result = await bindWeChatOfficial(code)
              if (cancelled) return
              if (!result.success) {
                throw new Error(result.message || t('Request failed'))
              }
              toast.success(t('Binding successful!'))
              onSuccessRef.current()
              onOpenChange(false)
              return
            }
            timer = setTimeout(poll, 1800)
          } catch (error) {
            if (axios.isAxiosError(error) && error.response?.status === 429) {
              timer = setTimeout(poll, 5000)
              return
            }
            if (!cancelled) {
              const message =
                error instanceof Error ? error.message : t('Request failed')
              setStatusMessage(message)
              toast.error(message)
            }
          }
        }
        timer = setTimeout(poll, 1200)
      } catch (error) {
        if (!cancelled) {
          const message =
            error instanceof Error ? error.message : t('Request failed')
          setStatusMessage(message)
          toast.error(message)
        }
      } finally {
        if (!cancelled) setLoading(false)
      }
    }

    void load()
    return () => {
      cancelled = true
      if (timer) clearTimeout(timer)
    }
  }, [officialEnabled, open, onOpenChange, t])

  if (!officialEnabled) {
    return (
      <Dialog
        open={open}
        onOpenChange={onOpenChange}
        title={t('Bind WeChat Account')}
        description={t('Scan the QR code with WeChat to bind your account')}
        contentClassName='sm:max-w-md'
        contentHeight='auto'
        bodyClassName='space-y-4'
      >
        <div className='space-y-4 py-4'>
          <Alert>
            <QrCode className='h-4 w-4' />
            <AlertDescription>
              {t('This feature requires server-side WeChat configuration')}
            </AlertDescription>
          </Alert>
        </div>
      </Dialog>
    )
  }

  return (
    <Dialog
      open={open}
      onOpenChange={onOpenChange}
      title={t('Bind WeChat Account')}
      description={t('Scan the QR code with WeChat to bind your account')}
      contentClassName='sm:max-w-md'
      contentHeight='auto'
      bodyClassName='space-y-4'
      footer={
        <Button
          type='button'
          variant='outline'
          onClick={() => onOpenChange(false)}
        >
          {t('Cancel')}
        </Button>
      }
    >
      <div className='space-y-4 py-4'>
        <Alert>
          <QrCode className='h-4 w-4' />
          <AlertDescription>
            {t('After scanning, the binding will complete automatically')}
          </AlertDescription>
        </Alert>
        <div className='bg-background mx-auto flex size-52 items-center justify-center rounded-lg border'>
          {imageUrl ? (
            <img
              src={imageUrl}
              alt={t('WeChat login QR code')}
              className='size-48 object-contain'
            />
          ) : (
            <p className='text-muted-foreground px-4 text-center text-sm'>
              {loading
                ? t('Creating QR code...')
                : t('QR code is not available')}
            </p>
          )}
        </div>
        <p
          aria-live='polite'
          className='text-muted-foreground text-center text-sm'
        >
          {statusMessage || t('Waiting for WeChat scan...')}
        </p>
      </div>
    </Dialog>
  )
}
