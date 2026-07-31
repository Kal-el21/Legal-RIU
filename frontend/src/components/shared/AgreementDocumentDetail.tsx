import { useCallback, useEffect, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { ArrowLeft, Download, FileText } from 'lucide-react'
import { agreementService, type AgreementDocument } from '@/services/agreement-document.service'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import StatusBadge from '@/components/common/StatusBadge'
import { formatDateTime } from '@/lib/utils'
import type { SubmissionStatus } from '@/types'

interface Props { apiBase?: string; approver?: boolean }

const approverFieldLabels: Record<string, string> = {
  agreement_number: 'Nomor Perjanjian',
  signing_place: 'Tempat Penandatanganan',
  signing_date: 'Tanggal Penandatanganan',
  party_one_signatory_name: 'Nama Pejabat Pihak Pertama',
  party_one_signatory_position: 'Jabatan Pejabat Pihak Pertama',
  pic: 'PIC',
  phone: 'Telepon',
  email: 'Email',
}

const formFieldLabels: Record<string, string> = {
  nomor_pihak_kedua: 'Nomor Pihak Kedua',
  tempat_ttd: 'Usulan Tempat Penandatanganan',
  pihak_kedua_nama: 'Nama Perusahaan',
  pihak_kedua_bidang: 'Bidang Usaha',
  pihak_kedua_alamat: 'Alamat',
  pihak_kedua_telepon: 'Telepon',
  pihak_kedua_email: 'Email',
  pihak_kedua_pic: 'PIC',
  pihak_kedua_pejabat: 'Nama Pejabat',
  pihak_kedua_jabatan: 'Jabatan Pejabat',
  jenis_pekerjaan: 'Jenis Pekerjaan',
  ruang_lingkup: 'Ruang Lingkup Pekerjaan',
  surat_penawaran_nomor: 'Nomor Surat Penawaran',
  surat_penawaran_perihal: 'Perihal Surat Penawaran',
  surat_penawaran_tanggal: 'Tanggal Surat Penawaran',
  surat_penunjukan_nomor: 'Nomor Surat Penunjukan',
  surat_penunjukan_perihal: 'Perihal Surat Penunjukan',
  surat_penunjukan_tanggal: 'Tanggal Surat Penunjukan',
  jangka_waktu_mulai: 'Tanggal Mulai',
  jangka_waktu_selesai: 'Tanggal Selesai',
  nilai_kontrak: 'Nilai Kontrak (Rupiah)',
  termin_1_persen: 'Termin 1 (%)',
  termin_1_nilai: 'Nilai Termin 1',
  termin_2_persen: 'Termin 2 (%)',
  termin_2_nilai: 'Nilai Termin 2',
  bank: 'Bank',
  nomor_rekening: 'Nomor Rekening',
  atas_nama: 'Atas Nama',
}

export default function AgreementDocumentDetail({ apiBase = '', approver = false }: Props) {
  const navigate = useNavigate()
  const { id } = useParams()
  const [document, setDocument] = useState<AgreementDocument>()
  const [note, setNote] = useState('')
  const [meta, setMeta] = useState<Record<string, string>>({})
  const [error, setError] = useState('')
  const [previewUrl, setPreviewUrl] = useState('')
  const [previewError, setPreviewError] = useState('')
  const [previewRevision, setPreviewRevision] = useState(0)
  const [previewLoading, setPreviewLoading] = useState(true)
  const objectUrlRef = useRef('')

  useEffect(() => {
    if (!id) return
    console.log('[AgreementDetail] detail effect mulai -> GET detail', { id, apiBase })
    let active = true

    void (async () => {
      try {
        const value = await agreementService.get(id, apiBase)
        if (!active) return
        console.log('[AgreementDetail] detail sukses', { id, status: value?.status })
        // Master hanya tersedia untuk approver (admin/legal). User biasa tidak punya akses endpoint ini.
        const master = apiBase
          ? await agreementService.master().catch(() => ({}) as Record<string, string>)
          : ({} as Record<string, string>)
        if (!active) return
        setDocument(value)
        setMeta({
          agreement_number: String(master.default_agreement_number || value.agreement_number || ''),
          signing_place: String(master.default_signing_place || value.form_data.tempat_ttd || ''),
          signing_date: String(value.form_data.tanggal_ttd || ''),
          party_one_signatory_name: String(master.default_signatory_name || value.form_data.pihak_pertama_pejabat || ''),
          party_one_signatory_position: String(master.default_signatory_position || value.form_data.pihak_pertama_jabatan || ''),
          pic: String(master.pic || ''),
          phone: String(master.phone || ''),
          email: String(master.email || ''),
        })
        setError('')
      } catch (reason) {
        if (!active) return
        console.error('[AgreementDetail] detail GAGAL', { id, reason })
        setError(reason instanceof Error ? reason.message : 'Gagal memuat pengajuan')
      }
    })()

    return () => {
      active = false
    }
  }, [apiBase, id])

  useEffect(() => {
    if (!id) return
    console.log('[AgreementDetail] preview effect mulai -> GET preview', { id, apiBase })
    let active = true

    void agreementService.preview(apiBase, id)
      .then((blob) => {
        if (!active) return
        console.log('[AgreementDetail] preview sukses, blob size=', blob.size)
        const url = URL.createObjectURL(blob)
        if (objectUrlRef.current) URL.revokeObjectURL(objectUrlRef.current)
        objectUrlRef.current = url
        setPreviewError('')
        setPreviewUrl(url)
        setPreviewLoading(false)
      })
      .catch((reason: unknown) => {
        if (!active) return
        console.error('[AgreementDetail] preview GAGAL', { id, reason })
        setPreviewError(reason instanceof Error ? reason.message : 'Gagal memuat preview PDF')
        setPreviewLoading(false)
      })

    return () => {
      active = false
    }
  }, [apiBase, id, previewRevision])

  useEffect(() => () => {
    if (objectUrlRef.current) URL.revokeObjectURL(objectUrlRef.current)
  }, [])

  const load = useCallback(async () => {
    if (!id) return
    try {
      const value = await agreementService.get(id, apiBase)
      const master = await agreementService.master().catch(() => ({}) as Record<string, string>)
      setDocument(value)
      setMeta({
        agreement_number: String(master.default_agreement_number || value.agreement_number || ''),
        signing_place: String(master.default_signing_place || value.form_data.tempat_ttd || ''),
        signing_date: String(value.form_data.tanggal_ttd || ''),
        party_one_signatory_name: String(master.default_signatory_name || value.form_data.pihak_pertama_pejabat || ''),
        party_one_signatory_position: String(master.default_signatory_position || value.form_data.pihak_pertama_jabatan || ''),
        pic: String(master.pic || ''),
        phone: String(master.phone || ''),
        email: String(master.email || ''),
      })
      setError('')
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : 'Gagal memuat pengajuan')
    }
  }, [apiBase, id])

  const changeStatus = async (status: string) => {
    if (!id) return
    try {
      if (status === 'COMPLETED') await agreementService.meta(apiBase, id, meta)
      await agreementService.status(apiBase, id, status, note)
      await load()
      setPreviewRevision((value) => value + 1)
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : 'Gagal memproses pengajuan')
    }
  }

  if (error && !document) return <div className="p-6 text-red-600">{error}</div>
  if (!document) return <div className="p-12 text-center text-gray-400">Memuat data...</div>

  return <div className="p-6 max-w-6xl mx-auto">
    <button onClick={() => navigate(-1)} className="p-2 rounded-lg hover:bg-gray-100 mt-0.5 mb-2" title="Kembali">
      <ArrowLeft className="w-5 h-5 text-gray-600" />
    </button>
    <div className="flex items-start gap-3">
      <div className="flex-1">
        <div className="flex items-center gap-3 flex-wrap">
          <span className="text-xs font-mono font-medium bg-gray-100 text-gray-600 px-2 py-1 rounded">
            {document.ticket_number}
          </span>
          <StatusBadge status={document.status as SubmissionStatus} />
        </div>
        <h1 className="text-2xl font-bold mt-2" style={{ color: '#0B2545' }}>{document.document_type_code} · {document.ticket_number}</h1>
        <p className="text-sm text-gray-500 mt-0.5">
          Diajukan oleh: {document.user?.full_name || '-'} · Dibuat {document.created_at ? formatDateTime(document.created_at) : '-'}
        </p>
      </div>
    </div>
    {error && <div className="mt-4 rounded-lg bg-red-50 p-3 text-sm text-red-700">{error}</div>}
    <div className="grid lg:grid-cols-2 gap-5 mt-5">
      <div className="bg-white rounded-2xl border border-gray-100 p-6">
        <h2 className="font-semibold mb-3" style={{ color: '#0B2545' }}>Data Pengajuan</h2>
        {Object.entries(document.form_data).map(([key, value]) => <div key={key} className="grid grid-cols-2 text-sm py-1.5 border-b border-gray-50"><span className="text-gray-500">{formFieldLabels[key] || key.split('_').join(' ')}</span><span className="text-gray-900">{String(value)}</span></div>)}
        <h2 className="font-semibold mt-5 mb-3" style={{ color: '#0B2545' }}>Lampiran</h2>
        {document.attachments && document.attachments.length > 0 ? (
          <div className="space-y-2">
            {document.attachments.map((attachment) => (
              <div key={attachment.id} className="flex items-center gap-3 p-3 rounded-lg bg-gray-50 border border-gray-100">
                <FileText className="w-4 h-4 text-gray-400 flex-shrink-0" />
                <p className="flex-1 min-w-0 text-sm text-gray-700 truncate">{attachment.file_name}</p>
                <a
                  href={`${agreementService.fileUrl(apiBase, id!, 'preview').replace('/preview', `/attachments/${attachment.id}`)}`}
                  download
                  className="p-1.5 rounded-lg hover:bg-gray-200 text-gray-400 hover:text-gray-600"
                  title="Unduh lampiran"
                >
                  <Download className="w-4 h-4" />
                </a>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-sm text-gray-400">Belum ada lampiran.</p>
        )}
      </div>
      <div>
        {approver && <div className="bg-white rounded-2xl border border-gray-100 p-6 mb-4">
          <h2 className="font-semibold mb-4" style={{ color: '#0B2545' }}>Data Persetujuan</h2>
          <div className="space-y-3">
            {Object.entries(meta).map(([key, value]) => (
              <div key={key}>
                <Label className="mb-1.5 text-gray-700">{approverFieldLabels[key] || key.split('_').join(' ')}</Label>
                <Input type={key === 'signing_date' ? 'date' : 'text'} value={value} onChange={(event) => setMeta({ ...meta, [key]: event.target.value })} />
              </div>
            ))}
            <div>
              <Label className="mb-1.5 text-gray-700">Catatan Revisi/Penolakan</Label>
              <Textarea placeholder="Catatan revisi/penolakan" value={note} onChange={(event) => setNote(event.target.value)} />
            </div>
          </div>
          <div className="flex gap-2 flex-wrap mt-4">
            <Button variant="outline" onClick={() => void changeStatus('UNDER_REVIEW')}>Mulai Pemeriksaan</Button>
            <Button variant="outline" onClick={() => void changeStatus('NEED_REVISION')}>Kembalikan</Button>
            <Button variant="outline" onClick={() => void changeStatus('REJECTED')}>Tolak</Button>
            <Button className="text-white transition hover:brightness-95" style={{ background: '#C8102E' }} onClick={() => void changeStatus('COMPLETED')}>Setujui</Button>
          </div>
        </div>}
        {previewError && <div className="rounded-2xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">Preview PDF gagal dimuat: {previewError}</div>}
        {!previewError && previewLoading && <div className="h-[650px] rounded-2xl border border-gray-100 bg-gray-50 flex items-center justify-center text-sm text-gray-500">Menyiapkan preview PDF...</div>}
        {previewUrl && <iframe title="preview" className="w-full h-[650px] rounded-2xl border border-gray-100" src={previewUrl} />}
        {document.status === 'COMPLETED' && <div className="flex gap-2 mt-3">
          <a href={agreementService.fileUrl(apiBase, id!, 'pdf')}><Button className="flex items-center gap-2 text-white transition hover:brightness-95" style={{ background: '#C8102E' }}><Download className="w-4 h-4" /> Download PDF</Button></a>
          {approver && <a href={agreementService.fileUrl(apiBase, id!, 'docx')}><Button variant="outline" className="flex items-center gap-2"><Download className="w-4 h-4" /> Download DOCX</Button></a>}
        </div>}
      </div>
    </div>
  </div>
}
