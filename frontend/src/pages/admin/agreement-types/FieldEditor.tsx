import { ArrowDown, ArrowUp, Plus, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import type { AgreementFieldDef, AgreementSectionDef } from '@/services/agreement-template.service'

const TIPE_FIELD: Record<string, string> = {
  text: 'Teks singkat',
  textarea: 'Teks panjang',
  email: 'Email',
  tel: 'Telepon',
  number: 'Angka',
  money: 'Uang (Rp)',
  percent: 'Persen',
  date: 'Tanggal',
  select: 'Pilihan',
  list: 'Daftar poin',
  group: 'Grup berulang (mis. termin)',
}

// Sub-field grup hanya bertipe satu nilai (bukan daftar atau grup), sama dengan backend.
const TIPE_SUB = Object.fromEntries(Object.entries(TIPE_FIELD).filter(([t]) => t !== 'list' && t !== 'group'))

// Placeholder yang dihasilkan satu isian, sama dengan aturan backend (tokenField).
function placeholderField(f: AgreementFieldDef): string[] {
  const kode = f.name.toUpperCase()
  if (!kode) return []
  if (f.type === 'group') {
    return [`${kode}_JUMLAH`, `${kode}_URUTAN`, ...(f.fields ?? []).flatMap((s) => placeholderField({ ...s, name: s.name ? `${f.name}_${s.name}` : '' }))]
  }
  if (f.type === 'date') return [kode, `${kode}_HARI`, `${kode}_TERBILANG`]
  if (['money', 'number', 'percent'].includes(f.type)) return [kode, `${kode}_TERBILANG`]
  return [kode]
}

function pindah<T>(list: T[], i: number, arah: -1 | 1): T[] {
  const j = i + arah
  if (j < 0 || j >= list.length) return list
  const next = [...list]
  ;[next[i], next[j]] = [next[j], next[i]]
  return next
}

// Mengganti jenis isian: sub-field dan "membagi nilai" hanya untuk grup.
function gantiTipe(f: AgreementFieldDef, type: string): AgreementFieldDef {
  if (type === 'group') return { ...f, type, fields: f.fields ?? [] }
  return { ...f, type, fields: undefined, split_of: undefined }
}

interface BarisProps {
  f: AgreementFieldDef
  ubah: (field: Partial<AgreementFieldDef>) => void
  naik?: () => void
  turun?: () => void
  hapus: () => void
  tipe: Record<string, string>
  // Awalan kode di Word untuk sub-field, mis. "TERMIN_".
  awalan?: string
}

// Satu baris isian: kode, label, pilihan atau saran, jenis, wajib, dan urutan.
function BarisField({ f, ubah, naik, turun, hapus, tipe, awalan = '' }: BarisProps) {
  return (
    <div className="grid gap-3 md:grid-cols-[minmax(0,1.1fr)_minmax(0,1.4fr)_minmax(0,0.9fr)_auto_auto] items-start">
      <div className="space-y-1">
        <Label className="text-xs text-gray-500">{awalan ? 'Kode sub-field' : 'Kode di Word'}</Label>
        <div className="flex items-center gap-1">
          {awalan && <span className="font-mono text-xs text-gray-400">{awalan}</span>}
          <Input
            value={f.name.toUpperCase()}
            onChange={(e) => ubah({ name: e.target.value.toLowerCase().replace(/\s+/g, '_') })}
            placeholder={awalan ? 'PERSEN' : 'NAMA_MITRA'}
            className="font-mono text-xs"
          />
        </div>
        {!awalan && f.type !== 'group' && <p className="text-[11px] text-gray-400 break-all">{placeholderField(f).map((p) => `{{${p}}}`).join(' ')}</p>}
      </div>
      <div className="space-y-1">
        <Label className="text-xs text-gray-500">Label di form</Label>
        <Input value={f.label} onChange={(e) => ubah({ label: e.target.value })} placeholder={awalan ? 'Persen' : 'Nama Mitra'} />
        {(f.type === 'select' || f.type === 'text') && (
          <Textarea
            value={(f.options ?? []).join('\n')}
            onChange={(e) => ubah({ options: e.target.value.split('\n') })}
            placeholder={f.type === 'select' ? 'Satu pilihan per baris' : 'Saran isian (opsional), satu per baris; pengaju tetap bisa mengetik sendiri'}
            className="mt-2 text-sm"
            aria-label={`${f.type === 'select' ? 'Pilihan' : 'Saran'} untuk ${f.label || f.name}`}
          />
        )}
      </div>
      <div className="space-y-1">
        <Label className="text-xs text-gray-500">Jenis isian</Label>
        <select
          value={f.type}
          onChange={(e) => ubah(gantiTipe(f, e.target.value))}
          className="h-9 w-full rounded-md border border-input bg-white px-2 text-sm"
        >
          {Object.entries(tipe).map(([value, label]) => <option key={value} value={value}>{label}</option>)}
        </select>
      </div>
      <label className="flex items-center gap-2 text-sm text-gray-700 md:pt-7">
        <input type="checkbox" checked={f.required} onChange={(e) => ubah({ required: e.target.checked })} className="rounded border-gray-300" />
        Wajib
      </label>
      <div className="flex gap-1 md:pt-6">
        <Button type="button" variant="outline" size="icon-sm" onClick={naik} disabled={!naik} title="Naikkan isian" aria-label="Naikkan isian">
          <ArrowUp className="w-4 h-4" />
        </Button>
        <Button type="button" variant="outline" size="icon-sm" onClick={turun} disabled={!turun} title="Turunkan isian" aria-label="Turunkan isian">
          <ArrowDown className="w-4 h-4" />
        </Button>
        <Button type="button" variant="outline" size="icon-sm" onClick={hapus} title="Hapus isian" aria-label="Hapus isian" className="text-red-600 hover:bg-red-50 hover:text-red-700">
          <Trash2 className="w-4 h-4" />
        </Button>
      </div>
    </div>
  )
}

// Sub-field grup dan pilihan "membagi nilai" (termin): baris grup mendapat persen
// sisa dan nilai rupiah yang dihitung dari field uang yang dipilih.
function SubField({ f, ubah, uang }: { f: AgreementFieldDef; ubah: (field: Partial<AgreementFieldDef>) => void; uang: AgreementFieldDef[] }) {
  const subs = f.fields ?? []
  const ubahSub = (k: number, field: Partial<AgreementFieldDef>) => ubah({ fields: subs.map((s, i) => (i === k ? { ...s, ...field } : s)) })
  const awalan = `${f.name.toUpperCase() || 'GRUP'}_`
  return (
    <div className="mt-3 rounded-lg border border-dashed border-gray-300 bg-gray-50/60 p-3 space-y-3">
      <p className="text-xs text-gray-600">
        Isian per baris. Di Word, tulis semua placeholder per baris dalam <b>satu paragraf</b>; paragraf itu diulang sekali per baris.
      </p>
      {subs.map((s, k) => (
        <BarisField
          key={k}
          f={s}
          awalan={awalan}
          tipe={TIPE_SUB}
          ubah={(field) => ubahSub(k, field)}
          naik={k > 0 ? () => ubah({ fields: pindah(subs, k, -1) }) : undefined}
          turun={k < subs.length - 1 ? () => ubah({ fields: pindah(subs, k, 1) }) : undefined}
          hapus={() => ubah({ fields: subs.filter((_, i) => i !== k) })}
        />
      ))}
      <Button type="button" variant="outline" size="sm" className="bg-white" onClick={() => ubah({ fields: [...subs, { name: '', label: '', type: 'text', required: true }] })}>
        <Plus className="w-4 h-4" /> Tambah sub-field
      </Button>
      <div className="max-w-md space-y-1">
        <Label className="text-xs text-gray-500">Membagi nilai</Label>
        <select value={f.split_of ?? ''} onChange={(e) => ubah({ split_of: e.target.value || undefined })} className="h-9 w-full rounded-md border border-input bg-white px-2 text-sm">
          <option value="">Tidak membagi nilai</option>
          {uang.filter((u) => u.name).map((u) => <option key={u.name} value={u.name}>{u.label || u.name.toUpperCase()}</option>)}
        </select>
        {f.split_of && <p className="text-[11px] text-gray-500">Butuh tepat satu sub-field Persen dan satu Uang. Persen baris terakhir (sisa) dan nilai setiap baris dihitung otomatis.</p>}
      </div>
      <p className="text-[11px] text-gray-400 break-all">{placeholderField(f).map((p) => `{{${p}}}`).join(' ')}</p>
    </div>
  )
}

interface Props {
  sections: AgreementSectionDef[]
  onChange: (sections: AgreementSectionDef[]) => void
}

// Penyusun isian per bagian. Setiap bagian menjadi satu kartu di wizard
// pengajuan. Kode ditampilkan huruf besar seperti di Word dan disimpan huruf
// kecil oleh server.
export default function FieldEditor({ sections, onChange }: Props) {
  const ubahBagian = (i: number, bagian: AgreementSectionDef) => onChange(sections.map((s, k) => (k === i ? bagian : s)))
  const ubahField = (i: number, j: number, field: Partial<AgreementFieldDef>) =>
    ubahBagian(i, { ...sections[i], fields: sections[i].fields.map((f, k) => (k === j ? { ...f, ...field } : f)) })
  // Field uang di bagian mana pun boleh dibagi grup.
  const uang = sections.flatMap((s) => s.fields).filter((f) => f.type === 'money')

  return (
    <div className="space-y-4">
      {sections.map((bagian, i) => (
        <div key={i} className="rounded-xl border border-gray-200">
          <div className="flex items-center gap-2 px-4 py-3 border-b border-gray-100 bg-gray-50/60 rounded-t-xl">
            <Input
              aria-label={`Nama bagian ${i + 1}`}
              value={bagian.title}
              onChange={(e) => ubahBagian(i, { ...bagian, title: e.target.value })}
              placeholder="Nama bagian, mis. Pihak Kedua"
              className="font-medium bg-white"
            />
            <Button type="button" variant="outline" size="icon-sm" onClick={() => onChange(pindah(sections, i, -1))} disabled={i === 0} title="Naikkan bagian" aria-label="Naikkan bagian">
              <ArrowUp className="w-4 h-4" />
            </Button>
            <Button type="button" variant="outline" size="icon-sm" onClick={() => onChange(pindah(sections, i, 1))} disabled={i === sections.length - 1} title="Turunkan bagian" aria-label="Turunkan bagian">
              <ArrowDown className="w-4 h-4" />
            </Button>
            <Button
              type="button"
              variant="outline"
              size="icon-sm"
              onClick={() => (bagian.fields.length === 0 || window.confirm(`Hapus bagian "${bagian.title}" beserta ${bagian.fields.length} isiannya?`)) && onChange(sections.filter((_, k) => k !== i))}
              title="Hapus bagian"
              aria-label="Hapus bagian"
              className="text-red-600 hover:bg-red-50 hover:text-red-700"
            >
              <Trash2 className="w-4 h-4" />
            </Button>
          </div>

          <div className="divide-y divide-gray-100">
            {bagian.fields.map((f, j) => (
              <div key={j} className="px-4 py-3">
                <BarisField
                  f={f}
                  tipe={TIPE_FIELD}
                  ubah={(field) => ubahField(i, j, field)}
                  naik={j > 0 ? () => ubahBagian(i, { ...bagian, fields: pindah(bagian.fields, j, -1) }) : undefined}
                  turun={j < bagian.fields.length - 1 ? () => ubahBagian(i, { ...bagian, fields: pindah(bagian.fields, j, 1) }) : undefined}
                  hapus={() => ubahBagian(i, { ...bagian, fields: bagian.fields.filter((_, k) => k !== j) })}
                />
                {f.type === 'list' && <p className="mt-2 text-[11px] text-gray-500">Di Word, tulis <code className="font-mono">{`{{${f.name.toUpperCase() || 'KODE'}}}`}</code> di satu paragraf berformat numbered list; paragraf itu diulang sekali per poin.</p>}
                {f.type === 'group' && <SubField f={f} ubah={(field) => ubahField(i, j, field)} uang={uang} />}
              </div>
            ))}
          </div>

          <div className="px-4 py-3">
            <Button type="button" variant="outline" size="sm" onClick={() => ubahBagian(i, { ...bagian, fields: [...bagian.fields, { name: '', label: '', type: 'text', required: true }] })}>
              <Plus className="w-4 h-4" /> Tambah isian
            </Button>
          </div>
        </div>
      ))}

      <Button type="button" variant="outline" onClick={() => onChange([...sections, { title: '', fields: [] }])}>
        <Plus className="w-4 h-4" /> Tambah bagian
      </Button>
    </div>
  )
}
