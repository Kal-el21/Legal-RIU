import api from './api'
import type { ApiResponse, DocumentType } from '@/types'

export const documentTypeService = {
  getAll: async () => {
    const res = await api.get<ApiResponse<DocumentType[]>>('/admin/document-types')
    return res.data.data ?? []
  },

  create: async (data: { name: string; label: string }) => {
    const res = await api.post<ApiResponse<DocumentType>>('/admin/document-types', data)
    return res.data.data!
  },

  update: async (id: string, data: { name: string; label: string; is_active?: boolean }) => {
    const res = await api.put<ApiResponse<DocumentType>>(`/admin/document-types/${id}`, data)
    return res.data.data!
  },

  delete: async (id: string) => {
    await api.delete(`/admin/document-types/${id}`)
  },
}