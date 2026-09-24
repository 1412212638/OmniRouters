import {
  useCallback,
  useEffect,
  useRef,
  type IframeHTMLAttributes,
  type SyntheticEvent,
} from 'react'
import { useTranslation } from 'react-i18next'

import { useTheme } from '@/context/theme-provider'

import { dispatchEmbeddedContentScroll } from './external-content-events'

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
  const cleanupScrollBridgeRef = useRef<(() => void) | null>(null)
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

  const installSameOriginScrollBridge = useCallback(() => {
    cleanupScrollBridgeRef.current?.()
    cleanupScrollBridgeRef.current = null

    try {
      const frameWindow = iframeRef.current?.contentWindow
      const frameDocument = iframeRef.current?.contentDocument
      if (!frameWindow || !frameDocument) return

      const getScrollY = () => {
        const scrollingElement =
          frameDocument.scrollingElement as HTMLElement | null
        return Math.max(
          frameWindow.scrollY || 0,
          scrollingElement?.scrollTop || 0,
          frameDocument.documentElement?.scrollTop || 0,
          frameDocument.body?.scrollTop || 0
        )
      }

      const syncScroll = () => dispatchEmbeddedContentScroll(getScrollY())

      frameWindow.addEventListener('scroll', syncScroll, { passive: true })
      frameDocument.addEventListener('scroll', syncScroll, true)
      cleanupScrollBridgeRef.current = () => {
        frameWindow.removeEventListener('scroll', syncScroll)
        frameDocument.removeEventListener('scroll', syncScroll, true)
      }
      syncScroll()
    } catch {
      cleanupScrollBridgeRef.current = null
    }
  }, [])

  useEffect(() => {
    postFrameContext()
  }, [postFrameContext])

  useEffect(() => {
    return () => {
      cleanupScrollBridgeRef.current?.()
      dispatchEmbeddedContentScroll(0)
    }
  }, [])

  useEffect(() => {
    dispatchEmbeddedContentScroll(0)
  }, [props.src])

  useEffect(() => {
    const handleMessage = (event: MessageEvent) => {
      if (event.source !== iframeRef.current?.contentWindow) return

      const data = parseFrameMessage(event.data)
      if (!data) return

      const type = data?.type

      if (
        type === 'IFRAME_READY' ||
        type === 'IFRAME_REQUEST_THEME' ||
        type === 'OMNIROUTERS_IFRAME_READY' ||
        type === 'OMNIROUTERS_IFRAME_REQUEST_CONTEXT'
      ) {
        postFrameContext()
        return
      }

      if (type === 'OMNIROUTERS_IFRAME_SCROLL' || type === 'IFRAME_SCROLL') {
        const nextScrollY = Number(data.scrollY ?? data.y ?? 0)
        if (Number.isFinite(nextScrollY)) {
          dispatchEmbeddedContentScroll(nextScrollY)
        }
      }
    }

    window.addEventListener('message', handleMessage)
    return () => window.removeEventListener('message', handleMessage)
  }, [postFrameContext])

  const handleLoad = (event: SyntheticEvent<HTMLIFrameElement>) => {
    props.onLoad?.(event)
    dispatchEmbeddedContentScroll(0)
    postFrameContext()
    installSameOriginScrollBridge()
    window.setTimeout(postFrameContext, 80)
    window.setTimeout(installSameOriginScrollBridge, 80)
    window.setTimeout(postFrameContext, 300)
    window.setTimeout(installSameOriginScrollBridge, 300)
  }

  return <iframe {...props} ref={iframeRef} onLoad={handleLoad} />
}
