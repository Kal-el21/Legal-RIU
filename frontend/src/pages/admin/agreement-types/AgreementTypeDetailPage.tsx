import { useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { isAxiosError } from 'axios'
import { AlertTriangle, ArrowLeft, CheckCircle2, Copy, Download, Eye, FileUp, Info, Save, Trash2, Upload } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  useActivateTemplate,
  useAgreementTypes,
  useDeleteTemplate,
  usePlaceholders,
  useTemplateVersions,
  useUpdateAgreementType,
  useUploadTemplate,
} from '@/hooks/useAgreementType'
import { agreementTemplateService, type AgreementSectionDef, type AgreementTemplate, type TemplateValidation } from '@/services/agreement-template.service'
import type { AgreementType } from '@/services/agreement-type.service'
import FieldEditor from './FieldEditor'

const statusStyle: Record<AgreementTemplate['status'], string> = {
  ACTIVE: 'bg-green-50 text-green-700 border-green-200',
  DRAFT: 'bg-amber-50 text-amber-700 border-amber-200',
  ARCHIVED: 'bg-gray-100 text-gray-600 border-gray-200',
}
const statusLabel: Record<AgreementTemplate['status'], string> = { ACTIVE: 'Aktif', DRAFT: 'Draft', ARCHIVED: 'Diarsipkan' }

function formatDate(value?: string) {
  if (!value) return '-'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '-' : date.toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' })
}

// Bentuk pembanding isian: urutan kunci dan isian kosong tidak dianggap perubahan.
function sidik(sections?: AgreementSectionDef[] | null) {
  return JSON.stringify((sections ?? []).map((s) => ({
    title: s.title.trim(),
    fields: s.fields.map((f) => ({ name: f.name.trim().toLowerCase(), label: f.label.trim(), type: f.type, required: !!f.required, options: (f.options ?? []).map((o) => o.trim()).filter(Boolean) })),
  })))
}

// "Jadikan isian": placeholder dari Word yang belum punya isian ditambahkan sebagai
// teks singkat. Turunan (_TERBILANG, _HARI) yang dasarnya ikut ditambahkan dilewati,
// karena turunan muncul sendiri setelah tipenya diganti uang, angka, persen, atau tanggal.
function jadikanIsian(sections: AgreementSectionDef[], tokens: string[]): AgreementSectionDef[] {
  const kode = tokens.map((t) => t.replace(/^\{\{|\}\}$/g, ''))
  const dasar = new Set(kode)
  const baru = kode
    .filter((k) => !/_(TERBILANG|HARI)$/.test(k) || !dasar.has(k.replace(/_(TERBILANG|HARI)$/, '')))
    .map((k) => k === 'RUANG_LINGKUP_LIST'
      ? { name: 'ruang_lingkup', label: 'Ruang Lingkup', type: 'textarea', required: false }
      : { name: k.toLowerCase(), label: k.toLowerCase().split('_').map((w) => w.charAt(0).toUpperCase() + w.slice(1)).join(' '), type: 'text', required: false })
  return [...sections, { title: 'Isian baru dari Word', fields: baru }]
}

export default function AgreementTypeDetailPage() {
  const { code = '' } = useParams()
  const { data: types = [], isLoading } = useAgreementTypes()
  const jenis = types.find((t) => t.code === code)
  if (isLoading) return <div className="p-12 text-center text-gray-400">Memuat data...</div>
  if (!jenis) return <div className="p-12 text-center text-gray-500">Jenis perjanjian {code} tidak ditemukan.</div>
  // key: isian di penyusun dimulai ulang dari data tersimpan kalau pindah jenis.
  return <DetailJenis key={jenis.id} jenis={jenis} />
}

