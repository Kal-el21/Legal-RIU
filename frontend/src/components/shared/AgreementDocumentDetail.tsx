import { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { ArrowLeft, Loader2, Download, AlertCircle, FileText, Pencil } from 'lucide-react'
import { Button } from '@/components/ui/button'
import StatusBadge from '@/components/common/StatusBadge'
import { agreementDocumentService } from '@/services/agreement-document.service'
import type { AgreementDocument, SubmissionStatus } from '@/types'

interface Props {
  basePath: string
  canEdit?: boolean
}

export default function SharedAgreementDocumentDetail({ basePath, canEdit = false }: Props) {
  const { id } = useParams()
  const navigate = useNavigate()
  const [doc, setDoc] = useState<AgreementDocument | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!id) return
    setLoading(true)
    agreementDocumentService
      .getByID(id)
      .then(setDoc)
      .finally(() => setLoading(false))
  }, [id])

  const downloadFinal = async () => {
    if (!id) return
    const url = await agreementDocumentService.getFinalURL(id)
    window.open(url, '_blank')
  }

  if (loading) {
    return (
      <div className="p-6 max-w-4xl mx-auto flex items-center gap-2 text-gray-400">
        <Loader2 className="w-4 h-4 animate-spin" /> Memuat...
      </div>
    )
  }
  if (!doc) return <div className="p-6 text-gray-400">Dokumen tidak ditemukan.</div>

  const fd = doc.form_data || {}
  const isRevision = doc.status === 'NEED_REVISION' || doc.status === 'REJECTED'

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <Link to={basePath} className="inline-flex items-center gap-1.5 text-sm text-gray-500 hover:text-gray-700 mb-4">
        <ArrowLeft className="w-4 h-4" /> Kembali
      </Link>

      <div className="flex items-center justify-between mb-2">
        <div>
          <p className="text-xs font-mono text-gray-400">{doc.ticket_number}</p>
          <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>{fd.jenis_pekerjaan || 'Dokumen Perjanjian'}</h1>
        </div>
        <StatusBadge status={doc.status as SubmissionStatus} />
      </div>

      {isRevision && doc.admin_note && (
        <div className="my-4 p-4 rounded-xl bg-yellow-50 border border-yellow-200 flex gap-3">
          <AlertCircle className="w-5 h-5 text-yellow-600 flex-shrink-0 mt-0.5" />
          <div>
            <p className="text-sm font-semibold text-yellow-800">Catatan Revisi dari Approver</p>
            <p className="text-sm text-yellow-700 whitespace-pre-wrap">{doc.admin_note}</p>
            <Button className="mt-2 text-white" style={{ background: '#C8102E' }} size="sm"
              onClick={() => navigate(`${basePath}/${id}/edit`)}>
              <Pencil className="w-4 h-4" /> Perbaiki & Kirim Ulang
            </Button>
          </div>
        </div>
      )}

      <div className="bg-white rounded-2xl border border-gray-100 p-6 space-y-6">
        <Block title="Pihak Kedua">
          <Row label="Nama Perusahaan" value={fd.pihak_kedua_nama} />
          <Row label="Bidang Usaha" value={fd.pihak_kedua_bidang} />
          <Row label="Alamat" value={fd.pihak_kedua_alamat} />
          <Row label="Telepon" value={fd.pihak_kedua_telepon} />
          <Row label="Email" value={fd.pihak_kedua_email} />
          <Row label="PIC" value={fd.pihak_kedua_pic} />
          <Row label="Pejabat" value={fd.pihak_kedua_pejabat} />
          <Row label="Jabatan" value={fd.pihak_kedua_jabatan} />
        </Block>

        <Block title="Ketentuan Khusus">
          <Row label="Jenis Pekerjaan" value={fd.jenis_pekerjaan} />
          <Row label="Ruang Lingkup" value={fd.ruang_lingkup} />
          <Row label="Jangka Waktu" value={`${fd.jangka_waktu_mulai || '-'} s.d. ${fd.jangka_waktu_selesai || '-'}`} />
          <Row label="Nilai Kontrak" value={String(fd.nilai_kontrak || '')} />
          <Row label="Termin 1" value={`${fd.termin1_persen || ''}% / Rp ${fd.termin1_nilai || ''}`} />
          <Row label="Termin 2" value={`${fd.termin2_persen || ''}% / Rp ${fd.termin2_nilai || ''}`} />
          <Row label="Bank" value={fd.bank} />
          <Row label="No. Rekening" value={fd.nomor_rekening} />
          <Row label="Atas Nama" value={fd.atas_nama} />
        </Block>

        <Block title="Lampiran">
          {doc.attachments && doc.attachments.length > 0 ? (
            <ul className="space-y-2">
              {doc.attachments.map((a) => (
                <li key={a.id} className="flex items-center gap-2 text-sm text-gray-600">
                  <FileText className="w-4 h-4 text-gray-400" /> {a.file_name}
                </li>
              ))}
            </ul>
          ) : (
            <p className="text-sm text-gray-400">Tidak ada lampiran.</p>
          )}
        </Block>
      </div>

      <div className="mt-6 flex gap-3">
        {doc.status === 'COMPLETED' && (
          <Button className="flex items-center gap-2 text-white" style={{ background: '#C8102E' }} onClick={downloadFinal}>
            <Download className="w-4 h-4" /> Unduh Dokumen Final (PDF)
          </Button>
        )}
        {canEdit && (doc.status === 'SUBMITTED' || isRevision) && (
          <Button variant="outline" onClick={() => navigate(`${basePath}/${id}/edit`)}>
            <Pencil className="w-4 h-4" /> Edit
          </Button>
        )}
      </div>
    </div>
  )
}

function Block({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div>
      <h2 className="text-sm font-semibold uppercase tracking-wide text-gray-400 mb-3">{title}</h2>
      <div className="space-y-2">{children}</div>
    </div>
  )
}
function Row({ label, value }: { label: string; value?: string | number }) {
  return (
    <div className="flex gap-4 text-sm">
      <span className="w-44 flex-shrink-0 text-gray-500">{label}</span>
      <span className="text-gray-800 whitespace-pre-wrap">{value || '-'}</span>
    </div>
  )
}
