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

function buildFormData(data: CreateLegalOpinionData): FormData {
  const form = new FormData()
  const json = { ...data }
  delete (json as Record<string, unknown>).attachments
  // Send JSON fields as individual form fields
  Object.entries(json).forEach(([k, v]) => {
    if (v !== undefined && v !== null) form.append(k, String(v))
  })
  data.attachments?.forEach((f) => form.append('attachments', f))
  return form
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
    const form = new FormData()
    // send body fields as JSON string + files separately
    const { attachments, ...fields } = data
    Object.entries(fields).forEach(([k, v]) => {
      if (v !== undefined && v !== null) form.append(k, String(v))
    })
    attachments?.forEach((f) => form.append('attachments', f))
    const res = await api.post<ApiResponse<LegalOpinion>>('/legal-opinions', form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
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
    const res = await api.post<ApiResponse<LegalOpinion>>(`/legal-opinions/${id}/resubmit`, form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
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
    await api.post(`/admin/legal-opinions/${id}/result`, form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },
}

// suppress unused warning
void buildFormData