import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from '@tanstack/react-router'
import { Building2, Check, ChevronsUpDown, Plus, UserRound } from 'lucide-react'
import { useCallback, useEffect } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useStatus } from '@/hooks/use-status'
import { useAuthStore } from '@/stores/auth-store'

import { getOrganizations } from '../api'
import { useActiveOrganizationId, useWorkspaceStore } from '../workspace-store'

const isWorkspaceQuery = (query: { queryKey: readonly unknown[] }) => {
  const root = String(query.queryKey[0])
  return (
    root.startsWith('organization') ||
    [
      'keys',
      'api-key',
      'token',
      'tokens',
      'token-auto-groups',
      'user-groups',
      'user-models',
      'usage',
      'logs',
      'tasks',
      'dashboard',
      'dashboard-overview',
      'user-data',
      'quota',
      'wallet',
      'topup',
    ].includes(root)
  )
}

export function WorkspaceSwitcher() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const { status } = useStatus()
  const enabled = status?.organization_enabled !== false
  const userId = useAuthStore((state) => state.auth.user?.id)
  const activeId = useActiveOrganizationId()
  const selectOrganization = useWorkspaceStore(
    (state) => state.selectOrganization
  )
  const organizations = useQuery({
    queryKey: ['organizations', userId],
    queryFn: getOrganizations,
    enabled: Boolean(userId) && enabled,
    staleTime: 30_000,
  })
  const active = organizations.data?.find(
    (entry) => entry.organization.id === activeId
  )
  const clearWorkspaceQueries = useCallback(async () => {
    await queryClient.cancelQueries({ predicate: isWorkspaceQuery })
    queryClient.removeQueries({ predicate: isWorkspaceQuery })
  }, [queryClient])

  useEffect(() => {
    const unavailable =
      organizations.isSuccess &&
      !organizations.data.some((entry) => entry.organization.id === activeId)
    if (!userId || activeId === null || (enabled && !unavailable)) return
    let cancelled = false
    void clearWorkspaceQueries().then(() => {
      if (!cancelled) selectOrganization(userId, null)
    })
    return () => {
      cancelled = true
    }
  }, [
    activeId,
    clearWorkspaceQueries,
    enabled,
    organizations.data,
    organizations.isSuccess,
    selectOrganization,
    userId,
  ])

  const switchWorkspace = async (organizationId: number | null) => {
    if (!userId) return
    await clearWorkspaceQueries()
    selectOrganization(userId, organizationId)
    await navigate({
      to: organizationId === null ? '/dashboard' : '/organizations',
    })
  }

  if (!enabled) return null

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={<Button variant='ghost' className='max-w-44 gap-2 px-2' />}
      >
        {active ? (
          <Building2 data-icon='inline-start' />
        ) : (
          <UserRound data-icon='inline-start' />
        )}
        <span className='max-w-24 truncate sm:max-w-32'>
          {active?.organization.name ?? t('Personal space')}
        </span>
        <ChevronsUpDown data-icon='inline-end' />
      </DropdownMenuTrigger>
      <DropdownMenuContent align='start' className='min-w-60'>
        <DropdownMenuGroup>
          <DropdownMenuLabel>{t('Switch workspace')}</DropdownMenuLabel>
          <DropdownMenuItem onSelect={() => switchWorkspace(null)}>
            <UserRound />
            {t('Personal space')}
            {!active && <Check className='ms-auto' />}
          </DropdownMenuItem>
          {(organizations.data ?? []).map((entry) => (
            <DropdownMenuItem
              key={entry.organization.id}
              onSelect={() => switchWorkspace(entry.organization.id)}
            >
              <Building2 />
              <span className='truncate'>{entry.organization.name}</span>
              {activeId === entry.organization.id && (
                <Check className='ms-auto' />
              )}
            </DropdownMenuItem>
          ))}
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuGroup>
          <DropdownMenuItem
            onSelect={() => void navigate({ to: '/organizations' })}
          >
            <Plus />
            {t('Manage organizations')}
          </DropdownMenuItem>
        </DropdownMenuGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
