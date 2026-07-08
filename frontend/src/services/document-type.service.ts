import api from './api'
import type { ApiResponse, DocumentType, ImportResult } from '@/types'

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

  importExcel: async (file: File) => {
    const form = new FormData()
    form.append('file', file)
    const res = await api.post<ApiResponse<ImportResult>>('/admin/document-types/import', form)
    return res.data.data!
  },

  downloadTemplate: async () => {
    const res = await api.get('/admin/document-types/import/template', {
      responseType: 'blob',
    })
    return { blob: res.data, filename: 'document-type-template.xlsx' }
  },
}