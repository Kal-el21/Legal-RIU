import { useCallback, useEffect, useRef, useState, type ReactNode } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { AlertTriangle, ArrowLeft, Download, FileText } from 'lucide-react'
import { agreementService, type AgreementDocument, type AgreementSchema } from '@/services/agreement-document.service'
import type { FieldDef, Nilai } from '@/lib/agreement-draft'
import { barisGrup, daftar, dariFormData, ringkasNilai, rupiah, tanggal, teks } from '@/lib/agreement-isian'
import { LAINNYA, dataPersetujuan, isianAwalPersetujuan, rentangPenandatangan, type Penandatangan, type UsulanPenandatangan } from '@/lib/penandatangan'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import StatusBadge from '@/components/common/StatusBadge'
import RiwayatAktivitas from '@/components/shared/RiwayatAktivitas'
import { formatDate, formatDateTime } from '@/lib/utils'
import type { SubmissionStatus } from '@/types'

interface Props { apiBase?: string; approver?: boolean }

// Label selama diproses dan pesan setelah berhasil, per aksi status.
const statusActions: Record<string, { busy: string; done: string }> = {
  UNDER_REVIEW: { busy: 'Memproses...', done: 'Pemeriksaan dimulai' },
  NEED_REVISION: { busy: 'Memproses...', done: 'Pengajuan dikembalikan untuk revisi' },
  REJECTED: { busy: 'Memproses...', done: 'Pengajuan ditolak' },
  COMPLETED: { busy: 'Menyetujui... (menyiapkan PDF)', done: 'Perjanjian disetujui' },
}

const approverFieldLabels: Record<string, string> = {
  agreement_number: 'Nomor Perjanjian',
  signing_place: 'Tempat Penandatanganan',
  signing_date: 'Tanggal Penandatanganan',
  pic: 'PIC',
  phone: 'Telepon',
  email: 'Email',
}

// Pilihan penandatangan Pihak Pertama di Data Persetujuan (PR 11, P3): usulan
// terpilih lebih dulu, bisa diganti dari Daftar Penandatangan atau Lainnya.
function PilihPenandatangan({ meta, ubah, daftar, usulan }: { meta: Record<string, string>; ubah: (v: Record<string, string>) => void; daftar: Penandatangan[]; usulan?: UsulanPenandatangan | null }) {
  const rentang = rentangPenandatangan(daftar)
  const aktif = daftar.filter((p) => p.is_active)
  const pilihan = meta.party_one_signatory_id ?? ''
  const dikenal = pilihan === '' || pilihan === LAINNYA || aktif.some((p) => p.id === pilihan)
  const nilai = usulan?.value != null ? ` untuk nilai ${rupiah(String(usulan.value))}` : ' (jenis ini tanpa nilai utama, memakai default)'
  return (
    <div className="space-y-2">
      <Label htmlFor="penandatangan" className="text-gray-700">Penandatangan Pihak Pertama</Label>
      <select id="penandatangan" value={pilihan} onChange={(e) => ubah({ party_one_signatory_id: e.target.value })} className="h-9 w-full rounded-md border border-input bg-white px-2 text-sm">
        <option value="">Pilih penandatangan...</option>
        {!dikenal && <option value={pilihan} disabled>Pilihan sebelumnya sudah nonaktif; pilih ulang</option>}
        {aktif.map((p) => (
          <option key={p.id} value={p.id}>{p.name}, {p.position} ({rentang[p.id]}){p.id === usulan?.id ? ' - usulan' : ''}</option>
        ))}
        <option value={LAINNYA}>Lainnya (ketik nama dan jabatan)</option>
      </select>
      {usulan?.id ? (
        <p className="text-xs text-gray-500">Usulan{nilai}: {usulan.name}, {usulan.position}.</p>
      ) : (
        usulan && <p className="text-xs text-amber-700">Tidak ada penandatangan di Daftar{nilai}; pilih dari Daftar atau Lainnya.</p>
      )}
      {pilihan !== '' && usulan?.id && pilihan !== usulan.id && <p className="text-xs text-amber-700">Berbeda dari usulan; penggantian dicatat di audit log.</p>}
      {pilihan === LAINNYA && (
        <div className="grid gap-2 sm:grid-cols-2">
          <Input aria-label="Nama pejabat penandatangan" placeholder="Nama pejabat" value={meta.party_one_signatory_name ?? ''} onChange={(e) => ubah({ party_one_signatory_name: e.target.value })} />
          <Input aria-label="Jabatan pejabat penandatangan" placeholder="Jabatan, mis. Plt. Direktur" value={meta.party_one_signatory_position ?? ''} onChange={(e) => ubah({ party_one_signatory_position: e.target.value })} />
        </div>
      )}
    </div>
  )
}

