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

export function getReminderTypeLabel(item: ReminderItem) {
  return item.submission_type === 'document_review' ? 'Review Dokumen' : 'Legal Opinion'
}
