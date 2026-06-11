import api from './api'
import type { ApiResponse, UserDashboardStats, AdminDashboardStats, LegalOpinion, DocumentReview } from '@/types'

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
}