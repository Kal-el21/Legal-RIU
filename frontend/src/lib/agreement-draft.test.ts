import { describe, expect, it } from 'vitest'
import { bacaDraf, drafKeServer, hapusDraf, isianKosong, pindahkanIsian, type SectionDef } from './agreement-draft'

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

  it('memindahkan daftar dan grup; draf ruang lingkup teks lama menjadi isian baru', () => {
    const termin = (sub: string[]) => ({ name: 'termin', label: 'Termin', type: 'group', required: true, fields: sub.map((s) => teks(s)) })
    const v1: SectionDef[] = [{ title: 'Pekerjaan', fields: [{ name: 'ruang_lingkup', label: 'Ruang Lingkup', type: 'textarea', required: true }] }]
    const v2: SectionDef[] = [{ title: 'Pekerjaan', fields: [{ name: 'ruang_lingkup', label: 'Ruang Lingkup', type: 'list', required: true }, termin(['persen', 'syarat'])] }]
    const pindah = pindahkanIsian({ ruang_lingkup: 'a\nb' }, v1, v2)
    expect(pindah.isian).toEqual({})
    expect(pindah.baru).toEqual(['ruang_lingkup', 'termin'])

    const isi = { ruang_lingkup: ['a', 'b'], termin: [{ persen: '30', syarat: 'x' }] }
    expect(pindahkanIsian(isi, v2, v2)).toEqual({ isian: isi, baru: [], dihapus: [] })
    // Isian yang bentuknya tidak cocok dengan jenis field-nya (draf teks untuk field daftar) dibuang.
    expect(pindahkanIsian({ ruang_lingkup: 'a', termin: ['x'] }, v2, v2)).toEqual({ isian: {}, baru: ['ruang_lingkup', 'termin'], dihapus: [] })
    // Sub-field grup berubah: baris lama bisa tidak cocok, jadi isiannya dibuang.
    const v3: SectionDef[] = [{ title: 'Pekerjaan', fields: [v2[0].fields[0], termin(['persen', 'nilai', 'syarat'])] }]
    expect(pindahkanIsian(isi, v2, v3)).toEqual({ isian: { ruang_lingkup: ['a', 'b'] }, baru: ['termin'], dihapus: [] })
    // Daftar terisi yang dihapus disebut; daftar berisi poin kosong tidak.
    const v4: SectionDef[] = [{ title: 'Pekerjaan', fields: [] }]
    expect(pindahkanIsian(isi, v2, v4).dihapus).toEqual(['Ruang Lingkup', 'Termin'])
    expect(pindahkanIsian({ ruang_lingkup: [' '], termin: [{ syarat: '' }] }, v2, v4).dihapus).toEqual([])
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

describe('draf lama di browser', () => {
  // Draf PR 9 hanya dibaca untuk dipindah sekali ke server (PR Draf); tes menulisnya langsung.
  const penyimpanan = (isi: Record<string, string> = {}) => {
    const m = new Map(Object.entries(isi))
    return { getItem: (k: string) => m.get(k) ?? null, setItem: (k: string, v: string) => void m.set(k, v), removeItem: (k: string) => void m.delete(k) }
  }
  const draf = (nama: string) => JSON.stringify({ code: 'PKS', template_id: 'v1', sections: [], isian: { nama_mitra: nama } })

  it('draf pengguna lain di browser yang sama tidak terbuka', () => {
    const p = penyimpanan({ 'alea-draf-perjanjian:user-a': draf('PT A'), 'alea-draf-perjanjian:user-b': draf('PT B') })
    expect(bacaDraf('user-a', p)?.isian.nama_mitra).toBe('PT A')
    hapusDraf('user-a', p)
    expect(bacaDraf('user-a', p)).toBeNull()
    expect(bacaDraf('user-b', p)?.isian.nama_mitra).toBe('PT B')
  })

  it('tanpa pengguna atau isi rusak tidak membaca apa pun', () => {
    expect(bacaDraf('', penyimpanan({ 'alea-draf-perjanjian:': draf('PT A') }))).toBeNull()
    expect(bacaDraf('user-a', penyimpanan({ 'alea-draf-perjanjian:user-a': '{rusak' }))).toBeNull()
  })
})

describe('drafKeServer', () => {
  const bagian = (...fields: { name: string; label: string; type: string; required?: boolean }[]) => [{ title: 'Data', fields: fields.map((f) => ({ required: false, ...f })) }]
  const lama = { code: 'PKS', template_id: 'v1', sections: bagian({ name: 'nama', label: 'Nama', type: 'text' }, { name: 'kota', label: 'Kota', type: 'text' }), isian: { nama: 'PT A', kota: 'Jakarta' } }

  it('versi sama: isian dikirim apa adanya dengan versi aktif', () => {
    const { permintaan, pindah } = drafKeServer(lama, { code: 'PKS', template_id: 'v1', sections: lama.sections })
    expect(pindah).toBeNull()
    expect(permintaan).toEqual({ document_type_code: 'PKS', template_id: 'v1', form_data: { nama: 'PT A', kota: 'Jakarta' } })
  })

  it('versi diganti: isian dipindah ke field versi aktif, dan yang hilang serta yang baru dilaporkan', () => {
    const aktif = { code: 'PKS', template_id: 'v2', sections: bagian({ name: 'nama', label: 'Nama', type: 'text' }, { name: 'npwp', label: 'NPWP', type: 'text', required: true }) }
    const { permintaan, pindah } = drafKeServer(lama, aktif)
    expect(permintaan.template_id).toBe('v2')
    expect(permintaan.form_data).toEqual({ nama: 'PT A' })
    expect(pindah?.baru).toEqual(['npwp'])
    expect(pindah?.dihapus).toEqual(['Kota'])
  })
})
