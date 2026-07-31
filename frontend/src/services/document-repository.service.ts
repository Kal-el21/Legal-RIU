import api from './api'
import type { ApiResponse, RepositoryDocument } from '@/types'

export const documentRepositoryService = {
  getAll: async (params?: { feature_code?: string; search?: string }) => {
    const res = await api.get<ApiResponse<RepositoryDocument[]>>('/document-repositories', { params })
    return res.data.data ?? []
  },
  getByID: async (id: string) => {
    const res = await api.get<ApiResponse<RepositoryDocument>>(`/document-repositories/${id}`)
    return res.data.data!
  },
  download: async (id: string): Promise<{ blob: Blob; filename: string }> => {
    const res = await api.get(`/document-repositories/${id}/download`, {
      responseType: 'blob',
    })
    const contentDisposition = res.headers['content-disposition']
    let filename = 'download'
    if (contentDisposition) {
      const match = contentDisposition.match(/filename="?([^"]+)"?/i)
      if (match) filename = match[1]
    }
    return { blob: res.data, filename }
  },
}