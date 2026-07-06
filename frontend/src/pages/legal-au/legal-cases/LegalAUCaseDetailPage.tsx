import { useParams, Link } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useLegalCase } from '@/hooks/useLegalCase'
import { formatDateTime } from '@/lib/utils'
import CaseChronologySection from '@/pages/admin/legal-cases/components/CaseChronologySection'

export default function LegalAUCaseDetailPage() {
  const { id } = useParams<{ id: string }>()
  const { data: legalCase, isLoading } = useLegalCase(id || '')

  if (isLoading) {
    return <div className="p-6 text-center text-gray-400">Memuat data...</div>
  }

  if (!legalCase) {
    return <div className="p-6 text-center text-gray-500">Kasus tidak ditemukan</div>
  }

  return (
    <div className="p-6 max-w-5xl mx-auto">
      <div className="mb-6">
        <Link to="/legal-au/cases">
          <Button variant="ghost" className="mb-4 -ml-2 text-gray-600 hover:text-gray-900">
            <ArrowLeft className="w-4 h-4 mr-2" /> Kembali
          </Button>
        </Link>
        <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>{legalCase.case_name}</h1>
        <p className="text-sm text-gray-500 mt-1">{legalCase.company?.name}</p>
      </div>

      <div className="bg-white rounded-2xl border border-gray-100 p-6 mb-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Detail Kasus</h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">Jenis Kasus</p>
            <p className="text-sm text-gray-900 mt-1">{legalCase.case_type?.label || '-'}</p>
          </div>
          <div>
            <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">Kategori</p>
            <p className="text-sm text-gray-900 mt-1">{legalCase.category?.label || '-'}</p>
          </div>
          <div>
            <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">Status</p>
            <p className="text-sm text-gray-900 mt-1">{legalCase.current_status || '-'}</p>
          </div>
          <div>
            <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">Tanggal</p>
            <p className="text-sm text-gray-900 mt-1">{legalCase.case_date ? new Date(legalCase.case_date).toLocaleDateString('id-ID') : '-'}</p>
          </div>
          <div>
            <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">Tingkat</p>
            <p className="text-sm text-gray-900 mt-1">{legalCase.level || '-'}</p>
          </div>
          <div>
            <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">Pihak Terkait</p>
            <p className="text-sm text-gray-900 mt-1">{legalCase.related_party?.name || '-'}</p>
          </div>
        </div>

        {legalCase.case_summary && (
          <div className="mt-4">
            <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">Ringkasan</p>
            <p className="text-sm text-gray-700 mt-1">{legalCase.case_summary}</p>
          </div>
        )}

        {legalCase.specification && (
          <div className="mt-4">
            <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">Spesifikasi</p>
            <p className="text-sm text-gray-700 mt-1">{legalCase.specification}</p>
          </div>
        )}
      </div>

      {legalCase.current_status && (
        <div className="bg-white rounded-2xl border border-gray-100 p-6 mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Posisi Kasus</h2>
          <div className="flex items-start gap-3">
            <div className="mt-1 h-3 w-3 rounded-full bg-[#C8102E]" />
            <div>
              <p className="text-sm font-semibold text-gray-900">{legalCase.current_status || 'Belum ada status'}</p>
              {legalCase.status_updated_at && (
                <p className="mt-1 text-xs text-gray-400">Status terakhir diubah: {formatDateTime(legalCase.status_updated_at)}</p>
              )}
            </div>
          </div>
        </div>
      )}

      <CaseChronologySection caseId={id!} />
    </div>
  )
}