// Isian sistem yang ditulis approver; tidak ada di definisi field jenis mana pun.
const isianApproverLabels: Record<string, string> = {
  tanggal_ttd: 'Tanggal Penandatanganan',
  pihak_pertama_pejabat: 'Nama Pejabat Pihak Pertama',
  pihak_pertama_jabatan: 'Jabatan Pejabat Pihak Pertama',
  pihak_pertama_pic: 'PIC Pihak Pertama',
  pihak_pertama_telepon: 'Telepon Pihak Pertama',
  pihak_pertama_email: 'Email Pihak Pertama',
}

function NilaiIsian({ f, v }: { f: FieldDef; v: Nilai | undefined }) {
  if (f.type === 'list') {
    const poin = daftar(v).filter((x) => x.trim())
    return poin.length ? <ol className="list-decimal space-y-0.5 pl-5">{poin.map((p, i) => <li key={i}>{p}</li>)}</ol> : <>-</>
  }
  if (f.type === 'group') {
    const baris = barisGrup(v)
    if (!baris.length) return <>-</>
    return (
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-gray-100 text-left text-xs text-gray-500">
              <th className="py-1 pr-3 font-medium">#</th>
              {(f.fields ?? []).map((s) => <th key={s.name} className="py-1 pr-3 font-medium">{s.label}</th>)}
            </tr>
          </thead>
          <tbody>
            {baris.map((b, i) => (
              <tr key={i} className="border-b border-gray-50 last:border-0 align-top">
                <td className="py-1 pr-3 text-gray-500">{i + 1}</td>
                {(f.fields ?? []).map((s) => <td key={s.name} className="py-1 pr-3 whitespace-pre-line">{ringkasNilai(s, b[s.name])}</td>)}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    )
  }
  return <span className="whitespace-pre-line wrap-break-word">{ringkasNilai(f, v)}</span>
}

// Isian pengajuan dengan label dari versi template yang dikunci pengajuan (keputusan
// D4): daftar bernomor, grup (termin) sebagai tabel. Isian yang tidak ada di versi
// itu (isian approver, atau semua isian kalau skema gagal dimuat) tampil di akhir.
function DataPengajuan({ data, schema }: { data: Record<string, unknown>; schema?: AgreementSchema }) {
  const isian = dariFormData(data)
  const dikenal = new Set(schema?.sections.flatMap((s) => s.fields.map((f) => f.name)) ?? [])
  // Kode pilihan penandatangan hanya untuk sistem; namanya tampil di data persetujuan.
  const lain = Object.keys(isian).filter((k) => !dikenal.has(k) && k !== 'pihak_pertama_penandatangan_id')
  const baris = (label: string, isi: ReactNode, lebar: boolean) => (
    <div className={`text-sm py-1.5 border-b border-gray-50 ${lebar ? 'space-y-1' : 'grid grid-cols-2 gap-2'}`}>
      <span className="text-gray-500">{label}</span>
      <div className="text-gray-900">{isi}</div>
    </div>
  )
  return (
    <>
      {schema?.sections.map((s, i) => (
        <div key={s.title + i}>
          <h3 className="mt-3 mb-1 text-xs font-semibold uppercase tracking-wide text-gray-500">{s.title}</h3>
          {s.fields.map((f) => <div key={f.name}>{baris(f.label, <NilaiIsian f={f} v={isian[f.name]} />, f.type === 'group')}</div>)}
        </div>
      ))}
      {lain.length > 0 && schema && <h3 className="mt-3 mb-1 text-xs font-semibold uppercase tracking-wide text-gray-500">Lainnya</h3>}
      {lain.map((k) => {
        const v = isian[k]
        const isi = Array.isArray(v) ? v.map((x) => (typeof x === 'string' ? x : Object.values(x).join(' · '))).join('\n') : tanggal(teks(v)) || '-'
        return <div key={k}>{baris(isianApproverLabels[k] || k.split('_').join(' '), <span className="whitespace-pre-line">{isi}</span>, false)}</div>
      })}
    </>
  )
}

export default function AgreementDocumentDetail({ apiBase = '', approver = false }: Props) {
  const navigate = useNavigate()
  const { id } = useParams()
  const queryClient = useQueryClient()
  const [document, setDocument] = useState<AgreementDocument>()
  const [schema, setSchema] = useState<AgreementSchema>()
  const [note, setNote] = useState('')
  const [meta, setMeta] = useState<Record<string, string>>({})
  const [error, setError] = useState('')
  const [processing, setProcessing] = useState<string | null>(null)
  const [notice, setNotice] = useState('')
  const [previewUrl, setPreviewUrl] = useState('')
  const [previewError, setPreviewError] = useState('')
  const [previewRevision, setPreviewRevision] = useState(0)
  const [previewLoading, setPreviewLoading] = useState(true)
  const objectUrlRef = useRef('')
  const [galatMaster, setGalatMaster] = useState('')
  const [daftarTtd, setDaftarTtd] = useState<Penandatangan[]>([])

  // Master Pihak Pertama dan Daftar Penandatangan hanya untuk approver, lewat alamat
  // perannya (/admin atau /legal). Kalau gagal dimuat, approver diberi tahu, bukan
  // diam-diam diberi isian kosong.
  const muatMaster = useCallback(async (): Promise<Record<string, string>> => {
    if (!apiBase) return {}
    const [master, penandatangan] = await Promise.allSettled([agreementService.master(apiBase), agreementService.signatories(apiBase)])
    setDaftarTtd(penandatangan.status === 'fulfilled' ? penandatangan.value : [])
    const galat = [master, penandatangan].flatMap((r) => (r.status === 'rejected' ? [r.reason instanceof Error ? r.reason.message : 'gagal dimuat'] : []))
    setGalatMaster(galat.join('; '))
    return master.status === 'fulfilled' ? master.value : {}
  }, [apiBase])

  useEffect(() => {
    if (!id) return
    console.log('[AgreementDetail] detail effect mulai -> GET detail', { id, apiBase })
    let active = true

    void (async () => {
      try {
        // Skema hanya untuk label; kalau gagal dimuat, isian tampil dengan kodenya.
        const [value, skema] = await Promise.all([agreementService.get(id, apiBase), agreementService.schemaFor(id, apiBase).catch(() => undefined)])
        if (!active) return
        setSchema(skema)
        console.log('[AgreementDetail] detail sukses', { id, status: value?.status })
        const master = await muatMaster()
        if (!active) return
        setDocument(value)
        setMeta(isianAwalPersetujuan(value, master))
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
  }, [apiBase, id, muatMaster])

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
      const master = await muatMaster()
      setDocument(value)
      setMeta(isianAwalPersetujuan(value, master))
      setError('')
      // Aksi di halaman ini tidak lewat useMutation, jadi riwayatnya dimuat ulang di sini.
      void queryClient.invalidateQueries({ queryKey: ['riwayat', 'agreement_document'] })
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : 'Gagal memuat pengajuan')
    }
  }, [apiBase, id, muatMaster, queryClient])

  // Satu aksi pada satu waktu: persetujuan butuh beberapa detik untuk menyiapkan
  // PDF, dan tanpa kunci tombol yang ditekan dua kali mengirim dua persetujuan.
  const changeStatus = async (status: string) => {
    if (!id || processing) return
    setProcessing(status)
    setNotice('')
    setError('')
    try {
      if (status === 'COMPLETED') await agreementService.meta(apiBase, id, dataPersetujuan(meta))
      await agreementService.status(apiBase, id, status, note)
      await load()
      setPreviewRevision((value) => value + 1)
      setNotice(statusActions[status]?.done ?? 'Status berhasil diubah')
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : 'Gagal memproses pengajuan')
    } finally {
      setProcessing(null)
    }
  }

  const statusButtonLabel = (status: string, label: string) =>
    processing === status ? statusActions[status]?.busy ?? 'Memproses...' : label

  if (error && !document) return <div className="p-6 text-red-600">{error}</div>
  if (!document) return <div className="p-12 text-center text-gray-400">Memuat data...</div>

  return <div className="p-6 max-w-6xl mx-auto">
    <Button variant="outline" size="icon" onClick={() => navigate(-1)} title="Kembali" aria-label="Kembali" className="mt-0.5 mb-2">
      <ArrowLeft className="w-5 h-5 text-gray-600" />
    </Button>
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
    {approver && document.old_template && (() => {
      const lama = document.old_template
      return <div className="mt-4 flex gap-3 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800">
        <AlertTriangle className="w-4 h-4 mt-0.5 shrink-0" />
        <p>
          Pengajuan ini memakai template {document.document_type_code} v{lama.version}. Versi aktif saat ini v{lama.active_version}
          {lama.activated_at && <>, diaktifkan {formatDate(lama.activated_at)}</>}{lama.activated_by && <> oleh {lama.activated_by}</>}.
          {lama.active_note && <> Catatan v{lama.active_version}: &ldquo;{lama.active_note}&rdquo;.</>}
          {' '}Jika disetujui, dokumen final dibentuk dengan v{lama.version}. Untuk memakai v{lama.active_version}, tolak dan minta pengaju membuat pengajuan baru.
        </p>
      </div>
    })()}
    <div className="grid lg:grid-cols-2 gap-5 mt-5">
      <div className="bg-white rounded-2xl border border-gray-100 p-6">
        <h2 className="font-semibold mb-3" style={{ color: '#0B2545' }}>Data Pengajuan</h2>
        <DataPengajuan data={document.form_data} schema={schema} />
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
          {galatMaster && (
            <p role="alert" className="mb-4 flex gap-2 rounded-lg border border-amber-200 bg-amber-50 p-3 text-sm text-amber-800">
              <AlertTriangle className="w-4 h-4 mt-0.5 shrink-0" />
              <span>Data Pihak Pertama atau Daftar Penandatangan gagal dimuat ({galatMaster}), jadi isian di bawah tidak terisi dari master. Muat ulang halaman, atau isi manual.</span>
            </p>
          )}
          <div className="space-y-3">
            {Object.entries(meta).map(([key, value]) => key === 'party_one_signatory_id' ? (
              <PilihPenandatangan key={key} meta={meta} daftar={daftarTtd} usulan={document.signatory_suggestion} ubah={(v) => setMeta({ ...meta, ...v })} />
            ) : !key.startsWith('party_one_signatory_') && (
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
            <Button variant="outline" disabled={!!processing} onClick={() => void changeStatus('UNDER_REVIEW')}>{statusButtonLabel('UNDER_REVIEW', 'Mulai Pemeriksaan')}</Button>
            <Button variant="outline" disabled={!!processing} onClick={() => void changeStatus('NEED_REVISION')}>{statusButtonLabel('NEED_REVISION', 'Kembalikan')}</Button>
            <Button variant="outline" disabled={!!processing} onClick={() => void changeStatus('REJECTED')}>{statusButtonLabel('REJECTED', 'Tolak')}</Button>
            <Button className="text-white transition hover:brightness-95" style={{ background: '#C8102E' }} disabled={!!processing} onClick={() => void changeStatus('COMPLETED')}>{statusButtonLabel('COMPLETED', 'Setujui')}</Button>
          </div>
          {notice && <p className="mt-3 text-sm text-green-600">{notice}</p>}
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
    <div className="mt-6">
      <RiwayatAktivitas kunci={['agreement_document', apiBase, id!]} ambil={() => agreementService.history(id!, apiBase)} />
    </div>
  </div>
}
