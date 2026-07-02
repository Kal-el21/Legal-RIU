import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Edit, Eye, FileText, Plus, Scale, Search, Trash2, Clock, ChevronRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useDeleteLegalCase, useLatestLegalCase, useLegalCases, useLegalCase } from '@/hooks/useLegalCase'
import { formatDate, formatDateTime } from '@/lib/utils'
import { getLegalCaseRouteBase } from '@/services/legal-case.service'
import type { CaseChronology, LegalCase } from '@/types'
import LegalCaseFormDialog from './components/LegalCaseFormDialog'

const CASE_TYPES = [
  { label: 'Semua Jenis', value: 'ALL' },
  { label: 'Non Litigasi', value: 'NON_LITIGASI' },
  { label: 'Perdata', value: 'PERDATA' },
  { label: 'Pidana', value: 'PIDANA' },
  { label: 'Tipekor', value: 'TIPEKOR' },
  { label: 'Arbitrase', value: 'ARBITRASE' },
  { label: 'TUN', value: 'TUN' },
]

const CASE_TYPE_LABEL: Record<string, string> = {
  NON_LITIGASI: 'Non Litigasi',
  PERDATA: 'Perdata',
  PIDANA: 'Pidana',
  TIPEKOR: 'Tipekor',
  ARBITRASE: 'Arbitrase',
  TUN: 'TUN',
}

const CASE_TYPE_COLOR: Record<string, string> = {
  NON_LITIGASI: 'bg-sky-100 text-sky-700',
  PERDATA: 'bg-emerald-100 text-emerald-700',
  PIDANA: 'bg-red-100 text-red-700',
  TIPEKOR: 'bg-orange-100 text-orange-700',
  ARBITRASE: 'bg-violet-100 text-violet-700',
  TUN: 'bg-amber-100 text-amber-700',
}

const CASE_CATEGORY_COLOR: Record<string, string> = {
  Life: 'bg-blue-100 text-blue-700',
  BPPDAN: 'bg-purple-100 text-purple-700',
  Property: 'bg-green-100 text-green-700',
  COB: 'bg-amber-100 text-amber-700',
}

