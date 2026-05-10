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
