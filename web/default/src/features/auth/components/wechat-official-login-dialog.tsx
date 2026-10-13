import { useTranslation } from 'react-i18next'

import { Dialog } from '@/components/dialog'
import { Button } from '@/components/ui/button'

import { WeChatOfficialLoginView } from './wechat-official-login-view'

type WeChatOfficialLoginDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess: (data: unknown) => Promise<void> | void
}

export function WeChatOfficialLoginDialog({
  open,
  onOpenChange,
  onSuccess,
}: WeChatOfficialLoginDialogProps) {
  const { t } = useTranslation()

  return (
    <Dialog
      open={open}
      onOpenChange={onOpenChange}
      title={t('Official WeChat login')}
      contentClassName='max-w-sm'
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
      {open && (
        <WeChatOfficialLoginView
          onSuccess={async (data) => {
            await onSuccess(data)
            onOpenChange(false)
          }}
        />
      )}
    </Dialog>
  )
}
