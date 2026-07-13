import { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { ArrowLeft, Loader2, Eye, EyeOff, CheckCircle, RotateCcw, XCircle, AlertCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import StatusBadge from '@/components/common/StatusBadge'
import { agreementDocumentService } from '@/services/agreement-document.service'
import type { AgreementDocument, SubmissionStatus } from '@/types'

interface Props {
  basePath: string
}

export default function SharedAgreementDocumentReview({ basePath }: Props) {
  const { id } = useParams()
  const navigate = useNavigate()
  const [doc, setDoc] = useState<AgreementDocument | null>(null)
  const [loading, setLoading] = useState(true)
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)
  const [showPreview, setShowPreview] = useState(true)
  const [ppPejabat, setPpPejabat] = useState('')
  const [ppJabatan, setPpJabatan] = useState('')
  const [adminNote, setAdminNote] = useState('')
  const [nomorPP, setNomorPP] = useState('')
  const [tempatTtd, setTempatTtd] = useState('')
  const [tanggalTtd, setTanggalTtd] = useState('')
  const [busy, setBusy] = useState(false)

  const refresh = () => {
    if (!id) return
    setLoading(true)
    agreementDocumentService.getByID(id).then((d) => {
      setDoc(d)
      const fd = d.form_data || {}
      setNomorPP((fd.nomor_pihak_pertama as string) || d.pihak_pertama?.name || '')
      setTempatTtd((fd.tempat_ttd as string) || '')
      setTanggalTtd((fd.tanggal_ttd as string) || '')
      setPpPejabat(d.pihak_pertama_pejabat || d.pihak_pertama?.default_pejabat || '')
      setPpJabatan(d.pihak_pertama_jabatan || d.pihak_pertama?.default_jabatan || '')
    }).finally(() => setLoading(false))
  }

  useEffect(() => {
    refresh()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id])

  const loadPreview = async () => {
    if (!id) return
    const url = await agreementDocumentService.getPreviewURL(id)
    setPreviewUrl(url)
  }

  const action = async (fn: () => Promise<any>, doneMsg: string) => {
    if (!id) return
    setBusy(true)
    try {
      await agreementDocumentService.updateMeta(id, {
        nomor_pihak_pertama: nomorPP,
        tempat_ttd: tempatTtd,
        tanggal_ttd: tanggalTtd,
        pihak_pertama_pejabat: ppPejabat,
        pihak_pertama_jabatan: ppJabatan,
      })
      await fn()
      alert(doneMsg)
      navigate(basePath)
    } catch (e: any) {
      alert(e?.response?.data?.message || 'Gagal memproses')
    } finally {
      setBusy(false)
    }
  }

  if (loading) {
    return (
      <div className="p-6 flex items-center gap-2 text-gray-400">
        <Loader2 className="w-4 h-4 animate-spin" /> Memuat...
      </div>
    )
  }
  if (!doc) return <div className="p-6 text-gray-400">Dokumen tidak ditemukan.</div>

  const fd = doc.form_data || {}
  const canDecide = doc.status !== 'COMPLETED' && doc.status !== 'REJECTED'

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <Link to={basePath} className="inline-flex items-center gap-1.5 text-sm text-gray-500 hover:text-gray-700 mb-4">
        <ArrowLeft className="w-4 h-4" /> Kembali
      </Link>

      <div className="flex items-center justify-between mb-4">
        <div>
          <p className="text-xs font-mono text-gray-400">{doc.ticket_number}</p>
          <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>Review Dokumen Perjanjian</h1>
        </div>
        <StatusBadge status={doc.status as SubmissionStatus} />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Left: editor */}
        <div className="space-y-4">
          <div className="bg-white rounded-2xl border border-gray-100 p-6">
            <h2 className="text-sm font-semibold uppercase tracking-wide text-gray-400 mb-3">Data Pihak Pertama (Approver)</h2>
            <div className="grid grid-cols-1 gap-3">
              <Input label="Nomor Pihak Pertama" value={nomorPP} onChange={setNomorPP} />
              <Input label="Nama Pejabat" value={ppPejabat} onChange={setPpPejabat} />
              <Input label="Jabatan" value={ppJabatan} onChange={setPpJabatan} />
            </div>
            <p className="text-xs text-gray-400 mt-2">Perusahaan: {doc.pihak_pertama?.name || 'PT Reasuransi Indonesia Utama (Persero)'}</p>
          </div>

          <div className="bg-white rounded-2xl border border-gray-100 p-6">
            <h2 className="text-sm font-semibold uppercase tracking-wide text-gray-400 mb-3">Penandatanganan</h2>
            <div className="grid grid-cols-1 gap-3">
              <Input label="Tempat" value={tempatTtd} onChange={setTempatTtd} />
              <Input label="Tanggal" type="date" value={tanggalTtd} onChange={setTanggalTtd} />
            </div>
          </div>

          <div className="bg-white rounded-2xl border border-gray-100 p-6">
            <h2 className="text-sm font-semibold uppercase tracking-wide text-gray-400 mb-3">Ringkasan</h2>
            <Row label="Jenis Pekerjaan" value={fd.jenis_pekerjaan} />
            <Row label="Pihak Kedua" value={fd.pihak_kedua_nama} />
            <Row label="Ruang Lingkup" value={fd.ruang_lingkup} />
            <Row label="Nilai Kontrak" value={String(fd.nilai_kontrak || '')} />
            <Row label="Lampiran" value={doc.attachments?.length ? `${doc.attachments.length} file` : 'Tidak ada'} />
          </div>

          <div className="bg-white rounded-2xl border border-gray-100 p-6">
            <h2 className="text-sm font-semibold uppercase tracking-wide text-gray-400 mb-3">Catatan / Revisi</h2>
            <textarea
              value={adminNote}
              onChange={(e) => setAdminNote(e.target.value)}
              rows={3}
              placeholder="Catatan untuk requester (saat Return/Reject)"
              className="w-full px-3.5 py-2.5 rounded-xl border border-gray-200 focus:ring-2 focus:ring-red-100 outline-none text-sm"
            />
            {doc.admin_note && (
              <div className="mt-3 p-3 rounded-lg bg-gray-50 text-sm text-gray-600 flex gap-2">
                <AlertCircle className="w-4 h-4 text-gray-400 flex-shrink-0 mt-0.5" /> {doc.admin_note}
              </div>
            )}
          </div>

          {canDecide && (
            <div className="flex flex-wrap gap-3">
              <Button className="flex items-center gap-2 text-white" style={{ background: '#16a34a' }} disabled={busy}
                onClick={() => action(() => agreementDocumentService.approve(id!), 'Dokumen disetujui & PDF final dibuat')}>
                <CheckCircle className="w-4 h-4" /> Approve
              </Button>
              <Button className="flex items-center gap-2 text-white" style={{ background: '#d97706' }} disabled={busy}
                onClick={() => action(() => agreementDocumentService.returnForRevision(id!, adminNote), 'Dokumen dikembalikan untuk revisi')}>
                <RotateCcw className="w-4 h-4" /> Return
              </Button>
              <Button className="flex items-center gap-2 text-white" style={{ background: '#C8102E' }} disabled={busy}
                onClick={() => action(() => agreementDocumentService.reject(id!, adminNote), 'Dokumen ditolak')}>
                <XCircle className="w-4 h-4" /> Reject
              </Button>
            </div>
          )}
        </div>

        {/* Right: preview */}
        <div className="bg-white rounded-2xl border border-gray-100 p-4">
          <div className="flex items-center justify-between mb-3">
            <h2 className="text-sm font-semibold text-gray-500">Preview Dokumen (Watermark)</h2>
            <div className="flex gap-2">
              {showPreview && previewUrl && (
                <Button size="sm" variant="outline" onClick={() => { setShowPreview(false); setPreviewUrl(null); }}>
                  <EyeOff className="w-4 h-4" /> Tutup Preview
                </Button>
              )}
              <Button size="sm" variant="outline" onClick={loadPreview}><Eye className="w-4 h-4" /> Load Preview</Button>
            </div>
          </div>
          {showPreview && previewUrl ? (
            <iframe src={previewUrl} className="w-full h-[70vh] rounded-lg border border-gray-100" title="Preview" />
          ) : showPreview ? (
            <div className="h-[70vh] flex items-center justify-center text-gray-400 text-sm">
              Klik "Load Preview" untuk melihat dokumen lengkap (termasuk data Pihak Pertama).
            </div>
          ) : (
            <div className="h-[70vh] flex items-center justify-center text-gray-400 text-sm">
              Preview disembunyikan.{' '}
              <button onClick={() => setShowPreview(true)} className="ml-1 underline" style={{ color: '#C8102E' }}>
                Tampilkan Preview
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

function Input({ label, value, onChange, type = 'text' }: { label: string; value: string; onChange: (v: string) => void; type?: string }) {
  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 mb-1.5">{label}</label>
      <input type={type} value={value} onChange={(e) => onChange(e.target.value)}
        className="w-full px-3.5 py-2.5 rounded-xl border border-gray-200 focus:ring-2 focus:ring-red-100 focus:border-red-400 outline-none text-sm" />
    </div>
  )
}
function Row({ label, value }: { label: string; value?: string | number }) {
  return (
    <div className="flex gap-4 text-sm py-1">
      <span className="w-40 flex-shrink-0 text-gray-500">{label}</span>
      <span className="text-gray-800 whitespace-pre-wrap">{value || '-'}</span>
    </div>
  )
}
