import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { materialService, type MaterialFormData } from '@/services/material.service'

const KEYS = {
  all: ['materials'] as const,
  list: (params?: object) => [...KEYS.all, 'list', params] as const,
  detail: (id: string) => [...KEYS.all, 'detail', id] as const,
}

export function useMaterials(params?: { page?: number; limit?: number; search?: string }) {
  return useQuery({
    queryKey: KEYS.list(params),
    queryFn: () => materialService.getAll(params),
  })
}

export function useMaterial(id: string) {
  return useQuery({
    queryKey: KEYS.detail(id),
    queryFn: () => materialService.getByID(id),
    enabled: !!id,
  })
}

export function useCreateMaterial() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: MaterialFormData) => materialService.create(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.all })
    },
  })
}

export function useUpdateMaterial() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: MaterialFormData }) => materialService.update(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.all })
    },
  })
}

export function useDeleteMaterial() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => materialService.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.all })
    },
  })
}

export function useImportLegalMaterials() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (file: File) => materialService.importExcel(file),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.all })
    },
  })
}
