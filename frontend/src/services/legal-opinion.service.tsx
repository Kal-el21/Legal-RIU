import api from './api'
import { useAuthStore } from '@/store/auth.store'
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

function getLegalOpinionEndpoint() {
  const role = useAuthStore.getState().user?.role
  const pathname = typeof window === 'undefined' ? '' : window.location.pathname
  if (role === 'EXTERNAL' || pathname.startsWith('/external')) return '/external/legal-opinions'
  if (role === 'ADMIN' || pathname.startsWith('/admin')) return '/admin/legal-opinions'
  if (role === 'LEGAL' || pathname.startsWith('/legal')) return '/legal/legal-opinions'
  return '/legal-opinions'
}

export const legalOpinionService = {
  getAll: async (params?: { page?: number; limit?: number; status?: string }) => {
    const res = await api.get<ApiResponse<PaginatedData<LegalOpinion>>>(getLegalOpinionEndpoint(), { params })
    return res.data.data!
  },

  getByID: async (id: string) => {
    const res = await api.get<ApiResponse<LegalOpinion>>(`${getLegalOpinionEndpoint()}/${id}`)
    return res.data.data!
  },

  create: async (data: CreateLegalOpinionData) => {
    const { attachments, ...fields } = data
    const form = new FormData()

    // Append semua text fields
    Object.entries(fields).forEach(([k, v]) => {
      if (v !== undefined && v !== null && v !== '') {
        form.append(k, String(v))
      }
    })

    // Append files jika ada
    attachments?.forEach((f) => form.append('attachments', f))

    // JANGAN set Content-Type manual — browser otomatis set beserta boundary
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
    const res = await api.get<ApiResponse<{ url: string }>>(`${getLegalOpinionEndpoint()}/presign`, { params: { path } })
    return res.data.data!.url
  },

  downloadFile: async (path: string): Promise<{ blob: Blob; filename: string }> => {
    const res = await api.get(`${getLegalOpinionEndpoint()}/download`, {
      params: { path },
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

  adminUpdateStatus: async (id: string, data: { status: string; admin_note?: string }) => {
    await api.patch(`/admin/legal-opinions/${id}/status`, data)
  },

  adminUploadResult: async (id: string, file: File, notes?: string) => {
    const form = new FormData()
    form.append('result', file)
    if (notes) form.append('notes', notes)
    await api.post(`/admin/legal-opinions/${id}/result`, form)
  },

  adminDownloadPDF: async (id: string): Promise<{ blob: Blob; filename: string }> => {
    const res = await api.get(`${getLegalOpinionEndpoint()}/${id}/pdf`, {
      responseType: 'blob',
    })

    const contentDisposition = res.headers['content-disposition']
    let filename = `legal-opinion-${id}.pdf`
    if (contentDisposition) {
      const match = contentDisposition.match(/filename="?([^"]+)"?/i)
      if (match) filename = match[1]
    }

    return { blob: res.data, filename }
  },

  legalUpdateStatus: async (id: string, data: { status: string; admin_note?: string }) => {
    await api.patch(`/legal/legal-opinions/${id}/status`, data)
  },

  legalUploadResult: async (id: string, file: File, notes?: string) => {
    const form = new FormData()
    form.append('result', file)
    if (notes) form.append('notes', notes)
    await api.post(`/legal/legal-opinions/${id}/result`, form)
  },
}
