import api from './api'
import { useAuthStore } from '@/store/auth.store'
import type { ApiResponse, UserDashboardStats, AdminDashboardStats, LegalOpinion, DocumentReview, RemindersResponse } from '@/types'

export interface ReminderParams {
  page?: number
  limit?: number
}

export interface MarkReminderReadPayload {
  submission_type: string
  submission_id: string
}

function getReminderEndpoint() {
  const role = useAuthStore.getState().user?.role
  const pathname = typeof window === 'undefined' ? '' : window.location.pathname
  if (role === 'ADMIN' || pathname.startsWith('/admin')) return '/admin/dashboard/reminders'
  if (role === 'LEGAL' || pathname.startsWith('/legal')) return '/legal/dashboard/reminders'
  return '/dashboard/reminders'
}

export const dashboardService = {
  getUserStats: async (): Promise<UserDashboardStats> => {
    const res = await api.get<ApiResponse<UserDashboardStats>>('/dashboard/stats')
    return res.data.data!
  },

  getUserRecent: async (): Promise<{ legal_opinions: LegalOpinion[]; document_reviews: DocumentReview[] }> => {
    const res = await api.get('/dashboard/recent')
    return res.data.data
  },

  getAdminStats: async (): Promise<AdminDashboardStats> => {
    const res = await api.get<ApiResponse<AdminDashboardStats>>('/admin/dashboard/stats')
    return res.data.data!
  },

  getAdminRecent: async (): Promise<{ legal_opinions: LegalOpinion[]; document_reviews: DocumentReview[] }> => {
    const res = await api.get('/admin/dashboard/recent')
    return res.data.data
  },

  getLegalStats: async (): Promise<AdminDashboardStats> => {
    const res = await api.get<ApiResponse<AdminDashboardStats>>('/legal/dashboard/stats')
    return res.data.data!
  },

  getLegalRecent: async (): Promise<{ legal_opinions: LegalOpinion[]; document_reviews: DocumentReview[] }> => {
    const res = await api.get('/legal/dashboard/recent')
    return res.data.data
  },

  getExternalStats: async (): Promise<AdminDashboardStats> => {
    const res = await api.get<ApiResponse<AdminDashboardStats>>('/external/dashboard/stats')
    return res.data.data!
  },

  getExternalRecent: async (): Promise<{ legal_opinions: LegalOpinion[]; document_reviews: DocumentReview[] }> => {
    const res = await api.get('/external/dashboard/recent')
    return res.data.data
  },

  getReminders: async (params?: ReminderParams): Promise<RemindersResponse> => {
    const res = await api.get<ApiResponse<RemindersResponse>>(getReminderEndpoint(), { params })
    return res.data.data!
  },

  markReminderRead: async (payload: MarkReminderReadPayload): Promise<void> => {
    await api.patch(`${getReminderEndpoint()}/read`, payload)
  },

  markAllRemindersRead: async (): Promise<void> => {
    await api.patch(`${getReminderEndpoint()}/read-all`)
  },
}
