import { Link } from 'react-router-dom'
import {
  ArrowRight, Building2, FilePlus2, FileX2, LogIn, LogOut, PenLine, Pencil, RefreshCw, ShieldCheck, Trash2,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { jamWIB, labelAksi, labelEntitas, labelPeran, nadaAksi, perbandingan, pisahCatatan, tautanData, type NadaAksi } from '@/lib/audit'
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
  biru: 'bg-blue-50 text-blue-700',
  hijau: 'bg-green-50 text-green-700',
  abu: 'bg-gray-100 text-gray-700',
  merah: 'bg-red-50 text-red-700',
  ungu: 'bg-purple-50 text-purple-700',
  kuning: 'bg-amber-50 text-amber-800',
  teal: 'bg-teal-50 text-teal-700',
}

// Kartu satu baris audit: jam, aksi, keterangan, catatan approver, perubahan nilai, pelaku, dan tautan ke data.
// Dipakai halaman Audit Log dan Riwayat Aktivitas di halaman detail (panel kosong = tanpa tautan).
export default function BarisAudit({ log, panel }: { log: AuditLog; panel: string }) {
  const Ikon = IKON[log.action] ?? Pencil
  const { keterangan, catatan } = pisahCatatan(log.description)
  const banding = perbandingan(log)
  const tautan = tautanData(log, panel)
  const perangkat = [log.ip_address, log.user_agent].filter(Boolean).join(' · ')

  return (
    <li className="flex gap-4 px-5 py-4">
      <time dateTime={log.created_at} className="w-11 shrink-0 pt-1 text-xs tabular-nums text-gray-500">{jamWIB(log.created_at)}</time>
      <span className={cn('mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-full', WARNA[nadaAksi(log.action)])} aria-hidden="true">
        <Ikon className="h-3.5 w-3.5" />
      </span>
      <div className="min-w-0 flex-1">
        <p className="flex flex-wrap items-center gap-x-2 gap-y-1 text-sm">
          <span className={cn('rounded-full px-2 py-0.5 text-xs font-medium', WARNA[nadaAksi(log.action)])}>{labelAksi(log.action)}</span>
          <span className="text-gray-500">{labelEntitas(log.entity_type)}</span>
        </p>
        {keterangan && <p className="mt-1 break-words text-sm text-gray-900">{keterangan}</p>}
        {catatan && (
          <blockquote className="mt-2 rounded-r-lg border-l-4 border-gray-300 bg-gray-50 px-3 py-2 text-sm text-gray-700">
            <span className="sr-only">Catatan: </span>{catatan}
          </blockquote>
        )}
        {banding && (
          <p className="mt-1.5 flex flex-wrap items-center gap-1.5 text-xs">
            {banding.lama && <del className="text-gray-500">{banding.lama}</del>}
            {banding.lama && <ArrowRight className="h-3 w-3 text-gray-400" aria-label="menjadi" />}
            {banding.baru && <ins className="font-medium text-gray-900 no-underline">{banding.baru}</ins>}
          </p>
        )}
        <p className="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-gray-600">
          <span>
            {log.user?.full_name ?? 'Pengguna tidak dikenal'}
            {log.user?.role && <span className="text-gray-400"> · {labelPeran(log.user.role)}</span>}
          </span>
          {tautan && (
            <Link to={tautan} className="inline-flex items-center gap-1 font-medium text-[#0B2545] hover:underline">
              Buka data <ArrowRight className="h-3 w-3" aria-hidden="true" />
            </Link>
          )}
        </p>
        {perangkat && <p className="mt-1 truncate text-[11px] text-gray-400" title={perangkat}>{perangkat}</p>}
      </div>
    </li>
  )
}
