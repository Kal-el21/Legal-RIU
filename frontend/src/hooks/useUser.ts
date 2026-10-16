import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { userService, type CreateUserData, type UpdateUserData } from '@/services/user.service'

const KEYS = {
  all: ['users'] as const,
  list: (params?: object) => [...KEYS.all, 'list', params] as const,
}

export function useUsers(params?: { page?: number; limit?: number; search?: string }) {
  return useQuery({
    queryKey: KEYS.list(params),
    queryFn: () => userService.getAll(params),
  })
}

export function useCreateUser() {
  const qc = useQueryClient()
  return useMutation({
    meta: { success: 'User berhasil ditambahkan' },
    mutationFn: (data: CreateUserData) => userService.create(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEYS.all }),
  })
}

export function useUpdateUser() {
  const qc = useQueryClient()
  return useMutation({
    meta: { success: 'User berhasil disimpan' },
    mutationFn: ({ id, data }: { id: string; data: UpdateUserData }) => userService.update(id, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEYS.all }),
  })
}

export function useUpdateUserStatus() {
  const qc = useQueryClient()
  return useMutation({
    meta: { success: 'Status user berhasil diubah' },
    mutationFn: ({ id, status }: { id: string; status: 'ACTIVE' | 'INACTIVE' }) =>
      userService.updateStatus(id, status),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEYS.all }),
  })
}

export function useDeleteUser() {
  const qc = useQueryClient()
  return useMutation({
    meta: { success: 'User berhasil dihapus' },
    mutationFn: (id: string) => userService.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEYS.all }),
  })
}

export function useResetPassword() {
  return useMutation({
    meta: { success: 'Password berhasil direset' },
    mutationFn: ({ id, password }: { id: string; password: string }) =>
      userService.resetPassword(id, password),
  })
}