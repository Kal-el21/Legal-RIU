import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { Plus, FileText } from 'lucide-react'
import { Button } from '@/components/ui/button'
import StatusBadge from '@/components/common/StatusBadge'
import { formatDate } from '@/lib/utils'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/store/auth.store'
import { getRoleHome } from '@/routes/guards'
import type { SubmissionStatus } from '@/types'
import { agreementDocumentService } from '@/services/agreement-document.service'

const STATUS_FILTERS = [
  { label: 'Semua', value: '' },
  { label: 'Diajukan', value: 'SUBMITTED' },
  { label: 'Sedang Direview', value: 'UNDER_REVIEW' },
  { label: 'Perlu Revisi', value: 'NEED_REVISION' },
  { label: 'Ditolak', value: 'REJECTED' },
  { label: 'Diajukan Ulang', value: 'RESUBMITTED' },
  { label: 'Selesai', value: 'COMPLETED' },
]

interface Props {
  basePath: string
  title: string
  description: string
  showCreateButton?: boolean
  createPath?: string
  linkLabel?: string
  viewPermission?: string
  createPermission?: string
}

export default function SharedAgreementDocumentListPage({
  basePath,
  title,
  description,
  showCreateButton = false,
  createPath,
  linkLabel = 'Kelola →',
  viewPermission,
  createPermission,
}: Props) {
  const [status, setStatus] = useState('')
  const [page, setPage] = useState(1)
  const [items, setItems] = useState<any[]>([])
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [isLoading, setIsLoading] = useState(true)
  const navigate = useNavigate()
  const hasPermission = useAuthStore((state) => state.hasPermission)
  const role = useAuthStore((state) => state.user?.role)
  const canView = viewPermission ? hasPermission(viewPermission) : true
  const canCreate = createPermission ? hasPermission(createPermission) : false

  useEffect(() => {
    if (!canView) {
      navigate(getRoleHome(role), { replace: true })
    }
  }, [canView, navigate, role])

  useEffect(() => {
    let active = true
    setIsLoading(true)
    agreementDocumentService
      .getAll({ page, limit: 10, status })
      .then((res) => {
        if (!active) return
        setItems(res.items)
        setTotal(res.total)
        setTotalPages(res.total_pages)
      })
      .finally(() => active && setIsLoading(false))
    return () => {
      active = false
    }
  }, [page, status])

  const secondParty = (dr: any) =>
    (dr.form_data && (dr.form_data.pihak_kedua_nama as string)) || '-'

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="mb-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>{title}</h1>
            <p className="text-sm text-gray-500 mt-0.5">{description}</p>
          </div>
          {showCreateButton && canCreate && createPath && (
            <Link to={createPath}>
              <Button className="flex items-center gap-2 text-white" style={{ background: '#C8102E' }}>
                <Plus className="w-4 h-4" /> Buat Pengajuan
              </Button>
            </Link>
          )}
        </div>
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
        ) : !items.length ? (
          <div className="p-16 text-center">
            <div className="w-16 h-16 rounded-2xl bg-gray-100 flex items-center justify-center mx-auto mb-4">
              <FileText className="w-7 h-7 text-gray-400" />
            </div>
            <p className="font-medium text-gray-500">Belum ada pengajuan dokumen perjanjian</p>
          </div>
        ) : (
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-100" style={{ background: '#f8fafc' }}>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Ticket</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Jenis Pekerjaan</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Pihak Kedua</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Status</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Tanggal</th>
                <th className="px-6 py-3.5" />
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {items.map((dr) => (
                <tr key={dr.id} className="hover:bg-gray-50/50 transition-colors">
                  <td className="px-6 py-4">
                    <span className="text-xs font-mono font-medium text-gray-600 bg-gray-100 px-2 py-1 rounded">
                      {dr.ticket_number}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <p className="text-sm text-gray-700 max-w-[200px] truncate">{dr.form_data?.jenis_pekerjaan || '-'}</p>
                  </td>
                  <td className="px-6 py-4">
                    <p className="text-sm text-gray-500 max-w-[160px] truncate">{secondParty(dr)}</p>
                  </td>
                  <td className="px-6 py-4">
                    <StatusBadge status={dr.status as SubmissionStatus} />
                  </td>
                  <td className="px-6 py-4">
                    <p className="text-sm text-gray-500">{formatDate(dr.created_at)}</p>
                  </td>
                  <td className="px-6 py-4 text-right">
                    <Link to={`${basePath}/${dr.id}`}
                      className="text-xs font-medium hover:underline" style={{ color: '#C8102E' }}>
                      {linkLabel}
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
        {totalPages > 1 && (
          <div className="px-6 py-4 border-t border-gray-100 flex items-center justify-between">
            <p className="text-sm text-gray-500">{((page-1)*10)+1}–{Math.min(page*10, total)} dari {total}</p>
            <div className="flex gap-2">
              <Button variant="outline" size="sm" disabled={page===1} onClick={() => setPage(p=>p-1)}>Sebelumnya</Button>
              <Button variant="outline" size="sm" disabled={page===totalPages} onClick={() => setPage(p=>p+1)}>Berikutnya</Button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
