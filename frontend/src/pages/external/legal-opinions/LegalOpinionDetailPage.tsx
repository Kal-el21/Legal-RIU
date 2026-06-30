import { useParams, useNavigate } from 'react-router-dom'
import { ArrowLeft, Download, FileText } from 'lucide-react'
import { Button } from '@/components/ui/button'
import StatusBadge from '@/components/common/StatusBadge'
import { useLegalOpinion } from '@/hooks/useLegalOpinion'
import { useAuthStore } from '@/store/auth.store'
import { formatDateTime, formatFileSize } from '@/lib/utils'
import { legalOpinionService } from '@/services/legal-opinion.service'
import type { SubmissionStatus } from '@/types'

export default function ExternalLegalOpinionDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)
  const isExternal = user?.role === 'EXTERNAL'

  const { data: lo, isLoading } = useLegalOpinion(id!)

  const handleDownload = async (filePath: string) => {
    const { blob, filename } = await legalOpinionService.downloadFile(filePath)
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    a.click()
    URL.revokeObjectURL(url)
  }

  if (isLoading) return <div className="p-12 text-center text-gray-400">Memuat data...</div>
  if (!lo) return <div className="p-12 text-center text-gray-500">Pengajuan tidak ditemukan</div>

  const status = lo.status as SubmissionStatus
  const canDownload = isExternal && status === 'COMPLETED' && (lo.results?.length ?? 0) > 0

  return (
    <div className="p-6 max-w-3xl mx-auto">
      {/* Header */}
      <div className="flex items-start gap-3 mb-8">
        <button onClick={() => navigate(-1)} className="p-2 rounded-lg hover:bg-gray-100 mt-0.5">
          <ArrowLeft className="w-5 h-5 text-gray-600" />
        </button>
        <div className="flex-1">
          <div className="flex items-center gap-3 flex-wrap">
            <span className="text-xs font-mono font-medium bg-gray-100 text-gray-600 px-2 py-1 rounded">
              {lo.ticket_number}
            </span>
            <StatusBadge status={status} />
          </div>
          <h1 className="text-xl font-bold mt-2" style={{ color: '#0B2545' }}>{lo.title}</h1>
        </div>
      </div>

      <div className="space-y-5">
        {/* Admin note */}
        {lo.admin_note && (
          <div className="p-4 rounded-xl border-l-4 bg-orange-50 border-orange-400">
            <div className="flex items-center gap-2 mb-1">
              <FileText className="w-4 h-4 text-orange-500" />
              <p className="text-xs font-semibold text-orange-600 uppercase tracking-wide">Catatan Admin</p>
            </div>
            <p className="text-sm text-orange-800">{lo.admin_note}</p>
          </div>
        )}

        {/* Info pemohon */}
        <Card title="Informasi Pemohon">
          <Grid>
            <InfoRow label="Nama Lengkap" value={lo.requestor_name} />
            <InfoRow label="Posisi Jabatan" value={lo.requestor_position} />
            <InfoRow label="Divisi" value={lo.requestor_division} />
            <InfoRow label="Email" value={lo.requestor_email} />
            <InfoRow label="Nomor WhatsApp" value={lo.requestor_phone} />
            <InfoRow label="Tanggal Pengajuan" value={formatDateTime(lo.created_at)} />
          </Grid>
        </Card>

        {/* Detail permasalahan */}
        <Card title="Detail Permasalahan">
          <div className="space-y-4">
            <InfoRow label="Jenis Kajian" value={lo.legal_type === 'Lain-Lain' && lo.legal_type_other ? lo.legal_type_other : lo.legal_type} />
            <div>
              <p className="text-xs font-medium text-gray-500 mb-1">Kronologis</p>
              <p className="text-sm text-gray-800 whitespace-pre-wrap leading-relaxed">{lo.chronology}</p>
            </div>
            <div>
              <p className="text-xs font-medium text-gray-500 mb-1">Pertanyaan</p>
              <p className="text-sm text-gray-800 whitespace-pre-wrap leading-relaxed">{lo.question}</p>
            </div>
          </div>
        </Card>

        {/* Attachments */}
        {(lo.attachments?.length ?? 0) > 0 && (
          <Card title="Dokumen Pendukung">
            <div className="space-y-2">
              {lo.attachments?.map((att) => (
                <div key={att.id} className="flex items-center gap-3 p-3 rounded-lg bg-gray-50 border border-gray-100">
                  <FileText className="w-4 h-4 text-gray-400 flex-shrink-0" />
                  <div className="flex-1 min-w-0">
                    <p className="text-sm text-gray-700 truncate">{att.file_name}</p>
                    <p className="text-xs text-gray-400">{formatFileSize(att.file_size)} · Round {att.upload_round}</p>
                  </div>
                  <button onClick={() => handleDownload(att.file_path)}
                    className="p-1.5 rounded-lg hover:bg-gray-200 transition-colors text-gray-400 hover:text-gray-600">
                    <Download className="w-4 h-4" />
                  </button>
                </div>
              ))}
            </div>
          </Card>
        )}

        {/* Results — for COMPLETED */}
        {canDownload && (
          <Card title="Hasil Kajian">
            <div className="space-y-2">
              {lo.results?.map((res) => (
                <div key={res.id} className="flex items-center gap-3 p-3 rounded-lg bg-green-50 border border-green-100">
                  <FileText className="w-4 h-4 text-green-500 flex-shrink-0" />
                  <div className="flex-1 min-w-0">
                    <p className="text-sm text-gray-700 truncate">{res.file_name}</p>
                    {res.notes && <p className="text-xs text-gray-500 mt-0.5">{res.notes}</p>}
                  </div>
                  <Button size="sm" onClick={() => handleDownload(res.file_path)}
                    className="flex items-center gap-1.5 text-white text-xs" style={{ background: '#0B2545' }}>
                    <Download className="w-3.5 h-3.5" /> Unduh
                  </Button>
                </div>
              ))}
            </div>
          </Card>
        )}
      </div>
    </div>
  )
}

function Card({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="bg-white rounded-2xl border border-gray-100 p-6">
      <h3 className="text-sm font-semibold mb-4" style={{ color: '#0B2545' }}>{title}</h3>
      {children}
    </div>
  )
}

function Grid({ children }: { children: React.ReactNode }) {
  return <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">{children}</div>
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-xs font-medium text-gray-500 mb-0.5">{label}</p>
      <p className="text-sm text-gray-800">{value}</p>
    </div>
  )
}