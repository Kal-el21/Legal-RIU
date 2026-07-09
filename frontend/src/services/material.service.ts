import api from './api'
import type { ApiResponse, LegalMaterial, PaginatedData, ImportResult } from '@/types'

export interface MaterialFormData {
  title: string
  excerpt?: string
  content: string
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
    const res = await api.post<ApiResponse<LegalMaterial>>('/materials', data)
    return res.data.data!
  },

  update: async (id: string, data: MaterialFormData) => {
    const res = await api.put<ApiResponse<LegalMaterial>>(`/materials/${id}`, data)
    return res.data.data!
  },

  delete: async (id: string) => {
    await api.delete(`/materials/${id}`)
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
