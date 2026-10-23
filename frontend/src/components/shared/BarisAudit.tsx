import { Link } from 'react-router-dom'
import {
  ArrowRight, Building2, FilePlus2, FileX2, LogIn, LogOut, PenLine, Pencil, RefreshCw, ShieldCheck, Trash2,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { jamWIB, kodeStatus, labelAksi, labelEntitas, labelPeran, nadaAksi, perbandingan, pisahCatatan, tautanData, type NadaAksi } from '@/lib/audit'
import StatusBadge from '@/components/common/StatusBadge'
import type { AuditAction, AuditLog } from '@/types'

const IKON: Record<AuditAction, LucideIcon> = {
  STATUS_CHANGE: RefreshCw,
  FILE_UPLOAD: FilePlus2,
  USER_UPDATE: Pencil,
  LOGIN: LogIn,
  LOGOUT: LogOut,
  DELETE: Trash2,
  FILE_DELETE: FileX2,
  PERMISSION_UPDATE: ShieldCheck,
  PARENT_CHANGE: Building2,
  SIGNATORY_CHANGE: PenLine,
}

const WARNA: Record<NadaAksi, string> = {
  netral: 'bg-gray-100 text-gray-700',
  hapus: 'bg-red-50 text-red-700',
}

function BadgeAksi({ action }: { action: AuditAction }) {
  const Ikon = IKON[action] ?? Pencil
  return (
    <span className={cn('inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-xs font-medium', WARNA[nadaAksi(action)])}>
      <Ikon className="h-3 w-3" aria-hidden="true" />
      {labelAksi(action)}
    </span>
  )
}

// Nilai lama atau baru: status pengajuan memakai StatusBadge seperti halaman lain, nilai lain sebagai teks.
function Nilai({ asli, teks, lama }: { asli?: string; teks: string; lama?: boolean }) {
  const status = kodeStatus(asli)
  if (status) return <StatusBadge status={status} className={lama ? 'opacity-70' : undefined} />
  return lama ? <del className="text-gray-500">{teks}</del> : <ins className="font-medium text-gray-900 no-underline">{teks}</ins>
}

// IsiAktivitas: aksi, keterangan, nilai lama dan baru, serta catatan approver. Dipakai tabel Audit Log
// dan Riwayat Aktivitas di halaman detail.
export function IsiAktivitas({ log }: { log: AuditLog }) {
  const { keterangan, catatan } = pisahCatatan(log.description)
  const banding = perbandingan(log)
  return (
    <div className="min-w-0 space-y-1.5">
      <BadgeAksi action={log.action} />
      {keterangan && <p className="wrap-break-word text-sm text-gray-900">{keterangan}</p>}
      {banding && (
        <p className="flex flex-wrap items-center gap-1.5 text-xs">
          {banding.lama && <Nilai asli={log.old_value} teks={banding.lama} lama />}
          {banding.lama && <ArrowRight className="h-3 w-3 text-gray-400" aria-label="menjadi" />}
          {banding.baru && <Nilai asli={log.new_value} teks={banding.baru} />}
        </p>
      )}
      {catatan && (
        <blockquote className="rounded-r-lg border-l-4 border-gray-300 bg-gray-50 px-3 py-1.5 text-sm text-gray-700">
          <span className="sr-only">Catatan: </span>{catatan}
        </blockquote>
      )}
    </div>
  )
}

// Jam WIB, dengan IP di bawahnya; user agent cukup sebagai keterangan saat kursor diarahkan.
// IP kosong untuk yang bukan auditor (riwayat per data, keputusan L7).
export function WaktuDanPerangkat({ log }: { log: AuditLog }) {
  return (
    <>
      <time dateTime={log.created_at} className="block text-sm tabular-nums text-gray-900">{jamWIB(log.created_at)}</time>
      {log.ip_address && (
        <span className="block text-xs tabular-nums text-gray-400" title={log.user_agent || undefined}>{log.ip_address}</span>
      )}
    </>
  )
}

export function Pelaku({ log }: { log: AuditLog }) {
  return (
    <>
      <p className="text-sm font-medium text-gray-900">{log.user?.full_name ?? 'Pengguna tidak dikenal'}</p>
      {log.user?.role && <p className="text-xs text-gray-400">{labelPeran(log.user.role)}</p>}
    </>
  )
}

// Jenis data dan tautan "Lihat", gaya "Lihat Detail" di daftar pengajuan.
export function DataAudit({ log, panel }: { log: AuditLog; panel: string }) {
  const tautan = tautanData(log, panel)
  return (
    <>
      <p className="text-sm text-gray-500">{labelEntitas(log.entity_type)}</p>
      {tautan && (
        <Link to={tautan} className="text-xs font-medium hover:underline" style={{ color: '#C8102E' }}>
          Lihat<span className="sr-only"> {labelEntitas(log.entity_type)}</span> →
        </Link>
      )}
    </>
  )
}

// Baris daftar ringkas untuk Riwayat Aktivitas di halaman detail (tanpa tautan, pembaca sudah di datanya).
export default function BarisAudit({ log }: { log: AuditLog }) {
  return (
    <li className="flex gap-4 px-5 py-4">
      <div className="w-24 shrink-0"><WaktuDanPerangkat log={log} /></div>
      <div className="min-w-0 flex-1 space-y-2">
        <IsiAktivitas log={log} />
        <p className="text-xs text-gray-600">
          {log.user?.full_name ?? 'Pengguna tidak dikenal'}
          {log.user?.role && <span className="text-gray-400"> · {labelPeran(log.user.role)}</span>}
        </p>
      </div>
    </li>
  )
}
