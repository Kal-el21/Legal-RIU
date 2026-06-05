import api from './api'
import type { ApiResponse, AuthResponse, LoginRequest, User } from '@/types'

export const authService = {
  login: async (data: LoginRequest): Promise<AuthResponse> => {
    const res = await api.post<ApiResponse<AuthResponse>>('/auth/login', data)
    return res.data.data!
  },

  me: async (): Promise<User> => {
    const res = await api.get<ApiResponse<User>>('/auth/me')
    return res.data.data!
  },

  changePassword: async (data: { current_password: string; new_password: string }): Promise<void> => {
    await api.post('/auth/change-password', data)
  },
}