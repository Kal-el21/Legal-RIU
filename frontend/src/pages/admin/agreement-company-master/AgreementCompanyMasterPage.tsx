import { useCallback, useEffect, useState } from 'react'
import { Pencil, Plus, Save, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { agreementService } from '@/services/agreement-document.service'
import { angkaSaja, ribuan, rupiah } from '@/lib/agreement-isian'
import { rentangPenandatangan, type Penandatangan } from '@/lib/penandatangan'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

// Penandatangan diatur di Daftar Penandatangan sejak PR 11, jadi kolom penandatangan
// lama di master tidak ditampilkan lagi.
const fieldLabels: Record<string, string> = {
  name: 'Nama Perusahaan',
  address: 'Alamat',
  npwp: 'NPWP',
  phone: 'Telepon',
  email: 'Email',
  pic: 'PIC',
  default_signing_place: 'Tempat Penandatanganan (Default)',
  default_agreement_number: 'Nomor Perjanjian (Default)',
}

const fieldKeys = Object.keys(fieldLabels)

interface IsianPenandatangan {
  id?: string
  name: string
  position: string
  // Angka saja; kosong berarti tak terbatas.
  max_value: string
  is_default: boolean
  is_active: boolean
}

const kosong: IsianPenandatangan = { name: '', position: '', max_value: '', is_default: false, is_active: true }

function DaftarPenandatangan() {
  const [daftar, setDaftar] = useState<Penandatangan[]>([])
  const [galat, setGalat] = useState('')
  const [isian, setIsian] = useState<IsianPenandatangan | null>(null)
  const [menyimpan, setMenyimpan] = useState(false)

  const muat = useCallback(async () => {
    try {
      setDaftar(await agreementService.signatories())
      setGalat('')
    } catch (e) {
      setGalat(e instanceof Error ? e.message : 'Gagal memuat Daftar Penandatangan')
    }
  }, [])

  useEffect(() => {
    void agreementService.signatories()
      .then(setDaftar)
      .catch((e: Error) => setGalat(e.message))
  }, [])

  const rentang = rentangPenandatangan(daftar)
  const adaTakTerbatas = daftar.some((p) => p.is_active && p.max_value === null)
  const batasTertinggi = Math.max(0, ...daftar.filter((p) => p.is_active && p.max_value !== null).map((p) => p.max_value as number))

  const simpan = async () => {
    if (!isian) return
    setMenyimpan(true)
    setGalat('')
    const isi = { name: isian.name, position: isian.position, max_value: isian.max_value ? Number(isian.max_value) : null, is_default: isian.is_default }
    try {
      if (isian.id) await agreementService.updateSignatory(isian.id, { ...isi, is_active: isian.is_active })
      else await agreementService.createSignatory(isi)
      toast.success(isian.id ? 'Penandatangan diperbarui' : 'Penandatangan ditambahkan')
      setIsian(null)
      await muat()
    } catch (e) {
      setGalat(e instanceof Error ? e.message : 'Gagal menyimpan penandatangan')
    } finally {
      setMenyimpan(false)
    }
  }

  const hapus = async (p: Penandatangan) => {
    if (!window.confirm(`Hapus ${p.name} (${p.position}) dari Daftar Penandatangan? Dokumen yang sudah disetujui tidak berubah.`)) return
    try {
      await agreementService.deleteSignatory(p.id)
      toast.success('Penandatangan dihapus')
      await muat()
    } catch (e) {
      setGalat(e instanceof Error ? e.message : 'Gagal menghapus penandatangan')
    }
  }

  return (
    <section className="bg-white rounded-2xl border border-gray-100 p-6">
      <div className="flex flex-wrap items-center justify-between gap-3 mb-1">
        <h2 className="font-semibold" style={{ color: '#0B2545' }}>Daftar Penandatangan</h2>
        <Button variant="outline" size="sm" onClick={() => setIsian({ ...kosong })} disabled={!!isian}>
          <Plus className="w-4 h-4" /> Tambah penandatangan
        </Button>
      </div>
      <p className="text-sm text-gray-500 mb-4">
        Penandatangan Pihak Pertama dipilih otomatis menurut nilai kontrak (isian bertanda Nilai utama): batas terkecil yang masih
        lebih besar atau sama dengan nilai kontrak. Jenis perjanjian tanpa Nilai utama memakai penandatangan Default. Approver
        bisa menggantinya saat persetujuan.
      </p>
      {galat && <p role="alert" className="mb-4 rounded-lg bg-red-50 p-3 text-sm text-red-700">{galat}</p>}
      {!adaTakTerbatas && daftar.length > 0 && (
        <p className="mb-4 rounded-lg bg-amber-50 p-3 text-sm text-amber-800">
          Tidak ada penandatangan aktif yang tak terbatas: pengajuan di atas {rupiah(String(batasTertinggi))} harus dipilih approver sendiri.
        </p>
      )}

      {isian && (
        <div className="mb-5 rounded-xl border border-gray-200 p-4">
          <div className="grid gap-3 md:grid-cols-2">
            <div className="space-y-1">
              <Label htmlFor="ttd-nama">Nama</Label>
              <Input id="ttd-nama" value={isian.name} onChange={(e) => setIsian({ ...isian, name: e.target.value })} />
            </div>
            <div className="space-y-1">
              <Label htmlFor="ttd-jabatan">Jabatan</Label>
              <Input id="ttd-jabatan" value={isian.position} onChange={(e) => setIsian({ ...isian, position: e.target.value })} />
            </div>
            <div className="space-y-1">
              <Label htmlFor="ttd-batas">Nilai kontrak maksimum</Label>
              <div className="relative">
                <span className="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 text-sm text-gray-500">Rp</span>
                <Input id="ttd-batas" inputMode="numeric" autoComplete="off" className="pl-8" placeholder="Kosongkan untuk tak terbatas" value={ribuan(isian.max_value)} onChange={(e) => setIsian({ ...isian, max_value: angkaSaja(e.target.value) })} />
              </div>
              <p className="text-xs text-gray-500">Batas bawah mengikuti batas penandatangan sebelumnya. Kosong: tak terbatas (hanya satu).</p>
            </div>
            <div className="flex flex-col justify-center gap-2 text-sm text-gray-700">
              <label className="flex items-center gap-2">
                <input type="checkbox" checked={isian.is_default} onChange={(e) => setIsian({ ...isian, is_default: e.target.checked })} className="rounded border-gray-300" />
                Default (jenis perjanjian tanpa Nilai utama)
              </label>
              {isian.id && (
                <label className="flex items-center gap-2">
                  <input type="checkbox" checked={isian.is_active} onChange={(e) => setIsian({ ...isian, is_active: e.target.checked })} className="rounded border-gray-300" />
                  Aktif
                </label>
              )}
            </div>
          </div>
          <div className="mt-4 flex gap-2">
            <Button onClick={() => void simpan()} disabled={menyimpan} className="flex items-center gap-2 text-white" style={{ background: '#0B2545' }}>
              <Save className="w-4 h-4" /> {menyimpan ? 'Menyimpan...' : 'Simpan'}
            </Button>
            <Button variant="outline" onClick={() => { setIsian(null); setGalat('') }}>Batal</Button>
          </div>
        </div>
      )}

      {!daftar.length ? (
        <div className="p-8 text-center text-gray-400">Belum ada penandatangan.</div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-xs uppercase tracking-wide text-gray-500 border-b border-gray-100">
                <th className="px-3 py-2 font-medium">Nama</th>
                <th className="px-3 py-2 font-medium">Jabatan</th>
                <th className="px-3 py-2 font-medium">Rentang nilai kontrak</th>
                <th className="px-3 py-2 font-medium">Keterangan</th>
                <th className="px-3 py-2 font-medium text-right">Aksi</th>
              </tr>
            </thead>
            <tbody>
              {daftar.map((p) => (
                <tr key={p.id} className={`border-b border-gray-50 last:border-0 ${p.is_active ? '' : 'text-gray-400'}`}>
                  <td className="px-3 py-3 font-medium">{p.name}</td>
                  <td className="px-3 py-3">{p.position}</td>
                  <td className="px-3 py-3">{rentang[p.id]}</td>
                  <td className="px-3 py-3 space-x-1">
                    {p.is_default && <span className="inline-block rounded-full border border-blue-200 bg-blue-50 px-2 py-0.5 text-xs text-blue-700">Default</span>}
                    {!p.is_active && <span className="inline-block rounded-full border border-gray-200 bg-gray-100 px-2 py-0.5 text-xs text-gray-600">Nonaktif</span>}
                  </td>
                  <td className="px-3 py-3">
                    <div className="flex justify-end gap-2">
                      <Button variant="outline" size="icon-sm" title="Ubah" aria-label={`Ubah ${p.name}`} disabled={!!isian} onClick={() => setIsian({ id: p.id, name: p.name, position: p.position, max_value: p.max_value === null ? '' : String(p.max_value), is_default: p.is_default, is_active: p.is_active })}>
                        <Pencil className="w-4 h-4" />
                      </Button>
                      <Button variant="outline" size="icon-sm" title="Hapus" aria-label={`Hapus ${p.name}`} className="text-red-600 hover:bg-red-50 hover:text-red-700" onClick={() => void hapus(p)}>
                        <Trash2 className="w-4 h-4" />
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  )
}

export default function AgreementCompanyMasterPage() {
  const [value, setValue] = useState<Record<string, string>>({})
  const [message, setMessage] = useState('')
  const [messageType, setMessageType] = useState<'success' | 'error' | ''>('')
  const [isLoading, setIsLoading] = useState(true)
  const [isSaving, setIsSaving] = useState(false)

  useEffect(() => {
    void agreementService.master()
      .then((data) => setValue(data))
      .catch((error: Error) => { setMessage(error.message); setMessageType('error') })
      .finally(() => setIsLoading(false))
  }, [])

  const save = async () => {
    setIsSaving(true)
    setMessage('')
    setMessageType('')
    try {
      await agreementService.saveMaster(Object.fromEntries(fieldKeys.map((k) => [k, value[k] ?? ''])))
      setMessage('Data berhasil disimpan')
      setMessageType('success')
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'Gagal menyimpan data')
      setMessageType('error')
    } finally {
      setIsSaving(false)
    }
  }

  return <div className="p-6 max-w-7xl mx-auto space-y-6">
    <div>
      <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>Master Pihak Pertama</h1>
      <p className="text-sm text-gray-500 mt-0.5">Data perusahaan dan penandatangan Pihak Pertama di perjanjian</p>
    </div>
    <section className="bg-white rounded-2xl border border-gray-100 p-6">
      <h2 className="font-semibold mb-4" style={{ color: '#0B2545' }}>Data Perusahaan</h2>
      {message && <p className={`mb-4 text-sm ${messageType === 'error' ? 'text-red-600' : 'text-green-600'}`}>{message}</p>}
      {isLoading ? (
        <div className="p-8 text-center text-gray-400">Memuat data...</div>
      ) : (
        <div className="grid md:grid-cols-2 gap-4">
          {fieldKeys.map((key) => (
            <div key={key}>
              <Label className="mb-1.5 text-gray-700">{fieldLabels[key]}</Label>
              <Input value={value[key] || ''} onChange={(event) => setValue({ ...value, [key]: event.target.value })} />
            </div>
          ))}
        </div>
      )}
      <Button onClick={() => void save()} disabled={isSaving} className="mt-5 flex items-center gap-2 text-white transition hover:brightness-95" style={{ background: '#C8102E' }}>
        <Save className="w-4 h-4" /> {isSaving ? 'Menyimpan...' : 'Simpan'}
      </Button>
    </section>
    <DaftarPenandatangan />
  </div>
}
