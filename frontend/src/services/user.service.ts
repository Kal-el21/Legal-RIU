import api from './api'
import type { ApiResponse, User, PaginatedData } from '@/types'

export interface CreateUserData {
  full_name: string
  email: string
  password: string
  position: string
  division: string
  role: 'USER' | 'ADMIN'
}

export interface UpdateUserData {
  full_name: string
  position: string
  division: string
  role: 'USER' | 'ADMIN'
}

export const userService = {
  getAll: async (params?: { page?: number; limit?: number; search?: string }) => {
    const res = await api.get<ApiResponse<PaginatedData<User>>>('/admin/users', { params })
    return res.data.data!
  },

  create: async (data: CreateUserData) => {
    const res = await api.post<ApiResponse<User>>('/admin/users', data)
    return res.data.data!
  },

  update: async (id: string, data: UpdateUserData) => {
    const res = await api.put<ApiResponse<User>>(`/admin/users/${id}`, data)
    return res.data.data!
  },

  updateStatus: async (id: string, status: 'ACTIVE' | 'INACTIVE') => {
    await api.patch(`/admin/users/${id}/status`, { status })
  },

  delete: async (id: string) => {
    await api.delete(`/admin/users/${id}`)
  },

  resetPassword: async (id: string, new_password: string) => {
    await api.post(`/admin/users/${id}/reset-password`, { new_password })
  },
}