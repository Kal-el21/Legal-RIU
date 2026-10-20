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
}

// Placeholder yang dihasilkan satu isian, sama dengan aturan backend (fieldPlaceholders).
function placeholderField(f: AgreementFieldDef): string[] {
  const kode = f.name.toUpperCase()
  if (!kode) return []
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

interface Props {
  sections: AgreementSectionDef[]
  onChange: (sections: AgreementSectionDef[]) => void
}

// Penyusun isian per bagian. Setiap bagian menjadi satu langkah di wizard
// pengajuan. Kode ditampilkan huruf besar seperti di Word dan disimpan huruf
// kecil oleh server.
export default function FieldEditor({ sections, onChange }: Props) {
  const ubahBagian = (i: number, bagian: AgreementSectionDef) => onChange(sections.map((s, k) => (k === i ? bagian : s)))
  const ubahField = (i: number, j: number, field: Partial<AgreementFieldDef>) =>
    ubahBagian(i, { ...sections[i], fields: sections[i].fields.map((f, k) => (k === j ? { ...f, ...field } : f)) })

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
              <div key={j} className="px-4 py-3 grid gap-3 md:grid-cols-[minmax(0,1.1fr)_minmax(0,1.4fr)_minmax(0,0.9fr)_auto_auto] items-start">
                <div className="space-y-1">
                  <Label className="text-xs text-gray-500">Kode di Word</Label>
                  <Input
                    value={f.name.toUpperCase()}
                    onChange={(e) => ubahField(i, j, { name: e.target.value.toLowerCase().replace(/\s+/g, '_') })}
                    placeholder="NAMA_MITRA"
                    className="font-mono text-xs"
                  />
                  <p className="text-[11px] text-gray-400 break-all">{placeholderField(f).map((p) => `{{${p}}}`).join(' ')}</p>
                </div>
                <div className="space-y-1">
                  <Label className="text-xs text-gray-500">Label di form</Label>
                  <Input value={f.label} onChange={(e) => ubahField(i, j, { label: e.target.value })} placeholder="Nama Mitra" />
                  {f.type === 'select' && (
                    <Textarea
                      value={(f.options ?? []).join('\n')}
                      onChange={(e) => ubahField(i, j, { options: e.target.value.split('\n') })}
                      placeholder="Satu pilihan per baris"
                      className="mt-2 text-sm"
                      aria-label={`Pilihan untuk ${f.label || f.name}`}
                    />
                  )}
                </div>
                <div className="space-y-1">
                  <Label className="text-xs text-gray-500">Jenis isian</Label>
                  <select
                    value={f.type}
                    onChange={(e) => ubahField(i, j, { type: e.target.value })}
                    className="h-9 w-full rounded-md border border-input bg-white px-2 text-sm"
                  >
                    {Object.entries(TIPE_FIELD).map(([value, label]) => <option key={value} value={value}>{label}</option>)}
                  </select>
                </div>
                <label className="flex items-center gap-2 text-sm text-gray-700 md:pt-7">
                  <input type="checkbox" checked={f.required} onChange={(e) => ubahField(i, j, { required: e.target.checked })} className="rounded border-gray-300" />
                  Wajib
                </label>
                <div className="flex gap-1 md:pt-6">
                  <Button type="button" variant="outline" size="icon-sm" onClick={() => ubahBagian(i, { ...bagian, fields: pindah(bagian.fields, j, -1) })} disabled={j === 0} title="Naikkan isian" aria-label="Naikkan isian">
                    <ArrowUp className="w-4 h-4" />
                  </Button>
                  <Button type="button" variant="outline" size="icon-sm" onClick={() => ubahBagian(i, { ...bagian, fields: pindah(bagian.fields, j, 1) })} disabled={j === bagian.fields.length - 1} title="Turunkan isian" aria-label="Turunkan isian">
                    <ArrowDown className="w-4 h-4" />
                  </Button>
                  <Button type="button" variant="outline" size="icon-sm" onClick={() => ubahBagian(i, { ...bagian, fields: bagian.fields.filter((_, k) => k !== j) })} title="Hapus isian" aria-label="Hapus isian" className="text-red-600 hover:bg-red-50 hover:text-red-700">
                    <Trash2 className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            ))}
          </div>

          <div className="px-4 py-3">
            <Button type="button" variant="outline" size="sm" onClick={() => ubahBagian(i, { ...bagian, fields: [...bagian.fields, { name: '', label: '', type: 'text', required: false }] })}>
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
