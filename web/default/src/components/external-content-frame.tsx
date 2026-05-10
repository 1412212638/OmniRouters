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
import {
  useCallback,
  useEffect,
  useRef,
  type IframeHTMLAttributes,
  type SyntheticEvent,
} from 'react'
import { useTranslation } from 'react-i18next'
import { useTheme } from '@/context/theme-provider'

function parseFrameMessage(data: unknown): Record<string, unknown> | null {
  if (typeof data === 'string') {
    try {
      const parsed = JSON.parse(data)
      return parsed && typeof parsed === 'object'
        ? (parsed as Record<string, unknown>)
        : null
    } catch {
      return null
    }
  }

  return data && typeof data === 'object'
    ? (data as Record<string, unknown>)
    : null
}

export function ExternalContentFrame(
  props: IframeHTMLAttributes<HTMLIFrameElement>
) {
  const iframeRef = useRef<HTMLIFrameElement>(null)
  const { i18n } = useTranslation()
  const { resolvedTheme } = useTheme()

  const postFrameContext = useCallback(() => {
    const target = iframeRef.current?.contentWindow
    if (!target) return

    target.postMessage(
      {
        type: 'OMNIROUTERS_FRAME_CONTEXT',
        themeMode: resolvedTheme,
        lang: i18n.language,
      },
      '*'
    )
  }, [i18n.language, resolvedTheme])

  useEffect(() => {
    postFrameContext()
  }, [postFrameContext])

  useEffect(() => {
    const handleMessage = (event: MessageEvent) => {
      const data = parseFrameMessage(event.data)
      const type = data?.type
      if (
        type === 'IFRAME_READY' ||
        type === 'IFRAME_REQUEST_THEME' ||
        type === 'OMNIROUTERS_IFRAME_READY' ||
        type === 'OMNIROUTERS_IFRAME_REQUEST_CONTEXT'
      ) {
        postFrameContext()
      }
    }

    window.addEventListener('message', handleMessage)
    return () => window.removeEventListener('message', handleMessage)
  }, [postFrameContext])

  const handleLoad = (event: SyntheticEvent<HTMLIFrameElement>) => {
    props.onLoad?.(event)
    postFrameContext()
    window.setTimeout(postFrameContext, 80)
    window.setTimeout(postFrameContext, 300)
  }

  return <iframe {...props} ref={iframeRef} onLoad={handleLoad} />
}
