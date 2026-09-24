import { getLobeIcon } from '@/lib/lobe-icon'
import { cn } from '@/lib/utils'

import {
  resolvePluginIcon,
  textAvatarClass,
  type PluginIconInput,
} from '../lib/plugin-icon'

type PluginIconProps = {
  plugin: PluginIconInput
  size?: number
}

export function PluginIcon(props: PluginIconProps) {
  const size = props.size ?? 20
  const descriptor = resolvePluginIcon(props.plugin)
  if (descriptor.kind === 'lobe') {
    return <>{getLobeIcon(descriptor.name, size)}</>
  }
  return (
    <div
      aria-hidden='true'
      className={cn(
        'flex items-center justify-center rounded-md font-semibold select-none',
        textAvatarClass(descriptor.colorSeed)
      )}
      style={{
        width: size,
        height: size,
        fontSize: Math.max(8, Math.floor(size * 0.42)),
      }}
    >
      {descriptor.label}
    </div>
  )
}
