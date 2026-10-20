import { useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { isAxiosError } from 'axios'
import { AlertTriangle, ArrowDown, ArrowLeft, ArrowRight, ArrowUp, Check, Plus, RefreshCw, Send, Trash2, X } from 'lucide-react'
import { toast } from 'sonner'
import { agreementService, type AgreementDocument, type AgreementSchema } from '@/services/agreement-document.service'
import { bacaDraf, hapusDraf, isianKosong, pindahkanIsian, simpanDraf, type Baris, type FieldDef, type HasilPindah, type Isian, type Nilai } from '@/lib/agreement-draft'
import { angkaSaja, bacaAngka, barisGrup, daftar, dariFormData, kosongDi, ribuan, ringkasNilai, rupiah, siapkanIsian, subBagi, teks, tulisAngka } from '@/lib/agreement-isian'
import { useAuthStore } from '@/store/auth.store'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import StatusBadge from '@/components/common/StatusBadge'
import type { SubmissionStatus } from '@/types'

// Termin versi PKS sebelum PR 10 (termin 1 dan 2): termin kedua dihitung dari termin
// pertama. Masih dipakai revisi pengajuan yang dikunci ke versi itu (keputusan D1).
function hitungTermin(isian: Isian, adaTermin: boolean): Isian {
  if (!adaTermin) return isian
  const n = (v?: Nilai) => bacaAngka(teks(v))
  return {
    ...isian,
    termin_2_persen: String(Math.max(0, 100 - n(isian.termin_1_persen))),
    termin_2_nilai: String(Math.max(0, n(isian.nilai_kontrak) - n(isian.termin_1_nilai))),
  }
}

const OTOMATIS = new Set(['termin_2_persen', 'termin_2_nilai'])

// Gulir ke elemen sesudah layar digambar ulang (mis. setelah kembali dari Tinjau).
function gulirKe(id: string) {
  setTimeout(() => {
    const el = document.getElementById(id)
    el?.scrollIntoView({ behavior: 'smooth', block: 'center' })
    if (el instanceof HTMLInputElement || el instanceof HTMLTextAreaElement || el instanceof HTMLSelectElement) el.focus({ preventScroll: true })
  }, 50)
}

// Isian satu nilai. Angka memakai input teks (inputMode), bukan type=number, supaya
// scroll mouse di atasnya tidak mengubah angka. Uang tampil dengan titik ribuan dan
// disimpan sebagai angka saja; persen dan angka boleh berkoma.
function InputIsian({ f, id = f.name, nilai, ubah, readOnly }: { f: FieldDef; id?: string; nilai: string; ubah: (v: string) => void; readOnly?: boolean }) {
  if (f.type === 'textarea') {
    return <Textarea id={id} value={nilai} readOnly={readOnly} onChange={(e) => ubah(e.target.value)} placeholder={f.name === 'ruang_lingkup' ? 'Tulis satu poin ruang lingkup per baris' : undefined} />
  }
  if (f.type === 'select') {
    return (
      <select id={id} value={nilai} disabled={readOnly} onChange={(e) => ubah(e.target.value)} className="h-9 w-full rounded-md border border-input bg-white px-2 text-sm">
        <option value="">Pilih...</option>
        {(f.options ?? []).map((o) => <option key={o} value={o}>{o}</option>)}
      </select>
    )
  }
  const otomatis = readOnly || OTOMATIS.has(f.name)
  const gayaOtomatis = otomatis ? 'bg-gray-50 text-gray-600' : ''
  if (f.type === 'money') {
    return (
      <div className="relative">
        <span className="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 text-sm text-gray-500">Rp</span>
        <Input id={id} inputMode="numeric" autoComplete="off" readOnly={otomatis} className={`pl-8 ${gayaOtomatis}`} value={ribuan(nilai)} onChange={(e) => ubah(angkaSaja(e.target.value))} />
      </div>
    )
  }
  if (f.type === 'percent' || f.type === 'number') {
    return (
      <div className="relative">
        <Input id={id} inputMode="decimal" autoComplete="off" readOnly={otomatis} className={`${f.type === 'percent' ? 'pr-7' : ''} ${gayaOtomatis}`} value={nilai} onChange={(e) => ubah(e.target.value.replace(/[^\d.,]/g, ''))} />
        {f.type === 'percent' && <span className="pointer-events-none absolute right-2.5 top-1/2 -translate-y-1/2 text-sm text-gray-500">%</span>}
      </div>
    )
  }
  // Saran isian (Admin): bisa dipilih atau diketik bebas.
  const saran = f.type === 'text' && f.options?.length ? `${id}-saran` : undefined
  return (
    <>
      <Input id={id} type={f.type} list={saran} readOnly={otomatis} className={gayaOtomatis} value={nilai} onChange={(e) => ubah(e.target.value)} />
      {saran && <datalist id={saran}>{f.options!.map((o) => <option key={o} value={o} />)}</datalist>}
    </>
  )
}

// Daftar poin (tipe list): satu isian per poin, bisa ditambah, dihapus, dan digeser.
function InputDaftar({ f, poin, ubah }: { f: FieldDef; poin: string[]; ubah: (v: string[]) => void }) {
  const geser = (i: number, arah: number) => {
    const baru = [...poin]
    ;[baru[i], baru[i + arah]] = [baru[i + arah], baru[i]]
    ubah(baru)
  }
  return (
    <div className="space-y-2">
      {poin.map((p, i) => (
        <div key={i} className="flex items-center gap-1.5">
          <span className="w-6 shrink-0 text-right text-sm text-gray-500">{i + 1}.</span>
          <Input id={i === 0 ? f.name : `${f.name}-${i}`} aria-label={`${f.label} poin ${i + 1}`} value={p} onChange={(e) => ubah(poin.map((x, k) => (k === i ? e.target.value : x)))} />
          <Button type="button" variant="ghost" size="icon-sm" disabled={i === 0} onClick={() => geser(i, -1)} aria-label={`Naikkan poin ${i + 1}`} title="Naikkan"><ArrowUp className="w-4 h-4" /></Button>
          <Button type="button" variant="ghost" size="icon-sm" disabled={i === poin.length - 1} onClick={() => geser(i, 1)} aria-label={`Turunkan poin ${i + 1}`} title="Turunkan"><ArrowDown className="w-4 h-4" /></Button>
          <Button type="button" variant="ghost" size="icon-sm" disabled={poin.length === 1} onClick={() => ubah(poin.filter((_, k) => k !== i))} aria-label={`Hapus poin ${i + 1}`} title="Hapus" className="text-red-600 hover:bg-red-50 hover:text-red-700"><Trash2 className="w-4 h-4" /></Button>
        </div>
      ))}
      <Button type="button" variant="outline" size="sm" onClick={() => ubah([...poin, ''])}><Plus className="w-3.5 h-3.5" /> Tambah poin</Button>
    </div>
  )
}

// Grup berulang (tipe group), mis. termin: satu kartu per baris. Pada grup yang
// membagi nilai, persen baris terakhir (sisa) dan nilai setiap baris dihitung
// otomatis dan tampil langsung; galat pembagian menahan Lanjut ke Tinjau.
function InputGrup({ f, baris, galat, ubah }: { f: FieldDef; baris: Baris[]; galat?: string; ubah: (v: Baris[]) => void }) {
  const [persen, nilai] = f.split_of ? subBagi(f) : ['', '']
  const tambah = () => {
    // Persen baris yang tadinya terakhir adalah sisa hitungan, jadi dikosongkan untuk diisi.
    const lama = f.split_of && baris.length ? baris.map((b, i) => (i === baris.length - 1 ? { ...b, [persen]: '' } : b)) : baris
    ubah([...lama, {}])
  }
  const jumlah = baris.reduce((n, b) => ({ persen: n.persen + bacaAngka(b[persen] ?? ''), nilai: n.nilai + Number(angkaSaja(b[nilai] ?? '') || 0) }), { persen: 0, nilai: 0 })
  return (
    <div className="space-y-3">
      {!baris.length && <p className="text-sm text-gray-500">Belum ada baris.</p>}
      {baris.map((b, i) => (
        <div key={i} className="rounded-xl border border-gray-200 p-4">
          <div className="mb-3 flex items-center justify-between">
            <p className="text-sm font-medium text-gray-800">{f.label} {i + 1}</p>
            <Button type="button" variant="outline" size="xs" disabled={f.required && baris.length === 1} onClick={() => ubah(baris.filter((_, k) => k !== i))} className="text-red-600 hover:bg-red-50 hover:text-red-700">Hapus</Button>
          </div>
          <div className="grid gap-3 md:grid-cols-2">
            {(f.fields ?? []).map((s) => {
              const otomatis = s.name === nilai || (s.name === persen && i === baris.length - 1)
              const id = `${f.name}-${i}-${s.name}`
              return (
                <div key={s.name} className={s.type === 'textarea' || s.options?.length ? 'md:col-span-2' : ''}>
                  <Label htmlFor={id} className="mb-1 text-xs text-gray-600">
                    {s.label}{s.required && !otomatis && ' *'}{otomatis && <span className="font-normal text-gray-400"> (otomatis{s.name === persen ? ', sisa' : ''})</span>}
                  </Label>
                  <InputIsian f={s} id={id} nilai={b[s.name] ?? ''} readOnly={otomatis} ubah={(v) => ubah(baris.map((x, k) => (k === i ? { ...x, [s.name]: v } : x)))} />
                </div>
              )
            })}
          </div>
        </div>
      ))}
      {galat ? (
        <p role="alert" className="rounded-lg bg-red-50 p-3 text-sm text-red-700">{galat}</p>
      ) : (
        f.split_of && baris.length > 0 && <p className="text-sm text-gray-600">Jumlah: {tulisAngka(Math.round(jumlah.persen * 100) / 100)}% · {rupiah(String(jumlah.nilai))}</p>
      )}
      <Button type="button" variant="outline" size="sm" onClick={tambah}><Plus className="w-3.5 h-3.5" /> Tambah {f.label.toLowerCase()}</Button>
    </div>
  )
}

// Wizard pembuatan dan revisi pengajuan perjanjian (PR 9, keputusan E). Langkah 1
// "Isi data": satu halaman, satu kartu per bagian isian, lalu lampiran. Langkah 2
// "Tinjau": ringkasan dan PDF pratinjau sebelum dikirim. Isian pengajuan baru yang
// belum dikirim disimpan di browser per pengguna, beserta versi templatenya.
export default function AgreementWizard() {
  const { id } = useParams()
  const revisi = !!id
  const navigate = useNavigate()
  const userId = useAuthStore((s) => s.user?.id ?? '')

  const [types, setTypes] = useState<{ code: string; name: string }[]>([])
  const [schema, setSchema] = useState<AgreementSchema>()
  const [doc, setDoc] = useState<AgreementDocument>()
  const [isian, setIsian] = useState<Isian>({})
  const [files, setFiles] = useState<File[]>([])
  const [tinjau, setTinjau] = useState(false)
  const [tandaBaru, setTandaBaru] = useState<Set<string>>(new Set())
  const [pesanPindah, setPesanPindah] = useState<{ versi: boolean; baru: number; dihapus: string[] } | null>(null)
  const [kosong, setKosong] = useState<string[]>([])
  const [galat, setGalat] = useState('')
  const [pratinjau, setPratinjau] = useState('')
  const [galatPratinjau, setGalatPratinjau] = useState('')
  const [memuatPratinjau, setMemuatPratinjau] = useState(false)
  const [mengirim, setMengirim] = useState(false)
  const urlPratinjau = useRef('')

  const sections = useMemo(() => schema?.sections ?? [], [schema])
  const adaTermin = sections.some((s) => s.fields.some((f) => f.name === 'termin_2_persen'))
  // Isian yang tampil dan dikirim: daftar dan grup dilengkapi, termin dihitung.
  const disiapkan = useMemo(() => siapkanIsian(sections, isian), [sections, isian])
  const kirimIsian = useMemo(() => hitungTermin(disiapkan.isian, adaTermin), [disiapkan, adaTermin])
  const galatGrup = disiapkan.galat
  const semuaKosong = [...sections.flatMap((s) => isianKosong(s, kirimIsian)), ...Object.values(galatGrup)]

  // Pindahkan isian ke field versi lain. Versi diganti Admin: field baru ditandai dan
  // digulir ke tempatnya. Ganti jenis: cukup catatan isian yang tidak terpakai.
  const terapkanPindah = (hasil: HasilPindah, baru: AgreementSchema, versi: boolean) => {
    setSchema(baru)
    setIsian(hasil.isian)
    setTinjau(false)
    setTandaBaru(new Set(versi ? hasil.baru : []))
    setPesanPindah(versi || hasil.dihapus.length ? { versi, baru: versi ? hasil.baru.length : 0, dihapus: hasil.dihapus } : null)
    if (versi && hasil.baru.length) gulirKe(hasil.baru[0])
  }

  useEffect(() => {
    if (id) {
      Promise.all([agreementService.get(id), agreementService.schemaFor(id)])
        .then(([d, s]) => {
          setDoc(d)
          setSchema(s)
          setIsian(dariFormData(d.form_data))
        })
        .catch((e: Error) => setGalat(e.message))
      return
    }
    agreementService.types()
      .then(async (daftar) => {
        setTypes(daftar)
        const draf = bacaDraf(userId)
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
  }, [id, userId])

  useEffect(() => {
    if (!revisi && schema) simpanDraf(userId, { code: schema.code, template_id: schema.template_id, sections: schema.sections, isian })
  }, [revisi, userId, schema, isian])

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

  // Lanjut ke Tinjau memeriksa semua isian wajib dan pembagian termin sekaligus, lalu
  // menggulir ke isian bermasalah yang pertama.
  const keTinjau = () => {
    if (!schema) return setKosong(['Jenis perjanjian'])
    const pertama = sections.flatMap((s) => s.fields).find((f) => kosongDi(f, kirimIsian[f.name]).length > 0 || galatGrup[f.name])
    if (pertama) {
      setKosong(semuaKosong)
      gulirKe(pertama.name)
      return
    }
    setKosong([])
    setTinjau(true)
    window.scrollTo({ top: 0, behavior: 'smooth' })
    void bentukPratinjau()
  }

  const ubahBagian = (i: number) => {
    setTinjau(false)
    gulirKe(`bagian-${i}`)
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
        hapusDraf(userId)
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

  const ubah = (name: string, v: Nilai) => setIsian((lama) => ({ ...lama, [name]: v }))

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
        {[{ judul: 'Isi data', ke: false }, { judul: 'Tinjau', ke: true }].map((l, k) => (
          <li key={l.judul}>
            <button
              type="button"
              onClick={() => (l.ke ? keTinjau() : setTinjau(false))}
              disabled={l.ke && (!schema || semuaKosong.length > 0)}
              aria-current={tinjau === l.ke ? 'step' : undefined}
              className={`flex items-center gap-1.5 rounded-full border px-3 py-1 text-sm transition disabled:cursor-not-allowed disabled:opacity-50 ${tinjau === l.ke ? 'border-[#0B2545] bg-[#0B2545] text-white' : 'border-gray-200 bg-white text-gray-700 hover:bg-gray-50'}`}
            >
              {!l.ke && tinjau ? <Check className="w-3.5 h-3.5" /> : <span className="text-xs">{k + 1}</span>}
              {l.judul}
            </button>
          </li>
        ))}
      </ol>

      {pesanPindah && (
        <div className="mb-5 flex gap-3 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800">
          <AlertTriangle className="w-4 h-4 mt-0.5 shrink-0" />
          <div className="flex-1 space-y-1">
            {pesanPindah.versi && (
              <p className="font-medium">
                {pesanPindah.baru > 0
                  ? `Template sudah diperbarui Admin. ${pesanPindah.baru} isian baru ditandai "Baru di template ini"; isi atau periksa sebelum mengirim.`
                  : 'Template sudah diperbarui Admin. Tidak ada isian baru; periksa kembali isian Anda sebelum mengirim.'}
              </p>
            )}
            {pesanPindah.dihapus.length > 0 && <p>Isian berikut tidak ada di template ini dan tidak ikut dikirim: {pesanPindah.dihapus.join(', ')}.</p>}
          </div>
          <Button variant="ghost" size="icon-sm" onClick={() => setPesanPindah(null)} aria-label="Tutup pemberitahuan" title="Tutup">
            <X className="w-4 h-4" />
          </Button>
        </div>
      )}

      {!tinjau ? (
        <div className="space-y-5">
          {!revisi && types.length > 1 && (
            <section className="bg-white rounded-2xl border border-gray-100 p-6">
              <h2 className="text-lg font-semibold mb-3" style={{ color: '#0B2545' }}>Jenis perjanjian</h2>
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
            </section>
          )}
          {!revisi && !types.length && <p className="text-sm text-gray-500">Belum ada jenis perjanjian yang bisa diajukan.</p>}

          {sections.map((s, i) => (
            <section key={s.title + i} id={`bagian-${i}`} className="bg-white rounded-2xl border border-gray-100 p-6 scroll-mt-6">
              <h2 className="text-lg font-semibold mb-4" style={{ color: '#0B2545' }}>{s.title}</h2>
              <div className="grid md:grid-cols-2 gap-4">
                {s.fields.map((f) => (
                  <div key={f.name} id={f.type === 'group' ? f.name : undefined} className={['textarea', 'list', 'group'].includes(f.type) ? 'md:col-span-2 scroll-mt-6' : ''}>
                    <Label htmlFor={f.type === 'group' ? `${f.name}-0-${f.fields?.[0]?.name}` : f.name} className="mb-1.5 text-gray-700">
                      {f.label}{f.required && ' *'}
                      {tandaBaru.has(f.name) && <span className="ml-2 rounded-full bg-amber-100 px-2 py-0.5 text-[11px] font-medium text-amber-800">Baru di template ini</span>}
                    </Label>
                    {f.type === 'list' ? (
                      <InputDaftar f={f} poin={daftar(kirimIsian[f.name])} ubah={(v) => ubah(f.name, v)} />
                    ) : f.type === 'group' ? (
                      <InputGrup f={f} baris={barisGrup(kirimIsian[f.name])} galat={galatGrup[f.name]} ubah={(v) => ubah(f.name, v)} />
                    ) : (
                      <InputIsian f={f} nilai={teks(kirimIsian[f.name])} ubah={(v) => ubah(f.name, v)} />
                    )}
                  </div>
                ))}
              </div>
            </section>
          ))}

          {schema && (
            <section className="bg-white rounded-2xl border border-gray-100 p-6">
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
            </section>
          )}

          {kosong.length > 0 && <p className="rounded-lg bg-red-50 p-3 text-sm text-red-700">Lengkapi dulu: {kosong.join(', ')}.</p>}
          <div className="flex justify-end">
            <Button onClick={keTinjau} disabled={!schema} className="text-white" style={{ background: '#0B2545' }}>
              Lanjut ke Tinjau <ArrowRight className="w-4 h-4" />
            </Button>
          </div>
        </div>
      ) : (
        <div className="space-y-5">
          <section className="bg-white rounded-2xl border border-gray-100 p-6">
            <div className="grid gap-4 md:grid-cols-2">
              {sections.map((s, i) => (
                <div key={s.title + i} className="rounded-xl border border-gray-100 p-4">
                  <div className="flex items-center justify-between mb-2">
                    <h3 className="font-medium text-gray-900">{s.title}</h3>
                    <Button variant="outline" size="xs" onClick={() => ubahBagian(i)}>Ubah</Button>
                  </div>
                  <dl className="space-y-1 text-sm">
                    {s.fields.map((f) => (
                      <div key={f.name} className="flex gap-2">
                        <dt className="w-2/5 shrink-0 text-gray-500">{f.label}</dt>
                        <dd className="text-gray-800 whitespace-pre-line break-words">{ringkasNilai(f, kirimIsian[f.name])}</dd>
                      </div>
                    ))}
                  </dl>
                </div>
              ))}
            </div>
          </section>
          <section className="bg-white rounded-2xl border border-gray-100 p-6">
            <div className="flex items-center justify-between mb-2">
              <h3 className="font-medium text-gray-900">Pratinjau dokumen</h3>
              <Button variant="outline" size="sm" onClick={() => void bentukPratinjau()} disabled={memuatPratinjau}>
                <RefreshCw className={`w-4 h-4 ${memuatPratinjau ? 'animate-spin' : ''}`} /> Perbarui pratinjau
              </Button>
            </div>
            {galatPratinjau && <p className="mb-2 rounded-lg bg-red-50 p-3 text-sm text-red-700">{galatPratinjau}</p>}
            {memuatPratinjau && !pratinjau && <div className="rounded-xl border border-gray-100 p-12 text-center text-gray-400">Membentuk pratinjau...</div>}
            {pratinjau && <iframe title="Pratinjau dokumen perjanjian" src={pratinjau} className="h-[75vh] w-full rounded-xl border border-gray-200" />}
          </section>
          <div className="flex justify-between">
            <Button variant="outline" onClick={() => setTinjau(false)}>
              <ArrowLeft className="w-4 h-4" /> Kembali ke Isi data
            </Button>
            <Button onClick={() => void kirim()} disabled={mengirim} className="flex items-center gap-2 text-white" style={{ background: '#C8102E' }}>
              <Send className="w-4 h-4" /> {mengirim ? 'Mengirim...' : revisi ? 'Simpan dan Ajukan Ulang' : 'Kirim Pengajuan'}
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
