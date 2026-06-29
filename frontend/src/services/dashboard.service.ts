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
}