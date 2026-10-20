// Isian pengajuan perjanjian yang belum dikirim: pemindahan antarversi template
// dan penyimpanan di browser. Fungsi di sini murni (tanpa React) supaya bisa dites
// dengan Vitest.

export interface FieldDef {
  name: string
  label: string
  type: string
  required: boolean
  options?: string[]
}

export interface SectionDef {
  title: string
  fields: FieldDef[]
}

export type Isian = Record<string, string>

export interface HasilPindah {
  isian: Isian
  // Field yang perlu diisi atau diperiksa ulang: baru di versi ini, atau jenisnya berubah.
  baru: string[]
  // Label field yang sudah terisi tetapi tidak ada lagi di versi ini.
  dihapus: string[]
}

const semuaField = (sections: SectionDef[]) => sections.flatMap((s) => s.fields)

// pindahkanIsian memindahkan isian dari field versi lama ke field versi baru
// (versi template diganti Admin, atau pengaju berganti jenis perjanjian). Isian
// berkode sama ikut pindah. Kalau jenis field berubah, isiannya dibuang karena
// formatnya bisa tidak cocok (mis. teks menjadi uang), dan field ditandai baru.
export function pindahkanIsian(isian: Isian, lama: SectionDef[], baru: SectionDef[]): HasilPindah {
  const fieldLama = new Map(semuaField(lama).map((f) => [f.name, f]))
  const namaBaru = new Set(semuaField(baru).map((f) => f.name))
  const hasil: HasilPindah = { isian: {}, baru: [], dihapus: [] }
  for (const f of semuaField(baru)) {
    const l = fieldLama.get(f.name)
    const nilai = isian[f.name]
    if (!l || l.type !== f.type || (f.type === 'select' && nilai && !(f.options ?? []).includes(nilai))) {
      hasil.baru.push(f.name)
      continue
    }
    if (nilai !== undefined) hasil.isian[f.name] = nilai
  }
  hasil.dihapus = semuaField(lama)
    .filter((f) => !namaBaru.has(f.name) && (isian[f.name] ?? '').trim() !== '')
    .map((f) => f.label)
  return hasil
}

// isianKosong: label field wajib yang belum diisi di satu bagian.
export function isianKosong(bagian: SectionDef, isian: Isian): string[] {
  return bagian.fields.filter((f) => f.required && !(isian[f.name] ?? '').trim()).map((f) => f.label)
}

// Draf pengajuan baru yang belum dikirim, disimpan di browser. Versi template dan
// field-nya ikut disimpan supaya isian bisa dipindah kalau Admin mengganti versi.
export interface Draf {
  code: string
  template_id: string
  sections: SectionDef[]
  isian: Isian
}

const KUNCI_DRAF = 'alea-draf-perjanjian'

export function bacaDraf(): Draf | null {
  try {
    const raw = localStorage.getItem(KUNCI_DRAF)
    return raw ? (JSON.parse(raw) as Draf) : null
  } catch {
    return null
  }
}

export function simpanDraf(draf: Draf) {
  try {
    localStorage.setItem(KUNCI_DRAF, JSON.stringify(draf))
  } catch {
    // Penyimpanan browser tidak tersedia (mis. mode privat); draf tidak disimpan.
  }
}

export function hapusDraf() {
  try {
    localStorage.removeItem(KUNCI_DRAF)
  } catch {
    // Tidak ada yang perlu dihapus.
  }
}
