import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Search, Plus, Edit, Eye } from 'lucide-react'
import { agreementService, type AgreementDocument } from '@/services/agreement-document.service'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import StatusBadge from '@/components/common/StatusBadge'
import { formatDate, STATUS_LABEL } from '@/lib/utils'
import type { SubmissionStatus } from '@/types'

const STATUS_OPTIONS = ['ALL', 'SUBMITTED', 'UNDER_REVIEW', 'NEED_REVISION', 'REJECTED', 'RESUBMITTED', 'COMPLETED']

export default function AgreementDocumentList({ basePath, apiBase = '', requester = false }: { basePath: string; apiBase?: string; requester?: boolean }) {
  const [items, setItems] = useState<AgreementDocument[] | null>(null)
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState('ALL')
  const [dateFrom, setDateFrom] = useState('')

  useEffect(() => {
    void agreementService.list(apiBase, { search, status: status === 'ALL' ? '' : status, date_from: dateFrom }).then(setItems)
  }, [apiBase, search, status, dateFrom])

  return <div className="p-6 max-w-7xl mx-auto">
    <div className="flex items-center justify-between mb-6">
      <div>
        <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>Dokumen Perjanjian</h1>
        <p className="text-sm text-gray-500 mt-0.5">Pengajuan dan pembuatan dokumen perjanjian</p>
      </div>
      {requester && <Link to={`${basePath}/new`}><Button className="flex items-center gap-2 text-white transition hover:brightness-95" style={{ background: '#C8102E' }}><Plus className="w-4 h-4" /> Pengajuan Baru</Button></Link>}
    </div>

    <section className="mb-5 rounded-2xl border border-gray-100 bg-white p-4">
      <div className="grid grid-cols-1 gap-3 md:grid-cols-[1fr_180px_180px]">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <Input placeholder="Cari nama pemohon..." className="pl-9" value={search} onChange={(e) => setSearch(e.target.value)} />
        </div>
        <Input type="date" value={dateFrom} onChange={(e) => setDateFrom(e.target.value)} />
        <Select value={status} onValueChange={setStatus}>
          <SelectTrigger className="w-full"><SelectValue placeholder="Semua Status" /></SelectTrigger>
          <SelectContent>
            {STATUS_OPTIONS.map((s) => (
              <SelectItem key={s} value={s}>{s === 'ALL' ? 'Semua Status' : STATUS_LABEL[s as SubmissionStatus] ?? s}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
    </section>

    <div className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
      {items === null ? (
        <div className="p-12 text-center text-gray-400">Memuat data...</div>
      ) : !items.length ? (
        <div className="p-16 text-center">
          <p className="font-medium text-gray-500">Belum ada pengajuan.</p>
        </div>
      ) : (
        <table className="w-full">
          <thead>
            <tr className="border-b border-gray-100" style={{ background: '#f8fafc' }}>
              <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Ticket</th>
              <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Tipe</th>
              <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Nomor</th>
              <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Diajukan Oleh</th>
              <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Tanggal</th>
              <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Status</th>
              <th className="text-right px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Aksi</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-50">
            {items.map((item) => (
              <tr key={item.id} className="hover:bg-gray-50/50 transition-colors">
                <td className="px-6 py-4 text-sm text-gray-900">{item.ticket_number}</td>
                <td className="px-6 py-4 text-sm text-gray-700">{item.document_type_code}</td>
                <td className="px-6 py-4 text-sm text-gray-700">{item.agreement_number}</td>
                <td className="px-6 py-4 text-sm text-gray-700">{item.requester?.full_name || '-'}</td>
                <td className="px-6 py-4 text-sm text-gray-700">{item.created_at ? formatDate(item.created_at) : '-'}</td>
                <td className="px-6 py-4"><StatusBadge status={item.status as SubmissionStatus} /></td>
                <td className="px-6 py-4">
                  <div className="flex items-center gap-1 justify-end">
                    <Link to={`${basePath}/${item.id}`} title="Detail" className="p-1.5 rounded-lg hover:bg-gray-100 transition-colors text-gray-400 hover:text-gray-700"><Eye className="w-4 h-4" /></Link>
                    {requester && (item.status === 'SUBMITTED' || item.status === 'NEED_REVISION') && <Link to={`${basePath}/${item.id}/edit`} title={item.status === 'NEED_REVISION' ? 'Revisi' : 'Edit'} className="p-1.5 rounded-lg hover:bg-gray-100 transition-colors text-gray-400 hover:text-gray-700"><Edit className="w-4 h-4" /></Link>}
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  </div>
}
