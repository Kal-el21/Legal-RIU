import { describe, expect, it } from 'vitest'
import { isianKosong, pindahkanIsian, type SectionDef } from './agreement-draft'

const teks = (name: string, label = name, required = false) => ({ name, label, type: 'text', required })

describe('pindahkanIsian', () => {
  const lama: SectionDef[] = [{ title: 'Mitra', fields: [teks('nama_mitra', 'Nama Mitra', true), teks('alamat', 'Alamat'), teks('kota', 'Kota'), { name: 'nilai', label: 'Nilai', type: 'text', required: false }] }]

  it('memindahkan isian berkode sama, juga kalau bagiannya berpindah', () => {
    const baru: SectionDef[] = [{ title: 'Pihak Kedua', fields: [teks('alamat')] }, { title: 'Mitra', fields: [teks('nama_mitra', 'Nama Mitra', true)] }]
    const hasil = pindahkanIsian({ nama_mitra: 'PT A', alamat: 'Jl. B' }, lama, baru)
    expect(hasil.isian).toEqual({ nama_mitra: 'PT A', alamat: 'Jl. B' })
    expect(hasil.baru).toEqual([])
  })

  it('menandai field baru tanpa isian', () => {
    const baru: SectionDef[] = [{ title: 'Mitra', fields: [teks('nama_mitra'), teks('npwp', 'NPWP', true)] }]
    const hasil = pindahkanIsian({ nama_mitra: 'PT A' }, lama, baru)
    expect(hasil.isian).toEqual({ nama_mitra: 'PT A' })
    expect(hasil.baru).toEqual(['npwp'])
  })

  it('mencatat field terisi yang dihapus, dan mengabaikan field kosong yang dihapus', () => {
    const baru: SectionDef[] = [{ title: 'Mitra', fields: [teks('nama_mitra')] }]
    const hasil = pindahkanIsian({ nama_mitra: 'PT A', alamat: 'Jl. B', kota: ' ' }, lama, baru)
    expect(hasil.isian).toEqual({ nama_mitra: 'PT A' })
    expect(hasil.dihapus).toEqual(['Alamat'])
  })

  it('membuang isian yang jenis field-nya berubah dan menandainya baru', () => {
    const baru: SectionDef[] = [{ title: 'Mitra', fields: [teks('nama_mitra'), { name: 'nilai', label: 'Nilai', type: 'money', required: true }] }]
    const hasil = pindahkanIsian({ nama_mitra: 'PT A', nilai: 'satu juta' }, lama, baru)
    expect(hasil.isian).toEqual({ nama_mitra: 'PT A' })
    expect(hasil.baru).toEqual(['nilai'])
  })

  it('membuang pilihan yang tidak ada lagi di daftar pilihan', () => {
    const pilih = (options: string[]): SectionDef[] => [{ title: 'Mitra', fields: [{ name: 'jenis', label: 'Jenis', type: 'select', required: true, options }] }]
    expect(pindahkanIsian({ jenis: 'Koperasi' }, pilih(['Koperasi', 'PT']), pilih(['PT'])).baru).toEqual(['jenis'])
    expect(pindahkanIsian({ jenis: 'PT' }, pilih(['Koperasi', 'PT']), pilih(['PT'])).isian).toEqual({ jenis: 'PT' })
  })
})

describe('isianKosong', () => {
  it('menyebut label field wajib yang kosong atau hanya spasi', () => {
    const bagian: SectionDef = { title: 'Mitra', fields: [teks('nama_mitra', 'Nama Mitra', true), teks('pic', 'PIC'), teks('npwp', 'NPWP', true)] }
    expect(isianKosong(bagian, { nama_mitra: ' ', pic: '' })).toEqual(['Nama Mitra', 'NPWP'])
    expect(isianKosong(bagian, { nama_mitra: 'PT A', npwp: '01' })).toEqual([])
  })
})
