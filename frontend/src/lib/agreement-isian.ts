// Isian perjanjian di browser: format angka, pembagian termin, dan isian daftar
// dan grup. Pembagian termin sama persis dengan backend (bagiTermin di
// service/agreement_termin.go, dijaga kasus uji yang sama), supaya angka di layar
// sama dengan dokumen; server tetap menghitung ulang saat pratinjau dan simpan.
import type { Baris, FieldDef, Isian, Nilai, SectionDef } from './agreement-draft'

export const teks = (v: Nilai | undefined): string => (typeof v === 'string' ? v : '')
export const daftar = (v: Nilai | undefined): string[] => (Array.isArray(v) ? v.filter((x): x is string => typeof x === 'string') : [])
export const barisGrup = (v: Nilai | undefined): Baris[] =>
  Array.isArray(v) ? v.filter((x): x is Baris => typeof x === 'object' && x !== null) : []

// adaIsi: isian berisi sesuatu selain spasi (teks, poin, atau isian baris grup).
export function adaIsi(v: Nilai | undefined): boolean {
  if (typeof v === 'string') return v.trim() !== ''
  return daftar(v).some((x) => x.trim()) || barisGrup(v).some((b) => Object.values(b).some((x) => String(x).trim()))
}

// angkaSaja: hanya digit, untuk isian uang ("Rp 1.500.000" menjadi "1500000").
export const angkaSaja = (s: string) => s.replace(/\D/g, '').replace(/^0+(?=\d)/, '')
// ribuan: "1500000" menjadi "1.500.000".
export const ribuan = (s: string) => angkaSaja(s).replace(/\B(?=(\d{3})+(?!\d))/g, '.')
// rupiah: sama dengan dokumen ("Rp 1.500.000"); kosong menjadi "-".
export const rupiah = (s: string) => (angkaSaja(s) ? `Rp ${ribuan(s)}` : '-')
// bacaAngka: "33,5" atau "33.5" menjadi 33.5; kosong atau bukan angka menjadi 0.
export function bacaAngka(s: string): number {
  const n = Number(s.trim().replace(',', '.'))
  return s.trim() && Number.isFinite(n) ? n : 0
}
// tulisAngka: 33.5 menjadi "33,5", seperti di dokumen.
export const tulisAngka = (n: number) => String(n).replace('.', ',')

export interface BarisTermin {
  persen: number
  nilai: number
}

// Batas nilai kontrak backend: hitungan dalam basis poin harus muat di int64.
const KONTRAK_MAKS = 922337203685477

// bagiTermin membagi nilai kontrak ke termin menurut persennya (keputusan K5 dan
// D5), dengan pesan galat yang sama dengan backend. Persen dibulatkan ke dua
// desimal (dua kali bulat, supaya 1,005 menjadi 1,01); persen termin terakhir
// diganti sisanya; nilai dibulatkan ke rupiah dan termin terakhir menyerap
// selisihnya. Nilai dihitung per 10.000 rupiah supaya tetap tepat walau nilai
// kontrak dikali basis poin melewati batas bilangan bulat JavaScript.
export function bagiTermin(kontrak: number, persen: number[]): BarisTermin[] | string {
  if (!persen.length) return 'minimal satu termin'
  if (!Number.isSafeInteger(kontrak) || kontrak <= 0 || kontrak > KONTRAK_MAKS) return 'nilai kontrak harus lebih dari 0'
  const q = Math.floor(kontrak / 10000)
  const r = kontrak % 10000
  const hasil: BarisTermin[] = []
  let totalBP = 0
  let totalNilai = 0
  for (let i = 0; i < persen.length - 1; i++) {
    const bp = Math.round(Math.round(persen[i] * 1e6) / 1e4)
    if (!(bp > 0)) return `persen termin ${i + 1} harus lebih dari 0`
    totalBP += bp
    if (totalBP > 10000) return 'jumlah persen termin melebihi 100%'
    const nilai = q * bp + Math.floor((r * bp + 5000) / 10000)
    hasil.push({ persen: bp / 100, nilai })
    totalNilai += nilai
  }
  if (totalBP === 10000) return 'persen sudah 100%, tidak ada sisa untuk termin terakhir'
  hasil.push({ persen: (10000 - totalBP) / 100, nilai: kontrak - totalNilai })
  return hasil
}

// subBagi: kode sub-field persen dan sub-field uang grup yang membagi nilai.
export function subBagi(f: FieldDef): [persen: string, nilai: string] {
  const subs = f.fields ?? []
  return [subs.find((s) => s.type === 'percent')?.name ?? '', subs.find((s) => s.type === 'money')?.name ?? '']
}

