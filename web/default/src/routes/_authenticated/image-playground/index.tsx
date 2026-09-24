import { createFileRoute, redirect } from '@tanstack/react-router'

import { Main } from '@/components/layout'
import { ImagePlayground } from '@/features/image-playground'
import { isSidebarModuleEnabled } from '@/lib/nav-modules'

export const Route = createFileRoute('/_authenticated/image-playground/')({
  beforeLoad: () => {
    if (!isSidebarModuleEnabled('chat', 'image')) {
      throw redirect({ to: '/dashboard' })
    }
  },
  component: ImagePlaygroundPage,
})

function ImagePlaygroundPage() {
  return (
    <Main className='p-0'>
      <ImagePlayground />
    </Main>
  )
}
