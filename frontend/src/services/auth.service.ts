import api from './api'
import type { ApiResponse, AuthResponse, LoginRequest, User } from '@/types'

export const authService = {
  login: async (data: LoginRequest): Promise<AuthResponse> => {
    const res = await api.post<ApiResponse<AuthResponse>>('/auth/login', data)
    return res.data.data!
  },

  refresh: async (): Promise<AuthResponse> => {
    const res = await api.post<ApiResponse<AuthResponse>>('/auth/refresh')
    return res.data.data!
  },

  logout: async (): Promise<void> => {
    await api.post('/auth/logout')
  },

  me: async (): Promise<User> => {
    const res = await api.get<ApiResponse<User>>('/auth/me')
    return res.data.data!
  },

  changePassword: async (data: { current_password: string; new_password: string }): Promise<void> => {
    await api.post('/auth/change-password', data)
  },
}

export const settingsService = {
  updateProfile: async (data: { full_name: string; position: string; division: string }): Promise<User> => {
    const res = await api.put<ApiResponse<User>>('/settings/profile', data)
    return res.data.data!
  },
  updateNotifications: async (email_notifications: boolean) => {
    await api.put('/settings/notifications', { email_notifications })
  },
  toggle2FA: async (enabled: boolean, password: string) => {
    await api.put('/settings/two-fa', { enabled, password })
  },
}
