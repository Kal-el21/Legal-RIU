import api from './api'
import type { ApiResponse, Permission, UserPermissionAccess, UserPermissionOverride } from '@/types'

export const permissionService = {
  getCatalog: async () => {
    const res = await api.get<ApiResponse<Permission[]>>('/admin/permissions')
    return res.data.data ?? []
  },

  getUserAccess: async (userID: string) => {
    const res = await api.get<ApiResponse<UserPermissionAccess>>(`/admin/users/${userID}/permissions`)
    return res.data.data!
  },

  updateUserAccess: async (userID: string, overrides: Pick<UserPermissionOverride, 'code' | 'effect'>[]) => {
    const res = await api.put<ApiResponse<UserPermissionAccess>>(`/admin/users/${userID}/permissions`, {
      overrides,
    })
    return res.data.data!
  },
}
