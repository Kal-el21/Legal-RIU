import { useState } from 'react'
import { Link } from 'react-router-dom'
import { FileSearch } from 'lucide-react'
import { Button } from '@/components/ui/button'
import StatusBadge from '@/components/common/StatusBadge'
import { useDocumentReviews } from '@/hooks/useDocumentReview'
import { formatDate } from '@/lib/utils'
import type { SubmissionStatus } from '@/types'

const STATUS_FILTERS = [
  { label: 'Semua', value: '' },
  { label: 'Diajukan', value: 'SUBMITTED' },
  { label: 'Sedang Direview', value: 'UNDER_REVIEW' },
  { label: 'Perlu Revisi', value: 'NEED_REVISION' },
  { label: 'Ditolak', value: 'REJECTED' },
  { label: 'Diajukan Ulang', value: 'RESUBMITTED' },
  { label: 'Selesai', value: 'COMPLETED' },
]

export default function AdminReviewDocumentListPage() {
  const [status, setStatus] = useState('')
  const [page, setPage] = useState(1)
  const { data, isLoading } = useDocumentReviews({ page, limit: 10, status })

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>Manage Review Dokumen</h1>
        <p className="text-sm text-gray-500 mt-0.5">Kelola seluruh pengajuan Review Dokumen</p>
      </div>

      <div className="flex flex-wrap gap-2 mb-6">
        {STATUS_FILTERS.map((f) => (
          <button key={f.value} onClick={() => { setStatus(f.value); setPage(1) }}
            className={`px-3 py-1.5 rounded-lg text-xs font-medium transition-colors ${
              status === f.value ? 'text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
            }`}
            style={status === f.value ? { background: '#C8102E' } : {}}>
            {f.label}
          </button>
        ))}
      </div>

      <div className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
        {isLoading ? (
          <div className="p-12 text-center text-gray-400">Memuat data...</div>
        ) : !data?.items?.length ? (
          <div className="p-16 text-center">
            <div className="w-16 h-16 rounded-2xl bg-gray-100 flex items-center justify-center mx-auto mb-4">
              <FileSearch className="w-7 h-7 text-gray-400" />
            </div>
            <p className="font-medium text-gray-500">Belum ada pengajuan</p>
          </div>
        ) : (
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-100" style={{ background: '#f8fafc' }}>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Ticket</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Pemohon</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Nama Dokumen</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Jenis</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Pihak Kedua</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Status</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Tanggal</th>
                <th className="px-6 py-3.5" />
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {data.items.map((dr) => (
                <tr key={dr.id} className="hover:bg-gray-50/50 transition-colors">
                  <td className="px-6 py-4">
                    <span className="text-xs font-mono font-medium text-gray-600 bg-gray-100 px-2 py-1 rounded">
                      {dr.ticket_number}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <p className="text-sm font-medium text-gray-900">{dr.requestor_name}</p>
                    <p className="text-xs text-gray-400">{dr.requestor_division}</p>
                  </td>
                  <td className="px-6 py-4">
                    <p className="text-sm text-gray-700 max-w-[150px] truncate">{dr.document_name}</p>
                  </td>
                  <td className="px-6 py-4">
                    <p className="text-sm text-gray-500 max-w-[120px] truncate">{dr.document_type}</p>
                  </td>
                  <td className="px-6 py-4">
                    <p className="text-sm text-gray-500 max-w-[120px] truncate">{dr.second_party}</p>
                  </td>
                  <td className="px-6 py-4">
                    <StatusBadge status={dr.status as SubmissionStatus} />
                  </td>
                  <td className="px-6 py-4">
                    <p className="text-sm text-gray-500">{formatDate(dr.created_at)}</p>
                  </td>
                  <td className="px-6 py-4 text-right">
                    <Link to={`/admin/review-documents/${dr.id}`}
                      className="text-xs font-medium hover:underline" style={{ color: '#C8102E' }}>
                      Kelola →
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
        {data && data.total_pages > 1 && (
          <div className="px-6 py-4 border-t border-gray-100 flex items-center justify-between">
            <p className="text-sm text-gray-500">{((page-1)*10)+1}–{Math.min(page*10, data.total)} dari {data.total}</p>
            <div className="flex gap-2">
              <Button variant="outline" size="sm" disabled={page===1} onClick={() => setPage(p=>p-1)}>Sebelumnya</Button>
              <Button variant="outline" size="sm" disabled={page===data.total_pages} onClick={() => setPage(p=>p+1)}>Berikutnya</Button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}