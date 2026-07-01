import api from './api'
import type { ApiResponse, CaseChronology, Cedant, Division, LegalCase, PaginatedData, Regency } from '@/types'

export interface LegalCaseFilters {
  page?: number
  limit?: number
  search?: string
  status?: string
  case_type?: string
  level?: string
  date_from?: string
  date_to?: string
}

export interface LegalCaseFormData {
  case_name: string
  case_summary?: string
  related_party_id: string
  category: string
  specification?: string
  case_type: string
  technical_reserve?: string
  case_value: number
  pic: string
  document_link?: string
  current_status?: string
  case_date: string
  level: string
  additional_notes?: string
  location_regency_id: string
}

export interface ChronologyFormData {
  agenda_date: string
  agenda: string
  description?: string
  documents?: string[]
  files?: File[]
}

export const legalCaseService = {
  getAll: async (params?: LegalCaseFilters) => {
    const res = await api.get<ApiResponse<PaginatedData<LegalCase>>>('/admin/legal-cases', { params })
    return res.data.data!
  },

  getLatest: async () => {
    const res = await api.get<ApiResponse<LegalCase | null>>('/admin/legal-cases/latest')
    return res.data.data ?? null
  },

  getByID: async (id: string) => {
    const res = await api.get<ApiResponse<LegalCase>>(`/admin/legal-cases/${id}`)
    return res.data.data!
  },

  create: async (data: LegalCaseFormData) => {
    const res = await api.post<ApiResponse<LegalCase>>('/admin/legal-cases', data)
    return res.data.data!
  },

  update: async (id: string, data: LegalCaseFormData) => {
    const res = await api.put<ApiResponse<LegalCase>>(`/admin/legal-cases/${id}`, data)
    return res.data.data!
  },

  delete: async (id: string) => {
    await api.delete(`/admin/legal-cases/${id}`)
  },

  createChronology: async (caseID: string, data: ChronologyFormData) => {
    const form = new FormData()
    form.append('agenda_date', data.agenda_date)
    form.append('agenda', data.agenda)
    if (data.description) form.append('description', data.description)
    data.documents?.forEach((path) => form.append('document_paths', path))
    data.files?.forEach((file) => form.append('documents', file))

    const res = await api.post<ApiResponse<CaseChronology>>(`/admin/legal-cases/${caseID}/chronology`, form)
    return res.data.data!
  },

  updateChronology: async (caseID: string, chronologyID: string, data: ChronologyFormData) => {
    const form = new FormData()
    form.append('agenda_date', data.agenda_date)
    form.append('agenda', data.agenda)
    if (data.description) form.append('description', data.description)
    data.documents?.forEach((path) => form.append('document_paths', path))
    data.files?.forEach((file) => form.append('documents', file))

    const res = await api.put<ApiResponse<CaseChronology>>(`/admin/legal-cases/${caseID}/chronology/${chronologyID}`, form)
    return res.data.data!
  },

  deleteChronology: async (caseID: string, chronologyID: string) => {
    await api.delete(`/admin/legal-cases/${caseID}/chronology/${chronologyID}`)
  },

  uploadDocument: async (caseID: string, file: File) => {
    const form = new FormData()
    form.append('document', file)
    const res = await api.post<ApiResponse<LegalCase>>(`/admin/legal-cases/${caseID}/upload-document`, form)
    return res.data.data!
  },

  deleteDocument: async (caseID: string) => {
    const res = await api.delete(`/admin/legal-cases/${caseID}/document`)
    return res.data.data!
  },

  downloadFile: async (path: string): Promise<{ blob: Blob; filename: string }> => {
    const res = await api.get('/admin/legal-cases/download', {
      params: { path },
      responseType: 'blob',
    })

    const contentDisposition = res.headers['content-disposition']
    let filename = path.split('/').pop() || 'download'
    if (contentDisposition) {
      const match = contentDisposition.match(/filename="?([^"]+)"?/i)
      if (match) filename = match[1]
    }

    return { blob: res.data, filename }
  },

  getRegencies: async (params?: { search?: string; limit?: number }) => {
    const res = await api.get<ApiResponse<Regency[]>>('/admin/legal-cases/regencies', { params })
    return res.data.data ?? []
  },

  getCedants: async (params?: { search?: string; limit?: number }) => {
    const res = await api.get<ApiResponse<Cedant[]>>('/admin/legal-cases/cedants', { params })
    return res.data.data ?? []
  },

  createCedant: async (data: { name: string; description?: string }) => {
    const res = await api.post<ApiResponse<Cedant>>('/admin/legal-cases/cedants', data)
    return res.data.data!
  },

  getDivisions: async (params?: { search?: string }) => {
    const res = await api.get<ApiResponse<Division[]>>('/divisions', { params })
    return res.data.data ?? []
  },
}
