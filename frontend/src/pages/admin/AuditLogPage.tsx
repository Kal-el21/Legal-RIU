import { useCallback, useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useLocation, useNavigate, useSearchParams } from 'react-router-dom'
import { History, Search, X } from 'lucide-react'
import { auditLogService } from '@/services/audit-log.service'
import { cn } from '@/lib/utils'
import { AKSI, kelompokPerHari, labelPeran, PILIHAN_ENTITAS } from '@/lib/audit'
import BarisAudit from '@/components/shared/BarisAudit'
import { useAuthStore } from '@/store/auth.store'
import { getRoleHome } from '@/routes/role-home'
import { Button } from '@/components/ui/button'

const SARINGAN = ['action', 'entity_type', 'user_id', 'pelaku', 'search', 'date_from', 'date_to'] as const
const PER_HALAMAN = 20

// useTunda: nilai yang baru berubah sesudah pengguna berhenti mengetik.
function useTunda<T>(nilai: T, ms = 400): T {
  const [tunda, setTunda] = useState(nilai)
  useEffect(() => {
    const t = setTimeout(() => setTunda(nilai), ms)
    return () => clearTimeout(t)
  }, [nilai, ms])
  return tunda
}

const kelasIsian = 'w-full px-3 py-2 rounded-lg border border-gray-200 text-sm focus:outline-none focus:ring-2 focus:ring-red-500/20'

