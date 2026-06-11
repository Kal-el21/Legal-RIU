import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { documentReviewService, type CreateDocumentReviewData } from '@/services/document-review.service'

const KEYS = {
  all: ['document-reviews'] as const,
  list: (params?: object) => [...KEYS.all, 'list', params] as const,
  detail: (id: string) => [...KEYS.all, 'detail', id] as const,
}

export function useDocumentReviews(params?: { page?: number; limit?: number; status?: string }) {
  return useQuery({
    queryKey: KEYS.list(params),
    queryFn: () => documentReviewService.getAll(params),
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
    mutationFn: (data: CreateDocumentReviewData) => documentReviewService.create(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEYS.all }),
  })
}

export function useUpdateDocumentReview() {
  const qc = useQueryClient()
  return useMutation({
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
    mutationFn: (id: string) => documentReviewService.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEYS.all }),
  })
}

export function useResubmitDocumentReview() {
  const qc = useQueryClient()
  return useMutation({
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
    mutationFn: ({ id, status, admin_note }: { id: string; status: string; admin_note?: string }) =>
      documentReviewService.adminUpdateStatus(id, { status, admin_note }),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: KEYS.detail(id) })
      qc.invalidateQueries({ queryKey: KEYS.all })
    },
  })
}