// Fungsi tampilan Audit Log (PR Audit Log tahap L.2). Dipakai halaman Audit Log, lalu
// riwayat per data di halaman detail (L.4) dan Repositori Dokumen (tautan ke data).
import { STATUS_LABEL } from './utils'
import type { AuditAction, AuditLog, SubmissionStatus, UserRole } from '@/types'

export type NadaAksi = 'biru' | 'hijau' | 'abu' | 'merah' | 'ungu' | 'kuning' | 'teal'

export const AKSI: Record<AuditAction, { label: string; nada: NadaAksi }> = {
  STATUS_CHANGE: { label: 'Status diubah', nada: 'biru' },
  FILE_UPLOAD: { label: 'File diunggah', nada: 'hijau' },
  USER_UPDATE: { label: 'Data dibuat atau diubah', nada: 'abu' },
  LOGIN: { label: 'Masuk', nada: 'teal' },
  LOGOUT: { label: 'Keluar', nada: 'abu' },
  DELETE: { label: 'Dihapus', nada: 'merah' },
  FILE_DELETE: { label: 'File dihapus', nada: 'merah' },
  PERMISSION_UPDATE: { label: 'Izin diubah', nada: 'ungu' },
  PARENT_CHANGE: { label: 'Induk perusahaan diganti', nada: 'kuning' },
  SIGNATORY_CHANGE: { label: 'Penandatangan diubah', nada: 'kuning' },
}

export function labelAksi(aksi: string): string {
  return AKSI[aksi as AuditAction]?.label ?? aksi
}

export function nadaAksi(aksi: string): NadaAksi {
  return AKSI[aksi as AuditAction]?.nada ?? 'abu'
}

// Semua jenis entitas yang ditulis backend (SetAuditContext dan auditLogSvc.Log); divisi sejak L.3.
export const ENTITAS: Record<string, string> = {
  legal_opinion: 'Legal Opinion',
  document_review: 'Review Dokumen',
  agreement_document: 'Pengajuan perjanjian',
  agreement_template: 'Template perjanjian',
  agreement_type: 'Jenis perjanjian',
  agreement_signatory: 'Daftar Penandatangan',
  legal_case: 'Perkara',
  legal_material: 'Materi Legal',
  company: 'Perusahaan',
  division: 'Divisi',
  case_type: 'Jenis perkara',
  case_category: 'Kategori perkara',
  purpose_type: 'Tujuan pembuatan',
  document_type: 'Jenis dokumen',
  regency: 'Kabupaten/kota',
  cedant: 'Cedant',
  user: 'Pengguna',
  auth: 'Sesi login',
}

export function labelEntitas(jenis: string): string {
  return ENTITAS[jenis] ?? jenis
}

// Pilihan saringan jenis data, urut menurut labelnya.
export const PILIHAN_ENTITAS = Object.entries(ENTITAS)
  .map(([value, label]) => ({ value, label }))
  .sort((a, b) => a.label.localeCompare(b.label, 'id'))

const PERAN: Record<UserRole, string> = {
  ADMIN: 'Admin',
  LEGAL: 'Legal',
  LEGAL_AU: 'Legal AU',
  USER: 'User',
  EXTERNAL: 'External',
}

export function labelPeran(peran?: string): string {
  return PERAN[peran as UserRole] ?? peran ?? ''
}

// pisahCatatan memisahkan catatan approver dari keterangan. Backend menulisnya sebagai
// "<keterangan>. Catatan: <isi>" (perjanjian sejak PR A, LO/RD/perkara sejak L.3).
export function pisahCatatan(keterangan?: string): { keterangan: string; catatan?: string } {
  const teks = keterangan ?? ''
  const i = teks.indexOf('. Catatan: ')
  if (i < 0) return { keterangan: teks }
  return { keterangan: teks.slice(0, i), catatan: teks.slice(i + '. Catatan: '.length) }
}

const POLA_UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i

const tampilNilai = (v?: string) => (v ? STATUS_LABEL[v as SubmissionStatus] ?? v : undefined)

