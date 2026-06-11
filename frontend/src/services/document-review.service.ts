import api from './api'
import type { ApiResponse, DocumentReview, PaginatedData } from '@/types'

export interface CreateDocumentReviewData {
  requestor_name: string
  requestor_position: string
  requestor_division: string
  requestor_email: string
  requestor_phone: string
  document_name: string
  second_party: string
  third_party?: string
  document_type: string
  document_type_other?: string
  additional_note?: string
  attachments?: File[]
}

export const documentReviewService = {
  getAll: async (params?: { page?: number; limit?: number; status?: string }) => {
    const res = await api.get<ApiResponse<PaginatedData<DocumentReview>>>('/review-documents', { params })
    return res.data.data!
  },

  getByID: async (id: string) => {
    const res = await api.get<ApiResponse<DocumentReview>>(`/review-documents/${id}`)
    return res.data.data!
  },

  create: async (data: CreateDocumentReviewData) => {
    const { attachments, ...fields } = data
    const form = new FormData()
    Object.entries(fields).forEach(([k, v]) => {
      if (v !== undefined && v !== null) form.append(k, String(v))
    })
    attachments?.forEach((f) => form.append('attachments', f))
    const res = await api.post<ApiResponse<DocumentReview>>('/review-documents', form)
    return res.data.data!
  },

  update: async (id: string, data: Omit<CreateDocumentReviewData, 'attachments'>) => {
    const res = await api.put<ApiResponse<DocumentReview>>(`/review-documents/${id}`, data)
    return res.data.data!
  },

  delete: async (id: string) => {
    await api.delete(`/review-documents/${id}`)
  },

  resubmit: async (id: string, files?: File[]) => {
    const form = new FormData()
    files?.forEach((f) => form.append('attachments', f))
    const res = await api.post<ApiResponse<DocumentReview>>(`/review-documents/${id}/resubmit`, form)
    return res.data.data!
  },

  getPresignedURL: async (path: string) => {
    const res = await api.get<ApiResponse<{ url: string }>>('/review-documents/presign', { params: { path } })
    return res.data.data!.url
  },

  adminUpdateStatus: async (id: string, data: { status: string; admin_note?: string }) => {
    await api.patch(`/admin/review-documents/${id}/status`, data)
  },

  adminUploadResult: async (id: string, file: File, notes?: string) => {
    const form = new FormData()
    form.append('result', file)
    if (notes) form.append('notes', notes)
    await api.post(`/admin/review-documents/${id}/result`, form)
  },
}