function DetailJenis({ jenis }: { jenis: AgreementType }) {
  const code = jenis.code
  const { data: versions = [] } = useTemplateVersions(code)
  const { data: placeholders = [] } = usePlaceholders(code)
  const simpan = useUpdateAgreementType(code)
  const ubahStatus = useUpdateAgreementType(code, 'Status jenis perjanjian diubah')
  const unggah = useUploadTemplate(code)
  const aktifkan = useActivateTemplate(code)
  const hapus = useDeleteTemplate(code)

  const [draf, setDraf] = useState<AgreementSectionDef[]>(jenis.fields ?? [])
  const [nama, setNama] = useState(jenis.name)
  const [file, setFile] = useState<File | null>(null)
  const [namaVersi, setNamaVersi] = useState('')
  const [berlaku, setBerlaku] = useState('')
  const [catatan, setCatatan] = useState('')
  const [hasil, setHasil] = useState<TemplateValidation | null>(null)
  const [lihatToken, setLihatToken] = useState(false)

  const aktif = versions.find((v) => v.status === 'ACTIVE')
  const belumDisimpan = sidik(draf) !== sidik(jenis.fields) || nama.trim() !== jenis.name
  const bedaDariAktif = !!aktif && sidik(jenis.fields) !== sidik(aktif.fields)
  const kelompokToken = useMemo(() => {
    const groups = new Map<string, typeof placeholders>()
    placeholders.forEach((p) => groups.set(p.group || 'Lainnya', [...(groups.get(p.group || 'Lainnya') || []), p]))
    return [...groups.entries()]
  }, [placeholders])

  const galatUnggah = unggah.error
  const tolak: TemplateValidation | undefined = isAxiosError(galatUnggah) ? galatUnggah.response?.data?.errors : undefined

  const kirimVersi = (pakaiFile: boolean) => {
    setHasil(null)
    unggah.mutate(
      { file: pakaiFile ? file ?? undefined : undefined, payload: { code, name: namaVersi, note: catatan, effective_date: berlaku } },
      {
        onSuccess: (data) => {
          setHasil(data.validation)
          setFile(null)
          setNamaVersi('')
          setBerlaku('')
          setCatatan('')
        },
      },
    )
  }

  const konfirmasiAktif = (v: AgreementTemplate) => {
    const berjalan = aktif?.running_documents ?? 0
    const pesan = aktif && berjalan > 0
      ? `Aktifkan v${v.version}? ${berjalan} pengajuan yang sedang berjalan masih memakai v${aktif.version} dan akan tetap memakainya. Pengajuan baru memakai v${v.version}.`
      : `Aktifkan v${v.version}? Pengajuan baru akan memakai versi ini.`
    if (window.confirm(pesan)) aktifkan.mutate(v.id)
  }

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      <div>
        <Button asChild variant="outline" size="icon" className="mb-2" title="Kembali">
          <Link to="/admin/agreement-types" aria-label="Kembali"><ArrowLeft className="w-5 h-5 text-gray-600" /></Link>
        </Button>
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>{jenis.name}</h1>
            <p className="text-sm text-gray-500 mt-0.5">
              Kode <span className="font-mono">{jenis.code}</span> · {aktif ? `versi aktif v${aktif.version}` : 'belum ada versi aktif'} · {jenis.is_active ? 'ditawarkan ke pengaju' : 'nonaktif, tidak ditawarkan ke pengaju'}
            </p>
          </div>
          <Button variant="outline" disabled={ubahStatus.isPending} onClick={() => ubahStatus.mutate({ name: jenis.name, is_active: !jenis.is_active })}>
            {jenis.is_active ? 'Nonaktifkan' : 'Aktifkan jenis'}
          </Button>
        </div>
      </div>

      {(!aktif || bedaDariAktif) && (
        <div className="flex gap-3 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800">
          <AlertTriangle className="w-4 h-4 mt-0.5 shrink-0" />
          <p>
            {!aktif
              ? 'Jenis ini belum punya versi template aktif, jadi pengaju belum bisa memilihnya. Susun isian, unggah DOCX, lalu aktifkan versinya.'
              : 'Ada perubahan isian yang belum dipakai versi mana pun. Pengaju masih memakai isian versi aktif. Unggah DOCX atau Simpan sebagai versi baru, lalu aktifkan.'}
          </p>
        </div>
      )}

      <section className="bg-white rounded-2xl border border-gray-100 p-6">
        <div className="flex flex-wrap items-center justify-between gap-3 mb-1">
          <h2 className="font-semibold" style={{ color: '#0B2545' }}>Isian Form</h2>
          <Button
            onClick={() => simpan.mutate({ name: nama, fields: draf })}
            disabled={!belumDisimpan || simpan.isPending}
            className="flex items-center gap-2 text-white"
            style={{ background: '#0B2545' }}
          >
            <Save className="w-4 h-4" /> {simpan.isPending ? 'Menyimpan...' : 'Simpan isian'}
          </Button>
        </div>
        <p className="text-sm text-gray-500 mb-4">
          Setiap bagian menjadi satu langkah di form pengajuan. Kode isian ditulis di Word sebagai placeholder, mis. <code className="font-mono">{'{{NAMA_MITRA}}'}</code>.
          Perubahan di sini baru dipakai pengaju setelah ada versi template baru yang diaktifkan.
        </p>
        <div className="max-w-md mb-5 space-y-1.5">
          <Label htmlFor="nama-jenis">Nama jenis</Label>
          <Input id="nama-jenis" value={nama} onChange={(e) => setNama(e.target.value)} />
        </div>
        <FieldEditor sections={draf} onChange={setDraf} />
        {belumDisimpan && <p className="mt-4 text-sm text-amber-700">Ada perubahan yang belum disimpan.</p>}
      </section>

      <section className="bg-white rounded-2xl border border-gray-100 p-6">
        <h2 className="font-semibold flex items-center gap-2 mb-1" style={{ color: '#0B2545' }}>
          <FileUp className="w-4 h-4" /> Versi Template Baru
        </h2>
        <p className="text-sm text-gray-500 mb-4">
          DOCX dicocokkan dengan isian yang sudah disimpan, lalu diuji menjadi PDF. Tanpa file, <b>Simpan sebagai versi baru</b> memakai DOCX versi terakhir
          (mis. setelah mengganti label atau status wajib).
        </p>
        <div className="grid md:grid-cols-2 gap-4">
          <div className="space-y-1.5">
            <Label htmlFor="file-template">File Template (.docx)</Label>
            <Input id="file-template" type="file" accept=".docx" onChange={(e) => setFile(e.target.files?.[0] || null)} />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="nama-versi">Nama Versi</Label>
            <Input id="nama-versi" value={namaVersi} placeholder="mis. Revisi klausul denda" onChange={(e) => setNamaVersi(e.target.value)} />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="berlaku">Tanggal Berlaku</Label>
            <Input id="berlaku" type="date" value={berlaku} onChange={(e) => setBerlaku(e.target.value)} />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="catatan">Catatan Perubahan *</Label>
            <Textarea id="catatan" value={catatan} placeholder="Apa yang berubah dibanding versi sebelumnya" onChange={(e) => setCatatan(e.target.value)} />
          </div>
        </div>
        {belumDisimpan && <p className="mt-4 text-sm text-amber-700">Simpan isian dulu: versi baru memakai isian yang sudah disimpan.</p>}
        <div className="mt-5 flex flex-wrap gap-2">
          <Button
            onClick={() => kirimVersi(true)}
            disabled={!file || !catatan.trim() || belumDisimpan || unggah.isPending}
            className="flex items-center gap-2 text-white"
            style={{ background: '#C8102E' }}
          >
            <Upload className="w-4 h-4" /> {unggah.isPending ? 'Menguji template...' : 'Unggah & Uji'}
          </Button>
          <Button variant="outline" onClick={() => kirimVersi(false)} disabled={!versions.length || !catatan.trim() || belumDisimpan || unggah.isPending}>
            Simpan sebagai versi baru
          </Button>
        </div>

        {galatUnggah && (
          <div className="mt-4 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
            <p>{galatUnggah.message}</p>
            {!!tolak?.unknown_placeholders?.length && (
              <div className="mt-2">
                <p className="font-mono text-xs break-all">{tolak.unknown_placeholders.join(' ')}</p>
                <Button
                  variant="outline"
                  size="sm"
                  className="mt-2 bg-white"
                  onClick={() => {
                    setDraf(jadikanIsian(draf, tolak.unknown_placeholders))
                    unggah.reset()
                    toast.info('Isian ditambahkan sebagai teks singkat di bagian baru. Periksa jenisnya, simpan isian, lalu unggah ulang.')
                  }}
                >
                  Jadikan isian
                </Button>
              </div>
            )}
          </div>
        )}
        {hasil && hasil.warnings.length > 0 && (
          <div className="mt-4 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800 space-y-1">
            <p className="font-medium">Versi tersimpan sebagai Draft, dengan peringatan:</p>
            {hasil.warnings.map((w) => <p key={w}>{w}</p>)}
          </div>
        )}
      </section>

      <section className="bg-white rounded-2xl border border-gray-100 p-6">
        <button type="button" onClick={() => setLihatToken((v) => !v)} className="font-semibold flex items-center gap-2 w-full text-left" style={{ color: '#0B2545' }}>
          <Info className="w-4 h-4" /> Placeholder untuk Word
          <span className="ml-auto text-sm font-normal text-gray-400">{lihatToken ? 'Sembunyikan' : `Lihat ${placeholders.length} token`}</span>
        </button>
        <p className="text-sm text-gray-500 mt-1">Dibentuk dari isian yang sudah disimpan. Salin lalu tempel di Word sebagai teks biasa.</p>
        {lihatToken && (
          <div className="mt-4 space-y-5">
            {kelompokToken.map(([group, tokens]) => (
              <div key={group}>
                <h3 className="text-xs font-semibold uppercase tracking-wide text-gray-500 mb-2">{group}</h3>
                <div className="space-y-1.5">
                  {tokens.map((t) => (
                    <div key={t.token} className="flex flex-wrap items-center gap-x-3 gap-y-0.5 text-sm">
                      <Button
                        variant="outline"
                        size="xs"
                        className="font-mono"
                        title="Salin"
                        onClick={() => void navigator.clipboard.writeText(t.token).then(() => toast.success(`${t.token} disalin`))}
                      >
                        <Copy className="w-3 h-3" /> {t.token}
                      </Button>
                      <span className="text-gray-500">{t.description}</span>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        )}
      </section>

      <section className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
        <h2 className="font-semibold px-6 pt-6 pb-4" style={{ color: '#0B2545' }}>Riwayat Versi</h2>
        {!versions.length ? (
          <div className="p-8 text-center text-gray-400">Belum ada versi template.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-xs uppercase tracking-wide text-gray-500 border-b border-gray-100">
                  <th className="px-6 py-3 font-medium">Versi</th>
                  <th className="px-6 py-3 font-medium">Nama</th>
                  <th className="px-6 py-3 font-medium">Status</th>
                  <th className="px-6 py-3 font-medium">Dipakai</th>
                  <th className="px-6 py-3 font-medium">Diunggah</th>
                  <th className="px-6 py-3 font-medium text-right">Aksi</th>
                </tr>
              </thead>
              <tbody>
                {versions.map((v) => {
                  const bisaHapus = v.status !== 'ACTIVE' && v.running_documents === 0
                  return (
                    <tr key={v.id} className="border-b border-gray-50 last:border-0">
                      <td className="px-6 py-4 font-medium text-gray-800">v{v.version}</td>
                      <td className="px-6 py-4">
                        <p className="text-gray-800">{v.name}</p>
                        <p className="text-xs text-gray-400 mt-0.5">{v.file_name} · berlaku {formatDate(v.effective_date)}</p>
                        {v.note && <p className="text-xs text-gray-500 mt-1">{v.note}</p>}
                      </td>
                      <td className="px-6 py-4">
                        <span className={`inline-block text-xs px-2 py-0.5 rounded-full border ${statusStyle[v.status]}`}>{statusLabel[v.status]}</span>
                        {v.is_legacy && <p className="text-xs text-gray-500 mt-1">Format lama, hanya bisa diunduh</p>}
                      </td>
                      <td className="px-6 py-4 text-gray-600">{v.running_documents} pengajuan berjalan</td>
                      <td className="px-6 py-4 text-gray-600">
                        <p>{formatDate(v.created_at)}</p>
                        <p className="text-xs text-gray-400">{v.uploader?.full_name || 'Sistem'}</p>
                      </td>
                      <td className="px-6 py-4">
                        <div className="flex items-center justify-end gap-2">
                          {!v.is_legacy && (
                            <Button asChild variant="outline" size="icon-sm" title="Pratinjau PDF">
                              <a href={agreementTemplateService.fileUrl(v.id, 'preview')} target="_blank" rel="noreferrer" aria-label="Pratinjau PDF"><Eye className="w-4 h-4" /></a>
                            </Button>
                          )}
                          <Button asChild variant="outline" size="icon-sm" title="Unduh .docx">
                            <a href={agreementTemplateService.fileUrl(v.id, 'download')} aria-label="Unduh .docx"><Download className="w-4 h-4" /></a>
                          </Button>
                          <Button
                            variant="outline"
                            size="icon-sm"
                            disabled={!bisaHapus || hapus.isPending}
                            title={v.status === 'ACTIVE' ? 'Versi aktif tidak bisa dihapus' : v.running_documents > 0 ? 'Masih dipakai pengajuan berjalan' : 'Hapus versi'}
                            aria-label="Hapus versi"
                            className="text-red-600 hover:bg-red-50 hover:text-red-700"
                            onClick={() => window.confirm(`Hapus v${v.version}? Versi yang dihapus tidak bisa dipakai lagi.`) && hapus.mutate(v.id)}
                          >
                            <Trash2 className="w-4 h-4" />
                          </Button>
                          {v.status !== 'ACTIVE' && !v.is_legacy && (
                            <Button
                              onClick={() => konfirmasiAktif(v)}
                              disabled={aktifkan.isPending}
                              className="flex items-center gap-1.5 h-8 px-3 text-xs text-white"
                              style={{ background: '#0B2545' }}
                            >
                              <CheckCircle2 className="w-3.5 h-3.5" /> Aktifkan
                            </Button>
                          )}
                        </div>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </div>
  )
}
