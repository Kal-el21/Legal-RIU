import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { documentTypeService } from '@/services/document-type.service'

const KEYS = {
  all: ['document-types'] as const,
  list: () => [...KEYS.all, 'list'] as const,
}

export function useDocumentTypes() {
  return useQuery({
    queryKey: KEYS.list(),
    queryFn: () => documentTypeService.getAll(),
  })
}

export function useCreateDocumentType() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: { name: string; label: string }) => documentTypeService.create(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.all })
    },
  })
}

export function useUpdateDocumentType() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: { name: string; label: string; is_active?: boolean } }) => documentTypeService.update(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.all })
    },
  })
}

export function useDeleteDocumentType() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => documentTypeService.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.all })
    },
  })
}