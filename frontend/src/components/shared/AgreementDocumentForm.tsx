import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Send, ArrowLeft } from 'lucide-react'
import { agreementService, type AgreementSchema } from '@/services/agreement-document.service'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import StatusBadge from '@/components/common/StatusBadge'
import type { SubmissionStatus } from '@/types'

const REQUIRED_FIELDS = new Set([
  'pihak_kedua_nama',
  'pihak_kedua_bidang',
  'pihak_kedua_alamat',
  'pihak_kedua_pejabat',
  'pihak_kedua_jabatan',
  'jenis_pekerjaan',
  'ruang_lingkup',
  'jangka_waktu_mulai',
  'jangka_waktu_selesai',
  'nilai_kontrak',
  'termin_1_persen',
  'termin_1_nilai',
  'termin_2_persen',
  'termin_2_nilai',
  'bank',
  'nomor_rekening',
  'atas_nama',
])

function parseNum(v: unknown): number | null {
  if (v === undefined || v === null) return null
  const s = String(v).trim()
  if (s === '') return null
  const n = Number(s)
  return Number.isFinite(n) ? n : null
}

export default function AgreementDocumentForm() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [schema, setSchema] = useState<AgreementSchema>()
  const [form, setForm] = useState<Record<string, string>>({})
  const [files, setFiles] = useState<File[]>([])
  const [currentStatus, setCurrentStatus] = useState('')
  const [ticketNumber, setTicketNumber] = useState('')
  const [error, setError] = useState<string>()

  const labelOf = (name: string) =>
    schema?.sections.flatMap((s) => s.fields).find((f) => f.name === name)?.label ?? name.replace(/_/g, ' ')

  useEffect(() => {
    void agreementService.schema().then(setSchema)
    if (id) {
      void agreementService.get(id).then((doc) => {
        setCurrentStatus(doc.status)
        setTicketNumber(doc.ticket_number)
        const values: Record<string, string> = {}
        Object.entries(doc.form_data).forEach(([key, value]) => { values[key] = String(value ?? '') })
        setForm(values)
      })
    }
  }, [id])

  const deriveForm = (source: Record<string, string>): Record<string, string> => {
    const toNum = (v: string) => { const n = Number(v); return Number.isFinite(n) ? n : 0 }
    const persen1 = toNum(source.termin_1_persen ?? '')
    const nilai1 = toNum(source.termin_1_nilai ?? '')
    const kontrak = toNum(source.nilai_kontrak ?? '')
    const persen2 = Math.max(0, 100 - persen1)
    const nilai2 = Math.max(0, kontrak - nilai1)
    return {
      ...source,
      termin_2_persen: String(persen2),
      termin_2_nilai: String(nilai2),
    }
  }

  const derivedForm = deriveForm(form)

  const validate = (): string | undefined => {
    const persen1 = parseNum(derivedForm.termin_1_persen)
    const persen2 = parseNum(derivedForm.termin_2_persen)
    const nilai1 = parseNum(derivedForm.termin_1_nilai)
    const nilai2 = parseNum(derivedForm.termin_2_nilai)
    const kontrak = parseNum(derivedForm.nilai_kontrak)

    if (persen1 === null) return `${labelOf('termin_1_persen')} harus berupa angka`
    if (persen2 === null) return `${labelOf('termin_2_persen')} harus berupa angka`
    if (nilai1 === null) return `${labelOf('termin_1_nilai')} harus berupa angka`
    if (nilai2 === null) return `${labelOf('termin_2_nilai')} harus berupa angka`
    if (kontrak === null) return `${labelOf('nilai_kontrak')} harus berupa angka`

    if (Math.abs(persen1 + persen2 - 100) > 0.001) return 'Jumlah persentase termin harus 100%'
    if (Math.abs(nilai1 + nilai2 - kontrak) > 0.5) return 'Jumlah nilai termin harus sama dengan nilai kontrak'
    if (derivedForm.jangka_waktu_selesai && derivedForm.jangka_waktu_mulai && derivedForm.jangka_waktu_selesai < derivedForm.jangka_waktu_mulai) return 'Tanggal selesai tidak boleh sebelum tanggal mulai'

    for (const key of REQUIRED_FIELDS) {
      if (!derivedForm[key]?.trim()) return `${labelOf(key)} wajib diisi`
    }
    return undefined
  }

  const submit = async () => {
    setError(undefined)
    const validationError = validate()
    if (validationError) {
      setError(validationError)
      return
    }
    try {
      if (id) {
        await agreementService.update(id, derivedForm)
        if (currentStatus === 'NEED_REVISION') await agreementService.resubmit(id, files)
      } else {
        await agreementService.create({ document_type_code: 'PKS', form_data: derivedForm }, files)
      }
      setFiles([])
      navigate('/dashboard/agreement-documents')
    } catch (err) {
      setError((err as Error).message || 'Terjadi kesalahan saat mengirim pengajuan')
    }
  }

   return <div className="p-6 max-w-4xl mx-auto">
      <button onClick={() => navigate(-1)} className="p-2 rounded-lg hover:bg-gray-100 mt-0.5 mb-2" title="Kembali">
        <ArrowLeft className="w-5 h-5 text-gray-600" />
      </button>
       {id && (
        <div className="flex items-center gap-3 mb-4">
          <span className="text-xs font-mono font-medium bg-gray-100 text-gray-600 px-2 py-1 rounded">
            {ticketNumber}
          </span>
          <StatusBadge status={currentStatus as SubmissionStatus} />
        </div>
      )}
      <h1 className="text-2xl font-bold mb-6" style={{ color: '#0B2545' }}>{id ? 'Revisi' : 'Pengajuan'} Perjanjian Kerja Sama</h1>
     {error && <div className="mb-4 p-3 rounded-lg bg-red-50 text-sm text-red-700">{error}</div>}
     {schema?.sections.map((section) => <section key={section.title} className="bg-white rounded-2xl border border-gray-100 p-6 mb-5">
       <h2 className="text-lg font-semibold mb-4" style={{ color: '#0B2545' }}>{section.title}</h2>
       <div className="grid md:grid-cols-2 gap-4">
          {section.fields.map((field) => (
            <div key={field.name} className={field.type === 'textarea' ? 'md:col-span-2' : ''}>
              <Label className="mb-1.5 text-gray-700">{field.label}{field.required && ' *'}</Label>
              {field.type === 'textarea'
                ? <Textarea required={field.required} placeholder={field.name === 'ruang_lingkup' ? 'Tulis satu poin ruang lingkup per baris' : undefined} value={form[field.name] || ''} onChange={(e) => setForm({ ...form, [field.name]: e.target.value })} />
                : <Input readOnly={field.name === 'termin_2_persen' || field.name === 'termin_2_nilai'} required={field.required} type={field.type === 'money' || field.type === 'decimal' ? 'number' : field.type} value={form[field.name] || ''} onChange={(e) => setForm({ ...form, [field.name]: e.target.value })} />}
            </div>
          ))}
       </div>
     </section>)}
      <section className="bg-white rounded-2xl border border-gray-100 p-6 mb-6">
        <h2 className="text-lg font-semibold mb-3" style={{ color: '#0B2545' }}>Lampiran Tambahan</h2>
        <input type="file" multiple onChange={(e) => setFiles((prev) => [...prev, ...Array.from(e.target.files || [])])} className="block w-full text-sm text-gray-600 file:mr-3 file:rounded-lg file:border-0 file:bg-gray-100 file:px-3 file:py-1.5 file:text-sm file:font-medium file:text-gray-700 hover:file:bg-gray-200" />
        {files.length > 0 && (
          <ul className="mt-3 space-y-2">
            {files.map((file, index) => (
              <li key={`${file.name}-${index}`} className="flex items-center justify-between gap-3 rounded-lg border border-gray-100 bg-gray-50 px-3 py-2 text-sm">
                <span className="truncate text-gray-800">{file.name}</span>
                <button type="button" onClick={() => setFiles((prev) => prev.filter((_, i) => i !== index))} className="shrink-0 text-gray-400 hover:text-red-600" title="Hapus lampiran">
                  Hapus
                </button>
              </li>
            ))}
          </ul>
        )}
      </section>
     <Button onClick={submit} className="flex items-center gap-2 text-white transition hover:brightness-95" style={{ background: '#C8102E' }}>
       <Send className="w-4 h-4" /> {id ? 'Simpan dan Ajukan Ulang' : 'Kirim Pengajuan'}
     </Button>
   </div>
 }
