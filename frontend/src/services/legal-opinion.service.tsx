import api from './api'
import type { ApiResponse, LegalOpinion, PaginatedData } from '@/types'

export interface CreateLegalOpinionData {
  requestor_name: string
  requestor_position: string
  requestor_division: string
  requestor_email: string
  requestor_phone: string
  legal_type: string
  legal_type_other?: string
  title: string
  chronology: string
  question: string
  attachments?: File[]
}

export const legalOpinionService = {
  getAll: async (params?: { page?: number; limit?: number; status?: string }) => {
    const res = await api.get<ApiResponse<PaginatedData<LegalOpinion>>>('/legal-opinions', { params })
    return res.data.data!
  },

  getByID: async (id: string) => {
    const res = await api.get<ApiResponse<LegalOpinion>>(`/legal-opinions/${id}`)
    return res.data.data!
  },

  create: async (data: CreateLegalOpinionData) => {
    const { attachments, ...fields } = data
    const form = new FormData()
    Object.entries(fields).forEach(([k, v]) => {
      if (v !== undefined && v !== null) form.append(k, String(v))
    })
    attachments?.forEach((f) => form.append('attachments', f))
    const res = await api.post<ApiResponse<LegalOpinion>>('/legal-opinions', form)
    return res.data.data!
  },

  update: async (id: string, data: Omit<CreateLegalOpinionData, 'attachments'>) => {
    const res = await api.put<ApiResponse<LegalOpinion>>(`/legal-opinions/${id}`, data)
    return res.data.data!
  },

  delete: async (id: string) => {
    await api.delete(`/legal-opinions/${id}`)
  },

  resubmit: async (id: string, files?: File[]) => {
    const form = new FormData()
    files?.forEach((f) => form.append('attachments', f))
    const res = await api.post<ApiResponse<LegalOpinion>>(`/legal-opinions/${id}/resubmit`, form)
    return res.data.data!
  },

  getPresignedURL: async (path: string) => {
    const res = await api.get<ApiResponse<{ url: string }>>('/legal-opinions/presign', { params: { path } })
    return res.data.data!.url
  },

  // Admin
  adminUpdateStatus: async (id: string, data: { status: string; admin_note?: string }) => {
    await api.patch(`/admin/legal-opinions/${id}/status`, data)
  },

  adminUploadResult: async (id: string, file: File, notes?: string) => {
    const form = new FormData()
    form.append('result', file)
    if (notes) form.append('notes', notes)
    await api.post(`/admin/legal-opinions/${id}/result`, form)
  },
}