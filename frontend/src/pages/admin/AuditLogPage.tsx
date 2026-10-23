import { useCallback, useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useLocation, useNavigate, useSearchParams } from 'react-router-dom'
import { Filter, History, Search, X } from 'lucide-react'
import { auditLogService } from '@/services/audit-log.service'
import { cn } from '@/lib/utils'
import { AKSI, kelompokPerHari, labelAksi, labelEntitas, labelPeran, PILIHAN_ENTITAS } from '@/lib/audit'
import { DataAudit, IsiAktivitas, Pelaku, WaktuDanPerangkat } from '@/components/shared/BarisAudit'
import type { AuditAction } from '@/types'
import { useAuthStore } from '@/store/auth.store'
import { getRoleHome } from '@/routes/role-home'
import { Button } from '@/components/ui/button'

const SARINGAN = ['action', 'entity_type', 'user_id', 'pelaku', 'search', 'date_from', 'date_to'] as const
// Saringan di panel Filter; aksi yang tidak ada di chip cepat juga dipilih dari panel.
const SARINGAN_PANEL = ['entity_type', 'user_id', 'date_from', 'date_to'] as const
const AKSI_CEPAT: AuditAction[] = ['STATUS_CHANGE', 'FILE_UPLOAD', 'USER_UPDATE', 'DELETE', 'LOGIN']
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
const kelasTh = 'text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide'
// Tanggal saringan (YYYY-MM-DD) dibaca sebagai tanggal itu sendiri, tanpa geser zona waktu.
const tanggalSaringan = new Intl.DateTimeFormat('id-ID', { timeZone: 'UTC', day: 'numeric', month: 'long', year: 'numeric' })

function ChipSaringan({ label, onHapus }: { label: string; onHapus: () => void }) {
  return (
    <span className="inline-flex items-center gap-1.5 rounded-full bg-[#0B2545] py-1 pl-3 pr-1.5 text-xs font-medium text-white">
      {label}
      <button type="button" onClick={onHapus} aria-label={`Hapus saringan ${label}`} className="rounded-full p-0.5 hover:bg-white/20">
        <X className="h-3 w-3" />
      </button>
    </span>
  )
}

