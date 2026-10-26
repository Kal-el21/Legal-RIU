import { tautanData } from './audit'
import type { RepositoryDocument } from '@/types'

export const MODUL_DOKUMEN: Record<string, string> = {
  legal_opinion: 'Legal Opinion',
  document_review: 'Review Dokumen',
  agreement_document: 'Perjanjian',
  case_management: 'Perkara',
}

export const JENIS_DOKUMEN = [
  { value: 'legal_opinion_attachments', label: 'Lampiran Legal Opinion' },
  { value: 'legal_opinion_results', label: 'Hasil Legal Opinion' },
  { value: 'document_review_attachments', label: 'Lampiran Review Dokumen' },
  { value: 'document_review_results', label: 'Hasil Review Dokumen' },
  { value: 'agreement_attachments', label: 'Lampiran Perjanjian' },
  { value: 'agreement_documents:application/pdf', label: 'Perjanjian Final (PDF)' },
  { value: 'agreement_documents:application/vnd.openxmlformats-officedocument.wordprocessingml.document', label: 'Perjanjian Final (DOCX)' },
  { value: 'legal_cases', label: 'Dokumen Perkara' },
  { value: 'legal_cases_photo', label: 'Foto Perkara' },
  { value: 'case_chronologies', label: 'Lampiran Kronologi' },
] as const

export function labelModul(featureCode: string): string {
  return MODUL_DOKUMEN[featureCode] ?? 'Dokumen'
}

export function labelJenisDokumen(doc: Pick<RepositoryDocument, 'source_type' | 'mime_type'>): string {
  if (doc.source_type === 'agreement_documents') {
    return doc.mime_type === 'application/pdf' ? 'Perjanjian Final (PDF)' : 'Perjanjian Final (DOCX)'
  }
  return JENIS_DOKUMEN.find(({ value }) => value === doc.source_type)?.label ?? 'Dokumen'
}

export function filterJenisDokumen(value: string): { source_type?: string; mime_type?: string } {
  if (!value) return {}
  const [sourceType, mimeType] = value.split(':')
  return { source_type: sourceType, ...(mimeType ? { mime_type: mimeType } : {}) }
}

export function formatUkuranDokumen(bytes: number): string {
  if (bytes <= 0) return '-'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export function formatTanggalUnggah(date?: string | null): string {
  if (!date) return '-'
  const value = new Date(date)
  if (Number.isNaN(value.getTime())) return '-'
  return new Intl.DateTimeFormat('id-ID', {
    timeZone: 'Asia/Jakarta',
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  }).format(value)
}

export function labelPengunggah(type: string): string {
  if (type === 'requester') return 'Pengaju'
  if (type === 'approver') return 'Approver'
  return 'Tidak diketahui'
}

const ENTITAS_SUMBER: Record<string, string> = {
  legal_opinion_attachments: 'legal_opinion',
  legal_opinion_results: 'legal_opinion',
  document_review_attachments: 'document_review',
  document_review_results: 'document_review',
  agreement_attachments: 'agreement_document',
  agreement_documents: 'agreement_document',
  legal_cases: 'legal_case',
  legal_cases_photo: 'legal_case',
}

export function tautanDokumenSumber(doc: Pick<RepositoryDocument, 'source_type' | 'source_id'>, panel: string): string | null {
  const entityType = ENTITAS_SUMBER[doc.source_type]
  if (!entityType) return null
  return tautanData({ action: 'FILE_UPLOAD', entity_type: entityType, entity_id: doc.source_id }, panel)
}
