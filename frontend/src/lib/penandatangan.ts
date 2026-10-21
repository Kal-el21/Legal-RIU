// Daftar Penandatangan dan pilihan penandatangan di persetujuan (PR 11). Fungsi murni
// supaya bisa dites dengan Vitest.
import { rupiah } from './agreement-isian'

export interface Penandatangan {
  id: string
  name: string
  position: string
  // null: tak terbatas.
  max_value: number | null
  is_default: boolean
  is_active: boolean
}

// Usulan penandatangan dari backend (dihitung saat dibaca). id null: tidak ada yang cocok.
export interface UsulanPenandatangan {
  id: string | null
  name: string
  position: string
  value: number | null
}

// Nilai pilihan untuk penandatangan di luar Daftar (nama dan jabatan diketik approver).
export const LAINNYA = 'lainnya'

// rentangPenandatangan: rentang nilai kontrak tiap penandatangan aktif. Batas bawah
// diturunkan dari batas penandatangan aktif sebelumnya, mis. "di atas Rp 500.000.000
// sampai Rp 5.000.000.000"; penandatangan nonaktif ditulis "nonaktif".
export function rentangPenandatangan(daftar: Penandatangan[]): Record<string, string> {
  const out: Record<string, string> = {}
  const aktif = daftar.filter((p) => p.is_active).sort((a, b) => (a.max_value ?? Infinity) - (b.max_value ?? Infinity))
  let bawah: number | null = null
  for (const p of aktif) {
    const atas = p.max_value === null ? '' : rupiah(String(p.max_value))
    if (bawah === null) out[p.id] = atas ? `${rupiah('0')} sampai ${atas}` : 'semua nilai'
    else out[p.id] = atas ? `di atas ${rupiah(String(bawah))} sampai ${atas}` : `di atas ${rupiah(String(bawah))}`
    if (p.max_value !== null) bawah = p.max_value
  }
  for (const p of daftar) if (!p.is_active) out[p.id] = 'nonaktif'
  return out
}

interface DokumenPersetujuan {
  agreement_number?: string | null
  form_data: Record<string, unknown>
  signatory_suggestion?: UsulanPenandatangan | null
}

// isianAwalPersetujuan: isian awal Data Persetujuan. Isian yang sudah tersimpan di
// pengajuan didahulukan, baru bawaan master, sama dengan urutan backend
// (partyOneSnapshot). Sebelumnya bawaan master didahulukan, jadi menyimpan data
// persetujuan menimpa usulan tempat dari pengaju dan isian approver sebelumnya.
// Penandatangan: pilihan tersimpan, lalu Lainnya kalau namanya tersimpan, lalu usulan.
export function isianAwalPersetujuan(doc: DokumenPersetujuan, master: Record<string, string>): Record<string, string> {
  const isi = (k: string) => String(doc.form_data[k] ?? '').trim()
  return {
    agreement_number: (doc.agreement_number ?? '').trim() || master.default_agreement_number || '',
    signing_place: isi('tempat_ttd') || master.default_signing_place || '',
    signing_date: isi('tanggal_ttd'),
    party_one_signatory_id: isi('pihak_pertama_penandatangan_id') || (isi('pihak_pertama_pejabat') ? LAINNYA : doc.signatory_suggestion?.id ?? ''),
    party_one_signatory_name: isi('pihak_pertama_pejabat'),
    party_one_signatory_position: isi('pihak_pertama_jabatan'),
    pic: isi('pihak_pertama_pic') || master.pic || '',
    phone: isi('pihak_pertama_telepon') || master.phone || '',
    email: isi('pihak_pertama_email') || master.email || '',
  }
}

// dataPersetujuan: isian Data Persetujuan menjadi permintaan ke backend. Pilihan dari
// Daftar dikirim sebagai id; Lainnya sebagai nama dan jabatan dengan id kosong.
export function dataPersetujuan(meta: Record<string, string>): Record<string, string> {
  const lainnya = meta.party_one_signatory_id === LAINNYA
  return {
    ...meta,
    party_one_signatory_id: lainnya ? '' : meta.party_one_signatory_id ?? '',
    party_one_signatory_name: lainnya ? meta.party_one_signatory_name ?? '' : '',
    party_one_signatory_position: lainnya ? meta.party_one_signatory_position ?? '' : '',
  }
}
