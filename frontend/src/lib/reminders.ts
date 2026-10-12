import type { ReminderItem, UserRole } from '@/types'

// Sama dengan penjaga route /notifications di backend
export const NOTIFICATION_PERMISSIONS = [
  'legal_opinion.view.own',
  'legal_opinion.view.all',
  'document_review.view.own',
  'document_review.view.all',
  'agreement_document.view.own',
  'agreement_document.view.all',
]

// Modul pengajuan yang masuk lonceng: label dan bagian alamat halaman detailnya
const REMINDER_TYPES: Record<string, { label: string; segment: string }> = {
  legal_opinion: { label: 'Legal Opinion', segment: 'legal-opinions' },
  document_review: { label: 'Review Dokumen', segment: 'review-documents' },
  agreement_document: { label: 'Dokumen Perjanjian', segment: 'agreement-documents' },
}

const reminderType = (item: ReminderItem) => REMINDER_TYPES[item.submission_type] ?? REMINDER_TYPES.legal_opinion

export function getNotificationBasePath(role?: UserRole | null) {
  if (role === 'ADMIN') return '/admin'
  if (role === 'LEGAL') return '/legal'
  if (role === 'LEGAL_AU') return '/legal-au'
  if (role === 'EXTERNAL') return '/external'
  return '/dashboard'
}

export function getReminderDetailPath(item: ReminderItem, role?: UserRole | null) {
  return `${getNotificationBasePath(role)}/${reminderType(item).segment}/${item.id}`
}

// Kalimat item lonceng, mengikuti status pengajuan saat ini. Pemilik menerima
// status yang diubah pemroses; pemroses menerima pengajuan yang masuk.
export function getReminderMessage(item: ReminderItem) {
  switch (item.status) {
    case 'SUBMITTED':
      return `Pengajuan baru dari ${item.requester_name}`
    case 'RESUBMITTED':
      return `Diajukan ulang oleh ${item.requester_name}`
    case 'NEED_REVISION':
      return 'Perlu revisi dari Anda'
    case 'REJECTED':
      return 'Ditolak'
    case 'COMPLETED':
      return 'Selesai, hasil bisa diunduh'
    default:
      return 'Sedang direview'
  }
}

// Baris yang menunggu tindakan penerimanya: pemroses atas pengajuan yang masuk
// atau terlambat, pemilik atas permintaan revisi. Sisanya hanya kabar.
export function reminderNeedsAction(item: ReminderItem) {
  return item.warning_level !== 'NONE' || ['SUBMITTED', 'RESUBMITTED', 'NEED_REVISION'].includes(item.status)
}

// Latar baris notifikasi: netral untuk yang belum dibaca. Kuning dan merah hanya
// dipakai badge peringatan.
export function getReminderRowTone(item: ReminderItem) {
  return item.is_read
    ? 'hover:bg-gray-50 focus:bg-gray-50'
    : 'bg-navy/5 hover:bg-navy/10 focus:bg-navy/10'
}

export function getReminderTypeLabel(item: ReminderItem) {
  return reminderType(item).label
}
