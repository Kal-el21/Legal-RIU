import api from './api'
import { useAuthStore } from '@/store/auth.store'
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

export function getLegalCaseRouteBase() {
  const role = useAuthStore.getState().user?.role
  const pathname = typeof window === 'undefined' ? '' : window.location.pathname
  if (role === 'LEGAL' || pathname.startsWith('/legal')) return '/legal/legal-cases'
  if (role === 'EXTERNAL' || pathname.startsWith('/external')) return '/external/legal-cases'
  return '/admin/legal-cases'
}

export const legalCaseService = {
  getAll: async (params?: LegalCaseFilters) => {
    const res = await api.get<ApiResponse<PaginatedData<LegalCase>>>(getLegalCaseRouteBase(), { params })
    return res.data.data!
  },

  getLatest: async () => {
    const res = await api.get<ApiResponse<LegalCase | null>>(`${getLegalCaseRouteBase()}/latest`)
    return res.data.data ?? null
  },

  getByID: async (id: string) => {
    const res = await api.get<ApiResponse<LegalCase>>(`${getLegalCaseRouteBase()}/${id}`)
    return res.data.data!
  },

  create: async (data: LegalCaseFormData) => {
    const res = await api.post<ApiResponse<LegalCase>>(getLegalCaseRouteBase(), data)
    return res.data.data!
  },

  update: async (id: string, data: LegalCaseFormData) => {
    const res = await api.put<ApiResponse<LegalCase>>(`${getLegalCaseRouteBase()}/${id}`, data)
    return res.data.data!
  },

  delete: async (id: string) => {
    await api.delete(`${getLegalCaseRouteBase()}/${id}`)
  },

  createChronology: async (caseID: string, data: ChronologyFormData) => {
    const form = new FormData()
    form.append('agenda_date', data.agenda_date)
    form.append('agenda', data.agenda)
    if (data.description) form.append('description', data.description)
    data.documents?.forEach((path) => form.append('document_paths', path))
    data.files?.forEach((file) => form.append('documents', file))

    const res = await api.post<ApiResponse<CaseChronology>>(`${getLegalCaseRouteBase()}/${caseID}/chronology`, form)
    return res.data.data!
  },

  updateChronology: async (caseID: string, chronologyID: string, data: ChronologyFormData) => {
    const form = new FormData()
    form.append('agenda_date', data.agenda_date)
    form.append('agenda', data.agenda)
    if (data.description) form.append('description', data.description)
    data.documents?.forEach((path) => form.append('document_paths', path))
    data.files?.forEach((file) => form.append('documents', file))

    const res = await api.put<ApiResponse<CaseChronology>>(`${getLegalCaseRouteBase()}/${caseID}/chronology/${chronologyID}`, form)
    return res.data.data!
  },

  deleteChronology: async (caseID: string, chronologyID: string) => {
    await api.delete(`${getLegalCaseRouteBase()}/${caseID}/chronology/${chronologyID}`)
  },

  uploadDocument: async (caseID: string, file: File) => {
    const form = new FormData()
    form.append('document', file)
    const res = await api.post<ApiResponse<LegalCase>>(`${getLegalCaseRouteBase()}/${caseID}/upload-document`, form)
    return res.data.data!
  },

  deleteDocument: async (caseID: string) => {
    const res = await api.delete(`${getLegalCaseRouteBase()}/${caseID}/document`)
    return res.data.data!
  },

  downloadFile: async (path: string): Promise<{ blob: Blob; filename: string }> => {
    const res = await api.get(`${getLegalCaseRouteBase()}/download`, {
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
    const res = await api.get<ApiResponse<Regency[]>>(`${getLegalCaseRouteBase()}/regencies`, { params })
    return res.data.data ?? []
  },

  getCedants: async (params?: { search?: string; limit?: number }) => {
    const res = await api.get<ApiResponse<Cedant[]>>(`${getLegalCaseRouteBase()}/cedants`, { params })
    return res.data.data ?? []
  },

  createCedant: async (data: { name: string; description?: string }) => {
    const res = await api.post<ApiResponse<Cedant>>(`${getLegalCaseRouteBase()}/cedants`, data)
    return res.data.data!
  },

  getDivisions: async (params?: { search?: string }) => {
    const res = await api.get<ApiResponse<Division[]>>('/divisions', { params })
    return res.data.data ?? []
  },
}