export default function AuditLogPage() {
  const navigate = useNavigate()
  const { pathname } = useLocation()
  const hasPermission = useAuthStore((state) => state.hasPermission)
  const role = useAuthStore((state) => state.user?.role)
  const [params, setParams] = useSearchParams()
  const panel = '/' + pathname.split('/')[1]

  useEffect(() => {
    if (!hasPermission('audit_log.view')) {
      navigate(getRoleHome(role), { replace: true })
    }
  }, [hasPermission, navigate, role])

  // Saringan disimpan di URL supaya tampilan yang sama bisa dibagikan lewat tautan.
  const ubah = useCallback((isi: Record<string, string>) => {
    setParams((lama) => {
      const baru = new URLSearchParams(lama)
      for (const [k, v] of Object.entries(isi)) {
        if (v) baru.set(k, v)
        else baru.delete(k)
      }
      if (!('page' in isi)) baru.delete('page')
      return baru
    }, { replace: true })
  }, [setParams])

  const [cari, setCari] = useState(params.get('search') ?? '')
  const cariTunda = useTunda(cari)
  useEffect(() => {
    setParams((lama) => {
      if ((lama.get('search') ?? '') === cariTunda) return lama
      const baru = new URLSearchParams(lama)
      if (cariTunda) baru.set('search', cariTunda)
      else baru.delete('search')
      baru.delete('page')
      return baru
    }, { replace: true })
  }, [cariTunda, setParams])

  const [cariPelaku, setCariPelaku] = useState('')
  const cariPelakuTunda = useTunda(cariPelaku.trim(), 300)
  const { data: calonPelaku = [] } = useQuery({
    queryKey: ['audit-actors', panel, cariPelakuTunda],
    queryFn: () => auditLogService.getActors(cariPelakuTunda),
    enabled: cariPelakuTunda.length >= 2,
  })

  const page = Math.max(1, Number(params.get('page')) || 1)
  const saringan = Object.fromEntries(SARINGAN.filter((k) => k !== 'pelaku').map((k) => [k, params.get(k) ?? '']))
  const adaSaringan = SARINGAN.some((k) => params.get(k))

  const { data, isLoading, isError } = useQuery({
    queryKey: ['audit-logs', panel, page, saringan],
    queryFn: () => auditLogService.getAll({ ...saringan, page, limit: PER_HALAMAN }),
  })

  const items = data?.items ?? []
  const totalPages = data?.total_pages || 1
  const total = data?.total ?? 0
  const hari = kelompokPerHari(items)

  const hapusSaringan = () => {
    setCari('')
    setCariPelaku('')
    setParams(new URLSearchParams(), { replace: true })
  }

  return (
    <div className="p-6 max-w-5xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>Audit Log</h1>
        <p className="text-sm text-gray-500 mt-0.5">Siapa melakukan apa, dan kapan</p>
      </div>

      <section aria-label="Saringan" className="bg-white rounded-2xl border border-gray-100 p-5 space-y-4">
        <div className="relative">
          <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" aria-hidden="true" />
          <input
            type="search"
            value={cari}
            onChange={(e) => setCari(e.target.value)}
            placeholder="Cari keterangan, nilai, atau nama pelaku..."
            aria-label="Cari"
            className={cn(kelasIsian, 'pl-9')}
          />
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <label className="block">
            <span className="text-xs text-gray-500 mb-1 block">Aksi</span>
            <select value={params.get('action') ?? ''} onChange={(e) => ubah({ action: e.target.value })} className={kelasIsian}>
              <option value="">Semua aksi</option>
              {Object.entries(AKSI).map(([value, a]) => <option key={value} value={value}>{a.label}</option>)}
            </select>
          </label>
          <label className="block">
            <span className="text-xs text-gray-500 mb-1 block">Jenis data</span>
            <select value={params.get('entity_type') ?? ''} onChange={(e) => ubah({ entity_type: e.target.value })} className={kelasIsian}>
              <option value="">Semua jenis</option>
              {PILIHAN_ENTITAS.map((p) => <option key={p.value} value={p.value}>{p.label}</option>)}
            </select>
          </label>
          <label className="block">
            <span className="text-xs text-gray-500 mb-1 block">Dari tanggal</span>
            <input type="date" value={params.get('date_from') ?? ''} onChange={(e) => ubah({ date_from: e.target.value })} className={kelasIsian} />
          </label>
          <label className="block">
            <span className="text-xs text-gray-500 mb-1 block">Sampai tanggal</span>
            <input type="date" value={params.get('date_to') ?? ''} onChange={(e) => ubah({ date_to: e.target.value })} className={kelasIsian} />
          </label>
        </div>
        <div className="flex flex-wrap items-start gap-3">
          {params.get('user_id') ? (
            <span className="inline-flex items-center gap-1.5 rounded-full bg-[#0B2545] py-1 pl-3 pr-1.5 text-xs font-medium text-white">
              Pelaku: {params.get('pelaku') ?? 'dipilih'}
              <button type="button" onClick={() => ubah({ user_id: '', pelaku: '' })} aria-label="Hapus saringan pelaku" className="rounded-full p-0.5 hover:bg-white/20">
                <X className="h-3 w-3" />
              </button>
            </span>
          ) : (
            <div className="relative w-full sm:w-72">
              <input
                type="search"
                value={cariPelaku}
                onChange={(e) => setCariPelaku(e.target.value)}
                placeholder="Saring pelaku (ketik nama atau surel)..."
                aria-label="Saring pelaku"
                className={kelasIsian}
              />
              {cariPelakuTunda.length >= 2 && cariPelaku.trim() && (
                <ul className="absolute z-10 mt-1 w-full overflow-hidden rounded-lg border border-gray-200 bg-white shadow-lg">
                  {calonPelaku.length === 0 ? (
                    <li className="px-3 py-2 text-xs text-gray-500">Tidak ada pelaku yang cocok</li>
                  ) : calonPelaku.map((p) => (
                    <li key={p.id}>
                      <button
                        type="button"
                        onClick={() => { ubah({ user_id: p.id, pelaku: p.full_name }); setCariPelaku('') }}
                        className="w-full px-3 py-2 text-left text-sm hover:bg-gray-50"
                      >
                        <span className="block text-gray-900">{p.full_name}</span>
                        <span className="block text-xs text-gray-500">{p.email} · {labelPeran(p.role)}</span>
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </div>
          )}
          {adaSaringan && (
            <Button variant="outline" size="sm" onClick={hapusSaringan} className="text-red-600 hover:bg-red-50 hover:text-red-700">
              <X className="w-3 h-3" /> Hapus semua saringan
            </Button>
          )}
        </div>
      </section>

      {isLoading ? (
        <p className="py-12 text-center text-sm text-gray-400">Memuat data...</p>
      ) : isError ? (
        <p role="alert" className="py-12 text-center text-sm text-red-600">Audit log gagal dimuat. Muat ulang halaman ini.</p>
      ) : items.length === 0 ? (
        <div className="py-12 text-center">
          <History className="w-10 h-10 text-gray-300 mx-auto mb-2" aria-hidden="true" />
          <p className="text-sm text-gray-500">{adaSaringan ? 'Tidak ada aktivitas yang cocok dengan saringan' : 'Belum ada aktivitas'}</p>
        </div>
      ) : (
        <div className="space-y-4">
          {hari.map((h) => (
            <section key={h.tanggal} aria-labelledby={`hari-${h.tanggal}`} className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
              <h2 id={`hari-${h.tanggal}`} className="flex items-baseline justify-between border-b border-gray-100 px-5 py-3 text-sm font-semibold" style={{ color: '#0B2545' }}>
                {h.label}
                <span className="text-xs font-normal text-gray-400">{h.items.length} aktivitas</span>
              </h2>
              <ul className="divide-y divide-gray-50">
                {h.items.map((log) => <BarisAudit key={log.id} log={log} panel={panel} />)}
              </ul>
            </section>
          ))}
        </div>
      )}

      {totalPages > 1 && (
        <nav aria-label="Halaman" className="flex items-center justify-between">
          <p className="text-xs text-gray-500">
            Menampilkan {(page - 1) * PER_HALAMAN + 1} - {Math.min(page * PER_HALAMAN, total)} dari {total} aktivitas
          </p>
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" disabled={page === 1} onClick={() => ubah({ page: String(page - 1) })}>Sebelumnya</Button>
            <span className="text-xs text-gray-500">Halaman {page} dari {totalPages}</span>
            <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => ubah({ page: String(page + 1) })}>Selanjutnya</Button>
          </div>
        </nav>
      )}
    </div>
  )
}
