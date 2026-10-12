import type { ReminderItem, UserRole } from '@/types'

// Sama dengan penjaga route /notifications di backend
export const NOTIFICATION_PERMISSIONS = [
  'legal_opinion.view.own',
  'legal_opinion.view.all',
  'document_review.view.own',
  'document_review.view.all',
]

export function getNotificationBasePath(role?: UserRole | null) {
  if (role === 'ADMIN') return '/admin'
  if (role === 'LEGAL') return '/legal'
  if (role === 'LEGAL_AU') return '/legal-au'
  if (role === 'EXTERNAL') return '/external'
  return '/dashboard'
}

export function getReminderDetailPath(item: ReminderItem, role?: UserRole | null) {
  const basePath = getNotificationBasePath(role)
  const segment = item.submission_type === 'document_review' ? 'review-documents' : 'legal-opinions'
  return `${basePath}/${segment}/${item.id}`
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
    : 'bg-[#0B2545]/5 hover:bg-[#0B2545]/10 focus:bg-[#0B2545]/10'
}

export function getReminderTypeLabel(item: ReminderItem) {
  return item.submission_type === 'document_review' ? 'Review Dokumen' : 'Legal Opinion'
}