export default function AuditLogPage() {
  const navigate = useNavigate()
  const { pathname } = useLocation()
  const hasPermission = useAuthStore((state) => state.hasPermission)
  const role = useAuthStore((state) => state.user?.role)
  const [params, setParams] = useSearchParams()
  const panel = '/' + pathname.split('/')[1]
  const [bukaFilter, setBukaFilter] = useState(false)

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
  const aksi = params.get('action') ?? ''
  const aksiLain = aksi !== '' && !AKSI_CEPAT.includes(aksi as AuditAction)
  const jumlahSaringanPanel = SARINGAN_PANEL.filter((k) => params.get(k)).length + (aksiLain ? 1 : 0)

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

  // Ringkasan saringan panel saat panel tertutup, supaya tetap terlihat apa yang sedang berlaku.
  const chipAktif: { label: string; hapus: Record<string, string> }[] = []
  const jenis = params.get('entity_type')
  const dari = params.get('date_from')
  const sampai = params.get('date_to')
  if (aksiLain) chipAktif.push({ label: `Aksi: ${labelAksi(aksi)}`, hapus: { action: '' } })
  if (jenis) chipAktif.push({ label: `Jenis data: ${labelEntitas(jenis)}`, hapus: { entity_type: '' } })
  if (params.get('user_id')) chipAktif.push({ label: `Pelaku: ${params.get('pelaku') ?? 'dipilih'}`, hapus: { user_id: '', pelaku: '' } })
  if (dari) chipAktif.push({ label: `Dari ${tanggalSaringan.format(new Date(dari))}`, hapus: { date_from: '' } })
  if (sampai) chipAktif.push({ label: `Sampai ${tanggalSaringan.format(new Date(sampai))}`, hapus: { date_to: '' } })

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="mb-6 flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>Audit Log</h1>
          <p className="text-sm text-gray-500 mt-0.5">Siapa melakukan apa, dan kapan</p>
        </div>
        <Button variant="outline" onClick={() => setBukaFilter((b) => !b)} aria-expanded={bukaFilter} aria-controls="panel-filter" className="flex items-center gap-2">
          <Filter className="w-4 h-4" aria-hidden="true" />
          Filter
          {jumlahSaringanPanel > 0 && (
            <span className="rounded-full px-1.5 text-xs font-semibold text-white" style={{ background: '#C8102E' }}>{jumlahSaringanPanel}</span>
          )}
        </Button>
      </div>

      <div className="mb-4 flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <div className="flex flex-wrap gap-2" role="group" aria-label="Saring aksi">
          {[{ value: '', label: 'Semua' }, ...AKSI_CEPAT.map((a) => ({ value: a as string, label: labelAksi(a) }))].map((f) => (
            <button
              key={f.value || 'semua'}
              type="button"
              onClick={() => ubah({ action: f.value })}
              aria-pressed={aksi === f.value}
              className={cn('px-3 py-1.5 rounded-lg text-xs font-medium transition-colors', aksi === f.value ? 'text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200')}
              style={aksi === f.value ? { background: '#C8102E' } : undefined}
            >
              {f.label}
            </button>
          ))}
        </div>
        <div className="relative w-full lg:w-80">
          <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" aria-hidden="true" />
          <input
            type="search"
            value={cari}
            onChange={(e) => setCari(e.target.value)}
            placeholder="Cari keterangan, nilai, atau nama pelaku..."
            aria-label="Cari"
            className={cn(kelasIsian, 'pl-9 bg-white')}
          />
        </div>
      </div>

      {bukaFilter && (
        <section id="panel-filter" aria-label="Filter" className="mb-4 bg-white rounded-2xl border border-gray-100 p-5 space-y-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            <label className="block">
              <span className="text-xs text-gray-500 mb-1 block">Aksi</span>
              <select value={aksi} onChange={(e) => ubah({ action: e.target.value })} className={kelasIsian}>
                <option value="">Semua aksi</option>
                {Object.entries(AKSI).map(([value, a]) => <option key={value} value={value}>{a.label}</option>)}
              </select>
            </label>
            <label className="block">
              <span className="text-xs text-gray-500 mb-1 block">Jenis data</span>
              <select value={jenis ?? ''} onChange={(e) => ubah({ entity_type: e.target.value })} className={kelasIsian}>
                <option value="">Semua jenis</option>
                {PILIHAN_ENTITAS.map((p) => <option key={p.value} value={p.value}>{p.label}</option>)}
              </select>
            </label>
            <label className="block">
              <span className="text-xs text-gray-500 mb-1 block">Dari tanggal</span>
              <input type="date" value={dari ?? ''} onChange={(e) => ubah({ date_from: e.target.value })} className={kelasIsian} />
            </label>
            <label className="block">
              <span className="text-xs text-gray-500 mb-1 block">Sampai tanggal</span>
              <input type="date" value={sampai ?? ''} onChange={(e) => ubah({ date_to: e.target.value })} className={kelasIsian} />
            </label>
          </div>
          <div className="flex flex-wrap items-end gap-3">
            {params.get('user_id') ? (
              <div>
                <span className="text-xs text-gray-500 mb-1 block">Pelaku</span>
                <ChipSaringan label={params.get('pelaku') ?? 'dipilih'} onHapus={() => ubah({ user_id: '', pelaku: '' })} />
              </div>
            ) : (
              <div className="relative w-full sm:w-72">
                <label htmlFor="saring-pelaku" className="text-xs text-gray-500 mb-1 block">Pelaku</label>
                <input
                  id="saring-pelaku"
                  type="search"
                  value={cariPelaku}
                  onChange={(e) => setCariPelaku(e.target.value)}
                  placeholder="Ketik nama atau surel..."
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
      )}

      {!bukaFilter && adaSaringan && (
        <div className="mb-4 flex flex-wrap items-center gap-2">
          {chipAktif.map((c) => <ChipSaringan key={c.label} label={c.label} onHapus={() => ubah(c.hapus)} />)}
          <button type="button" onClick={hapusSaringan} className="text-xs font-medium text-red-600 hover:underline">Hapus semua saringan</button>
        </div>
      )}

      <div className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
        {isLoading ? (
          <div className="p-12 text-center text-gray-400">Memuat data...</div>
        ) : isError ? (
          <p role="alert" className="p-12 text-center text-sm text-red-600">Audit log gagal dimuat. Muat ulang halaman ini.</p>
        ) : items.length === 0 ? (
          <div className="p-16 text-center">
            <div className="w-16 h-16 rounded-2xl bg-gray-100 flex items-center justify-center mx-auto mb-4">
              <History className="w-7 h-7 text-gray-400" aria-hidden="true" />
            </div>
            <p className="font-medium text-gray-500">{adaSaringan ? 'Tidak ada aktivitas yang cocok dengan saringan' : 'Belum ada aktivitas'}</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full min-w-[760px]">
              <caption className="sr-only">Daftar aktivitas, dikelompokkan per hari</caption>
              <thead>
                <tr className="border-b border-gray-100" style={{ background: '#f8fafc' }}>
                  <th scope="col" className={cn(kelasTh, 'w-32')}>Waktu</th>
                  <th scope="col" className={cn(kelasTh, 'w-52')}>Pelaku</th>
                  <th scope="col" className={kelasTh}>Aktivitas</th>
                  <th scope="col" className={cn(kelasTh, 'w-44')}>Data</th>
                </tr>
              </thead>
              {hari.map((h) => (
                <tbody key={h.tanggal} className="divide-y divide-gray-50">
                  <tr className="bg-gray-50/70">
                    <th scope="colgroup" colSpan={4} className="px-6 py-2 text-left text-xs font-semibold" style={{ color: '#0B2545' }}>
                      {h.label}
                      <span className="font-normal text-gray-400"> · {h.items.length} aktivitas</span>
                    </th>
                  </tr>
                  {h.items.map((log) => (
                    <tr key={log.id} className="align-top hover:bg-gray-50/50 transition-colors">
                      <td className="px-6 py-4"><WaktuDanPerangkat log={log} /></td>
                      <td className="px-6 py-4"><Pelaku log={log} /></td>
                      <td className="px-6 py-4"><IsiAktivitas log={log} /></td>
                      <td className="px-6 py-4"><DataAudit log={log} panel={panel} /></td>
                    </tr>
                  ))}
                </tbody>
              ))}
            </table>
          </div>
        )}

        {totalPages > 1 && (
          <nav aria-label="Halaman" className="px-6 py-4 border-t border-gray-100 flex items-center justify-between">
            <p className="text-sm text-gray-500">
              {(page - 1) * PER_HALAMAN + 1}–{Math.min(page * PER_HALAMAN, total)} dari {total}
            </p>
            <div className="flex gap-2">
              <Button variant="outline" size="sm" disabled={page === 1} onClick={() => ubah({ page: String(page - 1) })}>Sebelumnya</Button>
              <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => ubah({ page: String(page + 1) })}>Berikutnya</Button>
            </div>
          </nav>
        )}
      </div>
    </div>
  )
}
