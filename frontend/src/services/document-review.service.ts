import api from './api'
import { formDraf } from '@/lib/draf'
import { namaDariHeader } from '@/lib/nama-unduhan'
import { useAuthStore } from '@/store/auth.store'
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

// Salinannya ada di backend/internal/router/alamat_frontend_test.go (awalanPerPeran); ubah keduanya.
function getDocumentReviewEndpoint() {
  const role = useAuthStore.getState().user?.role
  const pathname = typeof window === 'undefined' ? '' : window.location.pathname
  if (role === 'EXTERNAL' || pathname.startsWith('/external')) return '/external/review-documents'
  if (role === 'ADMIN' || pathname.startsWith('/admin')) return '/admin/review-documents'
  if (role === 'LEGAL' || pathname.startsWith('/legal')) return '/legal/review-documents'
  return '/review-documents'
}

export const documentReviewService = {
  getAll: async (params?: { page?: number; limit?: number; status?: string }) => {
    const res = await api.get<ApiResponse<PaginatedData<DocumentReview>>>(
      getDocumentReviewEndpoint(),
      { params }
    )
    return res.data.data!
  },

  getByID: async (id: string) => {
    const res = await api.get<ApiResponse<DocumentReview>>(`${getDocumentReviewEndpoint()}/${id}`)
    return res.data.data!
  },

  create: async (data: CreateDocumentReviewData) => {
    const { attachments, ...fields } = data
    const form = new FormData()

    Object.entries(fields).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        form.append(key, String(value))
      }
    })

    attachments?.forEach((file) => {
      form.append('attachments', file)
    })

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

  // Draf (PR Draf): hanya pemiliknya. Simpan draf tanpa id membuat draf baru.
  saveDraft: async (data: Omit<CreateDocumentReviewData, 'attachments'>, files: File[] = [], id?: string) => {
    const form = formDraf(data, files)
    const res = id
      ? await api.put<ApiResponse<DocumentReview>>(`/review-documents/${id}/draft`, form)
      : await api.post<ApiResponse<DocumentReview>>('/review-documents/drafts', form)
    return res.data.data!
  },

  deleteDraftAttachment: async (id: string, attachmentId: string) => {
    await api.delete(`/review-documents/${id}/attachments/${attachmentId}`)
  },

  submitDraft: async (id: string) => {
    const res = await api.post<ApiResponse<DocumentReview>>(`/review-documents/${id}/submit`)
    return res.data.data!
  },

  resubmit: async (id: string, files?: File[]) => {
    const form = new FormData()

    files?.forEach((file) => {
      form.append('attachments', file)
    })

    const res = await api.post<ApiResponse<DocumentReview>>(
      `/review-documents/${id}/resubmit`,
      form
    )

    return res.data.data!
  },


  downloadFile: async (path: string): Promise<{ blob: Blob; filename: string }> => {
    const res = await api.get(`${getDocumentReviewEndpoint()}/download`, {
      params: { path },
      responseType: 'blob',
    })

    const filename = namaDariHeader(res.headers['content-disposition'], 'download')

    return { blob: res.data, filename }
  },

  adminUpdateStatus: async (id: string, data: { status: string; admin_note?: string }) => {
    await api.patch(`/admin/review-documents/${id}/status`, data)
  },

  adminUploadResult: async (id: string, file: File, notes?: string): Promise<void> => {
    const form = new FormData()
    form.append('result', file)
    if (notes) form.append('notes', notes)
    await api.post(`/admin/review-documents/${id}/result`, form)
  },

  legalUpdateStatus: async (id: string, data: { status: string; admin_note?: string }) => {
    await api.patch(`/legal/review-documents/${id}/status`, data)
  },

  legalUploadResult: async (id: string, file: File, notes?: string): Promise<void> => {
    const form = new FormData()
    form.append('result', file)
    if (notes) form.append('notes', notes)
    await api.post(`/legal/review-documents/${id}/result`, form)
  },

  adminDownloadPDF: async (id: string): Promise<{ blob: Blob; filename: string }> => {
    const res = await api.get(`${getDocumentReviewEndpoint()}/${id}/pdf`, {
      responseType: 'blob',
    })

    const filename = namaDariHeader(res.headers['content-disposition'], `document-review-${id}.pdf`)

    return { blob: res.data, filename }
  },
}
