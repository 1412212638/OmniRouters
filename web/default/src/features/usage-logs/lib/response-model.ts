import type { ResponseModelInfo } from '../types'

// Upstream names are diagnostic only. Keep this comparison in sync with
// relay/common/response_model.go; do not use it to select billing or routing.
export function hasResponseModelMismatch(info?: ResponseModelInfo): boolean {
  if (!info?.returned_model) return false
  const returned = info.returned_model.toLowerCase()
  return ![info.requested_model, info.upstream_model].some((name) => {
    const expected = name.toLowerCase()
    return (
      expected !== '' &&
      (returned.startsWith(expected) || returned.endsWith(expected))
    )
  })
}
