import { ExternalLink, Gift, Loader2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

interface RedemptionCodeCardProps {
  redemptionCode: string
  onRedemptionCodeChange: (code: string) => void
  onRedeem: () => void
  redeeming: boolean
  topupLink?: string
}

export function RedemptionCodeCard({
  redemptionCode,
  onRedemptionCodeChange,
  onRedeem,
  redeeming,
  topupLink,
}: RedemptionCodeCardProps) {
  const { t } = useTranslation()

  return (
    <Card data-card-hover='false'>
      <CardContent className='space-y-3 p-4 sm:p-6'>
        <div className='flex items-center gap-2'>
          <Gift className='text-muted-foreground h-4 w-4' />
          <Label
            htmlFor='redemption-code'
            className='text-base font-semibold'
          >
            {t('Have a Code?')}
          </Label>
        </div>
        <div className='grid grid-cols-[minmax(0,1fr)_auto] gap-2'>
          <Input
            id='redemption-code'
            value={redemptionCode}
            onChange={(event) => onRedemptionCodeChange(event.target.value)}
            placeholder={t('Enter your redemption code')}
            className='h-10 min-w-0'
          />
          <Button
            type='button'
            onClick={onRedeem}
            disabled={redeeming}
            variant='outline'
            className='h-10 px-4'
          >
            {redeeming && <Loader2 className='mr-2 h-4 w-4 animate-spin' />}
            {t('Redeem')}
          </Button>
        </div>
        {topupLink && (
          <p className='text-muted-foreground text-xs'>
            {t('Need a redemption code?')}{' '}
            <a
              href={topupLink}
              target='_blank'
              rel='noopener noreferrer'
              className='inline-flex items-center gap-1 underline-offset-4 hover:underline'
            >
              {t('Get one here')}
              <ExternalLink className='h-3 w-3' />
            </a>
          </p>
        )}
      </CardContent>
    </Card>
  )
}
