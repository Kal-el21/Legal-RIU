import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { legalCaseService, type ChronologyFormData, type LegalCaseFilters, type LegalCaseFormData } from '@/services/legal-case.service'

const KEYS = {
  all: ['legal-cases'] as const,
  list: (params?: object) => [...KEYS.all, 'list', params] as const,
  latest: () => [...KEYS.all, 'latest'] as const,
  detail: (id: string) => [...KEYS.all, 'detail', id] as const,
  regencies: (params?: object) => [...KEYS.all, 'regencies', params] as const,
  cedants: (params?: object) => [...KEYS.all, 'cedants', params] as const,
  divisions: (params?: object) => [...KEYS.all, 'divisions', params] as const,
}

export function useLegalCases(params?: LegalCaseFilters) {
  return useQuery({
    queryKey: KEYS.list(params),
    queryFn: () => legalCaseService.getAll(params),
  })
}

export function useLatestLegalCase() {
  return useQuery({
    queryKey: KEYS.latest(),
    queryFn: () => legalCaseService.getLatest(),
  })
}

export function useLegalCase(id: string) {
  return useQuery({
    queryKey: KEYS.detail(id),
    queryFn: () => legalCaseService.getByID(id),
    enabled: !!id,
  })
}

export function useCreateLegalCase() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: LegalCaseFormData) => legalCaseService.create(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.all })
    },
  })
}

export function useUpdateLegalCase() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: LegalCaseFormData }) => legalCaseService.update(id, data),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: KEYS.detail(id) })
      qc.invalidateQueries({ queryKey: KEYS.all })
    },
  })
}

export function useDeleteLegalCase() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => legalCaseService.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.all })
    },
  })
}

export function useCreateCaseChronology(caseID: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: ChronologyFormData) => legalCaseService.createChronology(caseID, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.detail(caseID) })
      qc.invalidateQueries({ queryKey: KEYS.all })
    },
  })
}

export function useDeleteCaseChronology(caseID: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (chronologyID: string) => legalCaseService.deleteChronology(caseID, chronologyID),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.detail(caseID) })
    },
  })
}

export function useRegencies(params?: { search?: string; limit?: number }) {
  return useQuery({
    queryKey: KEYS.regencies(params),
    queryFn: () => legalCaseService.getRegencies(params),
  })
}

export function useCedants(params?: { search?: string; limit?: number }) {
  return useQuery({
    queryKey: KEYS.cedants(params),
    queryFn: () => legalCaseService.getCedants(params),
  })
}

export function useCreateCedant() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: { name: string; description?: string }) => legalCaseService.createCedant(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.all })
    },
  })
}

export function useDivisions(params?: { search?: string }) {
  return useQuery({
    queryKey: KEYS.divisions(params),
    queryFn: () => legalCaseService.getDivisions(params),
  })
}
