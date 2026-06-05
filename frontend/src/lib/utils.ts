import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'
import type { SubmissionStatus } from '@/types'

// ─── Tailwind class merger ────────────────────────────────────────────────────
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

// ─── Date formatting ──────────────────────────────────────────────────────────
export function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  })
}

export function formatDateTime(dateString: string): string {
  return new Date(dateString).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

// ─── File size formatting ─────────────────────────────────────────────────────
export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`
}

// ─── Status display helpers ───────────────────────────────────────────────────
export const STATUS_LABEL: Record<SubmissionStatus, string> = {
  SUBMITTED: 'Diajukan',
  UNDER_REVIEW: 'Sedang Direview',
  NEED_REVISION: 'Perlu Revisi',
  REJECTED: 'Ditolak',
  RESUBMITTED: 'Diajukan Ulang',
  COMPLETED: 'Selesai',
}

export const STATUS_COLOR: Record<SubmissionStatus, string> = {
  SUBMITTED: 'bg-blue-100 text-blue-700',
  UNDER_REVIEW: 'bg-yellow-100 text-yellow-700',
  NEED_REVISION: 'bg-orange-100 text-orange-700',
  REJECTED: 'bg-red-100 text-red-700',
  RESUBMITTED: 'bg-purple-100 text-purple-700',
  COMPLETED: 'bg-green-100 text-green-700',
}

export function getStatusLabel(status: SubmissionStatus): string {
  return STATUS_LABEL[status] ?? status
}

export function getStatusColor(status: SubmissionStatus): string {
  return STATUS_COLOR[status] ?? 'bg-gray-100 text-gray-700'
}

// ─── File validation ──────────────────────────────────────────────────────────
const ALLOWED_EXTENSIONS = ['.pdf', '.doc', '.docx']
const MAX_FILE_SIZE = 100 * 1024 * 1024 // 100 MB

export function validateFile(file: File): string | null {
  const ext = '.' + file.name.split('.').pop()?.toLowerCase()
  if (!ALLOWED_EXTENSIONS.includes(ext)) {
    return `Format file tidak didukung. Gunakan: PDF, DOC, DOCX`
  }
  if (file.size > MAX_FILE_SIZE) {
    return `Ukuran file melebihi batas maksimal 100 MB`
  }
  return null
}