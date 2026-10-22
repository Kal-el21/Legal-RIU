import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { documentReviewService, type CreateDocumentReviewData } from '@/services/document-review.service'

const KEYS = {
  all: ['document-reviews'] as const,
  list: (params?: object) => [...KEYS.all, 'list', params] as const,
  detail: (id: string) => [...KEYS.all, 'detail', id] as const,
}

export function useDocumentReviews(params?: { page?: number; limit?: number; status?: string }, enabled = true) {
  return useQuery({
    queryKey: KEYS.list(params),
    queryFn: () => documentReviewService.getAll(params),
    enabled,
  })
}

export function useDocumentReview(id: string) {
  return useQuery({
    queryKey: KEYS.detail(id),
    queryFn: () => documentReviewService.getByID(id),
    enabled: !!id,
  })
}

export function useCreateDocumentReview() {
  const qc = useQueryClient()
  return useMutation({
    meta: { success: 'Review dokumen berhasil diajukan' },
    mutationFn: (data: CreateDocumentReviewData) => documentReviewService.create(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEYS.all }),
  })
}

export function useUpdateDocumentReview() {
  const qc = useQueryClient()
  return useMutation({
    meta: { success: 'Review dokumen berhasil disimpan' },
    mutationFn: ({ id, data }: { id: string; data: Omit<CreateDocumentReviewData, 'attachments'> }) =>
      documentReviewService.update(id, data),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: KEYS.detail(id) })
      qc.invalidateQueries({ queryKey: KEYS.all })
    },
  })
}

export function useDeleteDocumentReview() {
  const qc = useQueryClient()
  return useMutation({
    meta: { success: 'Review dokumen berhasil dihapus' },
    mutationFn: (id: string) => documentReviewService.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEYS.all }),
  })
}

export function useResubmitDocumentReview() {
  const qc = useQueryClient()
  return useMutation({
    meta: { success: 'Review dokumen berhasil diajukan ulang' },
    mutationFn: ({ id, files }: { id: string; files?: File[] }) =>
      documentReviewService.resubmit(id, files),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: KEYS.detail(id) })
      qc.invalidateQueries({ queryKey: KEYS.all })
    },
  })
}

export function useAdminUpdateDocumentReviewStatus() {
  const qc = useQueryClient()
  return useMutation({
    meta: { success: 'Status berhasil diubah' },
    mutationFn: ({ id, status, admin_note }: { id: string; status: string; admin_note?: string }) =>
      documentReviewService.adminUpdateStatus(id, { status, admin_note }),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: KEYS.detail(id) })
      qc.invalidateQueries({ queryKey: KEYS.all })
    },
  })
}

export function useLegalUpdateDocumentReviewStatus() {
  const qc = useQueryClient()
  return useMutation({
    meta: { success: 'Status berhasil diubah' },
    mutationFn: ({ id, status, admin_note }: { id: string; status: string; admin_note?: string }) =>
      documentReviewService.legalUpdateStatus(id, { status, admin_note }),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: KEYS.detail(id) })
      qc.invalidateQueries({ queryKey: KEYS.all })
    },
  })
}

export function useAdminDownloadPDF() {
  return useMutation({
    mutationFn: (id: string) => documentReviewService.adminDownloadPDF(id),
  })
}

// ─── Draf (PR Draf) ───

export function useSaveDocumentReviewDraft() {
  const qc = useQueryClient()
  return useMutation({
    meta: { success: 'Draf tersimpan' },
    mutationFn: ({ id, data, files }: { id?: string; data: Omit<CreateDocumentReviewData, 'attachments'>; files?: File[] }) =>
      documentReviewService.saveDraft(data, files, id),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEYS.all }),
  })
}

// Mengirim draf: isian terakhir dan lampiran baru disimpan dulu, lalu dikirim.
export function useSubmitDocumentReviewDraft() {
  const qc = useQueryClient()
  return useMutation({
    meta: { success: 'Review dokumen berhasil diajukan' },
    mutationFn: async ({ id, data, files }: { id: string; data: Omit<CreateDocumentReviewData, 'attachments'>; files?: File[] }) => {
      await documentReviewService.saveDraft(data, files, id)
      return documentReviewService.submitDraft(id)
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: KEYS.all }),
  })
}

export function useDeleteDocumentReviewDraftAttachment() {
  const qc = useQueryClient()
  return useMutation({
    meta: { success: 'Lampiran dihapus' },
    mutationFn: ({ id, attachmentId }: { id: string; attachmentId: string }) =>
      documentReviewService.deleteDraftAttachment(id, attachmentId),
    onSuccess: (_, { id }) => qc.invalidateQueries({ queryKey: KEYS.detail(id) }),
  })
}
