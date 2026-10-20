import { useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { isAxiosError } from 'axios'
import { AlertTriangle, ArrowLeft, ArrowRight, Check, RefreshCw, Send, X } from 'lucide-react'
import { toast } from 'sonner'
import { agreementService, type AgreementDocument, type AgreementSchema } from '@/services/agreement-document.service'
import { bacaDraf, hapusDraf, isianKosong, pindahkanIsian, simpanDraf, type FieldDef, type HasilPindah, type Isian } from '@/lib/agreement-draft'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import StatusBadge from '@/components/common/StatusBadge'
import type { SubmissionStatus } from '@/types'

type Langkah = { jenis: 'jenis' } | { jenis: 'bagian'; i: number } | { jenis: 'lampiran' } | { jenis: 'tinjau' }

// Termin PKS: termin kedua dihitung dari termin pertama. Diganti grup termin di PR 10.
function hitungTermin(isian: Isian, adaTermin: boolean): Isian {
  if (!adaTermin) return isian
  const n = (v?: string) => (Number.isFinite(Number(v)) ? Number(v) : 0)
  return {
    ...isian,
    termin_2_persen: String(Math.max(0, 100 - n(isian.termin_1_persen))),
    termin_2_nilai: String(Math.max(0, n(isian.nilai_kontrak) - n(isian.termin_1_nilai))),
  }
}

const OTOMATIS = new Set(['termin_2_persen', 'termin_2_nilai'])

function InputIsian({ f, nilai, ubah }: { f: FieldDef; nilai: string; ubah: (v: string) => void }) {
  if (f.type === 'textarea') {
    return <Textarea id={f.name} value={nilai} onChange={(e) => ubah(e.target.value)} placeholder={f.name === 'ruang_lingkup' ? 'Tulis satu poin ruang lingkup per baris' : undefined} />
  }
  if (f.type === 'select') {
    return (
      <select id={f.name} value={nilai} onChange={(e) => ubah(e.target.value)} className="h-9 w-full rounded-md border border-input bg-white px-2 text-sm">
        <option value="">Pilih...</option>
        {(f.options ?? []).map((o) => <option key={o} value={o}>{o}</option>)}
      </select>
    )
  }
  const tipe = ['money', 'number', 'percent'].includes(f.type) ? 'number' : f.type
  return <Input id={f.name} type={tipe} step={f.type === 'money' ? undefined : 'any'} readOnly={OTOMATIS.has(f.name)} value={nilai} onChange={(e) => ubah(e.target.value)} />
}

// Wizard pembuatan dan revisi pengajuan perjanjian (PR 9 tahap 9.6). Langkah 1
// "Isi data": pilih jenis, satu langkah per bagian isian, lampiran. Langkah 2
// "Tinjau": ringkasan dan PDF pratinjau sebelum dikirim. Isian pengajuan baru
// yang belum dikirim disimpan di browser beserta versi templatenya.
export default function AgreementWizard() {
  const { id } = useParams()
  const revisi = !!id
  const navigate = useNavigate()

  const [types, setTypes] = useState<{ code: string; name: string }[]>([])
  const [schema, setSchema] = useState<AgreementSchema>()
  const [doc, setDoc] = useState<AgreementDocument>()
  const [isian, setIsian] = useState<Isian>({})
  const [files, setFiles] = useState<File[]>([])
  const [posisi, setPosisi] = useState(0)
  const [tandaBaru, setTandaBaru] = useState<Set<string>>(new Set())
  const [pesanPindah, setPesanPindah] = useState<{ versi: boolean; dihapus: string[] } | null>(null)
  const [galatLangkah, setGalatLangkah] = useState<string[]>([])
  const [galat, setGalat] = useState('')
  const [pratinjau, setPratinjau] = useState('')
  const [galatPratinjau, setGalatPratinjau] = useState('')
  const [memuatPratinjau, setMemuatPratinjau] = useState(false)
  const [mengirim, setMengirim] = useState(false)
  const urlPratinjau = useRef('')

  const sections = useMemo(() => schema?.sections ?? [], [schema])
  const adaTermin = sections.some((s) => s.fields.some((f) => f.name === 'termin_2_persen'))
  const kirimIsian = useMemo(() => hitungTermin(isian, adaTermin), [isian, adaTermin])
  const langkah: Langkah[] = useMemo(() => [
    ...(revisi ? [] : [{ jenis: 'jenis' } as const]),
    ...sections.map((_, i) => ({ jenis: 'bagian', i }) as const),
    { jenis: 'lampiran' } as const,
    { jenis: 'tinjau' } as const,
  ], [revisi, sections])
  const sekarang = langkah[Math.min(posisi, langkah.length - 1)]

  const lengkap = (l: Langkah) => (l.jenis === 'jenis' ? !!schema : l.jenis === 'bagian' ? isianKosong(sections[l.i], kirimIsian).length === 0 : true)
  const bisaKe = (k: number) => langkah.slice(0, k).every(lengkap)

  // Pindahkan isian ke field versi lain. Versi diganti Admin: field baru ditandai
  // dan pengaju diarahkan ke langkahnya. Ganti jenis: cukup catatan isian yang tidak terpakai.
  const terapkanPindah = (hasil: HasilPindah, baru: AgreementSchema, versi: boolean) => {
    setSchema(baru)
    setIsian(hasil.isian)
    setTandaBaru(new Set(versi ? hasil.baru : []))
    setPesanPindah(versi || hasil.dihapus.length ? { versi, dihapus: hasil.dihapus } : null)
    if (versi && hasil.baru.length) {
      const i = baru.sections.findIndex((s) => s.fields.some((f) => hasil.baru.includes(f.name)))
      if (i >= 0) setPosisi(i + (revisi ? 0 : 1))
    }
  }

  useEffect(() => {
    if (id) {
      Promise.all([agreementService.get(id), agreementService.schemaFor(id)])
        .then(([d, s]) => {
          setDoc(d)
          setSchema(s)
          setIsian(Object.fromEntries(Object.entries(d.form_data).map(([k, v]) => [k, String(v ?? '')])))
        })
        .catch((e: Error) => setGalat(e.message))
      return
    }
    agreementService.types()
      .then(async (daftar) => {
        setTypes(daftar)
        const draf = bacaDraf()
        const code = draf && daftar.some((t) => t.code === draf.code) ? draf.code : daftar.length === 1 ? daftar[0].code : ''
        if (!code) return
        const s = await agreementService.schema(code)
        if (draf?.code !== code) return setSchema(s)
        if (draf.template_id === s.template_id) {
          setSchema(s)
          setIsian(draf.isian)
        } else {
          terapkanPindah(pindahkanIsian(draf.isian, draf.sections, s.sections), s, true)
        }
      })
      .catch((e: Error) => setGalat(e.message))
    // terapkanPindah hanya memakai setter dan nilai awal, jadi cukup dijalankan sekali per halaman.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id])

  useEffect(() => {
    if (!revisi && schema) simpanDraf({ code: schema.code, template_id: schema.template_id, sections: schema.sections, isian })
  }, [revisi, schema, isian])

  useEffect(() => () => { if (urlPratinjau.current) URL.revokeObjectURL(urlPratinjau.current) }, [])

  const pilihJenis = async (code: string) => {
    if (code === schema?.code) return
    try {
      const s = await agreementService.schema(code)
      terapkanPindah(pindahkanIsian(isian, sections, s.sections), s, false)
    } catch (e) {
      toast.error((e as Error).message)
    }
  }

  // Versi template diganti Admin saat wizard terbuka (409): muat versi aktif dan pindahkan isian.
  const segarkanVersi = async () => {
    if (!schema) return
    const s = await agreementService.schema(schema.code)
    terapkanPindah(pindahkanIsian(isian, sections, s.sections), s, true)
    setPratinjau('')
  }

  const bentukPratinjau = async () => {
    if (!schema) return
    setMemuatPratinjau(true)
    setGalatPratinjau('')
    try {
      const blob = await agreementService.previewDraft({
        document_type_code: schema.code,
        template_id: schema.template_id,
        form_data: kirimIsian,
        attachment_names: files.map((f) => f.name),
        document_id: id,
      })
      if (urlPratinjau.current) URL.revokeObjectURL(urlPratinjau.current)
      urlPratinjau.current = URL.createObjectURL(blob)
      setPratinjau(urlPratinjau.current)
    } catch (e) {
      const err = e as Error & { status?: number }
      if (err.status === 409 && !revisi) await segarkanVersi()
      else setGalatPratinjau(err.message)
    } finally {
      setMemuatPratinjau(false)
    }
  }

  const pindahKe = (k: number) => {
    if (!bisaKe(k)) return
    setGalatLangkah([])
    setPosisi(k)
    if (langkah[k].jenis === 'tinjau') void bentukPratinjau()
  }

  const lanjut = () => {
    const kosong = sekarang.jenis === 'bagian' ? isianKosong(sections[sekarang.i], kirimIsian) : sekarang.jenis === 'jenis' && !schema ? ['Jenis perjanjian'] : []
    if (kosong.length) return setGalatLangkah(kosong)
    pindahKe(posisi + 1)
  }

  const kirim = async () => {
    if (!schema) return
    setMengirim(true)
    try {
      if (id) {
        await agreementService.update(id, kirimIsian)
        if (doc?.status === 'NEED_REVISION') await agreementService.resubmit(id, files)
      } else {
        await agreementService.create({ document_type_code: schema.code, template_id: schema.template_id, form_data: kirimIsian }, files)
        hapusDraf()
      }
      toast.success(revisi ? 'Revisi pengajuan tersimpan' : 'Pengajuan perjanjian terkirim')
      navigate('/dashboard/agreement-documents')
    } catch (e) {
      if (!revisi && isAxiosError(e) && e.response?.status === 409) await segarkanVersi()
      toast.error((e as Error).message || 'Gagal mengirim pengajuan', { duration: 8000 })
    } finally {
      setMengirim(false)
    }
  }

  const judulLangkah = (l: Langkah) => (l.jenis === 'jenis' ? 'Jenis' : l.jenis === 'bagian' ? sections[l.i].title : l.jenis === 'lampiran' ? 'Lampiran' : 'Tinjau')
  const ubah = (name: string, v: string) => setIsian((lama) => ({ ...lama, [name]: v }))

  if (galat) return <div className="p-12 text-center text-red-600">{galat}</div>

  return (
    <div className="p-6 max-w-5xl mx-auto">
      <Button variant="outline" size="icon" onClick={() => navigate(-1)} title="Kembali" aria-label="Kembali" className="mt-0.5 mb-2">
        <ArrowLeft className="w-5 h-5 text-gray-600" />
      </Button>
      {doc && (
        <div className="flex items-center gap-3 mb-4">
          <span className="text-xs font-mono font-medium bg-gray-100 text-gray-600 px-2 py-1 rounded">{doc.ticket_number}</span>
          <StatusBadge status={doc.status as SubmissionStatus} />
        </div>
      )}
      <h1 className="text-2xl font-bold mb-5" style={{ color: '#0B2545' }}>{revisi ? 'Revisi' : 'Pengajuan'} {schema?.name ?? 'Perjanjian'}</h1>

      <ol className="flex flex-wrap gap-2 mb-6" aria-label="Langkah pengajuan">
        {langkah.map((l, k) => {
          const aktif = k === posisi
          const selesai = k < posisi && lengkap(l)
          return (
            <li key={k}>
              <button
                type="button"
                onClick={() => pindahKe(k)}
                disabled={!bisaKe(k)}
                aria-current={aktif ? 'step' : undefined}
                className={`flex items-center gap-1.5 rounded-full border px-3 py-1 text-sm transition disabled:cursor-not-allowed disabled:opacity-50 ${aktif ? 'border-[#0B2545] bg-[#0B2545] text-white' : 'border-gray-200 bg-white text-gray-700 hover:bg-gray-50'}`}
              >
                {selesai ? <Check className="w-3.5 h-3.5" /> : <span className="text-xs">{l.jenis === 'tinjau' ? '2' : `1.${k + 1}`}</span>}
                {judulLangkah(l)}
              </button>
            </li>
          )
        })}
      </ol>

      {pesanPindah && (
        <div className="mb-5 flex gap-3 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800">
          <AlertTriangle className="w-4 h-4 mt-0.5 shrink-0" />
          <div className="flex-1 space-y-1">
            {pesanPindah.versi && <p className="font-medium">Template sudah diperbarui Admin. Periksa kembali isian Anda; isian yang baru ditandai.</p>}
            {pesanPindah.dihapus.length > 0 && <p>Isian berikut tidak ada di template ini dan tidak ikut dikirim: {pesanPindah.dihapus.join(', ')}.</p>}
          </div>
          <Button variant="ghost" size="icon-sm" onClick={() => setPesanPindah(null)} aria-label="Tutup pemberitahuan" title="Tutup">
            <X className="w-4 h-4" />
          </Button>
        </div>
      )}

      <section className="bg-white rounded-2xl border border-gray-100 p-6 mb-5">
        {sekarang.jenis === 'jenis' && (
          <div className="space-y-3">
            <h2 className="text-lg font-semibold" style={{ color: '#0B2545' }}>Pilih jenis perjanjian</h2>
            {!types.length && <p className="text-sm text-gray-500">Belum ada jenis perjanjian yang bisa diajukan.</p>}
            <div className="grid sm:grid-cols-2 gap-3">
              {types.map((t) => (
                <button
                  key={t.code}
                  type="button"
                  onClick={() => void pilihJenis(t.code)}
                  aria-pressed={schema?.code === t.code}
                  className={`rounded-xl border p-4 text-left transition ${schema?.code === t.code ? 'border-[#0B2545] ring-1 ring-[#0B2545]' : 'border-gray-200 hover:bg-gray-50'}`}
                >
                  <p className="font-medium text-gray-900">{t.name}</p>
                  <p className="text-xs font-mono text-gray-500">{t.code}</p>
                </button>
              ))}
            </div>
          </div>
        )}

        {sekarang.jenis === 'bagian' && (
          <div>
            <h2 className="text-lg font-semibold mb-4" style={{ color: '#0B2545' }}>{sections[sekarang.i].title}</h2>
            <div className="grid md:grid-cols-2 gap-4">
              {sections[sekarang.i].fields.map((f) => (
                <div key={f.name} className={f.type === 'textarea' ? 'md:col-span-2' : ''}>
                  <Label htmlFor={f.name} className="mb-1.5 text-gray-700">
                    {f.label}{f.required && ' *'}
                    {tandaBaru.has(f.name) && <span className="ml-2 rounded-full bg-amber-100 px-2 py-0.5 text-[11px] font-medium text-amber-800">Baru di template ini</span>}
                  </Label>
                  <InputIsian f={f} nilai={kirimIsian[f.name] ?? ''} ubah={(v) => ubah(f.name, v)} />
                </div>
              ))}
            </div>
          </div>
        )}

        {sekarang.jenis === 'lampiran' && (
          <div>
            <h2 className="text-lg font-semibold mb-3" style={{ color: '#0B2545' }}>Lampiran</h2>
            {!!doc?.attachments?.length && (
              <ul className="mb-3 space-y-1 text-sm text-gray-600">
                {doc.attachments.map((a) => <li key={a.id}>{a.file_name} <span className="text-gray-400">(sudah terlampir)</span></li>)}
              </ul>
            )}
            <input type="file" multiple aria-label="Pilih file lampiran" onChange={(e) => setFiles((lama) => [...lama, ...Array.from(e.target.files || [])])} className="block w-full text-sm text-gray-600 file:mr-3 file:rounded-lg file:border-0 file:bg-gray-100 file:px-3 file:py-1.5 file:text-sm file:font-medium file:text-gray-700 hover:file:bg-gray-200" />
            {files.length > 0 && (
              <ul className="mt-3 space-y-2">
                {files.map((file, index) => (
                  <li key={`${file.name}-${index}`} className="flex items-center justify-between gap-3 rounded-lg border border-gray-100 bg-gray-50 px-3 py-2 text-sm">
                    <span className="truncate text-gray-800">{file.name}</span>
                    <Button variant="outline" size="xs" type="button" onClick={() => setFiles((lama) => lama.filter((_, i) => i !== index))} aria-label="Hapus lampiran" className="shrink-0 text-red-600 hover:bg-red-50 hover:text-red-700">Hapus</Button>
                  </li>
                ))}
              </ul>
            )}
            {!revisi && <p className="mt-3 text-xs text-gray-400">File lampiran tidak ikut tersimpan di browser; pilih ulang kalau halaman ditutup sebelum dikirim.</p>}
          </div>
        )}

        {sekarang.jenis === 'tinjau' && (
          <div className="space-y-5">
            <div className="grid gap-4 md:grid-cols-2">
              {sections.map((s, i) => (
                <div key={s.title + i} className="rounded-xl border border-gray-100 p-4">
                  <div className="flex items-center justify-between mb-2">
                    <h3 className="font-medium text-gray-900">{s.title}</h3>
                    <Button variant="outline" size="xs" onClick={() => pindahKe(i + (revisi ? 0 : 1))}>Ubah</Button>
                  </div>
                  <dl className="space-y-1 text-sm">
                    {s.fields.map((f) => (
                      <div key={f.name} className="flex gap-2">
                        <dt className="w-2/5 shrink-0 text-gray-500">{f.label}</dt>
                        <dd className="text-gray-800 whitespace-pre-line break-words">{kirimIsian[f.name]?.trim() || '-'}</dd>
                      </div>
                    ))}
                  </dl>
                </div>
              ))}
            </div>
            <div>
              <div className="flex items-center justify-between mb-2">
                <h3 className="font-medium text-gray-900">Pratinjau dokumen</h3>
                <Button variant="outline" size="sm" onClick={() => void bentukPratinjau()} disabled={memuatPratinjau}>
                  <RefreshCw className={`w-4 h-4 ${memuatPratinjau ? 'animate-spin' : ''}`} /> Perbarui pratinjau
                </Button>
              </div>
              {galatPratinjau && <p className="mb-2 rounded-lg bg-red-50 p-3 text-sm text-red-700">{galatPratinjau}</p>}
              {memuatPratinjau && !pratinjau && <div className="rounded-xl border border-gray-100 p-12 text-center text-gray-400">Membentuk pratinjau...</div>}
              {pratinjau && <iframe title="Pratinjau dokumen perjanjian" src={pratinjau} className="h-[75vh] w-full rounded-xl border border-gray-200" />}
            </div>
          </div>
        )}

        {galatLangkah.length > 0 && (
          <p className="mt-4 rounded-lg bg-red-50 p-3 text-sm text-red-700">Lengkapi dulu: {galatLangkah.join(', ')}.</p>
        )}
      </section>

      <div className="flex justify-between">
        <Button variant="outline" onClick={() => pindahKe(posisi - 1)} disabled={posisi === 0}>
          <ArrowLeft className="w-4 h-4" /> Sebelumnya
        </Button>
        {sekarang.jenis === 'tinjau' ? (
          <Button onClick={() => void kirim()} disabled={mengirim} className="flex items-center gap-2 text-white" style={{ background: '#C8102E' }}>
            <Send className="w-4 h-4" /> {mengirim ? 'Mengirim...' : revisi ? 'Simpan dan Ajukan Ulang' : 'Kirim Pengajuan'}
          </Button>
        ) : (
          <Button onClick={lanjut} className="text-white" style={{ background: '#0B2545' }}>
            Lanjut <ArrowRight className="w-4 h-4" />
          </Button>
        )}
      </div>
    </div>
  )
}
