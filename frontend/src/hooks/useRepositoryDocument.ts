import { useQuery, useMutation } from '@tanstack/react-query'
import { documentRepositoryService } from '@/services/document-repository.service'

const KEYS = {
  all: ['repository-documents'] as const,
  list: (params?: { feature_code?: string; search?: string }) => [...KEYS.all, 'list', params] as const,
  detail: (id: string) => [...KEYS.all, 'detail', id] as const,
}

export function useRepositoryDocuments(params?: { feature_code?: string; search?: string }) {
  return useQuery({
    queryKey: KEYS.list(params),
    queryFn: () => documentRepositoryService.getAll(params),
  })
}

export function useRepositoryDocument(id: string) {
  return useQuery({
    queryKey: KEYS.detail(id),
    queryFn: () => documentRepositoryService.getByID(id),
    enabled: !!id,
  })
}

export function useDownloadRepositoryDocument() {
  return useMutation({
    mutationFn: (id: string) => documentRepositoryService.download(id),
  })
}