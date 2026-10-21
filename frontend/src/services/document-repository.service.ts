import api from './api'
import { namaDariHeader } from '@/lib/nama-unduhan'
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
    const filename = namaDariHeader(res.headers['content-disposition'], 'download')
    return { blob: res.data, filename }
  },
}