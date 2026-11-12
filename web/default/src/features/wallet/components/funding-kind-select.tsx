import { useTranslation } from 'react-i18next'

import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

import { fundingKinds, type FundingKind } from '../funding'

export function FundingKindSelect({
  value,
  onChange,
  disabled,
  allowUnclassified,
  id,
}: {
  value: FundingKind
  onChange: (value: FundingKind) => void
  disabled?: boolean
  allowUnclassified?: boolean
  id?: string
}) {
  const { t } = useTranslation()
  const labels = {
    paid: t('Paid credit'),
    gift: t('Gift credit'),
    adjustment: t('Balance adjustment'),
    unclassified: t('Unclassified credit'),
  }
  return (
    <Select
      value={value}
      onValueChange={(next) => {
        if (next) onChange(next)
      }}
      disabled={disabled}
    >
      <SelectTrigger id={id} className='w-full'>
        <SelectValue>{labels[value]}</SelectValue>
      </SelectTrigger>
      <SelectContent>
        <SelectGroup>
          {fundingKinds
            .filter((kind) => allowUnclassified || kind !== 'unclassified')
            .map((kind) => (
              <SelectItem key={kind} value={kind}>
                {labels[kind]}
              </SelectItem>
            ))}
        </SelectGroup>
      </SelectContent>
    </Select>
  )
}
