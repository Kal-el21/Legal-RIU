// Isian pengajuan perjanjian yang belum dikirim: pemindahan antarversi template
// dan penyimpanan di browser. Fungsi di sini murni (tanpa React) supaya bisa dites
// dengan Vitest.

import { adaIsi, kosongDi, teks } from './agreement-isian'

export interface FieldDef {
  name: string
  label: string
  type: string
  required: boolean
  // Pilihan tipe select, atau saran isian tipe text.
  options?: string[]
  // Sub-field tipe group, diisi sekali per baris.
  fields?: FieldDef[]
  // Kode field uang yang dibagi ke baris grup menurut persennya (termin).
  split_of?: string
}

export interface SectionDef {
  title: string
  fields: FieldDef[]
}

// Isian satu field: teks untuk field biasa, daftar poin untuk tipe list, dan baris
// sub-field untuk tipe group.
export type Baris = Record<string, string>
export type Nilai = string | string[] | Baris[]
export type Isian = Record<string, Nilai>

export interface HasilPindah {
  isian: Isian
  // Field yang perlu diisi atau diperiksa ulang: baru di versi ini, atau jenisnya berubah.
  baru: string[]
  // Label field yang sudah terisi tetapi tidak ada lagi di versi ini.
  dihapus: string[]
}

const semuaField = (sections: SectionDef[]) => sections.flatMap((s) => s.fields)

// Sub-field grup (kode dan jenis); grup dengan sub-field berbeda dianggap berubah.
const sidikSub = (f: FieldDef) => JSON.stringify((f.fields ?? []).map((s) => [s.name, s.type]))

// Bentuk isian cocok dengan jenis field-nya (mis. draf lama dengan ruang lingkup teks).
const bentukCocok = (f: FieldDef, v: Nilai) =>
  f.type === 'list' ? Array.isArray(v) && v.every((x) => typeof x === 'string')
    : f.type === 'group' ? Array.isArray(v) && v.every((x) => typeof x === 'object' && x !== null)
      : typeof v === 'string'

// pindahkanIsian memindahkan isian dari field versi lama ke field versi baru
// (versi template diganti Admin, atau pengaju berganti jenis perjanjian). Isian
// berkode sama ikut pindah. Kalau jenis field (atau sub-field grup) berubah,
// isiannya dibuang karena formatnya bisa tidak cocok (mis. teks menjadi uang, atau
// ruang lingkup satu kotak teks menjadi daftar poin), dan field ditandai baru.
export function pindahkanIsian(isian: Isian, lama: SectionDef[], baru: SectionDef[]): HasilPindah {
  const fieldLama = new Map(semuaField(lama).map((f) => [f.name, f]))
  const namaBaru = new Set(semuaField(baru).map((f) => f.name))
  const hasil: HasilPindah = { isian: {}, baru: [], dihapus: [] }
  for (const f of semuaField(baru)) {
    const l = fieldLama.get(f.name)
    const nilai = isian[f.name]
    const berubah = !l || l.type !== f.type || (f.type === 'group' && sidikSub(l) !== sidikSub(f)) ||
      (nilai !== undefined && !bentukCocok(f, nilai)) ||
      (f.type === 'select' && teks(nilai) && !(f.options ?? []).includes(teks(nilai)))
    if (berubah) {
      hasil.baru.push(f.name)
      continue
    }
    if (nilai !== undefined) hasil.isian[f.name] = nilai
  }
  hasil.dihapus = semuaField(lama)
    .filter((f) => !namaBaru.has(f.name) && adaIsi(isian[f.name]))
    .map((f) => f.label)
  return hasil
}

// isianKosong: isian wajib yang belum diisi di satu bagian, termasuk sub-field
// wajib per baris grup ("Termin Pembayaran 2: Dibayarkan").
export function isianKosong(bagian: SectionDef, isian: Isian): string[] {
  return bagian.fields.flatMap((f) => kosongDi(f, isian[f.name]))
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

// drafKeServer: isi permintaan untuk menyimpan draf lama di browser (PR 9) sekali sebagai
// draf di server (PR Draf). Kalau versi templatenya sudah diganti Admin, isian dipindah ke
// versi aktif lebih dulu; hasil pindahnya ikut dikembalikan untuk pemberitahuan di wizard.
export function drafKeServer(draf: Draf, aktif: { code: string; template_id: string; sections: SectionDef[] }) {
  const pindah = draf.template_id === aktif.template_id ? null : pindahkanIsian(draf.isian, draf.sections, aktif.sections)
  return {
    permintaan: { document_type_code: aktif.code, template_id: aktif.template_id, form_data: pindah ? pindah.isian : draf.isian },
    pindah,
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
