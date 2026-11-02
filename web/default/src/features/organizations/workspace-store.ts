import { create } from 'zustand'

import { useAuthStore } from '@/stores/auth-store'

const storageKey = 'organization_workspaces'

function readSelections(): Record<string, number> {
  try {
    const value: unknown = JSON.parse(
      window.localStorage.getItem(storageKey) || '{}'
    )
    if (!value || typeof value !== 'object' || Array.isArray(value)) return {}
    return Object.fromEntries(
      Object.entries(value).filter(
        ([, id]) => Number.isSafeInteger(id) && Number(id) > 0
      )
    )
  } catch {
    return {}
  }
}

interface WorkspaceState {
  selections: Record<string, number>
  selectOrganization: (userId: number, organizationId: number | null) => void
}

export const useWorkspaceStore = create<WorkspaceState>((set) => ({
  selections: readSelections(),
  selectOrganization: (userId, organizationId) =>
    set((state) => {
      const selections = { ...state.selections }
      if (organizationId === null) delete selections[String(userId)]
      else selections[String(userId)] = organizationId
      try {
        window.localStorage.setItem(storageKey, JSON.stringify(selections))
      } catch {
        // The current session can still switch spaces when storage is unavailable.
      }
      return { selections }
    }),
}))

export function useActiveOrganizationId(): number | null {
  const userId = useAuthStore((state) => state.auth.user?.id)
  return useWorkspaceStore((state) => getSelectedOrganizationId(state, userId))
}

function getSelectedOrganizationId(
  state: Pick<WorkspaceState, 'selections'>,
  userId: number | undefined
): number | null {
  return userId ? (state.selections[String(userId)] ?? null) : null
}

/**
 * Read the active workspace outside React components. API clients and SSE
 * helpers use this to keep dashboard requests in the selected workspace.
 */
export function getActiveOrganizationId(): number | null {
  const userId = useAuthStore.getState().auth.user?.id
  return getSelectedOrganizationId(useWorkspaceStore.getState(), userId)
}
