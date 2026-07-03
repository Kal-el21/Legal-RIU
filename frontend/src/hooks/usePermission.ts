import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { permissionService } from '@/services/permission.service'
import type { PermissionEffect } from '@/types'

const KEYS = {
  all: ['permissions'] as const,
  catalog: () => [...KEYS.all, 'catalog'] as const,
  user: (userID?: string) => [...KEYS.all, 'user', userID] as const,
}

export function usePermissionCatalog() {
  return useQuery({
    queryKey: KEYS.catalog(),
    queryFn: permissionService.getCatalog,
  })
}

export function useUserPermissions(userID?: string) {
  return useQuery({
    queryKey: KEYS.user(userID),
    queryFn: () => permissionService.getUserAccess(userID!),
    enabled: Boolean(userID),
  })
}

export function useUpdateUserPermissions(userID?: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (overrides: { code: string; effect: PermissionEffect }[]) =>
      permissionService.updateUserAccess(userID!, overrides),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.user(userID) })
      qc.invalidateQueries({ queryKey: ['users'] })
    },
  })
}
