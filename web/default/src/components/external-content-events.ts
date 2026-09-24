export const EMBEDDED_CONTENT_SCROLL_EVENT = 'omnirouters:embedded-scroll'

export type EmbeddedContentScrollEventDetail = {
  scrollY: number
}

export function dispatchEmbeddedContentScroll(scrollY: number) {
  window.dispatchEvent(
    new CustomEvent<EmbeddedContentScrollEventDetail>(
      EMBEDDED_CONTENT_SCROLL_EVENT,
      {
        detail: {
          scrollY: Math.max(0, scrollY),
        },
      }
    )
  )
}
