import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { ChevronDown, History } from 'lucide-react'
import { cn } from '@/lib/utils'
import { kelompokPerHari } from '@/lib/audit'
import BarisAudit from '@/components/shared/BarisAudit'
import type { AuditLog } from '@/types'

// Riwayat aktivitas satu data di halaman detail (PR Audit Log tahap L.4). Tertutup di awal dan baru
// dimuat saat dibuka. Tanpa tautan "Buka data", karena pembaca sudah di halaman datanya.
export default function RiwayatAktivitas({ kunci, ambil }: { kunci: string[]; ambil: () => Promise<AuditLog[]> }) {
  const [buka, setBuka] = useState(false)
  const { data = [], isLoading, isError } = useQuery({
    queryKey: ['riwayat', ...kunci],
    queryFn: ambil,
    enabled: buka,
  })
  const idIsi = `riwayat-${kunci.join('-')}`

  return (
    <section className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
      <h3>
        <button
          type="button"
          onClick={() => setBuka((b) => !b)}
          aria-expanded={buka}
          aria-controls={idIsi}
          className="flex w-full items-center gap-2 px-5 py-4 text-left text-sm font-semibold hover:bg-gray-50"
          style={{ color: '#0B2545' }}
        >
          <History className="h-4 w-4" aria-hidden="true" />
          Riwayat Aktivitas
          <ChevronDown className={cn('ml-auto h-4 w-4 text-gray-400 transition-transform', buka && 'rotate-180')} aria-hidden="true" />
        </button>
      </h3>
      {buka && (
        <div id={idIsi} className="border-t border-gray-100">
          {isLoading ? (
            <p className="px-5 py-6 text-center text-sm text-gray-400">Memuat riwayat...</p>
          ) : isError ? (
            <p role="alert" className="px-5 py-6 text-center text-sm text-red-600">Riwayat gagal dimuat. Tutup lalu buka lagi.</p>
          ) : data.length === 0 ? (
            <p className="px-5 py-6 text-center text-sm text-gray-500">Belum ada aktivitas yang tercatat</p>
          ) : (
            kelompokPerHari(data).map((h) => (
              <div key={h.tanggal}>
                <h4 className="bg-gray-50 px-5 py-2 text-xs font-semibold text-gray-600">{h.label}</h4>
                <ul className="divide-y divide-gray-50">
                  {h.items.map((log) => <BarisAudit key={log.id} log={log} panel="" />)}
                </ul>
              </div>
            ))
          )}
        </div>
      )}
    </section>
  )
}
