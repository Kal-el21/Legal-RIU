import { describe, expect, it } from 'vitest'
import { angkaSaja, bagiTermin, dariFormData, kosongDi, ribuan, ringkasNilai, rupiah, siapkanIsian, type BarisTermin } from './agreement-isian'
import type { FieldDef, SectionDef } from './agreement-draft'

// Kasus yang sama dengan TestPembagianTermin (backend/internal/service/agreement_termin_test.go),
// supaya hitungan browser dan server tidak berbeda.
describe('bagiTermin', () => {
  const kasus: [string, number, number[], BarisTermin[] | string][] = [
    ['SatuTermin', 750_000_000, [40], [{ persen: 100, nilai: 750_000_000 }]],
    ['TigaTerminBulat', 750_000_000, [30, 30, 0], [{ persen: 30, nilai: 225_000_000 }, { persen: 30, nilai: 225_000_000 }, { persen: 40, nilai: 300_000_000 }]],
    ['SepertigaDariSatuMiliar', 1_000_000_000, [33.33, 33.33, 0], [{ persen: 33.33, nilai: 333_300_000 }, { persen: 33.33, nilai: 333_300_000 }, { persen: 33.34, nilai: 333_400_000 }]],
    ['SelisihPembulatanKeTerminTerakhir', 1_000_000_001, [33.33, 33.33, 0], [{ persen: 33.33, nilai: 333_300_000 }, { persen: 33.33, nilai: 333_300_000 }, { persen: 33.34, nilai: 333_400_001 }]],
    ['NilaiKecilDibulatkan', 100, [33.33, 33.33, 0], [{ persen: 33.33, nilai: 33 }, { persen: 33.33, nilai: 33 }, { persen: 33.34, nilai: 34 }]],
    ['NilaiDibulatkanKeAtas', 1_000_005, [33.33, 0], [{ persen: 33.33, nilai: 333_302 }, { persen: 66.67, nilai: 666_703 }]],
    ['PersenLebihDariDuaDesimal', 1_000_000, [12.345, 0], [{ persen: 12.35, nilai: 123_500 }, { persen: 87.65, nilai: 876_500 }]],
    ['PersenSetengahDibulatkanKeAtas', 1_000_000, [1.005, 0], [{ persen: 1.01, nilai: 10_100 }, { persen: 98.99, nilai: 989_900 }]],
    ['PersenTerakhirDiabaikan', 1_000_000, [25, 99], [{ persen: 25, nilai: 250_000 }, { persen: 75, nilai: 750_000 }]],
    ['TanpaTermin', 1_000_000, [], 'minimal satu termin'],
    ['PersenMelebihi100', 1_000_000, [60, 50, 0], 'jumlah persen termin melebihi 100%'],
    ['PersenTepat100TanpaSisa', 1_000_000, [50, 50, 0], 'persen sudah 100%, tidak ada sisa untuk termin terakhir'],
    ['PersenNol', 1_000_000, [0, 0], 'persen termin 1 harus lebih dari 0'],
    ['PersenNegatif', 1_000_000, [-10, 0], 'persen termin 1 harus lebih dari 0'],
    ['NilaiKontrakNol', 0, [50, 0], 'nilai kontrak harus lebih dari 0'],
  ]
  it.each(kasus)('%s', (_, kontrak, persen, harap) => {
    const hasil = bagiTermin(kontrak, persen)
    expect(hasil).toEqual(harap)
    if (typeof hasil !== 'string') expect(hasil.reduce((n, b) => n + b.nilai, 0)).toBe(kontrak)
  })

  it('tetap tepat untuk nilai kontrak triliunan', () => {
    // Nilai kontrak x basis poin melewati batas bilangan bulat JavaScript; dihitung
    // langsung, termin pertama meleset satu rupiah (999.900.001.668).
    const hasil = bagiTermin(3_000_000_005_003, [33.33, 0]) as BarisTermin[]
    expect(hasil).toEqual([{ persen: 33.33, nilai: 999_900_001_667 }, { persen: 66.67, nilai: 2_000_100_003_336 }])
  })
})

describe('format rupiah', () => {
  it('menulis titik ribuan seperti di dokumen dan menyimpan angka saja', () => {
    expect(angkaSaja('Rp 1.500.000')).toBe('1500000')
    expect(angkaSaja('007')).toBe('7')
    expect(ribuan('750000000')).toBe('750.000.000')
    expect(ribuan('999')).toBe('999')
    expect(rupiah('750000000')).toBe('Rp 750.000.000')
    expect(rupiah('')).toBe('-')
  })
})

