import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { legalOpinionService, type CreateLegalOpinionData } from '@/services/legal-opinion.service'

const KEYS = {
  all: ['legal-opinions'] as const,
  list: (params?: object) => [...KEYS.all, 'list', params] as const,
  detail: (id: string) => [...KEYS.all, 'detail', id] as const,
}

export function useLegalOpinions(params?: { page?: number; limit?: number; status?: string }, enabled = true) {
  return useQuery({
    queryKey: KEYS.list(params),
    queryFn: () => legalOpinionService.getAll(params),
    enabled,
  })
}

export function useLegalOpinion(id: string) {
  return useQuery({
    queryKey: KEYS.detail(id),
    queryFn: () => legalOpinionService.getByID(id),
    enabled: !!id,
  })
}

export function useCreateLegalOpinion() {
  const qc = useQueryClient()
  return useMutation({
    meta: { success: 'Legal opinion berhasil diajukan' },
    mutationFn: (data: CreateLegalOpinionData) => legalOpinionService.create(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEYS.all }),
  })
}

export function useUpdateLegalOpinion() {
  const qc = useQueryClient()
  return useMutation({
    meta: { success: 'Legal opinion berhasil disimpan' },
    mutationFn: ({ id, data }: { id: string; data: Omit<CreateLegalOpinionData, 'attachments'> }) =>
      legalOpinionService.update(id, data),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: KEYS.detail(id) })
      qc.invalidateQueries({ queryKey: KEYS.all })
    },
  })
}

export function useDeleteLegalOpinion() {
  const qc = useQueryClient()
  return useMutation({
    meta: { success: 'Legal opinion berhasil dihapus' },
    mutationFn: (id: string) => legalOpinionService.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEYS.all }),
  })
}

export function useResubmitLegalOpinion() {
  const qc = useQueryClient()
  return useMutation({
    meta: { success: 'Legal opinion berhasil diajukan ulang' },
    mutationFn: ({ id, files }: { id: string; files?: File[] }) =>
      legalOpinionService.resubmit(id, files),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: KEYS.detail(id) })
      qc.invalidateQueries({ queryKey: KEYS.all })
    },
  })
}

export function useAdminUpdateStatus() {
  const qc = useQueryClient()
  return useMutation({
    meta: { success: 'Status berhasil diubah' },
    mutationFn: ({ id, status, admin_note }: { id: string; status: string; admin_note?: string }) =>
      legalOpinionService.adminUpdateStatus(id, { status, admin_note }),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: KEYS.detail(id) })
      qc.invalidateQueries({ queryKey: KEYS.all })
    },
  })
}

export function useLegalUpdateStatus() {
  const qc = useQueryClient()
  return useMutation({
    meta: { success: 'Status berhasil diubah' },
    mutationFn: ({ id, status, admin_note }: { id: string; status: string; admin_note?: string }) =>
      legalOpinionService.legalUpdateStatus(id, { status, admin_note }),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: KEYS.detail(id) })
      qc.invalidateQueries({ queryKey: KEYS.all })
    },
  })
}

export function useAdminDownloadPDF() {
  return useMutation({
    mutationFn: (id: string) => legalOpinionService.adminDownloadPDF(id),
  })
}

// ─── Draf (PR Draf) ───

export function useSaveLegalOpinionDraft() {
  const qc = useQueryClient()
  return useMutation({
    meta: { success: 'Draf tersimpan' },
    mutationFn: ({ id, data, files }: { id?: string; data: Omit<CreateLegalOpinionData, 'attachments'>; files?: File[] }) =>
      legalOpinionService.saveDraft(data, files, id),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEYS.all }),
  })
}

// Mengirim draf: isian terakhir dan lampiran baru disimpan dulu, lalu dikirim.
export function useSubmitLegalOpinionDraft() {
  const qc = useQueryClient()
  return useMutation({
    meta: { success: 'Legal opinion berhasil diajukan' },
    mutationFn: async ({ id, data, files }: { id: string; data: Omit<CreateLegalOpinionData, 'attachments'>; files?: File[] }) => {
      await legalOpinionService.saveDraft(data, files, id)
      return legalOpinionService.submitDraft(id)
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: KEYS.all }),
  })
}

export function useDeleteLegalOpinionDraftAttachment() {
  const qc = useQueryClient()
  return useMutation({
    meta: { success: 'Lampiran dihapus' },
    mutationFn: ({ id, attachmentId }: { id: string; attachmentId: string }) =>
      legalOpinionService.deleteDraftAttachment(id, attachmentId),
    onSuccess: (_, { id }) => qc.invalidateQueries({ queryKey: KEYS.detail(id) }),
  })
}
