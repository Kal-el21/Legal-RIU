import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { agreementTypeService, type AgreementTypePayload } from '@/services/agreement-type.service'
import { agreementTemplateService, type UploadTemplatePayload } from '@/services/agreement-template.service'

const KEYS = {
  types: ['agreement-types'] as const,
  versions: (code: string) => ['agreement-templates', code] as const,
  placeholders: (code: string) => ['agreement-templates', code, 'placeholders'] as const,
}

export function useAgreementTypes() {
  return useQuery({ queryKey: KEYS.types, queryFn: agreementTypeService.list })
}

export function useCreateAgreementType() {
  const qc = useQueryClient()
  return useMutation({
    meta: { success: 'Jenis perjanjian berhasil dibuat' },
    mutationFn: (payload: AgreementTypePayload) => agreementTypeService.create(payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEYS.types }),
  })
}

export function useUpdateAgreementType(code: string, success = 'Jenis perjanjian berhasil disimpan') {
  const qc = useQueryClient()
  return useMutation({
    meta: { success },
    mutationFn: (payload: AgreementTypePayload) => agreementTypeService.update(code, payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.types })
      qc.invalidateQueries({ queryKey: KEYS.placeholders(code) })
    },
  })
}

export function useTemplateVersions(code: string) {
  return useQuery({ queryKey: KEYS.versions(code), queryFn: () => agreementTemplateService.list(code) })
}

export function usePlaceholders(code: string) {
  return useQuery({ queryKey: KEYS.placeholders(code), queryFn: () => agreementTemplateService.placeholders(code) })
}

// Penolakan pencocokan (placeholder tanpa field) ditampilkan di halaman beserta
// daftarnya, jadi tidak perlu notifikasi galat tambahan.
export function useUploadTemplate(code: string) {
  const qc = useQueryClient()
  return useMutation({
    meta: { success: 'Template lolos uji coba dan tersimpan sebagai Draft', errorToast: false },
    mutationFn: ({ file, payload }: { file?: File; payload: UploadTemplatePayload }) =>
      file ? agreementTemplateService.upload(file, payload) : agreementTemplateService.newVersion(payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEYS.versions(code) }),
  })
}

export function useActivateTemplate(code: string) {
  const qc = useQueryClient()
  return useMutation({
    meta: { success: 'Versi template diaktifkan untuk pengajuan baru' },
    mutationFn: (id: string) => agreementTemplateService.activate(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEYS.versions(code) })
      qc.invalidateQueries({ queryKey: KEYS.types })
    },
  })
}

export function useDeleteTemplate(code: string) {
  const qc = useQueryClient()
  return useMutation({
    meta: { success: 'Versi template dihapus' },
    mutationFn: (id: string) => agreementTemplateService.remove(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: KEYS.versions(code) }),
  })
}
