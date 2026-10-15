import axios from 'axios'
import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import {
  createWeChatOfficialQr,
  loginWithWeChatOfficial,
  pollWeChatOfficialLogin,
} from '@/features/auth/api'
import { getAffiliateCode } from '@/features/auth/lib/storage'

type WeChatOfficialLoginViewProps = {
  onSuccess: (data: unknown) => Promise<void> | void
}

export function WeChatOfficialLoginView({
  onSuccess,
}: WeChatOfficialLoginViewProps) {
  const { t } = useTranslation()
  const [imageUrl, setImageUrl] = useState('')
  const [loading, setLoading] = useState(true)
  const [errorMessage, setErrorMessage] = useState('')
  const onSuccessRef = useRef(onSuccess)

  useEffect(() => {
    onSuccessRef.current = onSuccess
  }, [onSuccess])

  useEffect(() => {
    let cancelled = false
    let timer: ReturnType<typeof setTimeout> | undefined

    const loadQr = async () => {
      setLoading(true)
      try {
        const response = await createWeChatOfficialQr()
        if (cancelled) return
        if (!response.success || !response.data) {
          throw new Error(
            response.message || t('Failed to create WeChat QR code')
          )
        }
        setImageUrl(response.data.image_url)
        setErrorMessage('')

        const qrCode = response.data.code
        const poll = async () => {
          if (cancelled) return
          try {
            const status = await pollWeChatOfficialLogin(qrCode)
            if (cancelled) return
            if (!status.success) {
              throw new Error(status.message || t('QR code has expired'))
            }
            if (status.data?.status === 'consumed') {
              throw new Error(t('This QR code has already been used'))
            }
            if (status.data?.status === 'ready') {
              const login = await loginWithWeChatOfficial(
                qrCode,
                getAffiliateCode()
              )
              if (cancelled) return
              if (!login.success) {
                throw new Error(login.message || t('Login failed'))
              }
              await onSuccessRef.current(login.data)
              if (!cancelled) toast.success(t('Signed in via WeChat'))
              return
            }
            timer = setTimeout(poll, 1800)
          } catch (error) {
            if (axios.isAxiosError(error) && error.response?.status === 429) {
              const retryAfter = Number(error.response.headers?.['retry-after'])
              const delay =
                Number.isFinite(retryAfter) && retryAfter > 0
                  ? retryAfter * 1000
                  : 5000
              timer = setTimeout(poll, delay)
              return
            }
            if (!cancelled) {
              const message =
                error instanceof Error ? error.message : t('Login failed')
              setErrorMessage(message)
              toast.error(message)
            }
          }
        }
        timer = setTimeout(poll, 1200)
      } catch (error) {
        if (!cancelled) {
          const message =
            error instanceof Error ? error.message : t('Login failed')
          setErrorMessage(message)
          toast.error(message)
        }
      } finally {
        if (!cancelled) setLoading(false)
      }
    }

    void loadQr()
    return () => {
      cancelled = true
      if (timer) clearTimeout(timer)
    }
  }, [t])

  return (
    <div className='grid justify-items-center gap-4 text-center'>
      <p className='text-muted-foreground text-sm'>
        {t('Scan the QR code with WeChat to continue.')}
      </p>
      <div className='bg-background flex size-52 items-center justify-center rounded-lg border'>
        {imageUrl ? (
          <img
            src={imageUrl}
            alt={t('Official WeChat login QR code')}
            className='size-48 object-contain'
          />
        ) : (
          <p className='text-muted-foreground px-4 text-sm'>
            {loading ? t('Creating QR code...') : t('QR code is not available')}
          </p>
        )}
      </div>
      <p aria-live='polite' className='text-muted-foreground text-sm'>
        {errorMessage || t('Waiting for WeChat scan...')}
      </p>
    </div>
  )
}
