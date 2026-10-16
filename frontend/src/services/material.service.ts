import api from './api'
import { useAuthStore } from '@/store/auth.store'
import type { ApiResponse, LegalMaterial, PaginatedData, ImportResult } from '@/types'

export interface MaterialFormData {
  title: string
  excerpt?: string
  content: string
}

// Membaca materi lewat /materials (semua peran); menulis lewat alamat per peran.
// Salinannya ada di backend/internal/router/alamat_frontend_test.go (awalanPerPeran); ubah keduanya.
function getMaterialEndpoint() {
  const role = useAuthStore.getState().user?.role
  const pathname = typeof window === 'undefined' ? '' : window.location.pathname
  // Legal AU diperiksa sebelum LEGAL: halaman Legal AU beralamat /legal-au/..., yang juga diawali "/legal".
  if (role === 'LEGAL_AU' || pathname.startsWith('/legal-au')) return '/legal-au/materials'
  if (role === 'LEGAL' || pathname.startsWith('/legal/')) return '/legal/materials'
  return '/admin/materials'
}

export const materialService = {
  getAll: async (params?: { page?: number; limit?: number; search?: string }) => {
    const res = await api.get<ApiResponse<PaginatedData<LegalMaterial>>>('/materials', { params })
    return res.data.data!
  },

  getByID: async (id: string) => {
    const res = await api.get<ApiResponse<LegalMaterial>>(`/materials/${id}`)
    return res.data.data!
  },

  create: async (data: MaterialFormData) => {
    const res = await api.post<ApiResponse<LegalMaterial>>(getMaterialEndpoint(), data)
    return res.data.data!
  },

  update: async (id: string, data: MaterialFormData) => {
    const res = await api.put<ApiResponse<LegalMaterial>>(`${getMaterialEndpoint()}/${id}`, data)
    return res.data.data!
  },

  delete: async (id: string) => {
    await api.delete(`${getMaterialEndpoint()}/${id}`)
  },

  importExcel: async (file: File) => {
    const form = new FormData()
    form.append('file', file)
    const res = await api.post<ApiResponse<ImportResult>>('/admin/materials/import', form)
    return res.data.data!
  },

  downloadTemplate: async () => {
    const res = await api.get('/admin/materials/import/template', {
      responseType: 'blob',
    })
    return { blob: res.data, filename: 'legal-material-template.xlsx' }
  },
}
