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

// Draf pengajuan baru yang belum dikirim, disimpan di browser per pengguna supaya
// di komputer bersama draf pengguna lain tidak terbuka. Versi template dan field-nya
// ikut disimpan supaya isian bisa dipindah kalau Admin mengganti versi.
export interface Draf {
  code: string
  template_id: string
  sections: SectionDef[]
  isian: Isian
}

type Penyimpanan = Pick<Storage, 'getItem' | 'setItem' | 'removeItem'>

const kunciDraf = (userId: string) => `alea-draf-perjanjian:${userId}`

// Penyimpanan bisa tidak tersedia (mis. mode privat, atau di luar browser); draf
// lalu tidak disimpan.
const browser = (): Penyimpanan | null => {
  try {
    return globalThis.localStorage ?? null
  } catch {
    return null
  }
}

export function bacaDraf(userId: string, penyimpanan = browser()): Draf | null {
  if (!userId || !penyimpanan) return null
  try {
    const raw = penyimpanan.getItem(kunciDraf(userId))
    return raw ? (JSON.parse(raw) as Draf) : null
  } catch {
    return null
  }
}

export function simpanDraf(userId: string, draf: Draf, penyimpanan = browser()) {
  if (!userId || !penyimpanan) return
  try {
    penyimpanan.setItem(kunciDraf(userId), JSON.stringify(draf))
  } catch {
    // Penyimpanan penuh atau ditolak; draf tidak disimpan.
  }
}

export function hapusDraf(userId: string, penyimpanan = browser()) {
  if (!userId || !penyimpanan) return
  try {
    penyimpanan.removeItem(kunciDraf(userId))
  } catch {
    // Tidak ada yang perlu dihapus.
  }
}
