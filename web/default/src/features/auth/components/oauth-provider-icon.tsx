/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { useState, type ComponentType, type SVGProps } from 'react'
import {
  IconDiscord,
  IconDocker,
  IconFacebook,
  IconFigma,
  IconGithub,
  IconGitlab,
  IconGmail,
  IconLinuxDo,
  IconMedium,
  IconNotion,
  IconSkype,
  IconSlack,
  IconStripe,
  IconTelegram,
  IconTrello,
  IconWeChat,
  IconWhatsapp,
  IconZoom,
} from '@/assets/brand-icons'
import { cn } from '@/lib/utils'

type IconComponent = ComponentType<SVGProps<SVGSVGElement>>

const BRAND_ICON_MAP: Record<string, IconComponent> = {
  discord: IconDiscord,
  docker: IconDocker,
  facebook: IconFacebook,
  figma: IconFigma,
  github: IconGithub,
  'github-enterprise': IconGithub,
  gitlab: IconGitlab,
  gmail: IconGmail,
  google: IconGmail,
  linuxdo: IconLinuxDo,
  'linux-do': IconLinuxDo,
  medium: IconMedium,
  notion: IconNotion,
  skype: IconSkype,
  slack: IconSlack,
  stripe: IconStripe,
  telegram: IconTelegram,
  trello: IconTrello,
  wechat: IconWeChat,
  'we-chat': IconWeChat,
  whatsapp: IconWhatsapp,
  zoom: IconZoom,
}

type OAuthProviderIconProps = {
  icon?: string
  name?: string
  className?: string
}

function isHttpUrl(value: string) {
  return /^https?:\/\//i.test(value)
}

function normalizeIconKey(value: string) {
  return value.trim().toLowerCase().replace(/[\s_]+/g, '-')
}

function getFallbackText(icon: string, name?: string) {
  const source = (name || icon).trim()
  const chars = Array.from(source)
  if (chars.length === 0) return null
  if (chars.length <= 2 && !/^[a-z0-9-]+$/i.test(source)) return source
  return chars[0].toUpperCase()
}

export function OAuthProviderIcon({
  icon,
  name,
  className,
}: OAuthProviderIconProps) {
  const value = icon?.trim()
  const [failedUrl, setFailedUrl] = useState<string | null>(null)

  if (!value) return null

  if (isHttpUrl(value) && failedUrl !== value) {
    return (
      <img
        src={value}
        alt=''
        aria-hidden='true'
        referrerPolicy='no-referrer'
        loading='lazy'
        className={cn('shrink-0 object-contain', className)}
        onError={() => setFailedUrl(value)}
      />
    )
  }

  const Icon = BRAND_ICON_MAP[normalizeIconKey(value)]
  if (Icon) {
    return <Icon aria-hidden='true' className={cn('shrink-0', className)} />
  }

  const fallback = getFallbackText(value, name)
  if (!fallback) return null

  return (
    <span
      aria-hidden='true'
      className={cn(
        'inline-flex shrink-0 items-center justify-center rounded-sm text-xs font-semibold leading-none',
        className
      )}
    >
      {fallback}
    </span>
  )
}