const termin: FieldDef = {
  name: 'termin', label: 'Termin', type: 'group', required: true, split_of: 'nilai_kontrak',
  fields: [
    { name: 'persen', label: 'Persen', type: 'percent', required: true },
    { name: 'nilai', label: 'Nilai', type: 'money', required: true },
    { name: 'syarat', label: 'Dibayarkan', type: 'text', required: true },
  ],
}
const poin: FieldDef = { name: 'ruang_lingkup', label: 'Ruang Lingkup', type: 'list', required: true }
const bagian: SectionDef[] = [{ title: 'Pembayaran', fields: [{ name: 'nilai_kontrak', label: 'Nilai Kontrak', type: 'money', required: true }, termin, poin] }]

describe('siapkanIsian', () => {
  it('menghitung persen termin terakhir dan nilai setiap termin dari nilai kontrak', () => {
    const { isian, galat } = siapkanIsian(bagian, {
      nilai_kontrak: '750000000',
      termin: [{ persen: '30', nilai: '1', syarat: 'a' }, { persen: '30,5', syarat: 'b' }, { persen: '99', syarat: 'c' }],
    })
    expect(galat).toEqual({})
    expect(isian.termin).toEqual([
      { persen: '30', nilai: '225000000', syarat: 'a' },
      { persen: '30,5', nilai: '228750000', syarat: 'b' },
      { persen: '39,5', nilai: '296250000', syarat: 'c' },
    ])
  })

  it('memulai daftar dengan satu poin dan grup dengan satu baris (100%)', () => {
    const { isian } = siapkanIsian(bagian, { nilai_kontrak: '1000' })
    expect(isian.ruang_lingkup).toEqual([''])
    expect(isian.termin).toEqual([{ persen: '100', nilai: '1000' }])
  })

  it('pembagian yang tidak sah menjadi galat dengan pesan backend, dan nilainya dikosongkan', () => {
    const { isian, galat } = siapkanIsian(bagian, { nilai_kontrak: '1000', termin: [{ persen: '70' }, { persen: '40' }, {}] })
    expect(galat).toEqual({ termin: 'Termin: jumlah persen termin melebihi 100%' })
    expect(isian.termin).toEqual([{ persen: '70', nilai: '' }, { persen: '40', nilai: '' }, { nilai: '' }])
    expect(siapkanIsian(bagian, { termin: [{}] }).galat.termin).toBe('Termin: nilai kontrak harus lebih dari 0')
  })
})

describe('kosongDi', () => {
  it('menyebut sub-field wajib per baris; nilai dan persen terakhir yang dihitung tidak', () => {
    expect(kosongDi(termin, [{ persen: '30', nilai: '', syarat: 'a' }, { persen: '', nilai: '', syarat: ' ' }])).toEqual(['Termin 2: Dibayarkan'])
    expect(kosongDi(termin, [{ persen: '', syarat: 'a' }, { syarat: 'b' }])).toEqual(['Termin 1: Persen'])
    expect(kosongDi(termin, [])).toEqual(['Termin'])
  })

  it('daftar wajib minimal satu poin berisi', () => {
    expect(kosongDi(poin, ['', ' '])).toEqual(['Ruang Lingkup'])
    expect(kosongDi(poin, ['', 'Analisis'])).toEqual([])
  })
})

describe('ringkasNilai', () => {
  it('menulis daftar bernomor dan termin per baris seperti di dokumen', () => {
    expect(ringkasNilai(poin, ['Analisis', ' ', 'Pelatihan'])).toBe('1. Analisis\n2. Pelatihan')
    expect(ringkasNilai(termin, [{ persen: '30', nilai: '300000', syarat: 'setelah tanda tangan' }, { persen: '70', nilai: '700000', syarat: '' }]))
      .toBe('1. 30% · Rp 300.000 · setelah tanda tangan\n2. 70% · Rp 700.000')
    expect(ringkasNilai({ name: 'n', label: 'N', type: 'percent', required: false }, '33.34')).toBe('33,34%')
  })
})

describe('dariFormData', () => {
  it('membaca isian tersimpan: angka menjadi teks, daftar dan baris grup tetap daftar', () => {
    expect(dariFormData({ nilai_kontrak: 750000000, kosong: null, ruang_lingkup: ['a', 'b'], termin: [{ persen: 30, nilai: 225000000, syarat: 'x' }] })).toEqual({
      nilai_kontrak: '750000000', kosong: '', ruang_lingkup: ['a', 'b'], termin: [{ persen: '30', nilai: '225000000', syarat: 'x' }],
    })
  })
})
