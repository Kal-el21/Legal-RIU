import { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { ArrowLeft, Loader2, Save, Send, Eye } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { agreementDocumentService } from '@/services/agreement-document.service'
import { useAuthStore } from '@/store/auth.store'

interface Props {
  basePath: string
  isEdit?: boolean
}

const numberFields = ['nilai_kontrak', 'termin1_persen', 'termin1_nilai', 'termin2_persen', 'termin2_nilai']

function Field({ label, required, type = 'text', value, onChange }: { label: string; required?: boolean; type?: string; value: string; onChange: (v: string) => void }) {
  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 mb-1.5">
        {label} {required && <span className="text-red-500">*</span>}
      </label>
      <input
        type={type}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="w-full px-3.5 py-2.5 rounded-xl border border-gray-200 focus:ring-2 focus:ring-red-100 focus:border-red-400 outline-none text-sm"
      />
    </div>
  )
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="bg-white rounded-2xl border border-gray-100 p-6">
      <h2 className="text-sm font-semibold uppercase tracking-wide text-gray-400 mb-4">{title}</h2>
      {children}
    </div>
  )
}

export default function SharedAgreementDocumentForm({ basePath, isEdit = false }: Props) {
  const { id } = useParams()
  const navigate = useNavigate()
  const hasPermission = useAuthStore((state) => state.hasPermission)
  const canEdit = hasPermission('agreement_document.update.own')

  const [form, setForm] = useState<Record<string, string>>({})
  const [files, setFiles] = useState<File[]>([])
  const [loading, setLoading] = useState(false)
  const [initialLoading, setInitialLoading] = useState(isEdit)

  useEffect(() => {
    if (isEdit && id) {
      agreementDocumentService
        .getByID(id)
        .then((doc) => {
          const fd = doc.form_data || {}
          const mapped: Record<string, string> = {}
          Object.entries(fd).forEach(([k, v]) => {
            mapped[k] = v === null || v === undefined ? '' : String(v)
          })
          setForm(mapped)
        })
        .finally(() => setInitialLoading(false))
    }
  }, [id, isEdit])

  const set = (k: string, v: string) => setForm((f) => ({ ...f, [k]: v }))

  const submit = async () => {
    setLoading(true)
    try {
      const payload: any = { ...form }
      numberFields.forEach((k) => {
        if (payload[k] !== '' && payload[k] != null) payload[k] = Number(payload[k])
        else delete payload[k]
      })
      payload.attachments = files
      if (isEdit && id) {
        await agreementDocumentService.update(id, payload)
      } else {
        await agreementDocumentService.create(payload)
      }
      navigate(`${basePath}`)
    } catch (e: any) {
      alert(e?.response?.data?.message || 'Gagal menyimpan pengajuan')
    } finally {
      setLoading(false)
    }
  }

  const preview = async () => {
    if (!id && !isEdit) {
      alert('Simpan pengajuan terlebih dahulu untuk melihat preview')
      return
    }
    const docId = id || form._id
    if (!docId) {
      alert('Simpan pengajuan terlebih dahulu untuk melihat preview')
      return
    }
    const url = await agreementDocumentService.getPreviewURL(docId)
    window.open(url, '_blank')
  }

  if (initialLoading) {
    return (
      <div className="p-6 max-w-4xl mx-auto flex items-center gap-2 text-gray-400">
        <Loader2 className="w-4 h-4 animate-spin" /> Memuat...
      </div>
    )
  }

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <Link to={basePath} className="inline-flex items-center gap-1.5 text-sm text-gray-500 hover:text-gray-700 mb-4">
        <ArrowLeft className="w-4 h-4" /> Kembali
      </Link>
      <h1 className="text-2xl font-bold mb-1" style={{ color: '#0B2545' }}>
        {isEdit ? 'Edit Pengajuan Perjanjian' : 'Pengajuan Dokumen Perjanjian'}
      </h1>
      <p className="text-sm text-gray-500 mb-6">Isi data Pihak Kedua dan detail pekerjaan. Pihak Pertama diisi oleh approver.</p>

      <div className="space-y-6">
        <Section title="Informasi Umum">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Field label="Nomor Pihak Kedua" value={form.nomor_pihak_kedua || ''} onChange={(v) => set('nomor_pihak_kedua', v)} />
            <Field label="Tempat Penandatanganan" value={form.tempat_ttd || ''} onChange={(v) => set('tempat_ttd', v)} />
            <Field label="Tanggal Penandatanganan" type="date" value={form.tanggal_ttd || ''} onChange={(v) => set('tanggal_ttd', v)} />
          </div>
        </Section>

        <Section title="Pihak Kedua">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Field label="Nama Perusahaan" required value={form.pihak_kedua_nama || ''} onChange={(v) => set('pihak_kedua_nama', v)} />
            <Field label="Bidang Usaha" value={form.pihak_kedua_bidang || ''} onChange={(v) => set('pihak_kedua_bidang', v)} />
            <Field label="Alamat" value={form.pihak_kedua_alamat || ''} onChange={(v) => set('pihak_kedua_alamat', v)} />
            <Field label="Telepon" value={form.pihak_kedua_telepon || ''} onChange={(v) => set('pihak_kedua_telepon', v)} />
            <Field label="Email" value={form.pihak_kedua_email || ''} onChange={(v) => set('pihak_kedua_email', v)} />
            <Field label="PIC" value={form.pihak_kedua_pic || ''} onChange={(v) => set('pihak_kedua_pic', v)} />
            <Field label="Nama Pejabat" required value={form.pihak_kedua_pejabat || ''} onChange={(v) => set('pihak_kedua_pejabat', v)} />
            <Field label="Jabatan Pejabat" required value={form.pihak_kedua_jabatan || ''} onChange={(v) => set('pihak_kedua_jabatan', v)} />
          </div>
        </Section>

        <Section title="Dasar Hukum">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Field label="No. Surat Penawaran" value={form.surat_penawaran_nomor || ''} onChange={(v) => set('surat_penawaran_nomor', v)} />
            <Field label="Perihal Penawaran" value={form.surat_penawaran_perihal || ''} onChange={(v) => set('surat_penawaran_perihal', v)} />
            <Field label="Tanggal Penawaran" type="date" value={form.surat_penawaran_tanggal || ''} onChange={(v) => set('surat_penawaran_tanggal', v)} />
            <Field label="No. Surat Penunjukan" value={form.surat_penunjukan_nomor || ''} onChange={(v) => set('surat_penunjukan_nomor', v)} />
            <Field label="Perihal Penunjukan" value={form.surat_penunjukan_perihal || ''} onChange={(v) => set('surat_penunjukan_perihal', v)} />
            <Field label="Tanggal Penunjukan" type="date" value={form.surat_penunjukan_tanggal || ''} onChange={(v) => set('surat_penunjukan_tanggal', v)} />
          </div>
        </Section>

        <Section title="Ketentuan Khusus">
          <div className="space-y-4">
            <Field label="Jenis Pekerjaan" required value={form.jenis_pekerjaan || ''} onChange={(v) => set('jenis_pekerjaan', v)} />
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1.5">Ruang Lingkup Pekerjaan <span className="text-red-500">*</span></label>
              <textarea
                value={form.ruang_lingkup || ''}
                onChange={(e) => set('ruang_lingkup', e.target.value)}
                rows={3}
                className="w-full px-3.5 py-2.5 rounded-xl border border-gray-200 focus:ring-2 focus:ring-red-100 focus:border-red-400 outline-none text-sm"
              />
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Field label="Jangka Waktu Mulai" type="date" value={form.jangka_waktu_mulai || ''} onChange={(v) => set('jangka_waktu_mulai', v)} />
              <Field label="Jangka Waktu Selesai" type="date" value={form.jangka_waktu_selesai || ''} onChange={(v) => set('jangka_waktu_selesai', v)} />
              <Field label="Nilai Kontrak (Rp)" type="number" value={form.nilai_kontrak || ''} onChange={(v) => set('nilai_kontrak', v)} />
              <div />
              <Field label="Termin 1 (%)" type="number" value={form.termin1_persen || ''} onChange={(v) => set('termin1_persen', v)} />
              <Field label="Termin 1 (Rp)" type="number" value={form.termin1_nilai || ''} onChange={(v) => set('termin1_nilai', v)} />
              <Field label="Termin 2 (%)" type="number" value={form.termin2_persen || ''} onChange={(v) => set('termin2_persen', v)} />
              <Field label="Termin 2 (Rp)" type="number" value={form.termin2_nilai || ''} onChange={(v) => set('termin2_nilai', v)} />
              <Field label="Bank" value={form.bank || ''} onChange={(v) => set('bank', v)} />
              <Field label="No. Rekening" value={form.nomor_rekening || ''} onChange={(v) => set('nomor_rekening', v)} />
              <Field label="Atas Nama" value={form.atas_nama || ''} onChange={(v) => set('atas_nama', v)} />
            </div>
          </div>
        </Section>

        <Section title="Lampiran">
          <input
            type="file"
            multiple
            onChange={(e) => setFiles(Array.from(e.target.files || []))}
            className="block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-xl file:border-0 file:text-sm file:font-medium file:bg-red-50 file:text-red-600 hover:file:bg-red-100"
          />
        </Section>
      </div>

      <div className="mt-8 flex gap-3">
        <Button onClick={preview} disabled={loading} variant="outline" className="flex items-center gap-2">
          <Eye className="w-4 h-4" /> Preview PDF
        </Button>
        <Button onClick={submit} disabled={loading || (isEdit && !canEdit)}
          className="flex items-center gap-2" variant="outline">
          {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />} Simpan
        </Button>
        <Button onClick={submit} disabled={loading || (isEdit && !canEdit)}
          className="flex items-center gap-2 text-white" style={{ background: '#C8102E' }}>
          {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Send className="w-4 h-4" />} Kirim Pengajuan
        </Button>
      </div>
    </div>
  )
}