// siapkanIsian melengkapi isian yang akan ditampilkan dan dikirim: daftar kosong
// dimulai dengan satu poin, grup dengan satu baris, dan grup yang membagi nilai
// mendapat persen baris terakhir dan nilai setiap baris. Pembagian yang tidak sah
// dikembalikan sebagai galat per kode field, dengan nilai baris dikosongkan.
export function siapkanIsian(sections: SectionDef[], isian: Isian): { isian: Isian; galat: Record<string, string> } {
  const out: Isian = { ...isian }
  const galat: Record<string, string> = {}
  for (const f of sections.flatMap((s) => s.fields)) {
    if (f.type === 'list' && !Array.isArray(out[f.name])) out[f.name] = ['']
    if (f.type !== 'group') continue
    const baris = Array.isArray(out[f.name]) ? barisGrup(out[f.name]).map((b) => ({ ...b })) : [{}]
    out[f.name] = baris
    if (!f.split_of) continue
    const [persen, nilai] = subBagi(f)
    const hasil = bagiTermin(Number(angkaSaja(teks(out[f.split_of]))), baris.map((b) => bacaAngka(b[persen] ?? '')))
    if (typeof hasil === 'string') {
      galat[f.name] = `${f.label}: ${hasil}`
      baris.forEach((b) => (b[nilai] = ''))
      continue
    }
    baris.forEach((b, i) => {
      b[nilai] = String(hasil[i].nilai)
      if (i === baris.length - 1) b[persen] = tulisAngka(hasil[i].persen)
    })
  }
  return { isian: out, galat }
}

// kosongDi: label isian wajib yang belum diisi pada satu field. Pada grup yang
// membagi nilai, nilai setiap baris dan persen baris terakhir dihitung, jadi tidak
// diperiksa di sini.
export function kosongDi(f: FieldDef, v: Nilai | undefined): string[] {
  if (f.type === 'list') return f.required && !daftar(v).some((x) => x.trim()) ? [f.label] : []
  if (f.type !== 'group') return f.required && !teks(v).trim() ? [f.label] : []
  const baris = barisGrup(v)
  if (!baris.length) return f.required ? [f.label] : []
  const [persen, nilai] = f.split_of ? subBagi(f) : ['', '']
  return baris.flatMap((b, i) =>
    (f.fields ?? [])
      .filter((s) => s.required && s.name !== nilai && !(s.name === persen && i === baris.length - 1) && !(b[s.name] ?? '').trim())
      .map((s) => `${f.label} ${i + 1}: ${s.label}`),
  )
}

const BULAN = ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember']

// tanggal: "2026-08-17" menjadi "17 Agustus 2026" (tanpa zona waktu); selain itu apa adanya.
export function tanggal(s: string): string {
  const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(s.trim())
  return m && BULAN[Number(m[2]) - 1] ? `${Number(m[3])} ${BULAN[Number(m[2]) - 1]} ${m[1]}` : s.trim()
}

// ringkasNilai: isian satu field untuk ringkasan Tinjau dan halaman detail, ditulis
// seperti di dokumen.
export function ringkasNilai(f: FieldDef, v: Nilai | undefined): string {
  switch (f.type) {
    case 'date':
      return tanggal(teks(v)) || '-'
    case 'money':
      return rupiah(teks(v))
    case 'percent':
      return teks(v).trim() ? `${tulisAngka(bacaAngka(teks(v)))}%` : '-'
    case 'list': {
      const poin = daftar(v).filter((x) => x.trim())
      return poin.length ? poin.map((x, i) => `${i + 1}. ${x.trim()}`).join('\n') : '-'
    }
    case 'group': {
      const baris = barisGrup(v)
      if (!baris.length) return '-'
      return baris
        .map((b, i) => `${i + 1}. ${(f.fields ?? []).map((s) => ringkasNilai(s, b[s.name])).filter((x) => x !== '-').join(' · ')}`)
        .join('\n')
    }
    default:
      return teks(v).trim() || '-'
  }
}

// dariFormData: isian tersimpan (JSON pengajuan) menjadi isian wizard; angka
// menjadi teks, daftar dan baris grup tetap berbentuk daftar.
export function dariFormData(data: Record<string, unknown>): Isian {
  const tulis = (x: unknown) => (x === null || x === undefined ? '' : String(x))
  return Object.fromEntries(
    Object.entries(data).map(([k, v]): [string, Nilai] => {
      if (!Array.isArray(v)) return [k, tulis(v)]
      if (v.every((x) => typeof x === 'object' && x !== null)) {
        return [k, v.map((b) => Object.fromEntries(Object.entries(b as object).map(([a, x]) => [a, tulis(x)])))]
      }
      return [k, v.map(tulis)]
    }),
  )
}
