import api from './api'
import { namaDariHeader } from '@/lib/nama-unduhan'
import type { ApiResponse, PaginatedData, RepositoryDocument } from '@/types'

export const documentRepositoryService = {
  getAll: async (params?: {
    page?: number
    limit?: number
    feature_code?: string
    source_type?: string
    mime_type?: string
    date_from?: string
    date_to?: string
    search?: string
  }): Promise<PaginatedData<RepositoryDocument>> => {
    const res = await api.get<ApiResponse<PaginatedData<RepositoryDocument>>>('/document-repositories', { params })
    if (!res.data.data) throw new Error('Respons daftar dokumen tidak valid')
    return res.data.data
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