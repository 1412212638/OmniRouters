import { Send } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Dialog } from '@/components/dialog'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { startTelegramOAuth } from '@/features/auth/api'

// ============================================================================
// Telegram Bind Dialog Component
// ============================================================================

interface TelegramBindDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  botName: string
  onSuccess: () => void
}

export function TelegramBindDialog({
  open,
  onOpenChange,
  botName,
  onSuccess: _onSuccess,
}: TelegramBindDialogProps) {
  const { t } = useTranslation()
  return (
    <Dialog
      open={open}
      onOpenChange={onOpenChange}
      title={t('Bind Telegram Account')}
      description={t('Click the button below to bind your Telegram account')}
      contentClassName='sm:max-w-md'
      contentHeight='auto'
      bodyClassName='space-y-4'
    >
      <div className='space-y-4 py-4'>
        <Alert>
          <Send className='h-4 w-4' />
          <AlertDescription>
            {t(
              'You will be redirected to Telegram to complete the binding process.'
            )}
          </AlertDescription>
        </Alert>

        <div className='flex flex-col items-center justify-center gap-4 rounded-lg border p-6'>
          <div className='flex h-12 w-12 items-center justify-center rounded-xl bg-blue-100 dark:bg-blue-900'>
            <Send className='h-6 w-6 text-blue-600 dark:text-blue-400' />
          </div>

          <div className='text-center'>
            <p className='text-muted-foreground text-sm'>
              {t('Bot:')}{' '}
              <span className='font-mono font-semibold'>@{botName}</span>
            </p>
            <p className='text-muted-foreground mt-1 text-xs'>
              {t(
                "After clicking the button, you'll be asked to authorize the bot"
              )}
            </p>
          </div>

          <Button
            type='button'
            onClick={async () => {
              const result = await startTelegramOAuth('bind')
              const url = result.data?.authorization_url
              if (result.success && url) window.location.assign(url)
            }}
          >
            <Send className='mr-2 h-4 w-4' />
            {t('Continue with Telegram')}
          </Button>
        </div>

        <p className='text-muted-foreground text-center text-xs'>
          {t('The binding will complete automatically after authorization')}
        </p>
      </div>
    </Dialog>
  )
}