// perbandingan: nilai lama dan baru untuk ditampilkan, dengan status pengajuan dibaca
// labelnya. Nilai berupa ID (mis. perusahaan di ganti induk) tidak ditampilkan, karena
// keterangannya sudah menyebut namanya.
export function perbandingan(log: Pick<AuditLog, 'old_value' | 'new_value'>): { lama?: string; baru?: string } | null {
  const lama = tampilNilai(log.old_value)
  const baru = tampilNilai(log.new_value)
  if (!lama && !baru) return null
  if ([log.old_value, log.new_value].some((v) => v && POLA_UUID.test(v))) return null
  return { lama, baru }
}

const ZONA = 'Asia/Jakarta'
const kunciTanggal = (d: Date) => new Intl.DateTimeFormat('en-CA', { timeZone: ZONA }).format(d) // YYYY-MM-DD

export function jamWIB(iso: string): string {
  return new Intl.DateTimeFormat('id-ID', { timeZone: ZONA, hour: '2-digit', minute: '2-digit' }).format(new Date(iso))
}

// kelompokPerHari mengelompokkan baris per tanggal WIB, urutan baris dipertahankan.
// Label "Hari ini" dan "Kemarin" dihitung dari sekarang.
export function kelompokPerHari<T extends Pick<AuditLog, 'created_at'>>(
  baris: T[],
  sekarang: Date = new Date(),
): { tanggal: string; label: string; items: T[] }[] {
  const hariIni = kunciTanggal(sekarang)
  const kemarin = kunciTanggal(new Date(sekarang.getTime() - 24 * 60 * 60 * 1000))
  const formatLabel = new Intl.DateTimeFormat('id-ID', { timeZone: ZONA, weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })
  const hasil: { tanggal: string; label: string; items: T[] }[] = []
  for (const b of baris) {
    const d = new Date(b.created_at)
    const tanggal = kunciTanggal(d)
    let kelompok = hasil[hasil.length - 1]
    if (!kelompok || kelompok.tanggal !== tanggal) {
      const label = tanggal === hariIni ? 'Hari ini' : tanggal === kemarin ? 'Kemarin' : formatLabel.format(d)
      kelompok = { tanggal, label, items: [] }
      hasil.push(kelompok)
    }
    kelompok.items.push(b)
  }
  return hasil
}

const ID_KOSONG = '00000000-0000-0000-0000-000000000000'

// Halaman detail per data, sama alamatnya di panel admin dan legal.
const DETAIL: Record<string, string> = {
  legal_opinion: 'legal-opinions',
  document_review: 'review-documents',
  agreement_document: 'agreement-documents',
  legal_case: 'legal-cases',
  legal_material: 'materials',
}

// Data tanpa halaman detail per ID: tautannya ke halaman kelolanya (hanya ada di panel admin).
const KELOLA_ADMIN: Record<string, string> = {
  company: 'companies',
  division: 'divisions',
  case_type: 'case-types',
  case_category: 'case-categories',
  purpose_type: 'purpose-types',
  document_type: 'document-types',
  regency: 'regencies',
  cedant: 'cedants',
  user: 'users',
  agreement_type: 'agreement-types',
  agreement_template: 'agreement-types',
  agreement_signatory: 'agreement-company-master',
}

// tautanData: alamat data yang dicatat, menurut panel pengguna ('/admin' atau '/legal').
// Data yang dihapus, sesi login, dan panel lain tidak bertautan. Baris impor (ID kosong)
// menuju daftar datanya.
export function tautanData(log: Pick<AuditLog, 'action' | 'entity_type' | 'entity_id'>, panel: string): string | null {
  if (log.action === 'DELETE' || (panel !== '/admin' && panel !== '/legal')) return null
  const detail = DETAIL[log.entity_type]
  if (detail) return log.entity_id && log.entity_id !== ID_KOSONG ? `${panel}/${detail}/${log.entity_id}` : `${panel}/${detail}`
  const kelola = KELOLA_ADMIN[log.entity_type]
  return panel === '/admin' && kelola ? `/admin/${kelola}` : null
}
