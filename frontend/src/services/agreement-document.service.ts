import api from './api'
import { useAuthStore } from '@/store/auth.store'
import type { ApiResponse, AgreementDocument, PaginatedData } from '@/types'

export interface CreateAgreementDocumentData {
  nomor_pihak_pertama?: string
  nomor_pihak_kedua?: string
  tempat_ttd?: string
  tanggal_ttd?: string
  pihak_kedua_nama: string
  pihak_kedua_bidang?: string
  pihak_kedua_alamat?: string
  pihak_kedua_telepon?: string
  pihak_kedua_email?: string
  pihak_kedua_pic?: string
  pihak_kedua_pejabat: string
  pihak_kedua_jabatan: string
  jenis_pekerjaan: string
  surat_penawaran_nomor?: string
  surat_penawaran_perihal?: string
  surat_penawaran_tanggal?: string
  surat_penunjukan_nomor?: string
  surat_penunjukan_perihal?: string
  surat_penunjukan_tanggal?: string
  ruang_lingkup: string
  jangka_waktu_mulai?: string
  jangka_waktu_selesai?: string
  nilai_kontrak?: number
  termin1_persen?: number
  termin1_nilai?: number
  termin2_persen?: number
  termin2_nilai?: number
  bank?: string
  nomor_rekening?: string
  atas_nama?: string
  attachments?: File[]
}

function basePath() {
  const role = useAuthStore.getState().user?.role
  const pathname = typeof window === 'undefined' ? '' : window.location.pathname
  if (role === 'EXTERNAL' || pathname.startsWith('/external')) return '/external/agreement-documents'
  if (role === 'ADMIN' || pathname.startsWith('/admin')) return '/admin/agreement-documents'
  if (role === 'LEGAL' || pathname.startsWith('/legal')) return '/legal/agreement-documents'
  return '/agreement-documents'
}

export const agreementDocumentService = {
  getAll: async (params?: { page?: number; limit?: number; status?: string }) => {
    const res = await api.get<ApiResponse<PaginatedData<AgreementDocument>>>(basePath(), { params })
    return res.data.data!
  },

  getByID: async (id: string) => {
    const res = await api.get<ApiResponse<AgreementDocument>>(`${basePath()}/${id}`)
    return res.data.data!
  },

  create: async (data: CreateAgreementDocumentData) => {
    const { attachments, ...fields } = data
    const form = new FormData()
    Object.entries(fields).forEach(([key, value]) => {
      if (value !== undefined && value !== null) form.append(key, String(value))
    })
    attachments?.forEach((file) => form.append('attachments', file))
    const res = await api.post<ApiResponse<AgreementDocument>>('/agreement-documents', form)
    return res.data.data!
  },

  update: async (id: string, data: Omit<CreateAgreementDocumentData, 'attachments'>) => {
    const res = await api.put<ApiResponse<AgreementDocument>>(`/agreement-documents/${id}`, data)
    return res.data.data!
  },

  delete: async (id: string) => {
    await api.delete(`/agreement-documents/${id}`)
  },

  resubmit: async (id: string, files?: File[]) => {
    const form = new FormData()
    files?.forEach((file) => form.append('attachments', file))
    const res = await api.post<ApiResponse<AgreementDocument>>(
      `/agreement-documents/${id}/resubmit`,
      form
    )
    return res.data.data!
  },

  updatePihakPertama: async (id: string, data: { pihak_pertama_pejabat: string; pihak_pertama_jabatan: string }) => {
    const res = await api.put<ApiResponse<AgreementDocument>>(
      `/admin/agreement-documents/${id}/pihak-pertama`,
      data
    )
    return res.data.data!
  },

  updateMeta: async (id: string, data: { nomor_pihak_pertama?: string; tempat_ttd?: string; tanggal_ttd?: string; pihak_pertama_pejabat?: string; pihak_pertama_jabatan?: string }) => {
    const res = await api.put<ApiResponse<AgreementDocument>>(
      `/admin/agreement-documents/${id}/meta`,
      data
    )
    return res.data.data!
  },

  approve: async (id: string) => {
    const res = await api.post<ApiResponse<AgreementDocument>>(`/admin/agreement-documents/${id}/approve`)
    return res.data.data!
  },

  returnForRevision: async (id: string, adminNote: string) => {
    const res = await api.post<ApiResponse<AgreementDocument>>(`/admin/agreement-documents/${id}/return`, {
      admin_note: adminNote,
    })
    return res.data.data!
  },

  reject: async (id: string, adminNote: string) => {
    const res = await api.post<ApiResponse<AgreementDocument>>(`/admin/agreement-documents/${id}/reject`, {
      admin_note: adminNote,
    })
    return res.data.data!
  },

  // Approver-only watermarked preview (returns a blob URL).
  getPreviewURL: async (id: string) => {
    const role = useAuthStore.getState().user?.role
    const pathname = typeof window !== 'undefined' ? window.location.pathname : ''
    const prefix = (role === 'ADMIN' || pathname.startsWith('/admin')) ? '/admin' :
                   (role === 'LEGAL' || pathname.startsWith('/legal')) ? '/legal' : ''
    const res = await api.get(`${prefix}/agreement-documents/${id}/preview`, {
      responseType: 'blob',
    })
    return URL.createObjectURL(res.data)
  },

  // Final PDF download (only when COMPLETED).
  getFinalURL: async (id: string) => {
    const res = await api.get(`/agreement-documents/${id}/pdf`, {
      responseType: 'blob',
    })
    return URL.createObjectURL(res.data)
  },
}
