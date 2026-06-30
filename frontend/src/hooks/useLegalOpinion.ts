import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { legalOpinionService, type CreateLegalOpinionData } from '@/services/legal-opinion.service'

const KEYS = {
  all: ['legal-opinions'] as const,
  list: (params?: object) => [...KEYS.all, 'list', params] as const,
  detail: (id: string) => [...KEYS.all, 'detail', id] as const,
}

export function useLegalOpinions(params?: { page?: number; limit?: number; status?: string }) {
  return useQuery({
    queryKey: KEYS.list(params),
    queryFn: () => legalOpinionService.getAll(params),
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
    mutationFn: (data: CreateLegalOpinionData) => legalOpinionService.create(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEYS.all }),
  })
}

export function useUpdateLegalOpinion() {
  const qc = useQueryClient()
  return useMutation({
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
    mutationFn: (id: string) => legalOpinionService.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEYS.all }),
  })
}

export function useResubmitLegalOpinion() {
  const qc = useQueryClient()
  return useMutation({
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