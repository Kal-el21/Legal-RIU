import { describe, expect, it } from 'vitest'
import { LAINNYA, dataPersetujuan, isianAwalPersetujuan, rentangPenandatangan, type Penandatangan } from './penandatangan'

const p = (id: string, max_value: number | null, is_active = true): Penandatangan => ({ id, name: id, position: 'Jabatan ' + id, max_value, is_default: false, is_active })

describe('rentangPenandatangan', () => {
  it('menurunkan batas bawah dari batas penandatangan aktif sebelumnya, apa pun urutan daftarnya', () => {
    const rentang = rentangPenandatangan([p('andi', null), p('lama', 900_000_000, false), p('sari', 5_000_000_000), p('budi', 500_000_000)])
    expect(rentang).toEqual({
      budi: 'Rp 0 sampai Rp 500.000.000',
      sari: 'di atas Rp 500.000.000 sampai Rp 5.000.000.000',
      andi: 'di atas Rp 5.000.000.000',
      lama: 'nonaktif',
    })
  })

  it('satu penandatangan tak terbatas berlaku untuk semua nilai', () => {
    expect(rentangPenandatangan([p('andi', null)])).toEqual({ andi: 'semua nilai' })
  })
})

describe('isianAwalPersetujuan', () => {
  const master = { default_agreement_number: '***/PKS/2026', default_signing_place: 'Jakarta', pic: 'Divisi Legal', phone: '021', email: 'legal@contoh.co.id' }

  it('mendahulukan isian tersimpan di atas bawaan master, sama dengan backend', () => {
    const awal = isianAwalPersetujuan({
      agreement_number: '001/PKS/2026',
      form_data: { tempat_ttd: 'Bandung', tanggal_ttd: '2026-10-11', pihak_pertama_pic: 'Rina', pihak_pertama_telepon: '022', pihak_pertama_email: 'rina@contoh.co.id' },
    }, master)
    expect(awal).toMatchObject({ agreement_number: '001/PKS/2026', signing_place: 'Bandung', signing_date: '2026-10-11', pic: 'Rina', phone: '022', email: 'rina@contoh.co.id' })
  })

  it('memakai bawaan master kalau isiannya kosong', () => {
    expect(isianAwalPersetujuan({ form_data: { tempat_ttd: ' ' } }, master)).toMatchObject({ agreement_number: '***/PKS/2026', signing_place: 'Jakarta', pic: 'Divisi Legal' })
  })

  it('penandatangan: pilihan tersimpan, lalu Lainnya, lalu usulan', () => {
    const usulan = { id: 'sari', name: 'Sari', position: 'Direktur', value: 750_000_000 }
    expect(isianAwalPersetujuan({ form_data: { pihak_pertama_penandatangan_id: 'andi' }, signatory_suggestion: usulan }, master).party_one_signatory_id).toBe('andi')
    expect(isianAwalPersetujuan({ form_data: { pihak_pertama_pejabat: 'Rudi', pihak_pertama_jabatan: 'Plt. Direktur' }, signatory_suggestion: usulan }, master))
      .toMatchObject({ party_one_signatory_id: LAINNYA, party_one_signatory_name: 'Rudi', party_one_signatory_position: 'Plt. Direktur' })
    expect(isianAwalPersetujuan({ form_data: {}, signatory_suggestion: usulan }, master).party_one_signatory_id).toBe('sari')
    expect(isianAwalPersetujuan({ form_data: {}, signatory_suggestion: { ...usulan, id: null } }, master).party_one_signatory_id).toBe('')
  })
})

describe('dataPersetujuan', () => {
  it('pilihan dari Daftar dikirim sebagai id; Lainnya sebagai nama dan jabatan', () => {
    const meta = { signing_date: '2026-10-11', party_one_signatory_name: 'Rudi', party_one_signatory_position: 'Plt. Direktur' }
    expect(dataPersetujuan({ ...meta, party_one_signatory_id: 'andi' }))
      .toEqual({ signing_date: '2026-10-11', party_one_signatory_id: 'andi', party_one_signatory_name: '', party_one_signatory_position: '' })
    expect(dataPersetujuan({ ...meta, party_one_signatory_id: LAINNYA }))
      .toEqual({ signing_date: '2026-10-11', party_one_signatory_id: '', party_one_signatory_name: 'Rudi', party_one_signatory_position: 'Plt. Direktur' })
  })
})