export default function AdminLegalCaseListPage() {
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState('')
  const [level, setLevel] = useState('')
  const [caseType, setCaseType] = useState('ALL')
  const [dateFrom, setDateFrom] = useState('')
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingCase, setEditingCase] = useState<LegalCase | null>(null)
  const [hoveredCaseId, setHoveredCaseId] = useState<string | null>(null)

  const filters = {
    page,
    limit: 10,
    search,
    status,
    level,
    case_type: caseType === 'ALL' ? '' : caseType,
    date_from: dateFrom,
  }

  const { data, isLoading } = useLegalCases(filters)
  const { data: latest } = useLatestLegalCase()
  const deleteMutation = useDeleteLegalCase()
  const { data: hoveredCase } = useLegalCase(hoveredCaseId ?? '')
  const caseRouteBase = getLegalCaseRouteBase()

  const handleEdit = (legalCase: LegalCase) => {
    setEditingCase(legalCase)
    setDialogOpen(true)
  }

  const handleCreate = () => {
    setEditingCase(null)
    setDialogOpen(true)
  }

  const handleDelete = async (id: string) => {
    if (!window.confirm('Hapus kasus hukum ini?')) return
    await deleteMutation.mutateAsync(id)
  }

  const latestChronology = hoveredCase?.chronologies?.[0]

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <div className="flex items-center gap-2">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-[#0B2545]">
              <Scale className="h-4 w-4 text-white" />
            </div>
            <div>
              <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>Lembar Monitoring Legal</h1>
              <p className="text-sm text-gray-500 mt-0.5">Kelola kasus hukum dan kronologi sidang</p>
            </div>
          </div>
        </div>
        <Button onClick={handleCreate} className="text-white" style={{ background: '#C8102E' }}>
          <Plus className="h-4 w-4" />
          Tambah Kasus
        </Button>
      </div>

      <section className="mb-5 rounded-2xl border border-gray-100 bg-white p-5">
        <div className="mb-4 flex items-center gap-2">
          <FileText className="h-4 w-4 text-[#C8102E]" />
          <h2 className="text-sm font-semibold" style={{ color: '#0B2545' }}>Kasus Terbaru</h2>
        </div>
        {latest ? (
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
            <Info label="Tanggal" value={formatDate(latest.case_date)} />
            <Info label="Tingkat" value={latest.level || '-'} />
            <Info label="Catatan Tambahan" value={latest.additional_notes || '-'} />
          </div>
        ) : (
          <p className="text-sm text-gray-400">Belum ada kasus hukum.</p>
        )}
      </section>

      <section className="mb-5 rounded-2xl border border-gray-100 bg-white p-4">
        <div className="grid grid-cols-1 gap-3 md:grid-cols-[1fr_150px_150px_170px_150px]">
          <div className="relative">
            <Search className="absolute left-2.5 top-2 h-4 w-4 text-gray-400" />
            <Input value={search} onChange={(event) => { setSearch(event.target.value); setPage(1) }} className="pl-8" placeholder="Cari kasus..." />
          </div>
          <Input value={dateFrom} onChange={(event) => { setDateFrom(event.target.value); setPage(1) }} type="date" />
          <Input value={level} onChange={(event) => { setLevel(event.target.value); setPage(1) }} placeholder="Tingkat" />
          <Input value={status} onChange={(event) => { setStatus(event.target.value); setPage(1) }} placeholder="Status" />
          <Select value={caseType} onValueChange={(value) => { setCaseType(value); setPage(1) }}>
            <SelectTrigger className="w-full"><SelectValue /></SelectTrigger>
            <SelectContent>
              {CASE_TYPES.map((item) => <SelectItem key={item.value} value={item.value}>{item.label}</SelectItem>)}
            </SelectContent>
          </Select>
        </div>
      </section>

      <div className="overflow-hidden rounded-2xl border border-gray-100 bg-white">
        {isLoading ? (
          <div className="p-12 text-center text-gray-400">Memuat data...</div>
        ) : !data?.items?.length ? (
          <div className="p-16 text-center">
            <div className="w-16 h-16 rounded-2xl bg-gray-100 flex items-center justify-center mx-auto mb-4">
              <Scale className="w-7 h-7 text-gray-400" />
            </div>
            <p className="font-medium text-gray-500">Belum ada kasus hukum</p>
          </div>
        ) : (
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-100" style={{ background: '#f8fafc' }}>
                <th className="w-16 px-6 py-3.5 text-left text-xs font-semibold uppercase tracking-wide text-gray-500">No</th>
                <th className="px-6 py-3.5 text-left text-xs font-semibold uppercase tracking-wide text-gray-500">Nama Kasus</th>
                <th className="px-6 py-3.5 text-left text-xs font-semibold uppercase tracking-wide text-gray-500">Kategori</th>
                <th className="px-6 py-3.5 text-left text-xs font-semibold uppercase tracking-wide text-gray-500">Jenis Kasus</th>
                <th className="px-6 py-3.5 text-left text-xs font-semibold uppercase tracking-wide text-gray-500">Status</th>
                <th className="px-6 py-3.5 text-left text-xs font-semibold uppercase tracking-wide text-gray-500">Lokasi</th>
                <th className="px-6 py-3.5 text-right text-xs font-semibold uppercase tracking-wide text-gray-500">Aksi</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {data.items.map((item, index) => (
                <tr
                  key={item.id}
                  className="hover:bg-gray-50/50 relative"
                  onMouseEnter={() => setHoveredCaseId(item.id)}
                  onMouseLeave={() => setHoveredCaseId(null)}
                >
                  <td className="px-6 py-4 text-sm text-gray-500">{((page - 1) * 10) + index + 1}</td>
                  <td className="px-6 py-4">
                    <p className="max-w-[260px] truncate text-sm font-medium text-gray-900">{item.case_name}</p>
                    <p className="text-xs text-gray-400">{formatDate(item.case_date)} - {item.level}</p>
                  </td>
                  <td className="px-6 py-4">
                    <span className={`inline-flex rounded-full px-2 py-1 text-xs font-medium ${CASE_CATEGORY_COLOR[item.category] ?? 'bg-gray-100 text-gray-700'}`}>
                      {item.category}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <CaseTypeBadge type={item.case_type} />
                  </td>
                  <td className="px-6 py-4">
                    <span className="inline-flex max-w-[150px] truncate rounded-full bg-gray-100 px-2 py-1 text-xs font-medium text-gray-600">
                      {item.current_status || '-'}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <p className="max-w-[180px] truncate text-sm text-gray-500">{item.location_regency?.label ?? '-'}</p>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex justify-end gap-1">
                      <Link to={`${caseRouteBase}/${item.id}`} className="rounded-lg p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-700" title="View">
                        <Eye className="h-4 w-4" />
                      </Link>
                      <button onClick={() => handleEdit(item)} className="rounded-lg p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-700" title="Edit">
                        <Edit className="h-4 w-4" />
                      </button>
                      <button onClick={() => handleDelete(item.id)} className="rounded-lg p-2 text-gray-400 hover:bg-red-50 hover:text-red-600" title="Delete">
                        <Trash2 className="h-4 w-4" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}

        {data && data.total_pages > 1 && (
          <div className="flex items-center justify-between border-t border-gray-100 px-6 py-4">
            <p className="text-sm text-gray-500">{((page - 1) * 10) + 1}-{Math.min(page * 10, data.total)} dari {data.total}</p>
            <div className="flex gap-2">
              <Button variant="outline" size="sm" disabled={page === 1} onClick={() => setPage((current) => current - 1)}>Sebelumnya</Button>
              <Button variant="outline" size="sm" disabled={page === data.total_pages} onClick={() => setPage((current) => current + 1)}>Berikutnya</Button>
            </div>
          </div>
        )}
      </div>

      {hoveredCaseId && hoveredCase && (
        <HoverPreview
          caseData={hoveredCase}
          latestChronology={latestChronology}
          caseRouteBase={caseRouteBase}
        />
      )}

      <LegalCaseFormDialog open={dialogOpen} onOpenChange={setDialogOpen} legalCase={editingCase} />
    </div>
  )
}

function HoverPreview({ caseData, latestChronology, caseRouteBase }: { caseData: LegalCase; latestChronology?: CaseChronology; caseRouteBase: string }) {
  return (
    <div className="fixed right-6 top-1/2 z-50 w-80 -translate-y-1/2 rounded-xl border border-gray-200 bg-white shadow-xl">
      <div className="border-b border-gray-100 px-4 py-3">
        <h3 className="text-sm font-semibold" style={{ color: '#0B2545' }}>Posisi Kasus</h3>
        <p className="text-xs text-gray-500 mt-0.5">{caseData.case_name}</p>
      </div>

      <div className="p-4 space-y-4">
        <div>
          <p className="text-xs text-gray-400 mb-1">Status Terkini</p>
          <span className="inline-flex rounded-full bg-gray-100 px-2 py-1 text-xs font-medium text-gray-700">
            {caseData.current_status || 'Belum ada status'}
          </span>
        </div>

        <div>
          <p className="text-xs text-gray-400 mb-2 flex items-center gap-1">
            <Clock className="h-3 w-3" />
            Kronologi Terbaru
          </p>
          {latestChronology ? (
            <div className="space-y-3">
              <div className="flex gap-3">
                <div className="mt-1.5 h-2 w-2 shrink-0 rounded-full bg-[#C8102E]" />
                <div>
                  <p className="text-xs text-gray-400">{formatDate(latestChronology.agenda_date)}</p>
                  <p className="text-sm font-medium text-gray-800">{latestChronology.agenda}</p>
                  {latestChronology.description && (
                    <p className="mt-0.5 text-xs text-gray-500 line-clamp-2">{latestChronology.description}</p>
                  )}
                </div>
              </div>
              {caseData.chronologies && caseData.chronologies.length > 1 && (
                <div className="pl-5 space-y-2">
                  {caseData.chronologies.slice(1, 4).map((chronology) => (
                    <div key={chronology.id} className="flex gap-2 border-l-2 border-gray-100 pl-3">
                      <div>
                        <p className="text-xs text-gray-400">{formatDate(chronology.agenda_date)}</p>
                        <p className="text-xs font-medium text-gray-600">{chronology.agenda}</p>
                      </div>
                    </div>
                  ))}
                </div>
              )}
              {caseData.chronologies && caseData.chronologies.length > 4 && (
                <Link
                  to={`${caseRouteBase}/${caseData.id}`}
                  className="inline-flex items-center gap-1 text-xs text-[#C8102E] hover:underline pl-5"
                >
                  Lihat semua kronologi
                  <ChevronRight className="h-3 w-3" />
                </Link>
              )}
            </div>
          ) : (
            <p className="text-xs text-gray-400">Belum ada kronologi sidang</p>
          )}
        </div>

        <div className="pt-2 border-t border-gray-50">
          <p className="text-xs text-gray-400">
            Update terakhir: {formatDateTime(caseData.updated_at)}
          </p>
        </div>
      </div>
    </div>
  )
}

function CaseTypeBadge({ type }: { type: string }) {
  return (
    <span className={`inline-flex rounded-full px-2 py-1 text-xs font-medium ${CASE_TYPE_COLOR[type] ?? 'bg-gray-100 text-gray-700'}`}>
      {CASE_TYPE_LABEL[type] ?? type}
    </span>
  )
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-xs text-gray-400">{label}</p>
      <p className="mt-1 line-clamp-2 text-sm font-medium text-gray-800">{value}</p>
    </div>
  )
}
